//go:build windows

package pcsc

const attributeClassPerformance = 0x7ffe

const (
	AttributeProtocolTypes Attribute = attributeClassProtocol<<16 | 0x0120

	AttributePerformanceNumberOfTransmissions Attribute = attributeClassPerformance<<16 | 0x0001
	AttributePerformanceBytesTransmitted      Attribute = attributeClassPerformance<<16 | 0x0002
	AttributePerformanceTransmissionTime      Attribute = attributeClassPerformance<<16 | 0x0003

	AttributeDeviceFriendlyName = AttributeDeviceFriendlyNameW
	AttributeDeviceSystemName   = AttributeDeviceSystemNameW
)
