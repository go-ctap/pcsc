//go:build darwin || linux

package pcsc

import (
	"bytes"
	"context"
	"errors"
	"iter"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

const (
	scardScopeSystem      = 2
	scardShareShared      = 2
	scardLeaveCard        = 0
	scardProtocolAny      = scardDWORD(ProtocolT0 | ProtocolT1)
	scardEInsufficientBuf = uint32(0x80100008)
	maxResponseBufSize    = 65538
)

type scardIORequest struct {
	Protocol scardDWORD
	Length   scardDWORD
}

var (
	scardEstablishContext func(scardDWORD, uintptr, uintptr, *scardContext) scardResult
	scardReleaseContext   func(scardContext) scardResult
	scardListReaders      func(scardContext, uintptr, uintptr, *scardDWORD) scardResult
	scardConnect          func(scardContext, uintptr, scardDWORD, scardDWORD, *scardHandle, *scardDWORD) scardResult
	scardDisconnect       func(scardHandle, scardDWORD) scardResult
	scardStatus           func(scardHandle, uintptr, *scardDWORD, *scardDWORD, *scardDWORD, uintptr, *scardDWORD) scardResult
	scardTransmit         func(scardHandle, *scardIORequest, uintptr, scardDWORD, *scardIORequest, uintptr, *scardDWORD) scardResult
	scardCancel           func(scardContext) scardResult
)

func scardError(op string, code scardResult) error {
	return pcscError(op, uint32(code))
}

func init() {
	lib, err := purego.Dlopen(pcscLibrary, purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		panic("pcsc: load native library: " + err.Error())
	}
	purego.RegisterLibFunc(&scardEstablishContext, lib, "SCardEstablishContext")
	purego.RegisterLibFunc(&scardReleaseContext, lib, "SCardReleaseContext")
	purego.RegisterLibFunc(&scardListReaders, lib, "SCardListReaders")
	purego.RegisterLibFunc(&scardConnect, lib, "SCardConnect")
	purego.RegisterLibFunc(&scardDisconnect, lib, "SCardDisconnect")
	purego.RegisterLibFunc(&scardStatus, lib, "SCardStatus")
	purego.RegisterLibFunc(&scardTransmit, lib, "SCardTransmit")
	purego.RegisterLibFunc(&scardCancel, lib, "SCardCancel")
}

func withContext(fn func(scardContext) error) error {
	var ctx scardContext
	if err := scardError("SCardEstablishContext", scardEstablishContext(scardScopeSystem, 0, 0, &ctx)); err != nil {
		return err
	}

	err := fn(ctx)
	releaseErr := scardError("SCardReleaseContext", scardReleaseContext(ctx))

	return errors.Join(err, releaseErr)
}

func enumerate() iter.Seq2[*ReaderInfo, error] {
	return func(yield func(*ReaderInfo, error) bool) {
		var names []string
		err := withContext(func(pcscCtx scardContext) error {
			var size scardDWORD

			if err := scardError("SCardListReaders", scardListReaders(pcscCtx, 0, 0, &size)); err != nil {
				if e := new(Error); errors.As(err, &e) && e.Code == 0x8010002e {
					return nil
				}

				return err
			}

			if size == 0 {
				return nil
			}
			buf := make([]byte, size)
			if err := scardError("SCardListReaders", scardListReaders(pcscCtx, 0, uintptr(unsafe.Pointer(unsafe.SliceData(buf))), &size)); err != nil {
				return err
			}

			names = parseMultiString(buf[:min(int(size), len(buf))])

			return nil
		})

		if err != nil {
			yield(nil, err)
			return
		}

		for _, name := range names {
			if !yield(&ReaderInfo{Name: name}, nil) {
				return
			}
		}
	}
}

func parseMultiString(buf []byte) []string {
	var out []string
	for len(buf) > 0 {
		i := bytes.IndexByte(buf, 0)
		if i < 0 {
			i = len(buf)
		}

		if i == 0 {
			break
		}

		out = append(out, string(buf[:i]))

		if i == len(buf) {
			break
		}

		buf = buf[i+1:]
	}
	return out
}

type card struct {
	mu       sync.Mutex
	context  scardContext
	handle   scardHandle
	protocol Protocol
	closed   bool
}

func open(reader string) (Card, error) {
	var pcscCtx scardContext
	if err := scardError("SCardEstablishContext", scardEstablishContext(scardScopeSystem, 0, 0, &pcscCtx)); err != nil {
		return nil, err
	}
	name := append([]byte(reader), 0)
	var handle scardHandle
	var protocol scardDWORD

	code := scardConnect(pcscCtx, uintptr(unsafe.Pointer(unsafe.SliceData(name))), scardShareShared, scardProtocolAny, &handle, &protocol)
	runtime.KeepAlive(name)

	if err := scardError("SCardConnect", code); err != nil {
		_ = scardError("SCardReleaseContext", scardReleaseContext(pcscCtx))
		return nil, err
	}

	return &card{context: pcscCtx, handle: handle, protocol: Protocol(uint32(protocol))}, nil
}

func (c *card) Transmit(ctx context.Context, apdu []byte) ([]byte, error) {
	c.mu.Lock()
	if err := ctx.Err(); err != nil {
		c.mu.Unlock()
		return nil, err
	}
	if c.closed {
		c.mu.Unlock()
		return nil, errors.New("pcsc: card closed")
	}

	result := make(chan transmitResult, 1)
	go func() {
		defer c.mu.Unlock()

		request := scardIORequest{Protocol: scardDWORD(c.protocol), Length: scardDWORD(unsafe.Sizeof(scardIORequest{}))}
		response := make([]byte, maxResponseBufSize)
		size := scardDWORD(len(response))
		code := scardTransmit(c.handle, &request, uintptr(unsafe.Pointer(unsafe.SliceData(apdu))), scardDWORD(len(apdu)), nil, uintptr(unsafe.Pointer(unsafe.SliceData(response))), &size)
		runtime.KeepAlive(apdu)

		if err := scardError("SCardTransmit", code); err != nil {
			result <- transmitResult{err: err}
			return
		}

		result <- transmitResult{response: bytes.Clone(response[:min(int(size), len(response))])}
	}()

	select {
	case <-ctx.Done():
		_ = scardCancel(c.context)
		return nil, ctx.Err()
	case r := <-result:
		return r.response, r.err
	}
}

func (c *card) Status() (*CardStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, errors.New("pcsc: card closed")
	}

	var readerLen, atrLen scardDWORD
	var state, protocol scardDWORD

	code := scardStatus(c.handle, 0, &readerLen, &state, &protocol, 0, &atrLen)
	if code != 0 && uint32(code) != scardEInsufficientBuf {
		return nil, scardError("SCardStatus", code)
	}

	reader := make([]byte, readerLen)
	atr := make([]byte, atrLen)

	code = scardStatus(c.handle, slicePointer(reader), &readerLen, &state, &protocol, slicePointer(atr), &atrLen)
	if err := scardError("SCardStatus", code); err != nil {
		return nil, err
	}

	return &CardStatus{
		Reader:   firstString(reader),
		State:    uint32(state),
		Protocol: Protocol(uint32(protocol)),
		ATR:      bytes.Clone(atr[:min(int(atrLen), len(atr))]),
	}, nil
}

func slicePointer(b []byte) uintptr {
	if len(b) == 0 {
		return 0
	}

	return uintptr(unsafe.Pointer(unsafe.SliceData(b)))
}

func firstString(b []byte) string {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		b = b[:i]
	}

	return string(b)
}

func (c *card) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	return errors.Join(scardError("SCardDisconnect", scardDisconnect(c.handle, scardLeaveCard)), scardError("SCardReleaseContext", scardReleaseContext(c.context)))
}
