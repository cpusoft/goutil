package certutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// 全局测试资源
var (
	testTempDir        string
	validRootCertFile  string
	validChildCertFile string
	expiredCertFile    string
	invalidCertFile    string
	largeFile          string
	emptyFile          string
	validCRLFile       string
	expiredCRLFile     string
	mismatchCRLFile    string
	invalidCRLFile     string
	testRootKey        *rsa.PrivateKey
)

// TestMain 初始化测试资源，执行完测试后清理
func TestMain(m *testing.M) {
	var err error
	// 创建临时目录
	testTempDir, err = os.MkdirTemp("", "certutil-test-*")
	if err != nil {
		fmt.Println("创建测试临时目录失败:", err)
		os.Exit(1)
	}

	// 初始化测试资源
	if err = setupTestResources(); err != nil {
		fmt.Println("初始化测试资源失败:", err)
		os.RemoveAll(testTempDir)
		os.Exit(1)
	}

	// 执行测试
	exitCode := m.Run()

	// 清理资源
	os.RemoveAll(testTempDir)
	os.Exit(exitCode)
}

// 修复CRL生成逻辑：移除错误的reasonCode扩展，简化CRL结构
func setupTestResources() error {
	var err error

	// 1. 生成RSA密钥对
	testRootKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("生成根密钥失败: %w", err)
	}
	childKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("生成子密钥失败: %w", err)
	}
	badKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("生成错误密钥失败: %w", err)
	}

	// 2. 生成根证书模板
	rootTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test Root CA",
			Organization: []string{"CertUtil Test Org"},
		},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	// 生成根证书（DER格式）
	rootCertBytes, err := x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate, &testRootKey.PublicKey, testRootKey)
	if err != nil {
		return fmt.Errorf("生成根证书失败: %w", err)
	}
	validRootCertFile = filepath.Join(testTempDir, "root.cer")
	if err = os.WriteFile(validRootCertFile, rootCertBytes, 0600); err != nil {
		return fmt.Errorf("写入根证书失败: %w", err)
	}

	// 解析根证书
	rootCert, err := x509.ParseCertificate(rootCertBytes)
	if err != nil {
		return fmt.Errorf("解析根证书失败: %w", err)
	}

	// 3. 生成有效子证书
	childTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName:   "Test Child Cert",
			Organization: []string{"CertUtil Test Org"},
		},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}
	childCertBytes, err := x509.CreateCertificate(rand.Reader, childTemplate, rootCert, &childKey.PublicKey, testRootKey)
	if err != nil {
		return fmt.Errorf("生成子证书失败: %w", err)
	}
	validChildCertFile = filepath.Join(testTempDir, "child.cer")
	if err = os.WriteFile(validChildCertFile, childCertBytes, 0600); err != nil {
		return fmt.Errorf("写入子证书失败: %w", err)
	}

	// 4. 生成过期证书
	expiredTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			CommonName:   "Expired Cert",
			Organization: []string{"CertUtil Test Org"},
		},
		NotBefore:             time.Now().Add(-48 * time.Hour),
		NotAfter:              time.Now().Add(-24 * time.Hour), // 已过期
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}
	expiredCertBytes, err := x509.CreateCertificate(rand.Reader, expiredTemplate, rootCert, &childKey.PublicKey, testRootKey)
	if err != nil {
		return fmt.Errorf("生成过期证书失败: %w", err)
	}
	expiredCertFile = filepath.Join(testTempDir, "expired.cer")
	if err = os.WriteFile(expiredCertFile, expiredCertBytes, 0600); err != nil {
		return fmt.Errorf("写入过期证书失败: %w", err)
	}

	// 5. 生成无效证书（随机字节）
	invalidCertFile = filepath.Join(testTempDir, "invalid.cer")
	invalidBytes := make([]byte, 1024)
	rand.Read(invalidBytes)
	if err = os.WriteFile(invalidCertFile, invalidBytes, 0600); err != nil {
		return fmt.Errorf("写入无效证书失败: %w", err)
	}

	// 6. 生成超大文件（51MB，超过50MB限制）
	largeFile = filepath.Join(testTempDir, "large.file")
	largeFileFd, err := os.Create(largeFile)
	if err != nil {
		return fmt.Errorf("创建超大文件失败: %w", err)
	}
	largeFileFd.Write(make([]byte, 51*1024*1024))
	largeFileFd.Close()

	// 7. 生成空文件
	emptyFile = filepath.Join(testTempDir, "empty.file")
	emptyFd, err := os.Create(emptyFile)
	if err != nil {
		return fmt.Errorf("创建空文件失败: %w", err)
	}
	emptyFd.Close()

	// 8. 生成有效CRL - 修复：移除错误的reasonCode扩展
	validCRLTemplate := &x509.RevocationList{
		Number: big.NewInt(1),
		// 修复1：移除格式错误的reasonCode扩展，简化吊销证书列表
		RevokedCertificates: []pkix.RevokedCertificate{
			{
				SerialNumber:   big.NewInt(999),
				RevocationTime: time.Now().Add(-12 * time.Hour),
				// 移除错误的Extensions字段
			},
		},
		ThisUpdate:         time.Now().Add(-24 * time.Hour),
		NextUpdate:         time.Now().Add(24 * time.Hour),
		SignatureAlgorithm: x509.SHA256WithRSA,
	}
	validCRLBytes, err := x509.CreateRevocationList(rand.Reader, validCRLTemplate, rootCert, testRootKey)
	if err != nil {
		return fmt.Errorf("生成有效CRL失败: %w", err)
	}
	validCRLFile = filepath.Join(testTempDir, "valid.crl")
	if err = os.WriteFile(validCRLFile, validCRLBytes, 0600); err != nil {
		return fmt.Errorf("写入有效CRL失败: %w", err)
	}

	// 9. 生成过期CRL
	expiredCRLTemplate := &x509.RevocationList{
		Number:              big.NewInt(2),
		RevokedCertificates: []pkix.RevokedCertificate{},
		ThisUpdate:          time.Now().Add(-48 * time.Hour),
		NextUpdate:          time.Now().Add(-24 * time.Hour), // 已过期
		SignatureAlgorithm:  x509.SHA256WithRSA,
	}
	expiredCRLBytes, err := x509.CreateRevocationList(rand.Reader, expiredCRLTemplate, rootCert, testRootKey)
	if err != nil {
		return fmt.Errorf("生成过期CRL失败: %w", err)
	}
	expiredCRLFile = filepath.Join(testTempDir, "expired.crl")
	if err = os.WriteFile(expiredCRLFile, expiredCRLBytes, 0600); err != nil {
		return fmt.Errorf("写入过期CRL失败: %w", err)
	}

	// 10. 生成颁发者不匹配的CRL（用错误密钥签名）
	mismatchCRLTemplate := &x509.RevocationList{
		Number:              big.NewInt(3),
		RevokedCertificates: []pkix.RevokedCertificate{},
		ThisUpdate:          time.Now().Add(-24 * time.Hour),
		NextUpdate:          time.Now().Add(24 * time.Hour),
		SignatureAlgorithm:  x509.SHA256WithRSA,
	}
	mismatchCRLBytes, err := x509.CreateRevocationList(rand.Reader, mismatchCRLTemplate, rootCert, badKey)
	if err != nil {
		return fmt.Errorf("生成不匹配CRL失败: %w", err)
	}
	mismatchCRLFile = filepath.Join(testTempDir, "mismatch.crl")
	if err = os.WriteFile(mismatchCRLFile, mismatchCRLBytes, 0600); err != nil {
		return fmt.Errorf("写入不匹配CRL失败: %w", err)
	}

	// 11. 生成无效CRL（随机字节）
	invalidCRLFile = filepath.Join(testTempDir, "invalid.crl")
	if err = os.WriteFile(invalidCRLFile, invalidBytes, 0600); err != nil {
		return fmt.Errorf("写入无效CRL失败: %w", err)
	}

	return nil
}

