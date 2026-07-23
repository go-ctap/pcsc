//go:build darwin || linux

package pcsc

const ProtocolRaw Protocol = 0x0004

// pcsc-lite and Apple's PCSC framework report SCardStatus states as a bitmask.
const (
	CardStateUnknown    CardState = 0x0001
	CardStateAbsent     CardState = 0x0002
	CardStatePresent    CardState = 0x0004
	CardStateSwallowed  CardState = 0x0008
	CardStatePowered    CardState = 0x0010
	CardStateNegotiable CardState = 0x0020
	CardStateSpecific   CardState = 0x0040
)

// ControlCode converts a PC/SC function number to the platform control code
// accepted by Card.Control.
func ControlCode(function uint32) uint32 {
	return 0x42000000 + function
}
