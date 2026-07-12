//go:build darwin || linux || windows

package pcsc

import (
	"errors"
	"testing"
	"unsafe"
)

func TestTransmitDoesNotRetryAfterInsufficientBuffer(t *testing.T) {
	original := scardTransmit
	t.Cleanup(func() { scardTransmit = original })

	calls := 0
	scardTransmit = func(_ scardHandle, sendPCI *scardIORequest, _ uintptr, _ scardDWORD, _ *scardIORequest, _ uintptr, recvLength *scardDWORD) scardResult {
		calls++
		if sendPCI == nil {
			t.Fatal("send PCI is nil")
		}
		if got, want := sendPCI.Length, scardDWORD(unsafe.Sizeof(*sendPCI)); got != want {
			t.Errorf("send PCI length = %d, want %d", got, want)
		}
		if got := *recvLength; got != scardDWORD(maxResponseBufSize) {
			t.Errorf("receive buffer size = %d, want %d", got, maxResponseBufSize)
		}

		*recvLength = 8192
		return scardResultFromCodeForTest(scardEInsufficientBuf)
	}

	c := &card{handle: scardHandle(1), protocol: ProtocolT1}
	response, err := c.Transmit([]byte{0x00, 0xa4, 0x04, 0x00})
	if response != nil {
		t.Fatalf("response = %x, want nil", response)
	}

	var pcscErr *Error
	if !errors.As(err, &pcscErr) || pcscErr.Code != scardEInsufficientBuf {
		t.Fatalf("error = %v, want PC/SC insufficient buffer", err)
	}
	if calls != 1 {
		t.Fatalf("SCardTransmit calls = %d, want 1", calls)
	}
}

func scardResultFromCodeForTest(code uint32) scardResult {
	return scardResult(int32(code))
}
