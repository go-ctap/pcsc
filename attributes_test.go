package pcsc

import "testing"

func TestCanonicalAttributeValues(t *testing.T) {
	tests := []struct {
		name string
		got  Attribute
		want Attribute
	}{
		{name: "vendor IFD version", got: AttributeVendorIFDVersion, want: 0x00010102},
		{name: "vendor IFD serial", got: AttributeVendorIFDSerialNo, want: 0x00010103},
		{name: "default CLK", got: AttributeDefaultCLK, want: 0x00030121},
		{name: "current protocol", got: AttributeCurrentProtocolType, want: 0x00080201},
		{name: "ATR string", got: AttributeATRString, want: 0x00090303},
		{name: "ESC reset", got: AttributeESCReset, want: 0x0007a000},
		{name: "max input", got: AttributeMaxInput, want: 0x0007a007},
	}

	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("%s = 0x%x, want 0x%x", test.name, test.got, test.want)
		}
	}
}

func TestSystemAttributeValues(t *testing.T) {
	tests := []struct {
		name string
		got  Attribute
		want Attribute
	}{
		{name: "device unit", got: AttributeDeviceUnit, want: 0x7fff0001},
		{name: "friendly name A", got: AttributeDeviceFriendlyNameA, want: 0x7fff0003},
		{name: "system name W", got: AttributeDeviceSystemNameW, want: 0x7fff0006},
		{name: "supress T1 IFS", got: AttributeSupressT1IFSRequest, want: 0x7fff0007},
	}

	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("%s = 0x%x, want 0x%x", test.name, test.got, test.want)
		}
	}
}
