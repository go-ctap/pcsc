//go:build windows

package pcsc

import (
	"bytes"
	"errors"
	"iter"
	"runtime"
	"sync"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	scardScopeSystem      = 2
	scardShareShared      = 2
	scardLeaveCard        = 0
	scardProtocolAny      = uint32(ProtocolT0 | ProtocolT1)
	scardEInsufficientBuf = uint32(0x80100008)
	maxResponseBufSize    = 65538
)

type scardHandle = uintptr
type scardDWORD = uint32
type scardResult = uint32

type scardIORequest struct {
	Protocol scardDWORD
	Length   scardDWORD
}

var (
	modWinSCard               = windows.NewLazySystemDLL("winscard.dll")
	procSCardEstablishContext = modWinSCard.NewProc("SCardEstablishContext")
	procSCardReleaseContext   = modWinSCard.NewProc("SCardReleaseContext")
	procSCardListReadersW     = modWinSCard.NewProc("SCardListReadersW")
	procSCardConnectW         = modWinSCard.NewProc("SCardConnectW")
	procSCardDisconnect       = modWinSCard.NewProc("SCardDisconnect")
	procSCardStatusW          = modWinSCard.NewProc("SCardStatusW")
	procSCardTransmit         = modWinSCard.NewProc("SCardTransmit")
)

var scardTransmit = func(handle scardHandle, sendPCI *scardIORequest, sendBuffer uintptr, sendLength scardDWORD, recvPCI *scardIORequest, recvBuffer uintptr, recvLength *scardDWORD) scardResult {
	code := callCode(procSCardTransmit, handle, uintptr(unsafe.Pointer(sendPCI)), sendBuffer, uintptr(sendLength), uintptr(unsafe.Pointer(recvPCI)), recvBuffer, uintptr(unsafe.Pointer(recvLength)))
	runtime.KeepAlive(sendPCI)
	runtime.KeepAlive(recvPCI)
	runtime.KeepAlive(recvLength)
	return code
}

func callCode(proc *windows.LazyProc, args ...uintptr) uint32 {
	r, _, _ := proc.Call(args...)
	return uint32(r)
}

func establishContext() (uintptr, error) {
	var ctx uintptr
	code := callCode(procSCardEstablishContext, scardScopeSystem, 0, 0, uintptr(unsafe.Pointer(&ctx)))

	return ctx, pcscError("SCardEstablishContext", code)
}

func enumerate() iter.Seq2[*ReaderInfo, error] {
	return func(yield func(*ReaderInfo, error) bool) {
		ctx, err := establishContext()
		if err != nil {
			yield(nil, err)
			return
		}

		defer func() { _ = callCode(procSCardReleaseContext, ctx) }()

		var size uint32
		code := callCode(procSCardListReadersW, ctx, 0, 0, uintptr(unsafe.Pointer(&size)))
		if uint32(code) == 0x8010002e {
			return
		}

		if err := pcscError("SCardListReadersW", code); err != nil {
			yield(nil, err)
			return
		}

		if size == 0 {
			return
		}

		buf := make([]uint16, size)
		code = callCode(procSCardListReadersW, ctx, 0, uintptr(unsafe.Pointer(unsafe.SliceData(buf))), uintptr(unsafe.Pointer(&size)))
		if err := pcscError("SCardListReadersW", code); err != nil {
			yield(nil, err)
			return
		}

		for _, name := range parseUTF16MultiString(buf[:min(int(size), len(buf))]) {
			if !yield(&ReaderInfo{Name: name}, nil) {
				return
			}
		}
	}
}

func parseUTF16MultiString(buf []uint16) []string {
	var out []string
	for len(buf) > 0 {
		i := 0
		for i < len(buf) && buf[i] != 0 {
			i++
		}

		if i == 0 {
			break
		}

		out = append(out, string(utf16.Decode(buf[:i])))

		if i == len(buf) {
			break
		}

		buf = buf[i+1:]
	}
	return out
}

