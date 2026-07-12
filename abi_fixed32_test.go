//go:build darwin || windows

package pcsc

import "unsafe"

const (
	nativeDWORDSize       = uintptr(4)
	dwordSize             = unsafe.Sizeof(scardDWORD(0))
	protocolFieldSize     = unsafe.Sizeof(scardIORequest{}.Protocol)
	lengthFieldSize       = unsafe.Sizeof(scardIORequest{}.Length)
	ioRequestSize         = unsafe.Sizeof(scardIORequest{})
	ioRequestLengthOffset = unsafe.Offsetof(scardIORequest{}.Length)
)

var (
	_ [nativeDWORDSize - dwordSize]byte
	_ [dwordSize - nativeDWORDSize]byte
	_ [nativeDWORDSize - protocolFieldSize]byte
	_ [protocolFieldSize - nativeDWORDSize]byte
	_ [nativeDWORDSize - lengthFieldSize]byte
	_ [lengthFieldSize - nativeDWORDSize]byte
	_ [2*nativeDWORDSize - ioRequestSize]byte
	_ [ioRequestSize - 2*nativeDWORDSize]byte
	_ [nativeDWORDSize - ioRequestLengthOffset]byte
	_ [ioRequestLengthOffset - nativeDWORDSize]byte
)
