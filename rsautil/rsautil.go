package rsautil

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/wenzhenxi/gorsa"
)

// get rsa public/private key
func GenerateRsaKeys() (privateKey, publicKey []byte, err error) {
	// get private
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		belogs.Error("GenerateRsaKeys(): GenerateKey fail, err:", err)
		return nil, nil, err
	}
	derStream := x509.MarshalPKCS1PrivateKey(rsaPrivateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	privateKey = pem.EncodeToMemory(block)
	rsaPublicKey := &rsaPrivateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(rsaPublicKey)
	if err != nil {
		belogs.Error("GenerateRsaKeys(): MarshalPKIXPublicKey fail, err:", err)
		return nil, nil, err
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	publicKey = pem.EncodeToMemory(block)
	return
}

// using publickKey encrypt
func RsaEncrypt(plainData, publicKey []byte) (encryptedData []byte, err error) {
	// 1. 提前校验空参数
	if len(plainData) == 0 {
		belogs.Error("RsaEncrypt(): plainData is empty")
		return nil, errors.New("RsaEncrypt: plainData cannot be empty")
	}
	if len(publicKey) == 0 {
		belogs.Error("RsaEncrypt(): publicKey is empty")
		return nil, errors.New("RsaEncrypt: publicKey cannot be empty")
	}

	block, _ := pem.Decode(publicKey)
	if block == nil {
		belogs.Error("RsaEncrypt(): Decode publicKey fail:", convert.PrintBytes(plainData, 8))
		return nil, fmt.Errorf("RsaEncrypt: Fail to convert publickey to pem, publicKey length: %d", len(publicKey))
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		belogs.Error("RsaEncrypt(): ParsePKIXPublicKey fail:", convert.PrintBytes(plainData, 8), "err:", err)
		return nil, fmt.Errorf("RsaEncrypt: ParsePKIXPublicKey fail: %w", err)
	}

	// 2. 安全的类型断言（避免panic）
	rsaPublicKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		belogs.Error("RsaEncrypt(): pubInterface is not *rsa.PublicKey")
		return nil, errors.New("RsaEncrypt: public key is not RSA public key")
	}

	// 3. 检查RSA PKCS1v15加密长度限制（密钥长度-11）
	maxPlainLen := rsaPublicKey.Size() - 11
	if len(plainData) > maxPlainLen {
		belogs.Error("RsaEncrypt(): plainData too long, max:", maxPlainLen, "actual:", len(plainData))
		return nil, fmt.Errorf("RsaEncrypt: plainData too long (max %d bytes, actual %d bytes)", maxPlainLen, len(plainData))
	}

	encryptedData, err = rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, plainData)
	if err != nil {
		belogs.Error("RsaEncrypt(): EncryptPKCS1v15 fail:", convert.PrintBytes(plainData, 8), "err:", err)
		return nil, fmt.Errorf("RsaEncrypt: EncryptPKCS1v15 fail: %w", err)
	}
	return encryptedData, nil
}

// using privateKey to decrypt
// not log encryptedData and privateKey
func RsaDecrypt(encryptedData, privateKey []byte) (plainData []byte, err error) {
	// 提前校验空参数
	if len(encryptedData) == 0 {
		belogs.Error("RsaDecrypt(): encryptedData is empty")
		return nil, errors.New("RsaDecrypt: encryptedData cannot be empty")
	}
	if len(privateKey) == 0 {
		belogs.Error("RsaDecrypt(): privateKey is empty")
		return nil, errors.New("RsaDecrypt: privateKey cannot be empty")
	}

	block, _ := pem.Decode(privateKey)
	if block == nil {
		belogs.Error("RsaDecrypt(): Decode fail:")
		return nil, fmt.Errorf("RsaDecrypt: Fail to convert privatekey to pem, privateKey length: %d", len(privateKey))
	}

	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		belogs.Error("RsaDecrypt(): ParsePKCS1PrivateKey fail:", "err:", err)
		return nil, fmt.Errorf("RsaDecrypt: ParsePKCS1PrivateKey fail: %w", err)
	}

	plainData, err = rsa.DecryptPKCS1v15(rand.Reader, rsaPrivateKey, encryptedData)
	if err != nil {
		belogs.Error("RsaDecrypt(): DecryptPKCS1v15 fail:", "err:", err)
		return nil, fmt.Errorf("RsaDecrypt: DecryptPKCS1v15 fail: %w", err)
	}
	return plainData, nil
}

