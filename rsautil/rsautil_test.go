package rsautil

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
)

// 全局测试密钥（每个测试用例前重新生成，避免复用导致问题）
var (
	testPrivateKey []byte
	testPublicKey  []byte
	testData       = []byte("test rsa util: 你好，RSA加密测试 123456")
)

// 测试前置：生成新的测试密钥
func setupTest(t *testing.T) {
	var err error
	testPrivateKey, testPublicKey, err = GenerateRsaKeys()
	if err != nil {
		t.Fatalf("GenerateRsaKeys failed: %v", err)
	}
	if len(testPrivateKey) == 0 || len(testPublicKey) == 0 {
		t.Fatal("GenerateRsaKeys return empty key")
	}
}

// ========== 基础功能测试 ==========

// TestGenerateRsaKeys 测试密钥生成功能
func TestGenerateRsaKeys(t *testing.T) {
	// 正常场景
	private, public, err := GenerateRsaKeys()
	if err != nil {
		t.Errorf("GenerateRsaKeys should not return error, got: %v", err)
	}
	if len(private) == 0 || len(public) == 0 {
		t.Error("GenerateRsaKeys return empty private/public key")
	}
}

// TestRsaEncryptDecrypt 测试加密/解密核心功能（含正常、异常、临界值）
func TestRsaEncryptDecrypt(t *testing.T) {
	setupTest(t)

	// 测试场景1：正常加密解密
	encrypted, err := RsaEncrypt(testData, testPublicKey)
	if err != nil {
		t.Errorf("RsaEncrypt normal case failed: %v", err)
	}
	decrypted, err := RsaDecrypt(encrypted, testPrivateKey)
	if err != nil {
		t.Errorf("RsaDecrypt normal case failed: %v", err)
	}
	if !bytes.Equal(decrypted, testData) {
		t.Errorf("Decrypted data not match, want: %s, got: %s", testData, decrypted)
	}

	// 测试场景2：空参数异常
	// 空明文
	_, err = RsaEncrypt([]byte{}, testPublicKey)
	if err == nil {
		t.Error("RsaEncrypt should return error with empty plainData")
	}
	// 空公钥
	_, err = RsaEncrypt(testData, []byte{})
	if err == nil {
		t.Error("RsaEncrypt should return error with empty publicKey")
	}
	// 空密文
	_, err = RsaDecrypt([]byte{}, testPrivateKey)
	if err == nil {
		t.Error("RsaDecrypt should return error with empty encryptedData")
	}
	// 空私钥
	_, err = RsaDecrypt(encrypted, []byte{})
	if err == nil {
		t.Error("RsaDecrypt should return error with empty privateKey")
	}

	// 测试场景3：临界值（2048位RSA最大明文长度245字节）
	// 245字节正常
	maxNormalData := bytes.Repeat([]byte("a"), 245)
	encryptedMax, err := RsaEncrypt(maxNormalData, testPublicKey)
	if err != nil {
		t.Errorf("RsaEncrypt 245 bytes failed: %v", err)
	}
	decryptedMax, err := RsaDecrypt(encryptedMax, testPrivateKey)
	if err != nil {
		t.Errorf("RsaDecrypt 245 bytes failed: %v", err)
	}
	if !bytes.Equal(decryptedMax, maxNormalData) {
		t.Error("245 bytes decrypted data not match")
	}

	// 246字节（超过临界值）应该报错
	overMaxData := bytes.Repeat([]byte("a"), 246)
	_, err = RsaEncrypt(overMaxData, testPublicKey)
	if err == nil {
		t.Error("RsaEncrypt 246 bytes should return error")
	}
}

