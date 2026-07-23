//go:build darwin || linux

package pcsc

const (
	AttributeAsyncProtocolTypes Attribute = attributeClassProtocol<<16 | 0x0120
	AttributeSyncProtocolTypes  Attribute = attributeClassProtocol<<16 | 0x0126

	AttributeDeviceFriendlyName = AttributeDeviceFriendlyNameA
	AttributeDeviceSystemName   = AttributeDeviceSystemNameA
)
