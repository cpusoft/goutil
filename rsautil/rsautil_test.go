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
	PrivateKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1oUFD7aYYWh7p9HdNVYUM+R7zdkzsUMCM5EXkwxIV1FDR/3B
GQ5T1uAzSphdfqjgBxbfv9x5gmieTvKXW84f126VZBNXtIGXjfUDakOuGiqDdKl2
He9KBf9lfitwkLih9zYsWTuiBh8o7V3FFtM1i3DuzP4HCo57aUxtUy5b7ij8LQxZ
U0XwOXZHaszVl+nAOCL+0/+VmLJkyio3kiU9C0hJDd+j4tvJC3jyLipOYX50ID4G
PfZxIi7FCvCHKHBadfehlNcW9EtDpxPhGcyVpP7AsgvStQvyU0TZz7vhLHDfUtY/
nl7dlZ+5PWUmvuw4myHGzDmvQ6vcUVW7DRMGAQIDAQABAoIBAQCXklFrMtckLFEC
2LP2FaYcrFoVrlxp6TDLAr+ndMxAdfiWC2O+snLmpm9XS6Tz85qnJ7BcvglU7Vq9
6YaspU22SDpiBZC4x8Av22jYUo3XiyZq7bm5mPOynSw3I7Zbazl1lN9tBUeMD8Q5
Q0IYyI9SwS7ZxLtw6A+m7Qtp9J2b/ipURjVLyyW7UvIyq2qHfTA0YdN7FU8R3Y3X
MLHB0QGLUUTmnABmlYSbYh1hFvmXS6vDrkVUz1zgbbmOXF8bMI1fCyD1t45F15yD
uQUQ5dx9m1Z6rLw1Gjr/RgpngAX65oOlGtkYp+EYk1hGHAKmpGEr0vsskJc4zYI4
+/KAcW+BAoGBAOidUrKr/v/mOU2HINzHgVI+255nUgaGRezh6RMQnzwW6nz2KdZT
VkE/YTWCKh3zo5+i/uHddT/miT8jlPhBl2Te4fYPhBBZPXsOYcX3p2rLbsx96AD+
YxWmuQaIaRzk9zLgfBjOXTwL7lHpy8scTmW6gzskqRrELqZ1eMcUm2wpAoGBAOwV
/zv82AehZUJNkM9B4A5BtlgjTNUZinejRIJ8k6maeC2BIe8Gzb9Xj2xs6IBm7Ry9
DkPcalE+aySO+CYS6TtiKeSdJigzJyg7Kwq0C/M0GNZwP92ZvlKnV/4Patcl7+E2
VSnxwfSu3/We32vUd15yg5gpkqjyDnV/4USeCIYZAoGBALcoYRxkh5XRHl+oPbz5
rh8ndWAFtLWEdnyt6Qrk9Kyo0pvwbELhPbKEiDNMuYL5+2VQP2dzK8ZT7M91YfAU
HXQEd2F7GB6TVfCWA3CQrxdM9YI4xTw7EaPTsi6trC5fLzG1RqF1pD4Kmu2OrLPS
Jvy83mXsWObFgIH7T01aMYL5AoGAfnEZjetRWGTccrJQSHCjq38ORg5B7DANtR3A
Z5KJE2Ej1FtA7V/begtPSWba70ow3B91MGswleq0P5RC20FtoNxmS4bPFOCwrB9k
YgskC1FvrAnaarkY8fOmcO+Y7TnoS9ppqllM49t1H3vDdWEJvY/fYvOBFPLvQ4cG
A1YQgqECgYAFFan2RbUn7pUXyrRchPG9dknayEpaqK5a9CxmsAohJc1Gn5IwtWhQ
IV425nQHa4A2OlOo4D9mNVSVpufKqLfw/ld1kSd9fhD+wBgz+qvDy0JHXPU2W8Sy
fMaZ6pemtyzA9kmJ+2E1mBnjdY9rTTf4A3pEfNdrNziiI78fop6gAw==
-----END RSA PRIVATE KEY-----`

	PublicKey := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1oUFD7aYYWh7p9HdNVYU
M+R7zdkzsUMCM5EXkwxIV1FDR/3BGQ5T1uAzSphdfqjgBxbfv9x5gmieTvKXW84f
126VZBNXtIGXjfUDakOuGiqDdKl2He9KBf9lfitwkLih9zYsWTuiBh8o7V3FFtM1
i3DuzP4HCo57aUxtUy5b7ij8LQxZU0XwOXZHaszVl+nAOCL+0/+VmLJkyio3kiU9
C0hJDd+j4tvJC3jyLipOYX50ID4GPfZxIi7FCvCHKHBadfehlNcW9EtDpxPhGcyV
pP7AsgvStQvyU0TZz7vhLHDfUtY/nl7dlZ+5PWUmvuw4myHGzDmvQ6vcUVW7DRMG
AQIDAQAB
-----END PUBLIC KEY-----`
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
