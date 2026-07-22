//go:build darwin || linux

package pcsc

import (
	"slices"
	"testing"
)

func TestParseMultiString(t *testing.T) {
	got := parseMultiString([]byte("Reader A\x00Reader B\x00\x00junk"))
	want := []string{"Reader A", "Reader B"}
	if !slices.Equal(got, want) {
		t.Fatalf("parseMultiString() = %#v, want %#v", got, want)
	}
}

func TestUnixRawProtocolValue(t *testing.T) {
	if ProtocolRaw != 0x0004 {
		t.Fatalf("ProtocolRaw = 0x%x, want 0x0004", ProtocolRaw)
	}
}
