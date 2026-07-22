//go:build windows

package pcsc

import (
	"slices"
	"testing"
)

func TestParseUTF16MultiString(t *testing.T) {
	got := parseUTF16MultiString([]uint16{
		'R', 'e', 'a', 'd', 'e', 'r', ' ', 'A', 0,
		'R', 'e', 'a', 'd', 'e', 'r', ' ', 'B', 0,
		0,
		'j', 'u', 'n', 'k',
	})
	want := []string{"Reader A", "Reader B"}
	if !slices.Equal(got, want) {
		t.Fatalf("parseUTF16MultiString() = %#v, want %#v", got, want)
	}
}

func TestWindowsRawProtocolValue(t *testing.T) {
	if ProtocolRaw != 0x00010000 {
		t.Fatalf("ProtocolRaw = 0x%x, want 0x00010000", ProtocolRaw)
	}
}
