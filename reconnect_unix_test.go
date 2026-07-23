//go:build darwin || linux

package pcsc

import "testing"

func TestUnixReconnectAllowsDirectAndEject(t *testing.T) {
	if err := validateReconnectParameters(ShareModeDirect, DispositionEjectCard); err != nil {
		t.Fatalf("validateReconnectParameters: %v", err)
	}
}
