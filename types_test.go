//go:build darwin || linux || windows

package pcsc

import (
	"slices"
	"testing"
)

func TestNewCardStatusPreservesAllReaderNames(t *testing.T) {
	names := []string{"Reader", "Reader Alias"}
	status := newCardStatus(names, CardStatePresent, ProtocolT1, []byte{0x3b})

	if !slices.Equal(status.ReaderNames, names) {
		t.Fatalf("ReaderNames = %#v, want %#v", status.ReaderNames, names)
	}
}
