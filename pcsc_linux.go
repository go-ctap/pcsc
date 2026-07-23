//go:build linux

package pcsc

const (
	pcscLibrary             = "libpcsclite.so.1"
	scardControlSymbol      = "SCardControl"
	reconnectSupportsDirect = true
	reconnectSupportsEject  = true
)

// ProtocolT15 is the pcsc-lite T=15 protocol identifier.
const ProtocolT15 Protocol = 0x0008

const (
	scardReaderStateATRSize = 33
	scardReaderStatePacked  = false
)

// pcsc-lite handles are pointer-sized integer tokens on Unix platforms other than macOS.
type scardContext uintptr
type scardHandle uintptr

// pcsc-lite defines LONG and DWORD as long and unsigned long respectively.
// Go's int and uint follow the same ILP32/LP64 widths on supported Linux targets.
type scardResult = int
type scardDWORD = uint
