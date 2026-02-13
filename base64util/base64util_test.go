package base64util

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// -------------------------- 核心测试用例 --------------------------

// TestEncodeBase64 测试标准Base64编码
func TestEncodeBase64(t *testing.T) {
	// 场景1：正常字符串编码
	src := []byte("hello world")
	expected := "aGVsbG8gd29ybGQ="
	result := EncodeBase64(src)
	assert.Equal(t, expected, result)

	// 场景2：空输入编码
	srcEmpty := []byte("")
	expectedEmpty := ""
	resultEmpty := EncodeBase64(srcEmpty)
	assert.Equal(t, expectedEmpty, resultEmpty)

	// 场景3：含特殊字符的输入编码
	srcSpecial := []byte("test!@#$%^&*()")
	resultSpecial := EncodeBase64(srcSpecial)
	decoded, err := base64.StdEncoding.DecodeString(resultSpecial)
	assert.NoError(t, err)
	assert.Equal(t, srcSpecial, decoded)
}

// TestDecodeBase64 测试多格式Base64解码
func TestDecodeBase64(t *testing.T) {
	rawData := []byte("test base64 decode")

	// 场景1：标准Base64解码
	stdEncoded := base64.StdEncoding.EncodeToString(rawData)
	decodedStd, err := DecodeBase64(stdEncoded)
	assert.NoError(t, err)
	assert.Equal(t, rawData, decodedStd)

	// 场景2：RawStd编码（无填充）解码
	rawStdEncoded := base64.RawStdEncoding.EncodeToString(rawData)
	decodedRawStd, err := DecodeBase64(rawStdEncoded)
	assert.NoError(t, err)
	assert.Equal(t, rawData, decodedRawStd)

	// 场景3：URL安全编码解码
	urlEncoded := base64.URLEncoding.EncodeToString(rawData)
	decodedURL, err := DecodeBase64(urlEncoded)
	assert.NoError(t, err)
	assert.Equal(t, rawData, decodedURL)

	// 场景4：RawURL编码解码
	rawURLEncoded := base64.RawURLEncoding.EncodeToString(rawData)
	decodedRawURL, err := DecodeBase64(rawURLEncoded)
	assert.NoError(t, err)
	assert.Equal(t, rawData, decodedRawURL)

	// 场景5：无效Base64解码（返回错误）
	invalidBase64 := "invalid_123_###"
	decodedInvalid, err := DecodeBase64(invalidBase64)
	assert.Error(t, err)
	assert.Nil(t, decodedInvalid)

	// 场景6：空字符串解码
	decodedEmpty, err := DecodeBase64("")
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), decodedEmpty)
}

// TestTrimBase64 测试Base64文本清理
func TestTrimBase641(t *testing.T) {
	// 场景1：含换行、空格、制表符的文本
	dirtyStr := "  aGVsbG8\n\r\td29ybGQ=  "
	cleanStr := TrimBase64(dirtyStr)
	assert.Equal(t, "aGVsbG8d29ybGQ=", cleanStr)

	// 场景2：纯空白文本
	blankStr := " \n\r\t "
	trimmedBlank := TrimBase64(blankStr)
	assert.Equal(t, "", trimmedBlank)

	// 场景3：无需要清理的文本
	normalStr := "test123+/=-" // 含合法的-
	trimmedNormal := TrimBase64(normalStr)
	assert.Equal(t, normalStr, trimmedNormal)
}

