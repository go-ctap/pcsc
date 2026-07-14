# go-pcsc

Minimal, CGO-free PC/SC access for Go. Windows calls `winscard.dll` through
`golang.org/x/sys/windows`; macOS loads the PCSC framework and Linux loads
`libpcsclite.so.1` at runtime through `purego`.

The package intentionally exposes only the primitives needed by token clients:
reader enumeration, connect/disconnect, status/ATR and raw APDU exchange.

The opt-in hardware test expects a FIDO authenticator presented to a PC/SC
reader:

```sh
PCSC_TEST_CTAPNFC=1 go test -run TestCTAPNFC -v
```

```go
import "context"
import "github.com/go-ctap/pcsc"

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
	log.Printf("reader=%q ATR=%x protocol=%d", status.Reader, status.ATR, status.Protocol)

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

Canceling an in-flight `Transmit` issues a best-effort `SCardCancel` for the
card's PC/SC context and returns `ctx.Err()`.
PC/SC implementations and reader drivers do not consistently guarantee that an
in-flight APDU can be interrupted, so the native operation may continue after
`Transmit` returns. The card remains serialized until that operation finishes.
Do not automatically retry a canceled APDU: the card may already have processed
it.