// -------------------------- 功能测试 --------------------------

// TestReadFileToCer 测试证书读取功能（覆盖所有边界场景）
func TestReadFileToCer(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		wantErr bool
		errMsg  string
		desc    string
	}{
		{
			name:    "有效DER证书",
			file:    validRootCertFile,
			wantErr: false,
			desc:    "读取合法的DER格式根证书",
		},
		{
			name:    "超大文件（51MB）",
			file:    largeFile,
			wantErr: true,
			errMsg:  "file size exceeds maximum limit",
			desc:    "读取超过50MB限制的文件",
		},
		{
			name:    "空文件",
			file:    emptyFile,
			wantErr: true,
			desc:    "读取空文件",
		},
		{
			name:    "无效证书（随机字节）",
			file:    invalidCertFile,
			wantErr: true,
			desc:    "读取非证书格式的文件",
		},
		{
			name:    "不存在的文件",
			file:    filepath.Join(testTempDir, "nocert.cer"),
			wantErr: true,
			desc:    "读取不存在的文件",
		},
		{
			name:    "PEM格式证书",
			file:    createPEMCertFile(t, validRootCertFile),
			wantErr: false,
			desc:    "读取PEM格式的证书",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ReadFileToCer(tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFileToCer() 测试[%s]失败: 期望错误=%v, 实际错误=%v, 错误信息=%s",
					tt.desc, tt.wantErr, err != nil, err)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ReadFileToCer() 测试[%s]失败: 期望错误包含'%s', 实际错误='%s'",
					tt.desc, tt.errMsg, err.Error())
			}
		})
	}
}

