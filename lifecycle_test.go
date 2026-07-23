//go:build darwin || linux || windows

package pcsc

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBeginTransactionCancellationReturnsPromptly(t *testing.T) {
	originalBegin := scardBeginTransaction
	originalCancel := scardCancel
	originalEnd := scardEndTransaction
	t.Cleanup(func() {
		scardBeginTransaction = originalBegin
		scardCancel = originalCancel
		scardEndTransaction = originalEnd
	})

	started := make(chan struct{})
	release := make(chan struct{})
	cancelCalled := make(chan struct{})
	cleanupCalled := make(chan struct{})

	scardBeginTransaction = func(scardHandle) scardResult {
		close(started)
		<-release

		return 0
	}
	scardCancel = func(scardContext) scardResult {
		close(cancelCalled)

		return 0
	}
	scardEndTransaction = func(_ scardHandle, disposition scardDWORD) scardResult {
		if disposition != scardDWORD(LeaveCard) {
			t.Errorf("cleanup disposition = %d, want %d", disposition, LeaveCard)
		}
		close(cleanupCalled)

		return 0
	}

	ctx, cancel := context.WithCancel(context.Background())
	card := &Card{context: scardContext(1), handle: scardHandle(1)}
	result := make(chan error, 1)
	go func() {
		result <- card.BeginTransaction(ctx)
	}()

	<-started
	cancel()

	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("BeginTransaction error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("BeginTransaction did not return after cancellation")
	}

	select {
	case <-cancelCalled:
	case <-time.After(time.Second):
		t.Fatal("SCardCancel was not called")
	}

	close(release)
	select {
	case <-cleanupCalled:
	case <-time.After(time.Second):
		t.Fatal("late successful transaction was not ended")
	}
}

func TestEndTransactionForwardsDisposition(t *testing.T) {
	original := scardEndTransaction
	t.Cleanup(func() { scardEndTransaction = original })

	var gotDisposition scardDWORD
	scardEndTransaction = func(_ scardHandle, disposition scardDWORD) scardResult {
		gotDisposition = disposition

		return 0
	}

	card := &Card{handle: scardHandle(1)}
	if err := card.EndTransaction(ResetCard); err != nil {
		t.Fatalf("EndTransaction: %v", err)
	}
	if gotDisposition != scardDWORD(ResetCard) {
		t.Fatalf("disposition = %d, want %d", gotDisposition, ResetCard)
	}
}

func TestReconnectUpdatesActiveProtocol(t *testing.T) {
	original := scardReconnect
	t.Cleanup(func() { scardReconnect = original })

	scardReconnect = func(
		_ scardHandle,
		shareMode scardDWORD,
		preferredProtocols scardDWORD,
		initialization scardDWORD,
		activeProtocol *scardDWORD,
	) scardResult {
		if shareMode != scardDWORD(ShareExclusive) {
			t.Errorf("share mode = %d, want %d", shareMode, ShareExclusive)
		}
		if preferredProtocols != scardDWORD(ProtocolT1) {
			t.Errorf("preferred protocols = %d, want %d", preferredProtocols, ProtocolT1)
		}
		if initialization != scardDWORD(ResetCard) {
			t.Errorf("initialization = %d, want %d", initialization, ResetCard)
		}

		*activeProtocol = scardDWORD(ProtocolT1)

		return 0
	}

	card := &Card{handle: scardHandle(1), protocol: ProtocolT0}
	protocol, err := card.Reconnect(
		context.Background(),
		ShareExclusive,
		ProtocolT1,
		ResetCard,
	)
	if err != nil {
		t.Fatalf("Reconnect: %v", err)
	}
	if protocol != ProtocolT1 {
		t.Fatalf("returned protocol = %d, want %d", protocol, ProtocolT1)
	}
	if card.protocol != ProtocolT1 {
		t.Fatalf("stored protocol = %d, want %d", card.protocol, ProtocolT1)
	}
}

func TestClosedCardRejectsLifecycleOperations(t *testing.T) {
	card := &Card{closed: true}

	if err := card.BeginTransaction(context.Background()); !errors.Is(err, ErrClosed) {
		t.Fatalf("BeginTransaction error = %v, want ErrClosed", err)
	}
	if err := card.EndTransaction(LeaveCard); !errors.Is(err, ErrClosed) {
		t.Fatalf("EndTransaction error = %v, want ErrClosed", err)
	}
	if _, err := card.Reconnect(context.Background(), ShareShared, ProtocolT1, LeaveCard); !errors.Is(err, ErrClosed) {
		t.Fatalf("Reconnect error = %v, want ErrClosed", err)
	}
}
