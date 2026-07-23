package pcsc

import (
	"errors"
	"testing"
)

func TestPCSCErrorMatchesStableErrors(t *testing.T) {
	tests := []struct {
		code uint32
		want error
	}{
		{code: 0x80100002, want: ErrCanceled},
		{code: 0x80100008, want: ErrInsufficientBuffer},
		{code: 0x80100009, want: ErrUnknownReader},
		{code: 0x8010000a, want: ErrTimeout},
		{code: 0x8010000b, want: ErrSharingViolation},
		{code: 0x8010000c, want: ErrNoCard},
		{code: 0x8010000e, want: ErrCannotDispose},
		{code: 0x8010000f, want: ErrProtocolMismatch},
		{code: 0x80100010, want: ErrNotReady},
		{code: 0x80100016, want: ErrNotTransacted},
		{code: 0x80100017, want: ErrReaderUnavailable},
		{code: 0x8010001d, want: ErrNoService},
		{code: 0x8010001e, want: ErrServiceStopped},
		{code: 0x8010002e, want: ErrNoReaders},
		{code: 0x80100068, want: ErrCardReset},
		{code: 0x80100069, want: ErrCardRemoved},
	}

	for _, test := range tests {
		err := pcscError("test", test.code)
		if !errors.Is(err, test.want) {
			t.Errorf("pcscError(0x%08x) = %v, want errors.Is(_, %v)", test.code, err, test.want)
		}
	}
}
