package pcsc

import "sync"

// DeviceEvent describes a PC/SC reader or card-presence change.
type DeviceEvent struct {
	Type       DeviceEventType
	ReaderInfo *ReaderInfo
	// Err is non-nil when the receiver stops because its background PC/SC
	// operation failed. Such an event is the final item from Listen.
	Err error
}

type DeviceEventType string

const (
	DeviceEventReaderConnected    DeviceEventType = "reader-connected"
	DeviceEventReaderDisconnected DeviceEventType = "reader-disconnected"
	DeviceEventCardInserted       DeviceEventType = "card-inserted"
	DeviceEventCardRemoved        DeviceEventType = "card-removed"
)

type EventReceiver interface {
	Listen() <-chan DeviceEvent
	Close() error
}

// deviceEventQueue decouples the native PC/SC wait from event consumers.
// Send only retains the event and wakes the forwarding goroutine; it never
// waits for a consumer to receive from Listen.
type deviceEventQueue struct {
	mu      sync.Mutex
	pending []queuedDeviceEvent
	closed  bool

	out     chan DeviceEvent
	wake    chan struct{}
	done    chan struct{}
	stopped chan struct{}

	closeOnce sync.Once
}

type queuedDeviceEvent struct {
	event     DeviceEvent
	delivered chan bool
}

func newDeviceEventQueue() *deviceEventQueue {
	q := &deviceEventQueue{
		out:     make(chan DeviceEvent),
		wake:    make(chan struct{}, 1),
		done:    make(chan struct{}),
		stopped: make(chan struct{}),
	}
	go q.run()
	return q
}

// Send enqueues an event without waiting for a listener. It reports false once
// closing has started; calls racing with Close are safe.
func (q *deviceEventQueue) Send(event DeviceEvent) bool {
	return q.send(queuedDeviceEvent{event: event})
}

// SendTerminal enqueues a final event and waits until it is delivered or the
// queue is closed. Event receivers use it to make terminal errors observable
// before closing Listen.
func (q *deviceEventQueue) SendTerminal(event DeviceEvent) bool {
	delivered := make(chan bool, 1)
	if !q.send(queuedDeviceEvent{event: event, delivered: delivered}) {
		return false
	}

	return <-delivered
}

func (q *deviceEventQueue) send(event queuedDeviceEvent) bool {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		return false
	}
	q.pending = append(q.pending, event)
	q.mu.Unlock()

	select {
	case q.wake <- struct{}{}:
	default:
	}

	return true
}

func (q *deviceEventQueue) Listen() <-chan DeviceEvent {
	return q.out
}

// Close stops accepting events and cancels delivery blocked on an absent
// consumer. It waits for out to close, but never waits for out to drain.
func (q *deviceEventQueue) Close() {
	q.closeOnce.Do(func() {
		q.mu.Lock()
		q.closed = true
		pending := q.pending
		q.pending = nil
		q.mu.Unlock()

		for _, event := range pending {
			notifyDeviceEventDelivery(event, false)
		}
		close(q.done)
		<-q.stopped
	})
}

func (q *deviceEventQueue) run() {
	defer close(q.stopped)
	defer close(q.out)

	for {
		q.mu.Lock()
		if len(q.pending) == 0 {
			q.mu.Unlock()
			select {
			case <-q.wake:
				continue
			case <-q.done:
				return
			}
		}

		event := q.pending[0]
		q.pending[0] = queuedDeviceEvent{}
		if len(q.pending) == 1 {
			q.pending = nil
		} else {
			q.pending = q.pending[1:]
		}
		q.mu.Unlock()

		select {
		case q.out <- event.event:
			notifyDeviceEventDelivery(event, true)
		case <-q.done:
			notifyDeviceEventDelivery(event, false)
			return
		}
	}
}

func notifyDeviceEventDelivery(event queuedDeviceEvent, delivered bool) {
	if event.delivered != nil {
		event.delivered <- delivered
	}
}
