package rsautil

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/wenzhenxi/gorsa"
)

// get rsa public/private key
func GenerateRsaKeys() (privateKey, publicKey []byte, err error) {
	// get priviate
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

	block, _ := pem.Decode(publicKey)
	if block == nil {
		belogs.Error("RsaEncrypt(): Decode publicKey fail:", convert.PrintBytes(plainData, 8))
		return nil, errors.New("RsaEncrypt:Fail to convert publickey to pem")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		belogs.Error("RsaEncrypt(): ParsePKIXPublicKey fail:", convert.PrintBytes(plainData, 8))
		return nil, err
	}

	rsaPublicKey := pubInterface.(*rsa.PublicKey)
	encryptedData, err = rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, plainData)
	if err != nil {
		belogs.Error("RsaEncrypt(): EncryptPKCS1v15 fail:", convert.PrintBytes(plainData, 8))
		return nil, err
	}
	return encryptedData, nil
}

// using privateKey to decrypt
// not log encryptedData and privateKey
func RsaDecrypt(encryptedData, privateKey []byte) (plainData []byte, err error) {

	block, _ := pem.Decode(privateKey)
	if block == nil {
		belogs.Error("RsaDecrypt(): Decode fail:")
		return nil, errors.New("RsaDecrypt:Fail to convert privatekey to pem")
	}

	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		belogs.Error("RsaDecrypt(): ParsePKCS1PrivateKey fail:")
		return nil, err
	}

	plainData, err = rsa.DecryptPKCS1v15(rand.Reader, rsaPrivateKey, encryptedData)
	if err != nil {
		belogs.Error("RsaDecrypt(): DecryptPKCS1v15 fail:")
		return nil, err
	}
	return plainData, nil
}

func RsaEncryptByPublicKey(plainData []byte, publicKey string) (encryptedData []byte, err error) {

	grsa := gorsa.RSASecurity{}
	grsa.SetPublicKey(publicKey)
	belogs.Debug(publicKey)

	encryptedData, err = grsa.PubKeyENCTYPT(plainData)
	if err != nil {
		belogs.Error("RsaEncryptByPublicKey(): PubKeyENCTYPT fail:", err)
		return nil, err
	}
	return encryptedData, nil
}

func RsaDecryptByPrivateKey(encryptedData []byte, privateKey string) (plainData []byte, err error) {
	grsa := gorsa.RSASecurity{}
	err = grsa.SetPrivateKey(privateKey)
	if err != nil {
		belogs.Error("RsaDecryptByPrivateKey(): SetPrivateKey:", err)
		return nil, err
	}

	plainData, err = grsa.PriKeyDECRYPT(encryptedData)
	if err != nil {
		belogs.Error("RsaDecryptByPrivateKey(): PriKeyDECRYPT:", err)
		return nil, err
	}
	return plainData, nil
}

func RsaEncryptByPrivateKey(plainData []byte, privateKey string) (encryptedData []byte, err error) {
	grsa := gorsa.RSASecurity{}
	grsa.SetPrivateKey(privateKey)

	encryptedData, err = grsa.PriKeyENCTYPT(plainData)
	if err != nil {
		belogs.Error("RsaEncryptByPrivateKey(): PriKeyENCTYPT:", err)
		return nil, err
	}
	return encryptedData, nil
}

func RsaDecryptByPublicKey(encryptedData []byte, publicKey string) (plainData []byte, err error) {
	grsa := gorsa.RSASecurity{}
	err = grsa.SetPublicKey(publicKey)
	if err != nil {
		belogs.Error("RsaDecryptByPublicKey(): SetPublicKey:", err)
		return nil, err
	}

	plainData, err = grsa.PubKeyDECRYPT(encryptedData)
	if err != nil {
		belogs.Error("RsaDecryptByPublicKey(): PubKeyDECRYPT:", err)
		return nil, err
	}
	return plainData, nil
}

// rsa sign with sha256
// not log privateKey
func RsaSignWithSha256(plainData []byte, privateKey []byte) (signData []byte, err error) {

	sha256 := sha256.Sum256(plainData)
	block, _ := pem.Decode(privateKey)
	if block == nil {
		belogs.Error("RsaSignWithSha256(): Decode publicKey fail:", convert.PrintBytes(plainData, 8))
		return nil, errors.New("RsaSignWithSha256:Fail to convert privateKey to pem")
	}
	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		belogs.Error("RsaSignWithSha256(): ParsePKCS1PrivateKey fail, err:", err)
		return nil, err
	}

	signData, err = rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, sha256[:])
	if err != nil {
		belogs.Error("RsaSignWithSha256(): SignPKCS1v15 fail, err:", err)
		return nil, err
	}
	return signData, nil
}

// rsa ver with sha256
func RsaVerifyWithSha256(plainData, signData, publicKey []byte) (bool, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		belogs.Error("RsaVerifyWithSha256(): Decode publicKey fail:", convert.PrintBytes(plainData, 8))
		return false, errors.New("RsaVerifyWithSha256:Fail to convert publicKey to pem")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		belogs.Error("RsaVerifyWithSha256():ParsePKIXPublicKey fail:", convert.PrintBytes(plainData, 8))
		return false, err
	}

	sha256 := sha256.Sum256(plainData)
	err = rsa.VerifyPKCS1v15(pubKey.(*rsa.PublicKey), crypto.SHA256, sha256[:], signData)
	if err != nil {
		belogs.Error("RsaVerifyWithSha256():VerifyPKCS1v15 fail:", convert.PrintBytes(plainData, 8))
		return false, err
	}
	return true, nil
}