// TestReadFileToCrl 测试CRL读取功能（覆盖所有边界场景）
func TestReadFileToCrl(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		wantErr bool
		errMsg  string
		desc    string
	}{
		{
			name:    "有效CRL",
			file:    validCRLFile,
			wantErr: false,
			desc:    "读取合法的DER格式CRL",
		},
		{
			name:    "超大文件（51MB）",
			file:    largeFile,
			wantErr: true,
			errMsg:  "file size exceeds maximum limit",
			desc:    "读取超过50MB限制的CRL文件",
		},
		{
			name:    "空文件",
			file:    emptyFile,
			wantErr: true,
			errMsg:  "this file is empty, not a valid CRL",
			desc:    "读取空的CRL文件",
		},
		{
			name:    "无效CRL（随机字节）",
			file:    invalidCRLFile,
			wantErr: true,
			desc:    "读取非CRL格式的文件",
		},
		{
			name:    "不存在的文件",
			file:    filepath.Join(testTempDir, "nocrl.crl"),
			wantErr: true,
			desc:    "读取不存在的CRL文件",
		},
		{
			name:    "PEM格式CRL",
			file:    createPEMCRLFile(t, validCRLFile),
			wantErr: false,
			desc:    "读取PEM格式的CRL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ReadFileToCrl(tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFileToCrl() 测试[%s]失败: 期望错误=%v, 实际错误=%v, 错误信息=%s",
					tt.desc, tt.wantErr, err != nil, err)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ReadFileToCrl() 测试[%s]失败: 期望错误包含'%s', 实际错误='%s'",
					tt.desc, tt.errMsg, err.Error())
			}
		})
	}
}

// TestVerifyCerByX509 测试证书验证功能
func TestVerifyCerByX509(t *testing.T) {
	tests := []struct {
		name       string
		fatherFile string
		childFile  string
		wantResult string
		wantErr    bool
		errMsg     string
		desc       string
	}{
		{
			name:       "有效父子证书验证",
			fatherFile: validRootCertFile,
			childFile:  validChildCertFile,
			wantResult: "ok",
			wantErr:    false,
			desc:       "验证由根证书签发的子证书",
		},
		{
			name:       "过期证书验证",
			fatherFile: validRootCertFile,
			childFile:  expiredCertFile,
			wantResult: "fail",
			wantErr:    true,
			errMsg:     "certificate has expired or is not yet valid",
			desc:       "验证已过期的证书",
		},
		{
			name:       "无效子证书",
			fatherFile: validRootCertFile,
			childFile:  invalidCertFile,
			wantResult: "fail",
			wantErr:    true,
			desc:       "验证无效格式的子证书",
		},
		{
			name:       "不存在的父证书",
			fatherFile: filepath.Join(testTempDir, "nofather.cer"),
			childFile:  validChildCertFile,
			wantResult: "fail",
			wantErr:    true,
			desc:       "父证书文件不存在",
		},
		{
			name:       "不存在的子证书",
			fatherFile: validRootCertFile,
			childFile:  filepath.Join(testTempDir, "nochild.cer"),
			wantResult: "fail",
			wantErr:    true,
			desc:       "子证书文件不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := VerifyCerByX509(tt.fatherFile, tt.childFile)
			if result != tt.wantResult {
				t.Errorf("VerifyCerByX509() 测试[%s]失败: 期望结果='%s', 实际结果='%s'",
					tt.desc, tt.wantResult, result)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyCerByX509() 测试[%s]失败: 期望错误=%v, 实际错误=%v, 错误信息=%s",
					tt.desc, tt.wantErr, err != nil, err)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("VerifyCerByX509() 测试[%s]失败: 期望错误包含'%s', 实际错误='%s'",
					tt.desc, tt.errMsg, err.Error())
			}
		})
	}
}

