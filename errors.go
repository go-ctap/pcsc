package pcsc

import (
	"errors"
	"fmt"
)

var (
	// ErrUnavailable indicates that the platform PC/SC runtime could not be loaded.
	ErrUnavailable = errors.New("pcsc: unavailable")

	ErrCanceled           = errors.New("pcsc: canceled")
	ErrTimeout            = errors.New("pcsc: timeout")
	ErrInsufficientBuffer = errors.New("pcsc: insufficient buffer")
	ErrUnknownReader      = errors.New("pcsc: unknown reader")
	ErrNoCard             = errors.New("pcsc: no smart card")
	ErrCannotDispose      = errors.New("pcsc: cannot dispose")
	ErrProtocolMismatch   = errors.New("pcsc: protocol mismatch")
	ErrNotReady           = errors.New("pcsc: not ready")
	ErrNotTransacted      = errors.New("pcsc: not transacted")
	ErrReaderUnavailable  = errors.New("pcsc: reader unavailable")
	ErrCardReset          = errors.New("pcsc: card reset")
	ErrCardRemoved        = errors.New("pcsc: card removed")
	ErrSharingViolation   = errors.New("pcsc: sharing violation")
	ErrNoService          = errors.New("pcsc: no service")
	ErrServiceStopped     = errors.New("pcsc: service stopped")
	ErrNoReaders          = errors.New("pcsc: no readers available")
	ErrClosed             = errors.New("pcsc: card closed")
)

// Error is a PC/SC return code annotated with the operation which failed.
type Error struct {
	Operation string
	Code      uint32
}

func (e *Error) Error() string {
	if info, ok := errorsByCode[e.Code]; ok {
		return fmt.Sprintf("pcsc: %s: %s (0x%08x)", e.Operation, info.name, e.Code)
	}

	return fmt.Sprintf("pcsc: %s failed (0x%08x)", e.Operation, e.Code)
}

func (e *Error) Is(target error) bool {
	info, ok := errorsByCode[e.Code]

	return ok && info.err != nil && target == info.err
}

func pcscError(operation string, code uint32) error {
	if code == 0 {
		return nil
	}

	return &Error{Operation: operation, Code: code}
}

var errorsByCode = map[uint32]struct {
	name string
	err  error
}{
	0x80100002: {name: "cancelled", err: ErrCanceled},
	0x80100008: {name: "insufficient buffer", err: ErrInsufficientBuffer},
	0x80100009: {name: "unknown reader", err: ErrUnknownReader},
	0x8010000a: {name: "timeout", err: ErrTimeout},
	0x8010000b: {name: "sharing violation", err: ErrSharingViolation},
	0x8010000c: {name: "no smart card", err: ErrNoCard},
	0x8010000e: {name: "cannot dispose", err: ErrCannotDispose},
	0x8010000f: {name: "protocol mismatch", err: ErrProtocolMismatch},
	0x80100010: {name: "not ready", err: ErrNotReady},
	0x80100016: {name: "not transacted", err: ErrNotTransacted},
	0x80100017: {name: "reader unavailable", err: ErrReaderUnavailable},
	0x8010001d: {name: "no service", err: ErrNoService},
	0x8010001e: {name: "service stopped", err: ErrServiceStopped},
	0x8010002e: {name: "no readers available", err: ErrNoReaders},
	0x80100068: {name: "card reset", err: ErrCardReset},
	0x80100069: {name: "card removed", err: ErrCardRemoved},
}
