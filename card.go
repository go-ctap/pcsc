//go:build darwin || linux || windows

package pcsc

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"unsafe"
)

const (
	scardEInsufficientBuf = uint32(0x80100008)
	maxResponseBufSize    = 65538
)

type scardIORequest struct {
	Protocol scardDWORD
	Length   scardDWORD
}

// Card is a connection to a smart card. Context cancellation is best-effort:
// a driver may continue an in-flight operation after the method returns.
type Card struct {
	mu                    sync.Mutex
	context               scardContext
	handle                scardHandle
	protocol              Protocol
	disconnectDisposition Disposition
	closed                bool
}

func scardError(operation string, result scardResult) error {
	return pcscError(operation, uint32(result))
}

func runCardOperation[T any](
	card *Card,
	ctx context.Context,
	operation func() (T, error),
) (T, error) {
	var zero T

	card.mu.Lock()

	if err := ctx.Err(); err != nil {
		card.mu.Unlock()

		return zero, err
	}

	if card.closed {
		card.mu.Unlock()

		return zero, ErrClosed
	}

	type operationResult struct {
		value T
		err   error
	}
	result := make(chan operationResult, 1)
	go func() {
		defer card.mu.Unlock()

		value, err := operation()
		result <- operationResult{value: value, err: err}
	}()

	select {
	case <-ctx.Done():
		_ = cancelNativeContext(card.context)
		return zero, ctx.Err()
	case result := <-result:
		return result.value, result.err
	}
}

// BeginTransaction prevents other PC/SC applications from interleaving card
// operations until EndTransaction is called.
func (card *Card) BeginTransaction(ctx context.Context) error {
	card.mu.Lock()

	if err := ctx.Err(); err != nil {
		card.mu.Unlock()

		return err
	}

	if card.closed {
		card.mu.Unlock()

		return ErrClosed
	}

	result := make(chan error, 1)
	accepted := make(chan struct{})
	abandoned := make(chan struct{})
	go func() {
		defer card.mu.Unlock()

		err := scardError("SCardBeginTransaction", scardBeginTransaction(card.handle))
		select {
		case result <- err:
			select {
			case <-accepted:
			case <-abandoned:
				if err == nil {
					_ = scardEndTransaction(card.handle, scardDWORD(LeaveCard))
				}
			}
		case <-abandoned:
			if err == nil {
				_ = scardEndTransaction(card.handle, scardDWORD(LeaveCard))
			}
		}
	}()

	select {
	case <-ctx.Done():
		close(abandoned)
		_ = cancelNativeContext(card.context)

		return ctx.Err()
	case err := <-result:
		close(accepted)

		return err
	}
}

// EndTransaction releases a transaction and applies disposition to the card.
func (card *Card) EndTransaction(disposition Disposition) error {
	card.mu.Lock()
	defer card.mu.Unlock()

	if card.closed {
		return ErrClosed
	}

	return scardError(
		"SCardEndTransaction",
		scardEndTransaction(card.handle, scardDWORD(disposition)),
	)
}

// Reconnect changes the connection parameters and optionally resets the card.
func (card *Card) Reconnect(
	ctx context.Context,
	shareMode ShareMode,
	preferredProtocols Protocol,
	initialization Disposition,
) (Protocol, error) {
	return runCardOperation(card, ctx, func() (Protocol, error) {
		var protocol scardDWORD
		result := scardReconnect(
			card.handle,
			scardDWORD(shareMode),
			scardDWORD(preferredProtocols),
			scardDWORD(initialization),
			&protocol,
		)
		if err := scardError("SCardReconnect", result); err != nil {
			return ProtocolUndefined, err
		}

		card.protocol = Protocol(protocol)

		return card.protocol, nil
	})
}

// Transmit sends one raw APDU and returns the complete response, including
// SW1/SW2.
func (card *Card) Transmit(ctx context.Context, apdu []byte) ([]byte, error) {
	return runCardOperation(card, ctx, func() ([]byte, error) {
		request := scardIORequest{
			Protocol: scardDWORD(card.protocol),
			Length:   scardDWORD(unsafe.Sizeof(scardIORequest{})),
		}
		response := make([]byte, maxResponseBufSize)
		size := scardDWORD(len(response))
		result := scardTransmit(
			card.handle,
			&request,
			byteSlicePointer(apdu),
			scardDWORD(len(apdu)),
			nil,
			byteSlicePointer(response),
			&size,
		)
		runtime.KeepAlive(apdu)
		if err := scardError("SCardTransmit", result); err != nil {
			return nil, err
		}

		return response[:min(int(size), len(response))], nil
	})
}

// Control sends a reader-specific control request.
func (card *Card) Control(
	ctx context.Context,
	controlCode uint32,
	input []byte,
) ([]byte, error) {
	return runCardOperation(card, ctx, func() ([]byte, error) {
		response := make([]byte, maxResponseBufSize)
		var size scardDWORD
		result := scardControl(
			card.handle,
			scardDWORD(controlCode),
			byteSlicePointer(input),
			scardDWORD(len(input)),
			byteSlicePointer(response),
			scardDWORD(len(response)),
			&size,
		)
		runtime.KeepAlive(input)
		if err := scardError("SCardControl", result); err != nil {
			return nil, err
		}

		return response[:min(int(size), len(response))], nil
	})
}

// GetAttribute returns a reader or card attribute.
func (card *Card) GetAttribute(attribute Attribute) ([]byte, error) {
	card.mu.Lock()
	defer card.mu.Unlock()

	if card.closed {
		return nil, ErrClosed
	}

	var size scardDWORD
	result := scardGetAttrib(card.handle, scardDWORD(attribute), 0, &size)
	if result != 0 && uint32(result) != scardEInsufficientBuf {
		return nil, scardError("SCardGetAttrib", result)
	}
	if size == 0 {
		return nil, nil
	}

	value := make([]byte, size)
	result = scardGetAttrib(card.handle, scardDWORD(attribute), byteSlicePointer(value), &size)
	if err := scardError("SCardGetAttrib", result); err != nil {
		return nil, err
	}

	return value[:min(int(size), len(value))], nil
}

// SetAttribute sets a reader or card attribute.
func (card *Card) SetAttribute(attribute Attribute, value []byte) error {
	card.mu.Lock()
	defer card.mu.Unlock()

	if card.closed {
		return ErrClosed
	}

	result := scardSetAttrib(
		card.handle,
		scardDWORD(attribute),
		byteSlicePointer(value),
		scardDWORD(len(value)),
	)
	runtime.KeepAlive(value)

	return scardError("SCardSetAttrib", result)
}

// Close disconnects from the card and releases its PC/SC context.
func (card *Card) Close() error {
	card.mu.Lock()
	defer card.mu.Unlock()

	if card.closed {
		return nil
	}
	card.closed = true

	return errors.Join(
		scardError(
			"SCardDisconnect",
			scardDisconnect(card.handle, scardDWORD(card.disconnectDisposition)),
		),
		releaseNativeContext(card.context),
	)
}

func byteSlicePointer(value []byte) uintptr {
	if len(value) == 0 {
		return 0
	}

	return uintptr(unsafe.Pointer(unsafe.SliceData(value)))
}