// TestVerifyEeCertByX509 测试EE证书验证功能
func TestVerifyEeCertByX509(t *testing.T) {
	// 生成包含EE证书的测试文件
	eeCertBytes, err := os.ReadFile(validChildCertFile)
	if err != nil {
		t.Fatalf("读取EE证书失败: %v", err)
	}
	// 构造包含EE证书的大文件（证书在 [100, 100+len(eeCertBytes)] 位置）
	eeTestFile := filepath.Join(testTempDir, "ee_test.file")
	eeTestBytes := append(make([]byte, 100), eeCertBytes...)
	eeTestBytes = append(eeTestBytes, make([]byte, 50)...)
	if err = os.WriteFile(eeTestFile, eeTestBytes, 0600); err != nil {
		t.Fatalf("写入EE测试文件失败: %v", err)
	}

	tests := []struct {
		name       string
		fatherFile string
		mftRoaFile string
		start      uint64
		end        uint64
		wantResult string
		wantErr    bool
		errMsg     string
		desc       string
	}{
		{
			name:       "有效EE证书验证",
			fatherFile: validRootCertFile,
			mftRoaFile: eeTestFile,
			start:      100,
			end:        100 + uint64(len(eeCertBytes)),
			wantResult: "ok",
			wantErr:    false,
			desc:       "验证有效范围的EE证书",
		},
		{
			name:       "起始>=结束",
			fatherFile: validRootCertFile,
			mftRoaFile: eeTestFile,
			start:      200,
			end:        100,
			wantResult: "fail",
			wantErr:    true,
			errMsg:     "invalid EE certificate range",
			desc:       "EE证书起始位置大于结束位置",
		},
		{
			name:       "结束超出文件长度",
			fatherFile: validRootCertFile,
			mftRoaFile: eeTestFile,
			start:      100,
			end:        uint64(len(eeTestBytes) + 100),
			wantResult: "fail",
			wantErr:    true,
			errMsg:     "invalid EE certificate range",
			desc:       "EE证书结束位置超出文件长度",
		},
		{
			name:       "无效EE证书内容",
			fatherFile: validRootCertFile,
			mftRoaFile: eeTestFile,
			start:      0,
			end:        100,
			wantResult: "fail",
			wantErr:    true,
			desc:       "验证无效内容的EE证书范围",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := VerifyEeCertByX509(tt.fatherFile, tt.mftRoaFile, tt.start, tt.end)
			if result != tt.wantResult {
				t.Errorf("VerifyEeCertByX509() 测试[%s]失败: 期望结果='%s', 实际结果='%s'",
					tt.desc, tt.wantResult, result)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyEeCertByX509() 测试[%s]失败: 期望错误=%v, 实际错误=%v, 错误信息=%s",
					tt.desc, tt.wantErr, err != nil, err)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("VerifyEeCertByX509() 测试[%s]失败: 期望错误包含'%s', 实际错误='%s'",
					tt.desc, tt.errMsg, err.Error())
			}
		})
	}
}

// TestVerifyRootCerByOpenssl 测试OpenSSL根证书验证
func TestVerifyRootCerByOpenssl(t *testing.T) {
	// 跳过CI环境（无OpenSSL）
	if os.Getenv("CI") == "true" {
		t.Skip("CI环境跳过OpenSSL测试")
	}

	tests := []struct {
		name       string
		rootFile   string
		wantResult string
		wantErr    bool
		desc       string
	}{
		{
			name:       "有效根证书验证",
			rootFile:   validRootCertFile,
			wantResult: "ok",
			wantErr:    false,
			desc:       "验证合法的根证书",
		},
		{
			name:       "无效证书验证",
			rootFile:   invalidCertFile,
			wantResult: "fail",
			wantErr:    true,
			desc:       "验证无效格式的证书",
		},
		{
			name:       "不存在的文件",
			rootFile:   filepath.Join(testTempDir, "noroot.cer"),
			wantResult: "fail",
			wantErr:    true,
			desc:       "验证不存在的证书文件",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := VerifyRootCerByOpenssl(tt.rootFile)
			if result != tt.wantResult {
				t.Errorf("VerifyRootCerByOpenssl() 测试[%s]失败: 期望结果='%s', 实际结果='%s'",
					tt.desc, tt.wantResult, result)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyRootCerByOpenssl() 测试[%s]失败: 期望错误=%v, 实际错误=%v, 错误信息=%s",
					tt.desc, tt.wantErr, err != nil, err)
			}
		})
	}
}

