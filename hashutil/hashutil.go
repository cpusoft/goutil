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
	// 修复：变量名从md5改为h，避免与包名冲突
	h := md5.New()
	h.Write([]byte(data))
	md5Data := h.Sum([]byte(""))
	return hex.EncodeToString(md5Data)
}

// Calculate Hmac (基于MD5的HMAC)
func Hmac(key, data string) string {
	// 修复：变量名从hmac改为h，避免与包名冲突
	h := hmac.New(md5.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum([]byte("")))
}

// Calculate Sha1
func Sha1(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum([]byte("")))
}

func Sha256(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum([]byte("")))
}

func Sha256File(fileName string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	// 修复1：错误变量名从erro改为errCopy，规范命名
	// 修复2：拷贝出错时返回正确的错误（errCopy）而非文件打开的err
	_, errCopy := io.Copy(h, f)
	if errCopy != nil {
		return "", errCopy
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func HashFileByte(filePathName string) ([32]byte, error) {
	file, err := os.Open(filePathName)
	if err != nil {
		return [32]byte{}, err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return [32]byte{}, err
	}

	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}