type card struct {
	mu       sync.Mutex
	context  uintptr
	handle   uintptr
	protocol Protocol
	closed   bool
}

func open(reader string) (Card, error) {
	ctx, err := establishContext()
	if err != nil {
		return nil, err
	}

	name, err := windows.UTF16PtrFromString(reader)
	if err != nil {
		_ = callCode(procSCardReleaseContext, ctx)
		return nil, err
	}

	var handle uintptr
	var protocol uint32

	code := callCode(procSCardConnectW, ctx, uintptr(unsafe.Pointer(name)), scardShareShared, uintptr(scardProtocolAny), uintptr(unsafe.Pointer(&handle)), uintptr(unsafe.Pointer(&protocol)))
	runtime.KeepAlive(name)

	if err := pcscError("SCardConnectW", code); err != nil {
		_ = callCode(procSCardReleaseContext, ctx)
		return nil, err
	}

	return &card{context: ctx, handle: handle, protocol: Protocol(protocol)}, nil
}

func (c *card) Transmit(apdu []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, errors.New("pcsc: card closed")
	}

	request := scardIORequest{Protocol: scardDWORD(c.protocol), Length: scardDWORD(unsafe.Sizeof(scardIORequest{}))}
	response := make([]byte, maxResponseBufSize)
	size := scardDWORD(len(response))
	code := scardTransmit(c.handle, &request, uintptr(unsafe.Pointer(unsafe.SliceData(apdu))), scardDWORD(len(apdu)), nil, uintptr(unsafe.Pointer(unsafe.SliceData(response))), &size)
	runtime.KeepAlive(apdu)

	if err := pcscError("SCardTransmit", uint32(code)); err != nil {
		return nil, err
	}

	return bytes.Clone(response[:min(int(size), len(response))]), nil
}

func (c *card) Status() (*CardStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, errors.New("pcsc: card closed")
	}

	var readerLen, atrLen, state, protocol uint32

	code := callCode(procSCardStatusW, c.handle, 0, uintptr(unsafe.Pointer(&readerLen)), uintptr(unsafe.Pointer(&state)), uintptr(unsafe.Pointer(&protocol)), 0, uintptr(unsafe.Pointer(&atrLen)))
	if code != 0 && uint32(code) != scardEInsufficientBuf {
		return nil, pcscError("SCardStatusW", code)
	}

	reader := make([]uint16, readerLen)
	atr := make([]byte, atrLen)

	code = callCode(procSCardStatusW, c.handle, uint16SlicePointer(reader), uintptr(unsafe.Pointer(&readerLen)), uintptr(unsafe.Pointer(&state)), uintptr(unsafe.Pointer(&protocol)), byteSlicePointer(atr), uintptr(unsafe.Pointer(&atrLen)))
	if err := pcscError("SCardStatusW", code); err != nil {
		return nil, err
	}

	return &CardStatus{
		Reader:   firstUTF16String(reader),
		State:    state,
		Protocol: Protocol(protocol),
		ATR:      bytes.Clone(atr[:min(int(atrLen), len(atr))]),
	}, nil
}

func uint16SlicePointer(b []uint16) uintptr {
	if len(b) == 0 {
		return 0
	}

	return uintptr(unsafe.Pointer(unsafe.SliceData(b)))
}

func byteSlicePointer(b []byte) uintptr {
	if len(b) == 0 {
		return 0
	}

	return uintptr(unsafe.Pointer(unsafe.SliceData(b)))
}

func firstUTF16String(b []uint16) string {
	for i, v := range b {
		if v == 0 {
			b = b[:i]
			break
		}
	}

	return string(utf16.Decode(b))
}

func (c *card) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	return errors.Join(pcscError("SCardDisconnect", callCode(procSCardDisconnect, c.handle, scardLeaveCard)), pcscError("SCardReleaseContext", callCode(procSCardReleaseContext, c.context)))
}