// TestDecodeCertBase64 测试修复后的证书Base64解码（核心场景）
func TestDecodeCertBase64(t *testing.T) {
	// 原始证书内容对应的Base64
	validCertRaw := []byte("test certificate base64 decode")
	validCertStdBase64 := base64.StdEncoding.EncodeToString(validCertRaw)
	validCertURLBase64 := base64.URLEncoding.EncodeToString(validCertRaw) // 含-

	// 场景1：合法PEM证书（含BEGIN/END）解码
	validCertPEM := `-----BEGIN CERTIFICATE-----
` + validCertStdBase64 + `
-----END CERTIFICATE-----`
	decoded1, err := DecodeCertBase64([]byte(validCertPEM))
	assert.NoError(t, err)
	assert.Equal(t, validCertRaw, decoded1)

	// 场景2：URL安全的证书Base64（含-）解码（修复过度替换-的问题）
	validURLCertPEM := `-----BEGIN CERTIFICATE-----
` + validCertURLBase64 + `
-----END CERTIFICATE-----`
	decoded2, err := DecodeCertBase64([]byte(validURLCertPEM))
	assert.NoError(t, err)
	assert.Equal(t, validCertRaw, decoded2)

	// 场景3：二进制输入（直接返回原字节）
	binaryData := []byte{0x00, 0x01, 0x02, 0x7F} // 含<32的二进制字符
	decoded3, err := DecodeCertBase64(binaryData)
	assert.NoError(t, err)
	assert.Equal(t, binaryData, decoded3)

	// 场景4：修复后的二进制判断（全空格不再误判）
	fakeBinary := []byte("   ") // 全是空格（打印字符），判定为文本
	decoded4, err := DecodeCertBase64(fakeBinary)
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), decoded4) // 空文本解码返回空

	// 场景5：空输入解码（修复后返回空字节）
	decoded5, err := DecodeCertBase64([]byte(""))
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), decoded5)

	// 场景6：无BEGIN/END的纯证书Base64解码
	plainCertBase64 := []byte(validCertStdBase64)
	decoded6, err := DecodeCertBase64(plainCertBase64)
	assert.NoError(t, err)
	assert.Equal(t, validCertRaw, decoded6)

	// 场景7：二进制+文本混合（含<32字符）
	mixedData := []byte{0x00, 'a', 0x01, 'b'} // 含二进制字符，判定为二进制
	decoded7, err := DecodeCertBase64(mixedData)
	assert.NoError(t, err)
	assert.Equal(t, mixedData, decoded7)
}

// TestDecodeCertBase64_EdgeCases 测试边缘场景
func TestDecodeCertBase64_EdgeCases(t *testing.T) {
	// 场景1：含制表符/换行的PEM证书
	tabCertPEM := `-----BEGIN CERTIFICATE-----
	` + base64.StdEncoding.EncodeToString([]byte("test edge case")) + `
-----END CERTIFICATE-----`
	decoded1, err := DecodeCertBase64([]byte(tabCertPEM))
	assert.NoError(t, err)
	assert.Equal(t, []byte("test edge case"), decoded1)

	// 场景2：含多余空格和-的证书（保留合法-）
	dirtyCertPEM := `-----BEGIN CERTIFICATE-----
  dGVzdC1lZGdlLWNhc2UtLS0=  -
-----END CERTIFICATE-----`
	decoded2, err := DecodeCertBase64([]byte(dirtyCertPEM))
	assert.NoError(t, err)
	// 解码验证：dGVzdC1lZGdlLWNhc2UtLS0= → test-edge-case--
	assert.Equal(t, []byte("test-edge-case--"), decoded2)
}

