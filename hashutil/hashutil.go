package hashutil

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// Calculate Md5
func Md5(data string) string {
	md5 := md5.New()
	md5.Write([]byte(data))
	md5Data := md5.Sum([]byte(""))
	return hex.EncodeToString(md5Data)
}

// Calculate Hmac
func Hmac(key, data string) string {
	hmac := hmac.New(md5.New, []byte(key))
	hmac.Write([]byte(data))
	return hex.EncodeToString(hmac.Sum([]byte("")))
}

// Calculate Sha1
func Sha1(data []byte) string {
	sha1 := sha1.New()
	sha1.Write(data)
	return hex.EncodeToString(sha1.Sum([]byte("")))
}

func Sha256(data []byte) string {
	sha256 := sha256.New()
	sha256.Write(data)
	return hex.EncodeToString(sha256.Sum([]byte("")))
}

func Sha256File(fileName string) (string, error) {

	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()
	sha256 := sha256.New()
	_, erro := io.Copy(sha256, f)
	if erro != nil {
		return "", err
	}
	return hex.EncodeToString(sha256.Sum(nil)), nil
}