// TestVerifyCrlByX509 测试CRL验证功能
func TestVerifyCrlByX509(t *testing.T) {
	tests := []struct {
		name       string
		cerFile    string
		crlFile    string
		wantResult string
		wantErr    bool
		errMsg     string
		desc       string
	}{
		{
			name:       "有效CRL验证",
			cerFile:    validRootCertFile,
			crlFile:    validCRLFile,
			wantResult: "ok",
			wantErr:    false,
			desc:       "验证合法的CRL",
		},
		{
			name:       "过期CRL验证",
			cerFile:    validRootCertFile,
			crlFile:    expiredCRLFile,
			wantResult: "fail",
			wantErr:    true,
			errMsg:     "CRL has expired",
			desc:       "验证过期的CRL",
		},
		{
			name:       "颁发者不匹配CRL",
			cerFile:    validRootCertFile,
			crlFile:    mismatchCRLFile,
			wantResult: "fail",
			wantErr:    true,
			errMsg:     "CRL signature verification failed",
			desc:       "验证颁发者不匹配的CRL",
		},
		{
			name:       "无效CRL文件",
			cerFile:    validRootCertFile,
			crlFile:    invalidCRLFile,
			wantResult: "fail",
			wantErr:    true,
			desc:       "验证无效格式的CRL",
		},
		{
			name:       "不存在的证书文件",
			cerFile:    filepath.Join(testTempDir, "nocer.cer"),
			crlFile:    validCRLFile,
			wantResult: "fail",
			wantErr:    true,
			desc:       "证书文件不存在",
		},
		{
			name:       "不存在的CRL文件",
			cerFile:    validRootCertFile,
			crlFile:    filepath.Join(testTempDir, "nocrl.crl"),
			wantResult: "fail",
			wantErr:    true,
			desc:       "CRL文件不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := VerifyCrlByX509(tt.cerFile, tt.crlFile)
			if result != tt.wantResult {
				t.Errorf("VerifyCrlByX509() 测试[%s]失败: 期望结果='%s', 实际结果='%s'",
					tt.desc, tt.wantResult, result)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyCrlByX509() 测试[%s]失败: 期望错误=%v, 实际错误=%v, 错误信息=%s",
					tt.desc, tt.wantErr, err != nil, err)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("VerifyCrlByX509() 测试[%s]失败: 期望错误包含'%s', 实际错误='%s'",
					tt.desc, tt.errMsg, err.Error())
			}
		})
	}
}

// TestJudgeBelongNic 测试NIC归属判断函数
func TestJudgeBelongNic(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
		filePath string
		want     string
		desc     string
	}{
		{
			name:     "有效路径匹配",
			repoPath: "/repo/nic/",
			filePath: "/repo/nic/apnic/file.txt",
			want:     "apnic",
			desc:     "文件路径匹配仓库路径，返回一级目录",
		},
		{
			name:     "空仓库路径",
			repoPath: "",
			filePath: "/repo/nic/apnic/file.txt",
			want:     "",
			desc:     "仓库路径为空，返回空",
		},
		{
			name:     "空文件路径",
			repoPath: "/repo/nic/",
			filePath: "",
			want:     "",
			desc:     "文件路径为空，返回空",
		},
		{
			name:     "路径不匹配",
			repoPath: "/repo/nic/",
			filePath: "/other/path/apnic/file.txt",
			want:     "",
			desc:     "文件路径不包含仓库路径，返回空",
		},
		{
			name:     "路径无子目录",
			repoPath: "/repo/nic/",
			filePath: "/repo/nic/file.txt",
			want:     "file.txt",
			desc:     "文件路径无一级子目录，返回文件名",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JudgeBelongNic(tt.repoPath, tt.filePath)
			if got != tt.want {
				t.Errorf("JudgeBelongNic() 测试[%s]失败: 期望='%s', 实际='%s'",
					tt.desc, tt.want, got)
			}
		})
	}
}

// -------------------------- 临界值测试 --------------------------

