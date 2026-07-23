package pcsc

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"
)

func TestDeviceEventQueueBurstIsLosslessAndOrdered(t *testing.T) {
	queue := newDeviceEventQueue()
	const count = 1000
	for index := range count {
		if !queue.Send(DeviceEvent{ReaderInfo: &ReaderInfo{Name: fmt.Sprintf("reader-%04d", index)}}) {
			t.Fatalf("Send(%d) rejected", index)
		}
	}

	for index := range count {
		select {
		case event := <-queue.Listen():
			want := fmt.Sprintf("reader-%04d", index)
			if event.ReaderInfo == nil || event.ReaderInfo.Name != want {
				t.Fatalf("event %d = %#v, want reader %q", index, event, want)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out receiving event %d", index)
		}
	}

	queue.Close()
	if _, ok := <-queue.Listen(); ok {
		t.Fatal("event channel remains open after Close")
	}
}

func TestDeviceEventQueueCloseWithoutListener(t *testing.T) {
	queue := newDeviceEventQueue()
	for range 100 {
		if !queue.Send(DeviceEvent{}) {
			t.Fatal("Send rejected before Close")
		}
	}

	done := make(chan struct{})
	go func() {
		queue.Close()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Close blocked without a listener")
	}
	if queue.Send(DeviceEvent{}) {
		t.Fatal("Send accepted after Close")
	}
}

func TestDeviceEventQueueConcurrentSendAndClose(t *testing.T) {
	queue := newDeviceEventQueue()

	const senderCount = 32
	var senders sync.WaitGroup
	var firstSends sync.WaitGroup
	firstSends.Add(senderCount)
	senders.Add(senderCount)
	for range senderCount {
		go func() {
			defer senders.Done()

			queue.Send(DeviceEvent{})
			firstSends.Done()
			for queue.Send(DeviceEvent{}) {
			}
		}()
	}

	firstSends.Wait()

	const closerCount = 8
	var closers sync.WaitGroup
	closers.Add(closerCount)
	for range closerCount {
		go func() {
			defer closers.Done()

			queue.Close()
		}()
	}

	finished := make(chan struct{})
	go func() {
		senders.Wait()
		closers.Wait()
		close(finished)
	}()

	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("concurrent Send and Close did not finish")
	}

	if _, ok := <-queue.Listen(); ok {
		t.Fatal("event channel remains open after concurrent Close")
	}
}

func TestReconcileDeviceEvents(t *testing.T) {
	presentA := &ReaderInfo{Name: "A", State: ReaderStatePresent, ATR: []byte{1}}
	emptyB := &ReaderInfo{Name: "B", State: ReaderStateEmpty}
	presentB := &ReaderInfo{Name: "B", State: ReaderStatePresent, ATR: []byte{2}}
	replacedA := &ReaderInfo{Name: "A", State: ReaderStatePresent, ATR: []byte{3}}

	tests := []struct {
		name     string
		previous map[string]*ReaderInfo
		current  map[string]*ReaderInfo
		want     []DeviceEventType
	}{
		{
			name:    "initial snapshot",
			current: map[string]*ReaderInfo{"B": emptyB, "A": presentA},
			want: []DeviceEventType{
				DeviceEventReaderConnected,
				DeviceEventCardInserted,
				DeviceEventReaderConnected,
			},
		},
		{
			name:     "card inserted",
			previous: map[string]*ReaderInfo{"B": emptyB},
			current:  map[string]*ReaderInfo{"B": presentB},
			want:     []DeviceEventType{DeviceEventCardInserted},
		},
		{
			name:     "card replaced without empty observation",
			previous: map[string]*ReaderInfo{"A": presentA},
			current:  map[string]*ReaderInfo{"A": replacedA},
			want:     []DeviceEventType{DeviceEventCardRemoved, DeviceEventCardInserted},
		},
		{
			name:     "reader removed with card",
			previous: map[string]*ReaderInfo{"A": presentA},
			want:     []DeviceEventType{DeviceEventCardRemoved, DeviceEventReaderDisconnected},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			events := reconcileDeviceEvents(test.previous, test.current)
			got := make([]DeviceEventType, len(events))
			for index, event := range events {
				got[index] = event.Type
			}
			if !slices.Equal(got, test.want) {
				t.Fatalf("event types = %v, want %v", got, test.want)
			}
		})
	}
}

func TestSnapshotFromReaderStatesNormalizesChangedBit(t *testing.T) {
	states := []nativeReaderState{
		{name: "reader", eventState: ReaderStateChanged | ReaderStatePresent, atr: []byte{1, 2}},
		{name: pnpNotificationReader, eventState: ReaderStateChanged | ReaderStateUnknown},
	}
	readers, pnpState := snapshotFromReaderStates(states)
	if got := readers["reader"].State; got != ReaderStatePresent {
		t.Fatalf("reader state = 0x%x, want 0x%x", got, ReaderStatePresent)
	}
	if pnpState != ReaderStateUnknown {
		t.Fatalf("PnP state = 0x%x, want 0x%x", pnpState, ReaderStateUnknown)
	}
}

func TestStatusChangeFallsBackWhenPnPReaderIsUnsupported(t *testing.T) {
	original := getStatusChange
	t.Cleanup(func() { getStatusChange = original })

	calls := 0
	getStatusChange = func(_ scardContext, _ time.Duration, states []nativeReaderState) error {
		calls++
		pnp := &states[len(states)-1]
		if calls == 1 {
			if pnp.currentState != ReaderStateUnaware {
				t.Fatalf("initial PnP state = 0x%x, want unaware", pnp.currentState)
			}

			return &Error{Operation: "SCardGetStatusChange", Code: 0x80100009}
		}

		if pnp.currentState != ReaderStateIgnore {
			t.Fatalf("fallback PnP state = 0x%x, want ignore", pnp.currentState)
		}

		return ErrTimeout
	}

	states := []nativeReaderState{
		{name: "reader"},
		{name: pnpNotificationReader},
	}
	err := waitForStatusChange(scardContext(1), 0, states)
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("waitForStatusChange error = %v, want ErrTimeout", err)
	}
	if calls != 2 {
		t.Fatalf("status-change calls = %d, want 2", calls)
	}
}
