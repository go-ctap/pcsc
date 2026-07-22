//go:build windows

package pcsc

import (
	"errors"
	"runtime"
	"time"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	scardScopeSystem        = 2
	scardReaderStateATRSize = 36
	scardReaderStatePacked  = false
)

type scardHandle = uintptr
type scardContext = uintptr
type scardDWORD = uint32
type scardResult = uint32

var (
	modWinSCard               = windows.NewLazySystemDLL("winscard.dll")
	procSCardEstablishContext = modWinSCard.NewProc("SCardEstablishContext")
	procSCardReleaseContext   = modWinSCard.NewProc("SCardReleaseContext")
	procSCardListReadersW     = modWinSCard.NewProc("SCardListReadersW")
	procSCardConnectW         = modWinSCard.NewProc("SCardConnectW")
	procSCardReconnect        = modWinSCard.NewProc("SCardReconnect")
	procSCardDisconnect       = modWinSCard.NewProc("SCardDisconnect")
	procSCardBeginTransaction = modWinSCard.NewProc("SCardBeginTransaction")
	procSCardEndTransaction   = modWinSCard.NewProc("SCardEndTransaction")
	procSCardStatusW          = modWinSCard.NewProc("SCardStatusW")
	procSCardTransmit         = modWinSCard.NewProc("SCardTransmit")
	procSCardControl          = modWinSCard.NewProc("SCardControl")
	procSCardGetAttrib        = modWinSCard.NewProc("SCardGetAttrib")
	procSCardSetAttrib        = modWinSCard.NewProc("SCardSetAttrib")
	procSCardGetStatusChangeW = modWinSCard.NewProc("SCardGetStatusChangeW")
	procSCardCancel           = modWinSCard.NewProc("SCardCancel")
)

var scardTransmit = func(
	handle scardHandle,
	sendPCI *scardIORequest,
	sendBuffer uintptr,
	sendLength scardDWORD,
	receivePCI *scardIORequest,
	receiveBuffer uintptr,
	receiveLength *scardDWORD,
) scardResult {
	result := callCode(
		procSCardTransmit,
		handle,
		uintptr(unsafe.Pointer(sendPCI)),
		sendBuffer,
		uintptr(sendLength),
		uintptr(unsafe.Pointer(receivePCI)),
		receiveBuffer,
		uintptr(unsafe.Pointer(receiveLength)),
	)
	runtime.KeepAlive(sendPCI)
	runtime.KeepAlive(receivePCI)
	runtime.KeepAlive(receiveLength)

	return result
}

var scardCancel = func(context scardContext) scardResult {
	return scardResult(callCode(procSCardCancel, context))
}

var scardBeginTransaction = func(handle scardHandle) scardResult {
	return scardResult(callCode(procSCardBeginTransaction, handle))
}

var scardEndTransaction = func(handle scardHandle, disposition scardDWORD) scardResult {
	return scardResult(callCode(procSCardEndTransaction, handle, uintptr(disposition)))
}

var scardDisconnect = func(handle scardHandle, disposition scardDWORD) scardResult {
	return scardResult(callCode(procSCardDisconnect, handle, uintptr(disposition)))
}

var scardReconnect = func(
	handle scardHandle,
	shareMode scardDWORD,
	preferredProtocols scardDWORD,
	initialization scardDWORD,
	activeProtocol *scardDWORD,
) scardResult {
	result := callCode(
		procSCardReconnect,
		handle,
		uintptr(shareMode),
		uintptr(preferredProtocols),
		uintptr(initialization),
		uintptr(unsafe.Pointer(activeProtocol)),
	)
	runtime.KeepAlive(activeProtocol)

	return scardResult(result)
}

var scardControl = func(
	handle scardHandle,
	controlCode scardDWORD,
	input uintptr,
	inputLength scardDWORD,
	output uintptr,
	outputLength scardDWORD,
	bytesReturned *scardDWORD,
) scardResult {
	result := callCode(
		procSCardControl,
		handle,
		uintptr(controlCode),
		input,
		uintptr(inputLength),
		output,
		uintptr(outputLength),
		uintptr(unsafe.Pointer(bytesReturned)),
	)
	runtime.KeepAlive(bytesReturned)

	return scardResult(result)
}

var scardGetAttrib = func(
	handle scardHandle,
	attribute scardDWORD,
	value uintptr,
	valueLength *scardDWORD,
) scardResult {
	result := callCode(
		procSCardGetAttrib,
		handle,
		uintptr(attribute),
		value,
		uintptr(unsafe.Pointer(valueLength)),
	)
	runtime.KeepAlive(valueLength)

	return scardResult(result)
}

var scardSetAttrib = func(
	handle scardHandle,
	attribute scardDWORD,
	value uintptr,
	valueLength scardDWORD,
) scardResult {
	return scardResult(callCode(
		procSCardSetAttrib,
		handle,
		uintptr(attribute),
		value,
		uintptr(valueLength),
	))
}

func callCode(proc *windows.LazyProc, args ...uintptr) uint32 {
	r, _, _ := proc.Call(args...)
	return uint32(r)
}

