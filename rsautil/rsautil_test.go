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