// TestReadFileToCer_EdgeCases 证书读取临界值测试
func TestReadFileToCer_EdgeCases(t *testing.T) {
	// 测试50MB临界值（刚好50MB的证书文件）
	edgeFile := filepath.Join(testTempDir, "50mb_cer.file")
	edgeBytes := make([]byte, 50*1024*1024)
	// 前1024字节填充有效证书内容
	certBytes, _ := os.ReadFile(validRootCertFile)
	copy(edgeBytes, certBytes)
	if err := os.WriteFile(edgeFile, edgeBytes, 0600); err != nil {
		t.Fatalf("创建50MB临界文件失败: %v", err)
	}

	// 测试刚好50MB（应该成功）
	_, err := ReadFileToCer(edgeFile)
	if err != nil {
		t.Errorf("读取50MB临界文件失败: %v", err)
	}

	// 测试50MB+1字节（应该失败）
	overEdgeFile := filepath.Join(testTempDir, "50mb_plus1_cer.file")
	overEdgeBytes := make([]byte, 50*1024*1024+1)
	copy(overEdgeBytes, certBytes)
	if err := os.WriteFile(overEdgeFile, overEdgeBytes, 0600); err != nil {
		t.Fatalf("创建50MB+1字节文件失败: %v", err)
	}

	_, err = ReadFileToCer(overEdgeFile)
	if err == nil || !strings.Contains(err.Error(), "file size exceeds maximum limit") {
		t.Errorf("读取50MB+1字节文件未触发预期错误: %v", err)
	}
}

// TestVerifyCerByteByX509_EdgeCases 证书验证临界值测试
func TestVerifyCerByteByX509_EdgeCases(t *testing.T) {
	// 测试颁发者字符串匹配但原始字节不匹配的场景
	rootCertBytes, _ := os.ReadFile(validRootCertFile)
	rootCert, _ := x509.ParseCertificate(rootCertBytes)

	// 构造"字符串匹配、字节不匹配"的子证书
	childTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(4),
		Subject: pkix.Name{
			CommonName:   "Test Child Cert",
			Organization: []string{"CertUtil Test Org"},
		},
		// 手动构造Issuer（字符串相同但字节顺序不同）
		Issuer: pkix.Name{
			Organization: []string{"CertUtil Test Org"}, // 顺序调换
			CommonName:   "Test Root CA",
		},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}
	childKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	childCertBytes, _ := x509.CreateCertificate(rand.Reader, childTemplate, rootCert, &childKey.PublicKey, testRootKey)

	// 验证应该返回ok（字符串匹配）
	result, err := VerifyCerByteByX509(rootCertBytes, childCertBytes)
	if result != "ok" || err != nil {
		t.Errorf("颁发者字符串匹配但字节不匹配的场景验证失败: result=%s, err=%v", result, err)
	}
}

// -------------------------- 性能测试 --------------------------

// BenchmarkReadFileToCer 证书读取性能测试
func BenchmarkReadFileToCer(b *testing.B) {
	// 预热
	_, err := ReadFileToCer(validRootCertFile)
	if err != nil {
		b.Fatalf("预热失败: %v", err)
	}

	// 重置计时器
	b.ResetTimer()

	// 并发测试
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ReadFileToCer(validRootCertFile)
		}
	})
}

// BenchmarkReadFileToCrl CRL读取性能测试
func BenchmarkReadFileToCrl(b *testing.B) {
	// 预热
	_, err := ReadFileToCrl(validCRLFile)
	if err != nil {
		b.Fatalf("预热失败: %v", err)
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ReadFileToCrl(validCRLFile)
		}
	})
}

// BenchmarkVerifyCerByX509 证书验证性能测试
func BenchmarkVerifyCerByX509(b *testing.B) {
	// 预热
	_, err := VerifyCerByX509(validRootCertFile, validChildCertFile)
	if err != nil {
		b.Fatalf("预热失败: %v", err)
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			VerifyCerByX509(validRootCertFile, validChildCertFile)
		}
	})
}

// BenchmarkVerifyCrlByX509 CRL验证性能测试
func BenchmarkVerifyCrlByX509(b *testing.B) {
	// 预热
	_, err := VerifyCrlByX509(validRootCertFile, validCRLFile)
	if err != nil {
		b.Fatalf("预热失败: %v", err)
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			VerifyCrlByX509(validRootCertFile, validCRLFile)
		}
	})
}

