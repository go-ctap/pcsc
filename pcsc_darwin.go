//go:build darwin

package pcsc

const pcscLibrary = "/System/Library/Frameworks/PCSC.framework/PCSC"

const (
	scardReaderStateATRSize = 33
	scardReaderStatePacked  = true
)

// Apple's PCSC framework deliberately uses fixed-width 32-bit handles.
type scardContext int32
type scardHandle int32
type scardResult = int32
type scardDWORD = uint32
