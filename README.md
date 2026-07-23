# go-ctap/pcsc

[![Go Reference](https://pkg.go.dev/badge/github.com/go-ctap/pcsc.svg)](https://pkg.go.dev/github.com/go-ctap/pcsc)
[![Go](https://github.com/go-ctap/pcsc/actions/workflows/go.yml/badge.svg)](https://github.com/go-ctap/pcsc/actions/workflows/go.yml)

Minimal, CGO-free PC/SC access for Go. Windows calls `winscard.dll` through
`golang.org/x/sys/windows`; macOS loads the PCSC framework and Linux loads
`libpcsclite.so.1` at runtime through `purego`.

Importing the package does not load the native PC/SC runtime. If it is not
installed, `Enumerate` yields and `Open` returns `pcsc.ErrUnavailable`, so
applications can keep PC/SC support optional.

The package exposes the PC/SC reader and card lifecycle needed by hardware-token
clients:

- reader enumeration and ordered connection events;
- shared, exclusive and direct card connections;
- status and ATR inspection;
- raw APDU and reader-control exchange;
- transactions and reconnect;
- reader and card attributes.

It deliberately does not implement ISO 7816 APDU encoding, status-word
handling, command chaining or application protocols such as CTAP, PIV and
OpenPGP.

The generic lifecycle hardware test is read-only and works with any connected
smart card:

```sh
PCSC_TEST=1 go test -run TestPCSCLifecycle -v
```

The CTAP-over-NFC hardware test expects a FIDO authenticator presented to a
PC/SC reader:

```sh
PCSC_TEST_CTAPNFC=1 go test -run TestCTAPNFC -v
```

```go
import (
	"context"

	"github.com/go-ctap/pcsc"
)

ctx := context.Background()

for reader, err := range pcsc.Enumerate() {
	if err != nil {
		log.Fatal(err)
	}

	card, err := pcsc.Open(reader.Name)
	if err != nil {
		continue
	}
	defer card.Close()

	status, err := card.Status()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("reader=%q ATR=%x protocol=%d", status.ReaderName, status.ATR, status.Protocol)

	// SELECT the standard FIDO applet.
	response, err := card.Transmit(ctx, []byte{
		0x00, 0xa4, 0x04, 0x00, 0x08,
		0xa0, 0x00, 0x00, 0x06, 0x47, 0x2f, 0x00, 0x01,
		0x00,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("response=%x", response) // response includes SW1/SW2
}
```

`Events` first publishes the current reader and card snapshot and then live
changes. Each receiver owns an independent PC/SC context:

```go
receiver, err := pcsc.Events()
if err != nil {
	log.Fatal(err)
}
defer receiver.Close()

for event := range receiver.Listen() {
	log.Printf("type=%s reader=%q", event.Type, event.ReaderInfo.Name)
}
```

Transactions map directly to `SCardBeginTransaction` and
`SCardEndTransaction`. They prevent other PC/SC applications from interleaving
commands, but they do not provide rollback:

```go
if err := card.BeginTransaction(ctx); err != nil {
	log.Fatal(err)
}
defer card.EndTransaction(pcsc.LeaveCard)
```

Canceling an in-flight contextual operation issues a best-effort `SCardCancel`
for the card's PC/SC context and returns `ctx.Err()`.
PC/SC implementations and reader drivers do not consistently guarantee that an
in-flight operation can be interrupted, so it may continue after the Go method
returns. The card remains serialized until the native call finishes. Do not
automatically retry a canceled APDU or control request: the device may already
have processed it.