// -------------------------- 辅助函数 --------------------------

// createPEMCertFile 将DER证书转换为PEM格式
func createPEMCertFile(t *testing.T, derFile string) string {
	derBytes, err := os.ReadFile(derFile)
	if err != nil {
		t.Fatalf("读取DER证书失败: %v", err)
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})

	pemFile := filepath.Join(testTempDir, "pem_cert.cer")
	if err = os.WriteFile(pemFile, pemBytes, 0600); err != nil {
		t.Fatalf("写入PEM证书失败: %v", err)
	}

	return pemFile
}

// createPEMCRLFile 将DER CRL转换为PEM格式
func createPEMCRLFile(t *testing.T, derFile string) string {
	derBytes, err := os.ReadFile(derFile)
	if err != nil {
		t.Fatalf("读取DER CRL失败: %v", err)
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "X509 CRL",
		Bytes: derBytes,
	})

	pemFile := filepath.Join(testTempDir, "pem_crl.crl")
	if err = os.WriteFile(pemFile, pemBytes, 0600); err != nil {
		t.Fatalf("写入PEM CRL失败: %v", err)
	}

	return pemFile
}

// ////////  old tests
/*
func TestReadFileToCer1(t *testing.T) {
	path := `G:\Download\cert\asncer4_1\`
	f := path + `d57344.cer`
	p, err := ReadFileToCer(f)
	fmt.Println("\r\n\r\ncer:", f, p, err)

	f = path + `d57344.pem.cer`
	p, err = ReadFileToCer(f)
	fmt.Println("\r\n\r\npem.cer:", f, p, err)
}

func TestReadFileToCrl1(t *testing.T) {
	path := `G:\Download\cert\asncrl2\`
	fatherFile := path + `1.crl`
	p, err := ReadFileToCrl(fatherFile)
	fmt.Println(p, err)

}


func TestReadFileToByte(t *testing.T) {
	path := `G:\Download\cert\verify\2\`
	fatherFile := path + `inter.pem.cer`
	p, by, err := ReadFileToByte(fatherFile)
	fmt.Println(p, by, err)

	fatherFile = path + `inter.cer`
	p, by, err = ReadFileToByte(fatherFile)
	fmt.Println(p, by, err)
}


func TestVerifyCertByX5091(t *testing.T) {
	path := `G:\Download\cert\`
	fatherFile := path + `c0793683-aa07-4935-a2c2-ec423ea7dd0b.father.cer`
	childFile := path + `c922abf8-95b1-37f0-90cd-bdb125467e8e.ee.cer`

	result, err := VerifyCerByX509(fatherFile, childFile)
	fmt.Println(result, err)
}

func TestVerifyRootCertByX5091(t *testing.T) {
	path := `E:\Go\common-util\src\certutil\example\`
	root := path + `root.cer`
	//childFile := path + `inter.cer`

	result, err := VerifyRootCerByOpenssl(root)
	fmt.Println(result, err)
}

func TestVerifyEeCertByX5091(t *testing.T) {

//		/root/rpki/repo/repo/rpki.ripe.net/repository/DEFAULT/ec/49c449-2d9c-4fc9-b340-51a23ddb6410/1/
//		rtpKuIKhDn9Y8Zg6y9HhlQfmPsU.roa
//			"eeStart": 159,
//			"eeEnd": 1426


//			/root/rpki/repo/repo/rpki.ripe.net/repository/DEFAULT/
//			ACBRR9OW8JgDvUcuWBka9usiwvU.cer

	path := `G:\Download\cert\`
	fatherFile := path + `ohcWJIUz0QduJriNGOlBlT-lB9c.cer`
	childFile := path + `db42e932-926a-42bd-afdb-63320fa7ec40.roa`

	result, err := VerifyEeCertByX509(fatherFile, childFile, 838969, 1019659)
	fmt.Println(result, err)
}

func TestVerifyCrlByX5091(t *testing.T) {
	path := `G:\Download\cert\verify\4\`

	//cerFile := path + `inter.cer` //err
	cerFile := path + `bW-_qXU9uNhGQz21NR2ansB8lr0.cer` //ok
	crlFile := path + `bW-_qXU9uNhGQz21NR2ansB8lr0.crl`

	result, err := VerifyCrlByX509(cerFile, crlFile)
	fmt.Println(result, err)
}
*/
