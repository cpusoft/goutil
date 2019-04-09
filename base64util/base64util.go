package base64util

import (
	"encoding/base64"
)

func EncodeBase64(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func DecodeBase64(src string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(src)
}
