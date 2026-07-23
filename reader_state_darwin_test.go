//go:build darwin

package pcsc

import "testing"

func TestDarwinReaderStateABI(t *testing.T) {
	layout := newNativeReaderStateLayout(scardReaderStateATRSize, scardReaderStatePacked)
	if layout.stride != 61 {
		t.Fatalf("packed SCARD_READERSTATE stride = %d, want 61", layout.stride)
	}
}

func TestDarwinUsesStandardSCardControlSymbol(t *testing.T) {
	if scardControlSymbol != "SCardControl132" {
		t.Fatalf("SCardControl symbol = %q, want SCardControl132", scardControlSymbol)
	}
}
