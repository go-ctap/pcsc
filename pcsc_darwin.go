//go:build darwin

package pcsc

const (
	pcscLibrary             = "/System/Library/Frameworks/PCSC.framework/PCSC"
	scardControlSymbol      = "SCardControl132"
	reconnectSupportsDirect = true
	reconnectSupportsEject  = true
)

const (
	scardReaderStateATRSize = 33
	scardReaderStatePacked  = true
)

// Apple's PCSC framework deliberately uses fixed-width 32-bit handles.
type scardContext int32
type scardHandle int32
type scardResult = int32
type scardDWORD = uint32