func RsaEncryptByPublicKey(plainData []byte, publicKey string) (encryptedData []byte, err error) {
	// 提前校验空参数
	if len(plainData) == 0 {
		belogs.Error("RsaEncryptByPublicKey(): plainData is empty")
		return nil, errors.New("RsaEncryptByPublicKey: plainData cannot be empty")
	}
	if publicKey == "" {
		belogs.Error("RsaEncryptByPublicKey(): publicKey is empty string")
		return nil, errors.New("RsaEncryptByPublicKey: publicKey cannot be empty")
	}

	grsa := gorsa.RSASecurity{}
	grsa.SetPublicKey(publicKey)
	belogs.Debug("RsaEncryptByPublicKey(): publicKey:", publicKey)

	encryptedData, err = grsa.PubKeyENCTYPT(plainData)
	if err != nil {
		belogs.Error("RsaEncryptByPublicKey(): PubKeyENCTYPT fail:", err)
		return nil, fmt.Errorf("RsaEncryptByPublicKey: PubKeyENCTYPT fail: %w", err)
	}
	return encryptedData, nil
}

func RsaDecryptByPrivateKey(encryptedData []byte, privateKey string) (plainData []byte, err error) {
	// 提前校验空参数
	if len(encryptedData) == 0 {
		belogs.Error("RsaDecryptByPrivateKey(): encryptedData is empty")
		return nil, errors.New("RsaDecryptByPrivateKey: encryptedData cannot be empty")
	}
	if privateKey == "" {
		belogs.Error("RsaDecryptByPrivateKey(): privateKey is empty string")
		return nil, errors.New("RsaDecryptByPrivateKey: privateKey cannot be empty")
	}

	grsa := gorsa.RSASecurity{}
	err = grsa.SetPrivateKey(privateKey)
	if err != nil {
		belogs.Error("RsaDecryptByPrivateKey(): SetPrivateKey:", err)
		return nil, fmt.Errorf("RsaDecryptByPrivateKey: SetPrivateKey fail: %w", err)
	}

	plainData, err = grsa.PriKeyDECRYPT(encryptedData)
	if err != nil {
		belogs.Error("RsaDecryptByPrivateKey(): PriKeyDECRYPT:", err)
		return nil, fmt.Errorf("RsaDecryptByPrivateKey: PriKeyDECRYPT fail: %w", err)
	}
	return plainData, nil
}

func RsaEncryptByPrivateKey(plainData []byte, privateKey string) (encryptedData []byte, err error) {
	// 提前校验空参数
	if len(plainData) == 0 {
		belogs.Error("RsaEncryptByPrivateKey(): plainData is empty")
		return nil, errors.New("RsaEncryptByPrivateKey: plainData cannot be empty")
	}
	if privateKey == "" {
		belogs.Error("RsaEncryptByPrivateKey(): privateKey is empty string")
		return nil, errors.New("RsaEncryptByPrivateKey: privateKey cannot be empty")
	}

	grsa := gorsa.RSASecurity{}
	grsa.SetPrivateKey(privateKey)

	encryptedData, err = grsa.PriKeyENCTYPT(plainData)
	if err != nil {
		belogs.Error("RsaEncryptByPrivateKey(): PriKeyENCTYPT fail:", err)
		return nil, fmt.Errorf("RsaEncryptByPrivateKey: PriKeyENCTYPT fail: %w", err)
	}
	return encryptedData, nil
}

func RsaDecryptByPublicKey(encryptedData []byte, publicKey string) (plainData []byte, err error) {
	// 提前校验空参数
	if len(encryptedData) == 0 {
		belogs.Error("RsaDecryptByPublicKey(): encryptedData is empty")
		return nil, errors.New("RsaDecryptByPublicKey: encryptedData cannot be empty")
	}
	if publicKey == "" {
		belogs.Error("RsaDecryptByPublicKey(): publicKey is empty string")
		return nil, errors.New("RsaDecryptByPublicKey: publicKey cannot be empty")
	}

	grsa := gorsa.RSASecurity{}
	err = grsa.SetPublicKey(publicKey)
	if err != nil {
		belogs.Error("RsaDecryptByPublicKey(): SetPublicKey:", err)
		return nil, fmt.Errorf("RsaDecryptByPublicKey: SetPublicKey fail: %w", err)
	}

	plainData, err = grsa.PubKeyDECRYPT(encryptedData)
	if err != nil {
		belogs.Error("RsaDecryptByPublicKey(): PubKeyDECRYPT:", err)
		return nil, fmt.Errorf("RsaDecryptByPublicKey: PubKeyDECRYPT fail: %w", err)
	}
	return plainData, nil
}

