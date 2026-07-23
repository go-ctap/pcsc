//go:build windows

package pcsc

const (
	ProtocolRaw     Protocol = 0x00010000
	ProtocolDefault Protocol = 0x80000000
	ProtocolOptimal Protocol = ProtocolUndefined
	ProtocolTx      Protocol = ProtocolT0 | ProtocolT1
)

// Windows reports SCardStatus states as an enum rather than the bitmask used
// by pcsc-lite.
const (
	CardStateUnknown    CardState = 0
	CardStateAbsent     CardState = 1
	CardStatePresent    CardState = 2
	CardStateSwallowed  CardState = 3
	CardStatePowered    CardState = 4
	CardStateNegotiable CardState = 5
	CardStateSpecific   CardState = 6
)

// ControlCode converts a PC/SC function number to the platform control code
// accepted by Card.Control.
func ControlCode(function uint32) uint32 {
	const fileDeviceSmartCard = 0x31

	return fileDeviceSmartCard<<16 | function<<2
}