// TestRsaEncryptByPublicKeyDecryptByPrivateKey 测试基于字符串密钥的加密解密
func TestRsaEncryptByPublicKeyDecryptByPrivateKey(t *testing.T) {
	setupTest(t)

	// 转换为字符串密钥
	pubKeyStr := string(testPublicKey)
	privKeyStr := string(testPrivateKey)

	// 测试场景1：正常加密解密
	encrypted, err := RsaEncryptByPublicKey(testData, pubKeyStr)
	if err != nil {
		t.Errorf("RsaEncryptByPublicKey normal case failed: %v", err)
	}
	decrypted, err := RsaDecryptByPrivateKey(encrypted, privKeyStr)
	if err != nil {
		t.Errorf("RsaDecryptByPrivateKey normal case failed: %v", err)
	}
	if !bytes.Equal(decrypted, testData) {
		t.Error("RsaEncryptByPublicKey/RsaDecryptByPrivateKey data not match")
	}

	// 测试场景2：空参数异常
	// 空明文
	_, err = RsaEncryptByPublicKey([]byte{}, pubKeyStr)
	if err == nil {
		t.Error("RsaEncryptByPublicKey should return error with empty plainData")
	}
	// 空公钥字符串
	_, err = RsaEncryptByPublicKey(testData, "")
	if err == nil {
		t.Error("RsaEncryptByPublicKey should return error with empty publicKey string")
	}
	// 空密文
	_, err = RsaDecryptByPrivateKey([]byte{}, privKeyStr)
	if err == nil {
		t.Error("RsaDecryptByPrivateKey should return error with empty encryptedData")
	}
	// 空私钥字符串
	_, err = RsaDecryptByPrivateKey(encrypted, "")
	if err == nil {
		t.Error("RsaDecryptByPrivateKey should return error with empty privateKey string")
	}
}

// TestRsaEncryptByPrivateKeyDecryptByPublicKey 测试私钥加密/公钥解密（签名类场景）
func TestRsaEncryptByPrivateKeyDecryptByPublicKey(t *testing.T) {
	setupTest(t)

	// 转换为字符串密钥
	pubKeyStr := string(testPublicKey)
	privKeyStr := string(testPrivateKey)

	// 测试场景1：正常加密解密
	encrypted, err := RsaEncryptByPrivateKey(testData, privKeyStr)
	if err != nil {
		t.Errorf("RsaEncryptByPrivateKey normal case failed: %v", err)
	}
	decrypted, err := RsaDecryptByPublicKey(encrypted, pubKeyStr)
	if err != nil {
		t.Errorf("RsaDecryptByPublicKey normal case failed: %v", err)
	}
	if !bytes.Equal(decrypted, testData) {
		t.Error("RsaEncryptByPrivateKey/RsaDecryptByPublicKey data not match")
	}

	// 测试场景2：空参数异常
	// 空明文
	_, err = RsaEncryptByPrivateKey([]byte{}, privKeyStr)
	if err == nil {
		t.Error("RsaEncryptByPrivateKey should return error with empty plainData")
	}
	// 空私钥字符串
	_, err = RsaEncryptByPrivateKey(testData, "")
	if err == nil {
		t.Error("RsaEncryptByPrivateKey should return error with empty privateKey string")
	}
	// 空密文
	_, err = RsaDecryptByPublicKey([]byte{}, pubKeyStr)
	if err == nil {
		t.Error("RsaDecryptByPublicKey should return error with empty encryptedData")
	}
	// 空公钥字符串
	_, err = RsaDecryptByPublicKey(encrypted, "")
	if err == nil {
		t.Error("RsaDecryptByPublicKey should return error with empty publicKey string")
	}
}

// TestRsaSignVerifyWithSha256 测试签名/验签功能
func TestRsaSignVerifyWithSha256(t *testing.T) {
	setupTest(t)

	// 测试场景1：正常签名+验签
	sign, err := RsaSignWithSha256(testData, testPrivateKey)
	if err != nil {
		t.Errorf("RsaSignWithSha256 normal case failed: %v", err)
	}
	ok, err := RsaVerifyWithSha256(testData, sign, testPublicKey)
	if err != nil {
		t.Errorf("RsaVerifyWithSha256 normal case failed: %v", err)
	}
	if !ok {
		t.Error("RsaVerifyWithSha256 should return true for valid sign")
	}

	// 测试场景2：验签失败（篡改数据）
	modifiedData := []byte("modified test data")
	ok, err = RsaVerifyWithSha256(modifiedData, sign, testPublicKey)
	if err == nil && ok {
		t.Error("RsaVerifyWithSha256 should fail for modified data")
	}

	// 测试场景3：空参数异常
	// 空明文
	_, err = RsaSignWithSha256([]byte{}, testPrivateKey)
	if err == nil {
		t.Error("RsaSignWithSha256 should return error with empty plainData")
	}
	// 空私钥
	_, err = RsaSignWithSha256(testData, []byte{})
	if err == nil {
		t.Error("RsaSignWithSha256 should return error with empty privateKey")
	}
	// 空明文验签
	_, err = RsaVerifyWithSha256([]byte{}, sign, testPublicKey)
	if err == nil {
		t.Error("RsaVerifyWithSha256 should return error with empty plainData")
	}
	// 空签名
	_, err = RsaVerifyWithSha256(testData, []byte{}, testPublicKey)
	if err == nil {
		t.Error("RsaVerifyWithSha256 should return error with empty signData")
	}
	// 空公钥
	_, err = RsaVerifyWithSha256(testData, sign, []byte{})
	if err == nil {
		t.Error("RsaVerifyWithSha256 should return error with empty publicKey")
	}
}

