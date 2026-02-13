package base64util

import (
	"encoding/base64"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/stringutil"
)

func EncodeBase64(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func DecodeBase64(src string) ([]byte, error) {
	b, err := base64.StdEncoding.DecodeString(src)
	if err == nil {
		return b, nil
	}
	b, err = base64.RawStdEncoding.DecodeString(src)
	if err == nil {
		return b, nil
	}
	b, err = base64.URLEncoding.DecodeString(src)
	if err == nil {
		return b, nil
	}
	b, err = base64.RawURLEncoding.DecodeString(src)
	if err == nil {
		return b, nil
	}
	return nil, err
}

func TrimBase64(str string) string {
	str = stringutil.TrimNewLine(str)
	str = strings.Replace(str, " ", "", -1)
	str = strings.Replace(str, "\t", "", -1)
	str = strings.TrimSpace(str)
	return str
}

// rm 'BEGIN CERTIFICATE' and 'END CERTIFICATE' and newline
func DecodeCertBase64(oldBytes []byte) ([]byte, error) {

	if len(oldBytes) == 0 {
		return []byte{}, nil
	}
	isBinary := false
	for _, b := range oldBytes {
		//	if t < 32 && t != 9 && t != 10 && t != 13 {
		//		isBinary = true
		//		break
		//	}
		if !((b >= 32 && b <= 126) || b == 9 || b == 10 || b == 13) {
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
	txt = stringutil.TrimNewLine(txt)
	txt = strings.Replace(txt, "\t", "", -1)
	txt = strings.Replace(txt, " ", "", -1)
	belogs.Debug("DecodeCertBase64(): txt after Replace: %s", txt)
	newBytes, err := DecodeBase64(txt)
	return newBytes, err
}
