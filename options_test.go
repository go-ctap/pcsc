package pcsc

import "testing"

func TestOpenOptions(t *testing.T) {
	options := newOpenOptions(
		WithShareMode(ShareExclusive),
		WithPreferredProtocols(ProtocolT1),
		WithDisconnectDisposition(ResetCard),
	)

	if options.shareMode != ShareExclusive {
		t.Fatalf("share mode = %d, want %d", options.shareMode, ShareExclusive)
	}
	if options.preferredProtocols != ProtocolT1 {
		t.Fatalf("preferred protocols = %d, want %d", options.preferredProtocols, ProtocolT1)
	}
	if options.disconnectDisposition != ResetCard {
		t.Fatalf("disconnect disposition = %d, want %d", options.disconnectDisposition, ResetCard)
	}
}

func TestDirectOpenDefaultsToUndefinedProtocol(t *testing.T) {
	options := newOpenOptions(WithShareMode(ShareDirect))

	if options.preferredProtocols != ProtocolUndefined {
		t.Fatalf("preferred protocols = %d, want undefined", options.preferredProtocols)
	}
}