// rsa sign with sha256
// not log privateKey
func RsaSignWithSha256(plainData []byte, privateKey []byte) (signData []byte, err error) {
	// 提前校验空参数
	if len(plainData) == 0 {
		belogs.Error("RsaSignWithSha256(): plainData is empty")
		return nil, errors.New("RsaSignWithSha256: plainData cannot be empty")
	}
	if len(privateKey) == 0 {
		belogs.Error("RsaSignWithSha256(): privateKey is empty")
		return nil, errors.New("RsaSignWithSha256: privateKey cannot be empty")
	}

	// 4. 修复变量名遮蔽问题（sha256 -> hashSum）
	hashSum := sha256.Sum256(plainData)
	block, _ := pem.Decode(privateKey)
	if block == nil {
		// 5. 修复业务逻辑笔误：publicKey -> privateKey
		belogs.Error("RsaSignWithSha256(): Decode privateKey fail:", convert.PrintBytes(plainData, 8))
		return nil, fmt.Errorf("RsaSignWithSha256: Fail to convert privateKey to pem, privateKey length: %d", len(privateKey))
	}
	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		belogs.Error("RsaSignWithSha256(): ParsePKCS1PrivateKey fail, err:", err)
		return nil, fmt.Errorf("RsaSignWithSha256: ParsePKCS1PrivateKey fail: %w", err)
	}

	signData, err = rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, hashSum[:])
	if err != nil {
		belogs.Error("RsaSignWithSha256(): SignPKCS1v15 fail, err:", err)
		return nil, fmt.Errorf("RsaSignWithSha256: SignPKCS1v15 fail: %w", err)
	}
	return signData, nil
}

// rsa ver with sha256
func RsaVerifyWithSha256(plainData, signData, publicKey []byte) (bool, error) {
	// 提前校验空参数
	if len(plainData) == 0 {
		belogs.Error("RsaVerifyWithSha256(): plainData is empty")
		return false, errors.New("RsaVerifyWithSha256: plainData cannot be empty")
	}
	if len(signData) == 0 {
		belogs.Error("RsaVerifyWithSha256(): signData is empty")
		return false, errors.New("RsaVerifyWithSha256: signData cannot be empty")
	}
	if len(publicKey) == 0 {
		belogs.Error("RsaVerifyWithSha256(): publicKey is empty")
		return false, errors.New("RsaVerifyWithSha256: publicKey cannot be empty")
	}

	block, _ := pem.Decode(publicKey)
	if block == nil {
		belogs.Error("RsaVerifyWithSha256(): Decode publicKey fail:", convert.PrintBytes(plainData, 8))
		return false, fmt.Errorf("RsaVerifyWithSha256: Fail to convert publicKey to pem, publicKey length: %d", len(publicKey))
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		belogs.Error("RsaVerifyWithSha256():ParsePKIXPublicKey fail:", convert.PrintBytes(plainData, 8), "err:", err)
		return false, fmt.Errorf("RsaVerifyWithSha256: ParsePKIXPublicKey fail: %w", err)
	}

	// 安全的类型断言（避免panic）
	rsaPublicKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		belogs.Error("RsaVerifyWithSha256(): pubKey is not *rsa.PublicKey")
		return false, errors.New("RsaVerifyWithSha256: public key is not RSA public key")
	}

	// 修复变量名遮蔽问题
	hashSum := sha256.Sum256(plainData)
	err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, hashSum[:], signData)
	if err != nil {
		belogs.Error("RsaVerifyWithSha256():VerifyPKCS1v15 fail:", convert.PrintBytes(plainData, 8), "err:", err)
		return false, fmt.Errorf("RsaVerifyWithSha256: VerifyPKCS1v15 fail: %w", err)
	}
	return true, nil
}
