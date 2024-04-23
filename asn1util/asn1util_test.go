package asn1util

/*
func TestDecodeBitStringToAddress(t *testing.T) {
	//03 04 04 0a0020            addressPrefix 10.0.32/20
	data := []byte{0x03, 0x04, 0x04, 0x0a, 0x00, 0x20}
	address, err := DecodeBitStringToAddress(data, 1, false, false)
	fmt.Println("\n---------should 10.0.32/20:", address, err)

	//03 04 00 0a0040            addressPrefix 10.0.64/24
	data = []byte{0x03, 0x04, 0x00, 0x0a, 0x00, 0x40}
	address, err = DecodeBitStringToAddress(data, 1, false, false)
	fmt.Println("\n---------should 10.0.64/24:", address, err)

	//03 03 00 0a01              addressPrefix    10.1/16
	data = []byte{0x03, 0x03, 0x00, 0x0a, 0x01}
	address, err = DecodeBitStringToAddress(data, 1, false, false)
	fmt.Println("\n---------should 10.1/16:", address, err)

	//30 0c                      addressRange {
	//03 04 04 0a0230           min        10.2.48.0
	//03 04 00 0a0240           max        10.2.64.255
	data = []byte{0x03, 0x04, 0x04, 0x0a, 0x02, 0x30}
	address, err = DecodeBitStringToAddress(data, 1, true, true)
	fmt.Println("\n---------should 10.2.48.0:", address, err)

	data = []byte{0x03, 0x04, 0x00, 0x0a, 0x02, 0x40}
	address, err = DecodeBitStringToAddress(data, 1, true, false)
	fmt.Println("\n---------should 10.2.64.255:", address, err)
}
*/
