//go:build linux

package pcsc

import "testing"

func TestLinuxT15ProtocolValue(t *testing.T) {
	if ProtocolT15 != 0x0008 {
		t.Fatalf("ProtocolT15 = 0x%x, want 0x0008", ProtocolT15)
	}
}
