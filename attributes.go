package pcsc

const (
	attributeClassVendorInfo    = 1
	attributeClassCommunication = 2
	attributeClassProtocol      = 3
	attributeClassPower         = 4
	attributeClassSecurity      = 5
	attributeClassMechanical    = 6
	attributeClassVendorDefined = 7
	attributeClassIFDProtocol   = 8
	attributeClassICCState      = 9
	attributeClassSystem        = 0x7fff
)

const (
	AttributeVendorName        Attribute = attributeClassVendorInfo<<16 | 0x0100
	AttributeVendorIFDType     Attribute = attributeClassVendorInfo<<16 | 0x0101
	AttributeVendorIFDVersion  Attribute = attributeClassVendorInfo<<16 | 0x0102
	AttributeVendorIFDSerialNo Attribute = attributeClassVendorInfo<<16 | 0x0103
	AttributeChannelID         Attribute = attributeClassCommunication<<16 | 0x0110
	AttributeDefaultCLK        Attribute = attributeClassProtocol<<16 | 0x0121
	AttributeMaxCLK            Attribute = attributeClassProtocol<<16 | 0x0122
	AttributeDefaultDataRate   Attribute = attributeClassProtocol<<16 | 0x0123
	AttributeMaxDataRate       Attribute = attributeClassProtocol<<16 | 0x0124
	AttributeMaxIFSD           Attribute = attributeClassProtocol<<16 | 0x0125
)

const (
	AttributePowerManagementSupport Attribute = attributeClassPower<<16 | 0x0131
	AttributeUserToCardAuthDevice   Attribute = attributeClassSecurity<<16 | 0x0140
	AttributeUserAuthInputDevice    Attribute = attributeClassSecurity<<16 | 0x0142
	AttributeCharacteristics        Attribute = attributeClassMechanical<<16 | 0x0150
)

const (
	AttributeCurrentProtocolType Attribute = attributeClassIFDProtocol<<16 | 0x0201
	AttributeCurrentCLK          Attribute = attributeClassIFDProtocol<<16 | 0x0202
	AttributeCurrentF            Attribute = attributeClassIFDProtocol<<16 | 0x0203
	AttributeCurrentD            Attribute = attributeClassIFDProtocol<<16 | 0x0204
	AttributeCurrentN            Attribute = attributeClassIFDProtocol<<16 | 0x0205
	AttributeCurrentW            Attribute = attributeClassIFDProtocol<<16 | 0x0206
	AttributeCurrentIFSC         Attribute = attributeClassIFDProtocol<<16 | 0x0207
	AttributeCurrentIFSD         Attribute = attributeClassIFDProtocol<<16 | 0x0208
	AttributeCurrentBWT          Attribute = attributeClassIFDProtocol<<16 | 0x0209
	AttributeCurrentCWT          Attribute = attributeClassIFDProtocol<<16 | 0x020a
	AttributeCurrentEBCEncoding  Attribute = attributeClassIFDProtocol<<16 | 0x020b
	AttributeExtendedBWT         Attribute = attributeClassIFDProtocol<<16 | 0x020c
)

const (
	AttributeICCPresence        Attribute = attributeClassICCState<<16 | 0x0300
	AttributeICCInterfaceStatus Attribute = attributeClassICCState<<16 | 0x0301
	AttributeCurrentIOState     Attribute = attributeClassICCState<<16 | 0x0302
	AttributeATRString          Attribute = attributeClassICCState<<16 | 0x0303
	AttributeICCTypePerATR      Attribute = attributeClassICCState<<16 | 0x0304
)

const (
	AttributeESCReset       Attribute = attributeClassVendorDefined<<16 | 0xa000
	AttributeESCCancel      Attribute = attributeClassVendorDefined<<16 | 0xa003
	AttributeESCAuthRequest Attribute = attributeClassVendorDefined<<16 | 0xa005
	AttributeMaxInput       Attribute = attributeClassVendorDefined<<16 | 0xa007
)

const (
	AttributeDeviceUnit          Attribute = attributeClassSystem<<16 | 0x0001
	AttributeDeviceInUse         Attribute = attributeClassSystem<<16 | 0x0002
	AttributeDeviceFriendlyNameA Attribute = attributeClassSystem<<16 | 0x0003
	AttributeDeviceSystemNameA   Attribute = attributeClassSystem<<16 | 0x0004
	AttributeDeviceFriendlyNameW Attribute = attributeClassSystem<<16 | 0x0005
	AttributeDeviceSystemNameW   Attribute = attributeClassSystem<<16 | 0x0006

	// AttributeSupressT1IFSRequest retains the spelling used by the native
	// SCARD_ATTR_SUPRESS_T1_IFS_REQUEST constant.
	AttributeSupressT1IFSRequest Attribute = attributeClassSystem<<16 | 0x0007
)
