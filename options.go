package pcsc

type OpenOption func(*openOptions)

type openOptions struct {
	shareMode             ShareMode
	preferredProtocols    Protocol
	preferredProtocolsSet bool
	disconnectDisposition Disposition
}

func newOpenOptions(opts ...OpenOption) openOptions {
	options := openOptions{
		shareMode:             ShareModeShared,
		preferredProtocols:    ProtocolT0 | ProtocolT1,
		disconnectDisposition: DispositionLeaveCard,
	}

	for _, option := range opts {
		option(&options)
	}

	if options.shareMode == ShareModeDirect && !options.preferredProtocolsSet {
		options.preferredProtocols = ProtocolUndefined
	}

	return options
}

// WithShareMode configures how the card connection is shared with other
// applications.
func WithShareMode(mode ShareMode) OpenOption {
	return func(options *openOptions) {
		options.shareMode = mode
	}
}

// WithPreferredProtocols configures the protocols offered to SCardConnect.
func WithPreferredProtocols(protocols Protocol) OpenOption {
	return func(options *openOptions) {
		options.preferredProtocols = protocols
		options.preferredProtocolsSet = true
	}
}

// WithDisconnectDisposition configures what Close does with the card.
func WithDisconnectDisposition(disposition Disposition) OpenOption {
	return func(options *openOptions) {
		options.disconnectDisposition = disposition
	}
}