// ========== 性能测试（基准测试） ==========

// BenchmarkRsaEncrypt 加密性能测试
func BenchmarkRsaEncrypt(b *testing.B) {
	setupTest(&testing.T{}) // 复用前置逻辑
	b.ResetTimer()          // 重置计时器，排除setup耗时

	for i := 0; i < b.N; i++ {
		_, err := RsaEncrypt(testData, testPublicKey)
		if err != nil {
			b.Fatalf("BenchmarkRsaEncrypt failed: %v", err)
		}
	}
}

// BenchmarkRsaDecrypt 解密性能测试
func BenchmarkRsaDecrypt(b *testing.B) {
	setupTest(&testing.T{})
	// 提前加密数据，避免循环内重复加密
	encrypted, err := RsaEncrypt(testData, testPublicKey)
	if err != nil {
		b.Fatalf("Prepare encrypted data failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := RsaDecrypt(encrypted, testPrivateKey)
		if err != nil {
			b.Fatalf("BenchmarkRsaDecrypt failed: %v", err)
		}
	}
}

// BenchmarkRsaSignWithSha256 签名性能测试
func BenchmarkRsaSignWithSha256(b *testing.B) {
	setupTest(&testing.T{})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := RsaSignWithSha256(testData, testPrivateKey)
		if err != nil {
			b.Fatalf("BenchmarkRsaSignWithSha256 failed: %v", err)
		}
	}
}