/*
	func TestASA(t *testing.T) {
		base := `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`
		b, e := DecodeBase64(base)
		fmt.Println(len(b), e)

		file := "1.asa"
		e = fileutil.WriteBytesToFile(file, b)
		fmt.Println(e)
	}

	func TestROA(t *testing.T) {
		base := `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`
		b, e := DecodeBase64(base)
		fmt.Println(len(b), e)

		file := "1.roa"
		e = fileutil.WriteBytesToFile(file, b)
		fmt.Println(e)
	}

	func TestCer(t *testing.T) {
		base := `MIIDxTCCAq2gAwIBAgIUEtGCqPuwhmJXm32ULONuYtSmqT0wDQYJKoZIhvcNAQELBQAwMzExMC8GA1UEAxMoNEYwN0UwQ0NCRkMwMjM0MTk0NjdCM0I0QTVEODI4MjI1QkVDQkQ4MzAeFw0yMjA3MjIxMTAwMzZaFw0yMzA3MjExMTA1MzZaMDMxMTAvBgNVBAMTKDE3MzE2OTAzRjA2NzEyMjlFODgwOEJBOEU4QUIwMTA1RkE5MTVBMDcwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARPXQUwHE/o/evpob8gKl+yS2ntfrKUkhjwrJ7rORCE063TL5xHNieAfwwbNjdxc29nrjPgKlnbaB/i17Jsaryjo4IBmjCCAZYwHQYDVR0OBBYEFBcxaQPwZxIp6ICLqOirAQX6kVoHMB8GA1UdIwQYMBaAFE8H4My/wCNBlGeztKXYKCJb7L2DMA4GA1UdDwEB/wQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDHjB0BgNVHR8EbTBrMGmgZ6BlhmNyc3luYzovL3Rlc3RiZWQua3JpbGwuY2xvdWQvcmVwby9sb2NhbC10ZXN0YmVkLWNoaWxkLzAvNEYwN0UwQ0NCRkMwMjM0MTk0NjdCM0I0QTVEODI4MjI1QkVDQkQ4My5jcmwwcwYIKwYBBQUHAQEEZzBlMGMGCCsGAQUFBzAChldyc3luYzovL3Rlc3RiZWQua3JpbGwuY2xvdWQvcmVwby90ZXN0YmVkLzAvNEYwN0UwQ0NCRkMwMjM0MTk0NjdCM0I0QTVEODI4MjI1QkVDQkQ4My5jZXIwDgYIKwYBBQUHAQsEAjAAMBgGA1UdIAEB/wQOMAwwCgYIKwYBBQUHDgIwGgYIKwYBBQUHAQgBAf8ECzAJoAcwBQIDAzl5MA0GCSqGSIb3DQEBCwUAA4IBAQBf0zStQZCA9xpQHVbuDruuXz/eYDferFkU3jfZov+W+7MjWOvMSWHv/FSA0mDrLrTd8F8HUufvV3w6Pj0KEwKdCs1r2qYocUicupu0tA+eg8KkwmDiCNhlySUM0WB3KEHXHzVApEk0SIGFPIkUGIyiIRiZcn3L7qOJNoYjC89Vg9frm7z1YIiwmqiRyyXU706nHBhe193VXa9xSDNYcyQi4FY6JUnNR/WIy/Z5VnQOrvKMYkozbF+m2CihqbWe0N+XZxqdySwjT/7fHbAncUNDstNkYivw3yWdw9mbtN2jGFUygpyOwTBzim9Mb6oSRffCTPzgnzNwQjtfMj4luSeJ`
		b, e := DecodeBase64(base)
		fmt.Println(len(b), e)

		file := "ROUTER-00033979-17316903F0671229E8808BA8E8AB0105FA915A07.cer"
		e = fileutil.WriteBytesToFile(file, b)
		fmt.Println(e)
	}
*/
func TestBase64(t *testing.T) {
	base := `MIIHLgYJKoZIhvcNAQcCoIIHHzCCBxsCAQMxDTALBglghkgBZQMEAgEwSAYLKoZIhvcNAQkQARigOQQ3MDUCAwDMxzAuMCwEAgABMCYwCQMEALm4jQIBIDAGAwQAubn4MAYDBAC5ufkwCQMEALm+UAIBIKCCBQ0wggUJMIID8aADAgECAhIBjrKq/UK5U2lU4IzqAwjKNt8wDQYJKoZIhvcNAQELBQAwMzExMC8GA1UEAxMoNmM4ZmQxYThhZTU5OTZjMWU1NjkyYzFhOGM0MmJmZTljM2JhNTc0NTAeFw0yNDA0MDYwOTA4NTRaFw0yNTA3MDEwMDAwMDBaMDMxMTAvBgNVBAMTKGQ4NmZhNjczMGNhNmE2Njc5NWY1MWQwN2I0YjY0YThkMmEzZWY3YzYwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCGyLzaqADkIkKk32lYHDa5Kx4H4rPqisuqVRi7LpmtPqHGa/qXsRO4l9JtQXCoVVSTiEK25s4aaU613JXZHbgAJbmp3P+4XMddxMnwdj3QNHSjypiKDZQExrEdNKdsmXpU4ikDfitTuO/l3Jtn65fIFZOrJs2008//g96y/5BXm0O8W9oNik9V66Y95FuPdsXiGwQ6hyl8GTv5B7X0aOTAMf2QKpmP5wz3MBQHMu5PK4g+9SKFTj1tgJULGtowJonG4GBBRTVVU4a1UD91PHPTguN5Lfkdrnfb0vh+pcy52jBXZ3mWq0Gd7i/RKYxp7LaKltmWl/LmHuRdezuY8GVcqS7KDotCdXi39ftNZLd/YtrqeG`

	b, e := DecodeBase64(base)
	fmt.Println(len(b), e)
}

