package hashutil

import (
	"fmt"
	"strings"
	"testing"

	base64util "github.com/cpusoft/goutil/base64util"
	fileutil "github.com/cpusoft/goutil/fileutil"
)

func TestSha256(t *testing.T) {
	s := []byte{0x01, 0x02, 0x02}
	sh := Sha256(s)
	fmt.Println(sh)
}

func TestSha256File(t *testing.T) {
	s := `G:\Download\hwuc0uQdqwW9VSfJFQ6UiVlxZpY.crl`
	sh, err := Sha256File(s)
	fmt.Println(sh, err)

	b, err := fileutil.ReadFileToBytes(s)
	fmt.Println(len(b), err)
	sh = Sha256(b)
	fmt.Println(sh)

	base := base64util.EncodeBase64(b)
	fmt.Println(base)
	sh = Sha256([]byte(base))
	fmt.Println(sh)
}

func TestSha256Password(t *testing.T) {
	p := Sha256([]byte("2e869b49-50c8-487b-ab1a-67c87c77ccc0" + "abc123!@#"))
	fmt.Println(p)
}

func TestSha256String(t *testing.T) {
	s := `
        MIIDFjCCAf4CAQEwDQYJKoZIhvcNAQELBQAwRjERMA8GA1UEAxMIQTkxOUI2M0MxMTAvBgNV
BAUTKDI1ODVEQTBCOTgwQTQ3RkVCQTBFMjM1MjA1REVFRTQwMkYyMEIzQ0IXDTE5MTEyNTAzMDAxN1oX
DTE5MTEyNzAzMDAxN1owggFQMBMCAihzFw0xOTExMjEwOTAyMTVaMBMCAih0Fw0xOTExMjExNTAwMTBa
MBMCAih1Fw0xOTExMjEyMTAwMjFaMBMCAih2Fw0xOTExMjIwMzAwMDRaMBMCAih3Fw0xOTExMjIwOTAx
MjJaMBMCAih4Fw0xOTExMjIxNTAwMzRaMBMCAih5Fw0xOTExMjIyMDU5MjlaMBMCAih6Fw0xOTExMjMw
MzAwMzdaMBMCAih7Fw0xOTExMjMwOTAxMzhaMBMCAih8Fw0xOTExMjMxNTAxNTFaMBMCAih9Fw0xOTEx
MjMyMDU5NTJaMBMCAih+Fw0xOTExMjQwMzAxMDZaMBMCAih/Fw0xOTExMjQwOTAwNDRaMBMCAiiAFw0x
OTExMjQxNTAwMzhaMBMCAiiBFw0xOTExMjQyMDU5NDVaMBMCAiiCFw0xOTExMjUwMzAwMTZaoDAwLjAf
BgNVHSMEGDAWgBQlhdoLmApH/roOI1IF3u5ALyCzyzALBgNVHRQEBAICUP0wDQYJKoZIhvcNAQELBQAD
ggEBAFmUueWNFT9n56ZJlDGwbwDmgUMBuS87ypRi+xRk1+cuM4+nYE/pWpLejkB8+AObUhrAiiun1VQa
06oXIu14X+/YREkaquSxPh4K1oHJY/bBQRaUxOj6elvhSXiCaplc4TLV2voTBCYBW3SZR06U5exq9KIh
LiocMrCTZOWRvcKs0DfbZUoCx8fm0XGDTSwiLhEkcJmyT5BxkFbrZXcCwvminNgk/iPqNVDm/MOtISX6
KCOuSHDD6gScUamzCy2jCT5truL2iKrb8xk+Yp5SUAA2TnGV/c6ToLuGU4DRZ/vsTDY4eomfxH+yfRqI
MhT+jBXdOpyAl2OE6yWR15SqrRE=
      `
	sh := Sha256([]byte(s))
	fmt.Println(sh)
	s1 := strings.Replace(s, "\r", "", -1)
	s2 := strings.Replace(s1, "\n", "", -1)
	s3 := strings.TrimSpace(s2)
	sh = Sha256([]byte(s3))
	fmt.Println(sh)

}