// BenchmarkRsaVerifyWithSha256 验签性能测试
func BenchmarkRsaVerifyWithSha256(b *testing.B) {
	setupTest(&testing.T{})
	// 提前生成签名
	sign, err := RsaSignWithSha256(testData, testPrivateKey)
	if err != nil {
		b.Fatalf("Prepare sign data failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := RsaVerifyWithSha256(testData, sign, testPublicKey)
		if err != nil {
			b.Fatalf("BenchmarkRsaVerifyWithSha256 failed: %v", err)
		}
	}
}

// ////////////////////////////////////////////////////
// //////////////////////////////////////////////////////////
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

func TestRsa1(t *testing.T) {
	PrivateKey := "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAwh2VooRzDT5pT58O5BHbn1PxPnXzGRlgft5CsiC9vTMAmaaMiBXg8di7eUa07dEY9MWqlhzejKWDTRY+Acda1i50wzBKh/eTUzCM/bosdePs2x6sUkNoOpRsQqD+4DknFrDwt2VlMjeYyEAKxZirVB/y3kMX8g+Amj0veQ5Pm5rJWppHjtQJxfox+7FIlTJ8BMyFi6XJgXxXU5oc+zWZuZEl1zldfi/PBU7otnMMw+GoaKTPLTeQwRtnWxFCSRdWNUUnf5wXlAcN5Qx8MVklQI3DJrVjQk79kWhg3U7uxAX85273VlQcHCkRxp5ZMDkpFcfb7XqryspV7eMQfCt2vwIDAQABAoIBADQfuZSAOFywC5tDvL3lRbIM2lTJW1O8CrtGd2ZZgFmTnm+j10ybg2Gtrvmr0N2jLi5b/ah9bA0cTJugg1n67BtjMhtPllPYWQkXmmRvX4zwfSOBowgb7Zr9S+zASnBvKF3heWHlqjzHzRDIkZvpmOfoBFFGduGa5A+Gqn65JwtDFyD+Wy6uixqUAjzazCL3YoMqoJB4Grf8edZH0CFirPZy1YdTFU0+WS76ZU8I8VL/DykwN0uBt6RK7ijowsLgEWqsOBeiiae1QuRzeh9Rh5n4pEs15sf7KDsIhzArhZrCHE1WB3K33q9PwwBAL7KQ547dz0RF7IOKWuBB0V4aRlECgYEA0NU6/F7hpFdwE5znbQzOLTYxthFuZwft2GgnQK0WaHl77DP3WSCm70kqZUuzFXt3iFTAzRY/+umgbFRHN2E/f93+gMNbuTFjufo+nvmuAAggD5J8HFGYpgsbpnzXyG/zHTDzI1ODHhM+T5bIjjXirbUxwV/4UCKUsGoPDNQC26sCgYEA7fVkeEwdMmnPYU2x9O96xrJAGT6QQA4YJKEeODJEVzOxFUv3YZuKHmnmP1+4MUww8/AsCdUzEPfiDgcCMwC05BAHkF0xBGwjczKackFrj6eBtuXb+Npl3T0lVMx7XyU1A9sbxtl7wUTYOJIvdkYPsaQucGiZtZDkd7Y1kpkQXT0CgYEAqOS69t6ZqPsaZpJQTM69dK1O2QwR+PvdrVbW4CTcaZRO2AJTOl8BA6dtxUzKwkX/r1+0KmmjRv9pwhMLIcvhuj1FFshox0cde4za1mHiCp6Hp3B2NcT8KtXy/9wZ/D4mJeavzVM+SLWRgHbXLsR/1rMjUVyXi9/b1y1/jIVu5k8CgYBUTHewSj6ZqnRmIzEk9WXIWENu7gQKPTP+XfmnrN8bCVv1kHjt4j028ws3bkCBbl11PaNMRHQX0ckKcR8tVFXA6ZDUar8/stOILugaC+T/+jZwxdN8wFoP22aLOPmHxlWWrRuVAVzfJfV3bQpVWxKaOvCcr+GsOc1GP42RBpqOfQKBgEg5aAms+wawrwWFHivLjvAmGwC6sw5LDzj6UBsJTxWrDdCWc/o0vLyZL6YlFGpDmZPPU6h+gUwTlRwEHX6ItkO1zjY1Zcd1rNCSoPPrMFEdDNRh6ZN1wk3lsNG2tmRO4aF9zQkiXgJgjGbi3Jkd+fEWHEA6HQzWXdXECF/Hetk6\n-----END RSA PRIVATE KEY-----"

	PublicKey := "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwh2VooRzDT5pT58O5BHbn1PxPnXzGRlgft5CsiC9vTMAmaaMiBXg8di7eUa07dEY9MWqlhzejKWDTRY+Acda1i50wzBKh/eTUzCM/bosdePs2x6sUkNoOpRsQqD+4DknFrDwt2VlMjeYyEAKxZirVB/y3kMX8g+Amj0veQ5Pm5rJWppHjtQJxfox+7FIlTJ8BMyFi6XJgXxXU5oc+zWZuZEl1zldfi/PBU7otnMMw+GoaKTPLTeQwRtnWxFCSRdWNUUnf5wXlAcN5Qx8MVklQI3DJrVjQk79kWhg3U7uxAX85273VlQcHCkRxp5ZMDkpFcfb7XqryspV7eMQfCt2vwIDAQAB\n-----END PUBLIC KEY-----"
	var data = "aaabbbcccdd测试"

	encryptedData, err := RsaEncryptByPublicKey([]byte(data), PublicKey)
	fmt.Println("encryptedData:\n" + hex.EncodeToString(encryptedData))
	fmt.Println(err)

	sourceData, err := RsaDecryptByPrivateKey(encryptedData, PrivateKey)
	fmt.Println("sourceData:\n" + string(sourceData))
	fmt.Println(err)
	fmt.Println("-----------------------------------")

	encryptedData, err = RsaEncryptByPrivateKey([]byte(data), PrivateKey)
	fmt.Println("encryptedData:\n" + hex.EncodeToString(encryptedData))
	fmt.Println(err)

	sourceData, err = RsaDecryptByPublicKey(encryptedData, PublicKey)
	fmt.Println("sourceData:\n" + string(sourceData))
	fmt.Println(err)

}