func TestTrimBase64(t *testing.T) {
	base := `MIIIAQYJKoZIhvcNAQcCoIIH8jCCB+4CAQMx
	DTALBglghkgBZQMEAgEwLAYLKoZIhvcNAQkQ
	ARigHQQbMBkCAkB9MBMwEQQCAAEwCzAJAwQC
	EiPgAgEYoIIF/DCCBfgwggTgoAMCAQICFEUo
	qA/pCbWEOPpeTlI0XvuCvbQ7MA0GCSqGSIb3
	DQEBCwUAMD0xOzA5BgNVBAMTMmRmNmYzYjNh
	MzRiNjM4NmQxYTMyZDhmNGZhMzE3OGVmMzE4
	ODdkOGI0MjhkZmFhNDc2MB4XDTIzMDYxODAw
	MDAwMFoXDTIzMDcyMzIzNTk1OVowejFJMEcG
	A1UEBRNAMTcxYjcxMmI4NzhhY2ZmNWRhYjhk
	OGRhMWZhNTY4YWUyZDNlMDkyYjAyMWM3NWNh
	OTVkOTdkMTViYjc1YzVlMjEtMCsGA1UEAxMk
	NWYyNzYwNDUtNWI5Zi00NWVmLTkyM2QtZjNm
	Y2UyNGE2MjI1MIIBIjANBgkqhkiG9w0BAQEF
	AAOCAQ8AMIIBCgKCAQEAzwcb2b8W4V5y7tUh
	BX8JaeCjOJ3hHK4kMaK2ItVakjhXFRjy4oNS
	K2naPqmVvkPoVuKT3PKDH8GuwDNNUlWpqsIn
	aOqKxPIeDHarctY7/cTxyCiIUAI3yDlPO7Aj
	spz0fQUENj8veuesYIsYTlVGTK7LGMSJnLyx
	/7RJeRkkG5B+kR0UcyzlO946m2l4gjGjMfVY
	2V2gLGbYE8DHnT2BnZhwiCBDh7bj1bUcxOSI
	HKgMn8h0/Yk4/1y3j3ln7axLanQoXU0wk/ux
	cwYPZBsXBBsiUm//oIGOGWrd2KGB41QCTNDn
	CKF393bqjOEpEdoOgebu35FC30/9/ljCj7BO
	NQIDAQABo4ICsTCCAq0wHQYDVR0OBBYEFCWx
	EDuxxpnyZHM9wIaGvE0Oki6JMB8GA1UdIwQY
	MBaAFCWt00KwHreljq0ZkCaItUs/gfS4MA4G
	A1UdDwEB/wQEAwIHgDCB8wYIKwYBBQUHAQEE
	geYwgeMwgeAGCCsGAQUFBzAChoHTcnN5bmM6Ly9ycGtpLmFy
	aW4ubmV0L3JlcG9zaXRvcnkvYXJpbi1ycGtpLXRhLzVlNGEyM2VhLWU4MGEtNDAzZS1iMDhjLTIxNzFkYTIxN
	TdkMy8yYTI0Njk0Ny0yZDYyLTRhNmMtYmEwNS04NzE4N2YwMDk5YjIvODUxY2VmMTctMTMyYS00MzM3LWI3ZDEtYmYxNmE1MmZm
	ZDAzL2RmNmYzYjNhMzRiNjM4NmQxYTMyZDhmNGZhMzE3OGVmMzE4ODdkOGI0MjhkZmFhNDc2LmNlcjCBngYIKwYBBQUHAQsEgZEwgY
	4wgYsGCCsGAQUFBzALhn9yc3luYzovL3Jwa2ktcnN5bmMudXMtZWFzdC0yLmFtYXpvbmF3cy5jb20vdm9sdW1lL2Y3MDM2OTZlLWU0
	N2ItNGMyMC1iZDkzLTZmODA5MDRlNDJkMi80ZjJjMDdmOC00MGNkLTRjNzMtYmQ1OC0wZTRkYTkzYWQ5ZmIucm9hMIGIBgNVHR8EgYA
	wfjB8oHqgeIZ2cnN5bmM6Ly9ycGtpLXJzeW5jLnVzLWVhc3QtMi5hbWF6b25hd3MuY29tL3ZvbHVtZS9mNzAzNjk2ZS1lNDdiLTRjMjA
	tYmQ5My02ZjgwOTA0ZTQyZDIvdGpodEdqTFk5UG94ZU84eGlIMkxRbzM2cEhZLmNybDAYBgNVHSABAf8EDjAMMAoGCCsGAQUFBw4CM
	B8GCCsGAQUFBwEHAQH/BBAwDjAMBAIAATAGAwQCEiPgMA0GCSqGSIb3DQEBCwUAA4IBAQBHhRUNetw2iAg91NdixAiMfK+iXMNx
	y6qKUnHd7Uzs9zfMu1mr0e8zxKTDHjjpW+HNO50UiaFy/30YZ18TDHb11WCSb4qKiTyii3gfWCKW5iRyDRiQZQ
	qpO2E8ld3+VXmZwhoXSFKel7B+m/yCPqg5pvMEVDieZpIUpzbk8v7gHIZ5KIHnxRckJY9
	kozukO05jev58dOi287mZJUpBOiERHmPYz/MgtupYbaYL0Dd
	Mx6za0lclwGqD2thmA8N+Fo7gZxySXK4BHz8KDU6KweKUQrxyAVhvSSZ56nIE4hq9RAbbKIE
	sD9hi6zyudWn4Db7EJgASS0uZdbkZszpy4guEMYIBqjCCAaYCAQOAFCWxEDuxxpnyZHM9wIaGvE0Oki6JMAs
	GCWCGSAFlAwQCAaBrMBoGCSqGSIb3DQEJAzENBgsqhkiG9w0BCRABGDAcBgkqhkiG9w0BCQUxDxcNMjMwNjE4MDAwMDAw
	WjAvBgkqhkiG9w0BCQQxIgQgHQwB3sI8nNCdDWaw3MoyEjZE6S4QYRS/cn/jFtK2UbEwDQYJKoZIhvc
	NAQELBQAEggEAplvpUtLCyuiGK/RidsUhzcXFL4lWRfWoAcA3RcSUiUXyH4KHSwGGYdVE+o39YUNDEP7egVWv1NVv27A6B
	Xq1oDIdrrQesOkzlZ+wmW51yvfWRw5BmIRaq7P76TnHJZbDmQ6QPUP6AXT
	j1c4ZSdGTm2FGiYpFmh4mgl6Zf189pFzlHWsVdEOdP1w23Ij4UzEsirkbtVTXKzBvesmmYCltpcmxdi3axjdXQ/WZtitjP2rLBCi7F6AtU
	CBrIDqeKcZIy3z5DidfVoUYZ27lYW3e1M
	t4vjpsSrxOFxTa848aZ4NlN1rFiSGYfy2CvNviZLDbNp+AiuFprL8HQym9yVfLtg==`

	b := TrimBase64(base)
	fmt.Println(b)
}
