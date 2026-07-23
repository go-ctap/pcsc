//go:build darwin || linux

package pcsc

// pcsc-lite historically assigns SCARD_E_UNSUPPORTED_FEATURE the same value
// as SCARD_E_UNEXPECTED.
const unsupportedFeatureCode = uint32(0x8010001f)

var platformErrorsByCode = map[uint32]errorInfo{}
