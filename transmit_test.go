//go:build darwin || linux || windows

package pcsc

import (
	"context"
	"errors"
	"testing"
	"time"
	"unsafe"
)

func TestTransmitDoesNotRetryAfterInsufficientBuffer(t *testing.T) {
	original := scardTransmit
	t.Cleanup(func() { scardTransmit = original })

	calls := 0
	scardTransmit = func(
		_ scardHandle,
		sendPCI *scardIORequest,
		_ uintptr,
		_ scardDWORD,
		_ *scardIORequest,
		_ uintptr,
		recvLength *scardDWORD,
	) scardResult {
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

	c := &Card{handle: scardHandle(1), protocol: ProtocolT1}
	response, err := c.Transmit(context.Background(), []byte{0x00, 0xa4, 0x04, 0x00})
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

func TestTransmitCancellationReturnsPromptlyAndCancelsContext(t *testing.T) {
	originalTransmit := scardTransmit
	originalCancel := scardCancel
	t.Cleanup(func() {
		scardTransmit = originalTransmit
		scardCancel = originalCancel
	})

	started := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan struct{})
	cancelCalled := make(chan struct{})

	scardTransmit = func(
		_ scardHandle,
		_ *scardIORequest,
		_ uintptr,
		_ scardDWORD,
		_ *scardIORequest,
		_ uintptr,
		recvLength *scardDWORD,
	) scardResult {
		close(started)
		<-release
		*recvLength = 0
		close(finished)
		return 0
	}
	scardCancel = func(_ scardContext) scardResult {
		close(cancelCalled)
		return 0
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := &Card{context: scardContext(1), handle: scardHandle(1), protocol: ProtocolT1}
	result := make(chan error, 1)
	go func() {
		_, err := c.Transmit(ctx, []byte{0x00, 0xa4, 0x04, 0x00})
		result <- err
	}()

	<-started
	cancel()

	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Transmit did not return after context cancellation")
	}

	select {
	case <-cancelCalled:
	case <-time.After(time.Second):
		t.Fatal("SCardCancel was not called")
	}

	close(release)
	<-finished
}

func TestTransmitDoesNotStartAfterCancellationWhileQueued(t *testing.T) {
	original := scardTransmit
	t.Cleanup(func() { scardTransmit = original })

	scardTransmit = func(
		_ scardHandle,
		_ *scardIORequest,
		_ uintptr,
		_ scardDWORD,
		_ *scardIORequest,
		_ uintptr,
		_ *scardDWORD,
	) scardResult {
		t.Fatal("SCardTransmit called after cancellation while waiting for the card")
		return 0
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := &Card{handle: scardHandle(1), protocol: ProtocolT1}
	c.mu.Lock()

	result := make(chan error, 1)
	go func() {
		_, err := c.Transmit(ctx, []byte{0x00})
		result <- err
	}()

	cancel()
	c.mu.Unlock()
	err := <-result
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func scardResultFromCodeForTest(code uint32) scardResult {
	return scardResult(int32(code))
}