func establishNativeContext() (scardContext, error) {
	var context uintptr
	result := callCode(
		procSCardEstablishContext,
		scardScopeSystem,
		0,
		0,
		uintptr(unsafe.Pointer(&context)),
	)

	return context, pcscError("SCardEstablishContext", result)
}

func releaseNativeContext(context scardContext) error {
	return pcscError("SCardReleaseContext", callCode(procSCardReleaseContext, context))
}

func cancelNativeContext(context scardContext) error {
	return pcscError("SCardCancel", uint32(scardCancel(context)))
}

func listReadersNative(context scardContext) ([]string, error) {
	var size uint32
	code := callCode(procSCardListReadersW, context, 0, 0, uintptr(unsafe.Pointer(&size)))
	if err := pcscError("SCardListReadersW", code); err != nil {
		if errors.Is(err, ErrNoReaders) {
			return nil, nil
		}
		return nil, err
	}
	if size == 0 {
		return nil, nil
	}

	buffer := make([]uint16, size)
	code = callCode(
		procSCardListReadersW,
		context,
		0,
		uintptr(unsafe.Pointer(unsafe.SliceData(buffer))),
		uintptr(unsafe.Pointer(&size)),
	)
	if err := pcscError("SCardListReadersW", code); err != nil {
		return nil, err
	}

	return parseUTF16MultiString(buffer[:min(int(size), len(buffer))]), nil
}

func getStatusChangeNative(context scardContext, timeout time.Duration, states []nativeReaderState) error {
	names := make([][]uint16, len(states))
	namePointers := make([]uintptr, len(states))
	for index, state := range states {
		names[index] = append(utf16.Encode([]rune(state.name)), 0)
		namePointers[index] = uintptr(unsafe.Pointer(unsafe.SliceData(names[index])))
	}

	layout := newNativeReaderStateLayout(scardReaderStateATRSize, scardReaderStatePacked)
	buffer := layout.encode(states, namePointers)
	code := callCode(
		procSCardGetStatusChangeW,
		context,
		uintptr(durationMilliseconds(timeout)),
		uintptr(unsafe.Pointer(unsafe.SliceData(buffer))),
		uintptr(len(states)),
	)
	runtime.KeepAlive(names)

	if err := pcscError("SCardGetStatusChangeW", code); err != nil {
		return err
	}

	layout.decode(buffer, states)

	return nil
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

// Open connects to the card in reader. By default it uses shared access,
// negotiates T=0 or T=1, and leaves the card powered when closed.
func Open(reader string, opts ...OpenOption) (*Card, error) {
	options := newOpenOptions(opts...)
	context, err := establishNativeContext()
	if err != nil {
		return nil, err
	}

	name, err := windows.UTF16PtrFromString(reader)
	if err != nil {
		return nil, errors.Join(err, releaseNativeContext(context))
	}

	var handle uintptr
	var protocol uint32

	code := callCode(
		procSCardConnectW,
		context,
		uintptr(unsafe.Pointer(name)),
		uintptr(options.shareMode),
		uintptr(options.preferredProtocols),
		uintptr(unsafe.Pointer(&handle)),
		uintptr(unsafe.Pointer(&protocol)),
	)
	runtime.KeepAlive(name)

	if err := pcscError("SCardConnectW", code); err != nil {
		return nil, errors.Join(err, releaseNativeContext(context))
	}

	return &Card{
		context:               context,
		handle:                handle,
		protocol:              Protocol(protocol),
		disconnectDisposition: options.disconnectDisposition,
	}, nil
}

func (card *Card) Status() (*CardStatus, error) {
	card.mu.Lock()
	defer card.mu.Unlock()

	if card.closed {
		return nil, ErrClosed
	}

	var readerLen, atrLen, state, protocol uint32

	code := callCode(
		procSCardStatusW,
		card.handle,
		0,
		uintptr(unsafe.Pointer(&readerLen)),
		uintptr(unsafe.Pointer(&state)),
		uintptr(unsafe.Pointer(&protocol)),
		0,
		uintptr(unsafe.Pointer(&atrLen)),
	)
	if code != 0 && uint32(code) != scardEInsufficientBuf {
		return nil, pcscError("SCardStatusW", code)
	}

	reader := make([]uint16, readerLen)
	atr := make([]byte, atrLen)

	code = callCode(
		procSCardStatusW,
		card.handle,
		uintptr(unsafe.Pointer(unsafe.SliceData(reader))),
		uintptr(unsafe.Pointer(&readerLen)),
		uintptr(unsafe.Pointer(&state)),
		uintptr(unsafe.Pointer(&protocol)),
		byteSlicePointer(atr),
		uintptr(unsafe.Pointer(&atrLen)),
	)
	if err := pcscError("SCardStatusW", code); err != nil {
		return nil, err
	}
	for index, character := range reader {
		if character == 0 {
			reader = reader[:index]
			break
		}
	}

	return &CardStatus{
		ReaderName: string(utf16.Decode(reader)),
		State:      CardState(state),
		Protocol:   Protocol(protocol),
		ATR:        atr[:min(int(atrLen), len(atr))],
	}, nil
}
