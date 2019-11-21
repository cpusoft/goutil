package hashutil

import (
	"fmt"
	"testing"

	base64util "github.com/cpusoft/goutil/base64util"
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
}

func TestSha256String(t *testing.T) {
	s := `MIIDATCCAekCAQEwDQYJKoZIhvcNAQELBQAwRjERMA8GA1UEAxMIQTkxQUVBOEMxMTAvBgNVBAUTKDg3MEI5Q0QyRTQxREFCMDVCRDU1MjdDOTE1MEU5NDg5NTk3MTY2OTYXDTE5MTEyMTEwMzIyNFoXDTE5MTEyMzEwMzIyNFowggE7MBMCAhDbFw0xOTExMTcyMjI1MDdaMBMCAhDcFw0xOTExMTgwNDI3MTlaMBMCAhDdFw0xOTExMTgxMDMxMDZaMBMCAhDeFw0xOTExMTgxNjMxMTZaMBMCAhDfFw0xOTExMTgyMjI1MTZaMBMCAhDgFw0xOTExMTkwNDI4MThaMBMCAhDhFw0xOTExMTkxMDI5MDRaMBMCAhDiFw0xOTExMTkxNjI4MjBaMBMCAhDjFw0xOTExMTkyMjI1NDlaMBMCAhDkFw0xOTExMjAwNDE1MzNaMBMCAhDlFw0xOTExMjAxMDI4MjNaMBMCAhDmFw0xOTExMjAxNjMxMDFaMBMCAhDnFw0xOTExMjAyMjI4MDRaMBMCAhDoFw0xOTExMjEwNDMwNDFaMBMCAhDpFw0xOTExMjExMDMyMjRaoDAwLjAfBgNVHSMEGDAWgBSHC5zS5B2rBb1VJ8kVDpSJWXFmljALBgNVHRQEBAICIcswDQYJKoZIhvcNAQELBQADggEBAH9Qn6OL3mtoAWbZBRCk4Uc01DtX63pxeBMwYdxxnFom+/bp1iQ+b0L6O0xaxjctthQBS/uDHtYLFHTUPeOEVA/RXhZu45yltqd7JHjG1OSmyHUszzM7OLDFRO3ERHc+ebDufUx2nGa4K0+IXrTLKO0+K8SSOB+xpvtRxPHw/Bx7oLo66KDLOFPXnZ1V9WiWI2dJc8rVCE/w2bJ+oOuwBKi6gLq0ljEblulGKbP35acCIFTY1rkro5vDK0dXhxCup0PQX3gebGXcb8bVzFmFd4OC//IKTRJ19aI3niIO1dAa3kekRzbOA/t3nH8WiBwlXJ4Now82I7sqyEk0PBL+5cY=`
	sh := Sha256([]byte(s))
	fmt.Println(sh)

	ss, err := base64util.DecodeBase64(s)
	sh = Sha256(ss)
	fmt.Println(sh, err)

	sss := base64util.EncodeBase64(ss)
	sh = Sha256([]byte(sss))
	fmt.Println(sh)
}
