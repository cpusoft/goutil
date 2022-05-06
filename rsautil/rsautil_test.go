package rsautil

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestRsa(t *testing.T) {
	privateKey, publicKey, err := GenerateRsaKeys()
	fmt.Println("privateKey:\n" + string(privateKey))
	fmt.Println("publicKey:\n" + string(publicKey))
	fmt.Println(err)

	var data = "aaabbbcccdd测试"

	signData, err := RsaSignWithSha256([]byte(data), privateKey)
	fmt.Println("signData:\n" + hex.EncodeToString(signData))
	fmt.Println(err)
	r, err := RsaVerifyWithSha256([]byte(data), signData, publicKey)
	fmt.Println("verify:", r, err)

	encryptedData, err := RsaEncrypt([]byte(data), publicKey)
	fmt.Println("encryptedData:\n" + hex.EncodeToString(encryptedData))
	fmt.Println(err)
	sourceData, err := RsaDecrypt(encryptedData, privateKey)
	fmt.Println("sourceData:\n" + string(sourceData))
	fmt.Println(err)

	encryptedData, err = RsaEncrypt([]byte(data), privateKey)
	fmt.Println("encryptedData2:\n" + hex.EncodeToString(encryptedData))
	fmt.Println(err)
	sourceData, err = RsaDecrypt(encryptedData, publicKey)
	fmt.Println("sourceData2:\n" + string(sourceData))
	fmt.Println(err)
}
