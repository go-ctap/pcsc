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
)

const (
	AttributeVendorName      Attribute = attributeClassVendorInfo<<16 | 0x0100
	AttributeVendorIFDType   Attribute = attributeClassVendorInfo<<16 | 0x0101
	AttributeVendorVersion   Attribute = attributeClassVendorInfo<<16 | 0x0102
	AttributeVendorSerial    Attribute = attributeClassVendorInfo<<16 | 0x0103
	AttributeChannelID       Attribute = attributeClassCommunication<<16 | 0x0110
	AttributeAsyncProtocols  Attribute = attributeClassProtocol<<16 | 0x0120
	AttributeDefaultClock    Attribute = attributeClassProtocol<<16 | 0x0121
	AttributeMaximumClock    Attribute = attributeClassProtocol<<16 | 0x0122
	AttributeDefaultDataRate Attribute = attributeClassProtocol<<16 | 0x0123
	AttributeMaximumDataRate Attribute = attributeClassProtocol<<16 | 0x0124
	AttributeMaximumIFSD     Attribute = attributeClassProtocol<<16 | 0x0125
	AttributeSyncProtocols   Attribute = attributeClassProtocol<<16 | 0x0126
)

const (
	AttributePowerManagement      Attribute = attributeClassPower<<16 | 0x0131
	AttributeUserToCardAuthDevice Attribute = attributeClassSecurity<<16 | 0x0140
	AttributeUserAuthInputDevice  Attribute = attributeClassSecurity<<16 | 0x0142
	AttributeCharacteristics      Attribute = attributeClassMechanical<<16 | 0x0150
)

const (
	AttributeCurrentProtocol Attribute = attributeClassIFDProtocol<<16 | 0x0201
	AttributeCurrentClock    Attribute = attributeClassIFDProtocol<<16 | 0x0202
	AttributeCurrentF        Attribute = attributeClassIFDProtocol<<16 | 0x0203
	AttributeCurrentD        Attribute = attributeClassIFDProtocol<<16 | 0x0204
	AttributeCurrentN        Attribute = attributeClassIFDProtocol<<16 | 0x0205
	AttributeCurrentW        Attribute = attributeClassIFDProtocol<<16 | 0x0206
	AttributeCurrentIFSC     Attribute = attributeClassIFDProtocol<<16 | 0x0207
	AttributeCurrentIFSD     Attribute = attributeClassIFDProtocol<<16 | 0x0208
	AttributeCurrentBWT      Attribute = attributeClassIFDProtocol<<16 | 0x0209
	AttributeCurrentCWT      Attribute = attributeClassIFDProtocol<<16 | 0x020a
	AttributeCurrentEBC      Attribute = attributeClassIFDProtocol<<16 | 0x020b
	AttributeExtendedBWT     Attribute = attributeClassIFDProtocol<<16 | 0x020c
)

const (
	AttributeCardPresence  Attribute = attributeClassICCState<<16 | 0x0300
	AttributeCardInterface Attribute = attributeClassICCState<<16 | 0x0301
	AttributeCurrentIO     Attribute = attributeClassICCState<<16 | 0x0302
	AttributeATR           Attribute = attributeClassICCState<<16 | 0x0303
	AttributeCardType      Attribute = attributeClassICCState<<16 | 0x0304
)

const (
	AttributeVendorReset       Attribute = attributeClassVendorDefined<<16 | 0xa000
	AttributeVendorCancel      Attribute = attributeClassVendorDefined<<16 | 0xa003
	AttributeVendorAuthRequest Attribute = attributeClassVendorDefined<<16 | 0xa005
	AttributeMaximumInput      Attribute = attributeClassVendorDefined<<16 | 0xa007
)
