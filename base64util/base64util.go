package base64util

import (
	"encoding/base64"
	"strings"

	"github.com/cpusoft/goutil/belogs"
)

func EncodeBase64(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func DecodeBase64(src string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(src)
}

func TrimBase64(str string) string {
	str = strings.Replace(str, "\r", "", -1)
	str = strings.Replace(str, "\n", "", -1)
	str = strings.Replace(str, " ", "", -1)
	str = strings.Replace(str, "\t", "", -1)
	str = strings.TrimSpace(str)
	return str
}

// rm 'BEGIN CERTIFICATE' and 'END CERTIFICATE' and newline
func DecodeCertBase64(oldBytes []byte) ([]byte, error) {
	isBinary := false

	for _, b := range oldBytes {
		t := int(b)

		if t < 32 && t != 9 && t != 10 && t != 13 {
			isBinary = true
			break
		}
	}

	belogs.Debug("DecodeCertBase64(): isBinary:", isBinary)
	if isBinary {
		return oldBytes, nil
	}
	txt := string(oldBytes)
	txt = strings.Replace(txt, "-----BEGIN CERTIFICATE-----", "", -1)
	txt = strings.Replace(txt, "-----END CERTIFICATE-----", "", -1)
	txt = strings.Replace(txt, "-", "", -1)
	txt = strings.Replace(txt, " ", "", -1)
	txt = strings.Replace(txt, "\r", "", -1)
	txt = strings.Replace(txt, "\n", "", -1)
	belogs.Debug("DecodeCertBase64(): txt after Replace: %s", txt)
	newBytes, err := base64.StdEncoding.DecodeString(txt)
	return newBytes, err
}
