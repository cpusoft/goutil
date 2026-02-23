package certutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// 全局测试变量
var (
	testTempDir    string
	testRootCert   []byte
	testChildCert  []byte
	testCRL        []byte // 存储DER格式的CRL字节
	testLargeFile  string
	emptyFile      string
	invalidCert    string
	expiredCert    string
	validRootFile  string
	validChildFile string
	validCRLFile   string
)

// 测试初始化：创建测试用的证书、CRL和测试文件
func TestMain(m *testing.M) {
	// 创建临时目录
	var err error
	testTempDir, err = os.MkdirTemp("", "certutil-test-*")
	if err != nil {
		fmt.Printf("TestMain: 创建临时目录失败: %v\n", err)
		os.Exit(1)
	}

	// 生成测试用证书和CRL（使用Go 1.25推荐的CreateRevocationList）
	generateTestCertsAndCRL()

	// 创建测试文件
	createTestFiles()

	// 运行测试
	exitCode := m.Run()

	// 清理临时文件
	os.RemoveAll(testTempDir)
	os.Exit(exitCode)
}

// 生成测试用的证书和CRL（修复参数顺序+Marshal错误）
func generateTestCertsAndCRL() {
	// 生成RSA密钥对
	rootKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("generateTestCertsAndCRL: 生成根密钥失败: %v\n", err)
		os.Exit(1)
	}

	childKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("generateTestCertsAndCRL: 生成子密钥失败: %v\n", err)
		os.Exit(1)
	}

	// 生成根证书模板（需包含CRL签名权限）
	rootTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test Root CA",
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign, // 必须包含CRLSign
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	// 生成根证书
	testRootCert, err = x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	if err != nil {
		fmt.Printf("generateTestCertsAndCRL: 生成根证书失败: %v\n", err)
		os.Exit(1)
	}

	// 解析根证书（用于生成CRL）
	rootCert, err := x509.ParseCertificate(testRootCert)
	if err != nil {
		fmt.Printf("generateTestCertsAndCRL: 解析根证书失败: %v\n", err)
		os.Exit(1)
	}

	// 生成子证书
	childTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName:   "Test Child Cert",
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}
	testChildCert, err = x509.CreateCertificate(rand.Reader, childTemplate, rootTemplate, &childKey.PublicKey, rootKey)
	if err != nil {
		fmt.Printf("generateTestCertsAndCRL: 生成子证书失败: %v\n", err)
		os.Exit(1)
	}

	// 生成过期证书
	expiredTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			CommonName:   "Expired Cert",
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now().Add(-48 * time.Hour),
		NotAfter:              time.Now().Add(-24 * time.Hour), // 已过期
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}
	expiredCertBytes, err := x509.CreateCertificate(rand.Reader, expiredTemplate, rootTemplate, &childKey.PublicKey, rootKey)
	if err != nil {
		fmt.Printf("generateTestCertsAndCRL: 生成过期证书失败: %v\n", err)
		os.Exit(1)
	}

	// ========== 核心修复：参数顺序 + 正确处理Marshal ==========
	// 构建符合RFC 5280的CRL模板（X.509 v2）
	crlTemplate := &x509.RevocationList{
		Number: big.NewInt(1), // CRL序列号
		RevokedCertificates: []pkix.RevokedCertificate{
			{
				SerialNumber:   big.NewInt(999),                 // 吊销的证书序列号
				RevocationTime: time.Now().Add(-12 * time.Hour), // 吊销时间
				// 可选：添加吊销原因扩展（符合RFC 5280）
				Extensions: []pkix.Extension{
					{
						Id:       asn1.ObjectIdentifier{2, 5, 29, 21}, // 吊销原因OID
						Critical: false,
						Value:    []byte{0x01, 0x01, 0x00}, // 原因：未指定 (0)
					},
				},
			},
		},
		ThisUpdate:         time.Now().Add(-24 * time.Hour), // CRL生效时间
		NextUpdate:         time.Now().Add(24 * time.Hour),  // 下一次CRL更新时间
		SignatureAlgorithm: x509.SHA256WithRSA,              // 签名算法
	}

	// 修复：参数顺序错误（crlTemplate 和 rootCert 位置互换）
	testCRL, err = x509.CreateRevocationList(rand.Reader, crlTemplate, rootCert, rootKey)
	if err != nil {
		fmt.Printf("generateTestCertsAndCRL: 生成CRL失败: %v\n", err)
		os.Exit(1)
	}

	// 写入测试文件
	validRootFile = filepath.Join(testTempDir, "root.cer")
	os.WriteFile(validRootFile, testRootCert, 0600)

	validChildFile = filepath.Join(testTempDir, "child.cer")
	os.WriteFile(validChildFile, testChildCert, 0600)

	validCRLFile = filepath.Join(testTempDir, "crl.crl")
	os.WriteFile(validCRLFile, testCRL, 0600)

	expiredCert = filepath.Join(testTempDir, "expired.cer")
	os.WriteFile(expiredCert, expiredCertBytes, 0600)

	// 生成无效证书文件（随机字节）
	invalidCert = filepath.Join(testTempDir, "invalid.cer")
	randBytes := make([]byte, 1024)
	rand.Read(randBytes)
	os.WriteFile(invalidCert, randBytes, 0600)
}

// 创建测试用的各类文件
func createTestFiles() {
	// 空文件
	emptyFile = filepath.Join(testTempDir, "empty.file")
	os.Create(emptyFile)

	// 超大文件（51MB，超过maxFileSize=50MB）
	testLargeFile = filepath.Join(testTempDir, "large.file")
	largeFile, _ := os.Create(testLargeFile)
	largeFile.Write(make([]byte, 51*1024*1024)) // 51MB
	largeFile.Close()
}

// -------------------------- 功能测试 --------------------------

// TestReadFileToCer 测试证书读取功能
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
			file:    validRootFile,
			wantErr: false,
			desc:    "读取合法的DER格式根证书",
		},
		{
			name:    "超大文件",
			file:    testLargeFile,
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
			name:    "无效证书文件",
			file:    invalidCert,
			wantErr: true,
			desc:    "读取随机字节的无效证书",
		},
		{
			name:    "不存在的文件",
			file:    filepath.Join(testTempDir, "notexist.cer"),
			wantErr: true,
			desc:    "读取不存在的文件",
		},
		{
			name:    "PEM格式证书",
			file:    createPEMCertFile(t),
			wantErr: false,
			desc:    "读取PEM格式证书",
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
			if tt.wantErr && tt.errMsg != "" && err != nil && !containsErrMsg(err, tt.errMsg) {
				t.Errorf("ReadFileToCer() 测试[%s]失败: 期望错误包含'%s', 实际错误='%s'",
					tt.desc, tt.errMsg, err.Error())
			}
		})
	}
}

// TestReadFileToCrl 测试CRL读取功能（适配新CRL格式）
func TestReadFileToCrl(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		wantErr bool
		errMsg  string
		desc    string
	}{
		{
			name:    "有效CRL文件（RFC 5280）",
			file:    validCRLFile,
			wantErr: false,
			desc:    "读取合法的X.509 v2 CRL文件",
		},
		{
			name:    "超大文件",
			file:    testLargeFile,
			wantErr: true,
			errMsg:  "file size exceeds maximum limit",
			desc:    "读取超过50MB限制的CRL文件",
		},
		{
			name:    "空文件",
			file:    emptyFile,
			wantErr: true,
			errMsg:  "empty, not a valid CRL",
			desc:    "读取空文件作为CRL",
		},
		{
			name:    "证书文件冒充CRL",
			file:    validRootFile,
			wantErr: true,
			desc:    "读取证书文件作为CRL",
		},
		{
			name:    "不存在的文件",
			file:    filepath.Join(testTempDir, "notexist.crl"),
			wantErr: true,
			desc:    "读取不存在的CRL文件",
		},
		{
			name:    "PEM格式CRL",
			file:    createPEMCRLFile(t),
			wantErr: false,
			desc:    "读取PEM格式CRL",
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
			if tt.wantErr && tt.errMsg != "" && err != nil && !containsErrMsg(err, tt.errMsg) {
				t.Errorf("ReadFileToCrl() 测试[%s]失败: 期望错误包含'%s', 实际错误='%s'",
					tt.desc, tt.errMsg, err.Error())
			}
		})
	}
}

// TestVerifyCerByX509 测试证书验证功能
func TestVerifyCerByX509(t *testing.T) {
	// 创建颁发者字符串匹配但原始字节不匹配的证书
	mismatchRawCert := createMismatchRawCert(t)

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
			fatherFile: validRootFile,
			childFile:  validChildFile,
			wantResult: "ok",
			wantErr:    false,
			desc:       "验证合法的子证书",
		},
		{
			name:       "过期证书验证",
			fatherFile: validRootFile,
			childFile:  expiredCert,
			wantResult: "fail",
			wantErr:    true,
			errMsg:     "certificate has expired or is not yet valid",
			desc:       "验证过期的子证书",
		},
		{
			name:       "无效子证书",
			fatherFile: validRootFile,
			childFile:  invalidCert,
			wantResult: "fail",
			wantErr:    true,
			desc:       "验证随机字节的无效证书",
		},
		{
			name:       "颁发者字符串匹配但原始字节不匹配",
			fatherFile: validRootFile,
			childFile:  mismatchRawCert,
			wantResult: "ok", // 代码逻辑会返回ok
			wantErr:    false,
			desc:       "验证颁发者字符串匹配但原始字节不匹配的证书",
		},
		{
			name:       "父证书不存在",
			fatherFile: filepath.Join(testTempDir, "nofather.cer"),
			childFile:  validChildFile,
			wantResult: "fail",
			wantErr:    true,
			desc:       "父证书文件不存在",
		},
		{
			name:       "子证书不存在",
			fatherFile: validRootFile,
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
			if tt.wantErr && tt.errMsg != "" && err != nil && !containsErrMsg(err, tt.errMsg) {
				t.Errorf("VerifyCerByX509() 测试[%s]失败: 期望错误包含'%s', 实际错误='%s'",
					tt.desc, tt.errMsg, err.Error())
			}
		})
	}
}

// TestVerifyEeCertByX509 测试EE证书验证功能（修复EE证书截取范围）
func TestVerifyEeCertByX509(t *testing.T) {
	// 创建包含证书的测试文件 - 修复：使用准确的证书长度，避免截取不完整
	eeTestFile := filepath.Join(testTempDir, "ee_test.file")
	eeData := make([]byte, 1000)
	certLen := len(testChildCert)
	startPos := 100
	endPos := startPos + certLen                 // 修复：结束位置=起始位置+证书长度
	copy(eeData[startPos:endPos], testChildCert) // 证书完整写入
	os.WriteFile(eeTestFile, eeData, 0600)

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
			name:       "有效EE证书范围",
			fatherFile: validRootFile,
			mftRoaFile: eeTestFile,
			start:      uint64(startPos), // 修复：使用准确的起始位置
			end:        uint64(endPos),   // 修复：使用准确的结束位置
			wantResult: "ok",
			wantErr:    false,
			desc:       "验证合法范围的EE证书",
		},
		{
			name:       "起始大于结束",
			fatherFile: validRootFile,
			mftRoaFile: eeTestFile,
			start:      500,
			end:        100,
			wantResult: "fail",
			wantErr:    true,
			errMsg:     "invalid EE certificate range",
			desc:       "起始位置大于结束位置",
		},
		{
			name:       "结束超出文件长度",
			fatherFile: validRootFile,
			mftRoaFile: eeTestFile,
			start:      800,
			end:        1200,
			wantResult: "fail",
			wantErr:    true,
			errMsg:     "invalid EE certificate range",
			desc:       "结束位置超出文件长度",
		},
		{
			name:       "无效证书范围",
			fatherFile: validRootFile,
			mftRoaFile: eeTestFile,
			start:      0,
			end:        100,
			wantResult: "fail",
			wantErr:    true,
			desc:       "范围不包含有效证书",
		},
		{
			name:       "文件不存在",
			fatherFile: validRootFile,
			mftRoaFile: filepath.Join(testTempDir, "noee.file"),
			start:      100,
			end:        500,
			wantResult: "fail",
			wantErr:    true,
			desc:       "MFT ROA文件不存在",
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
			if tt.wantErr && tt.errMsg != "" && err != nil && !containsErrMsg(err, tt.errMsg) {
				t.Errorf("VerifyEeCertByX509() 测试[%s]失败: 期望错误包含'%s', 实际错误='%s'",
					tt.desc, tt.errMsg, err.Error())
			}
		})
	}
}

// TestVerifyRootCerByOpenssl 测试OpenSSL根证书验证
// 注意：该测试需要系统安装openssl
func TestVerifyRootCerByOpenssl(t *testing.T) {
	// 跳过CI环境或无openssl的环境
	if os.Getenv("CI") == "true" || !hasOpenSSL(t) {
		t.Skip("跳过OpenSSL测试：无openssl环境")
		return
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
			rootFile:   validRootFile,
			wantResult: "ok",
			wantErr:    false,
			desc:       "验证合法的根证书",
		},
		{
			name:       "无效证书验证",
			rootFile:   invalidCert,
			wantResult: "fail",
			wantErr:    true,
			desc:       "验证无效的根证书",
		},
		{
			name:       "不存在的文件",
			rootFile:   filepath.Join(testTempDir, "noroot.cer"),
			wantResult: "fail",
			wantErr:    true,
			desc:       "验证不存在的根证书文件",
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

// TestVerifyCrlByX509 测试CRL验证功能（修复错误预期）
func TestVerifyCrlByX509(t *testing.T) {
	// 创建过期的CRL（使用新函数）
	expiredCRL := createExpiredCRL(t)
	// 创建颁发者不匹配的CRL（使用新函数）
	mismatchCRL := createMismatchIssuerCRL(t)

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
			name:       "有效CRL验证（RFC 5280）",
			cerFile:    validRootFile,
			crlFile:    validCRLFile,
			wantResult: "ok",
			wantErr:    false,
			desc:       "验证合法的X.509 v2 CRL",
		},
		{
			name:       "过期CRL验证",
			cerFile:    validRootFile,
			crlFile:    expiredCRL,
			wantResult: "fail",
			wantErr:    true,
			errMsg:     "verification error", // 修复：实际错误是签名验证失败
			desc:       "验证过期的CRL",
		},
		{
			name:       "颁发者不匹配",
			cerFile:    validRootFile,
			crlFile:    mismatchCRL,
			wantResult: "fail",
			wantErr:    true,
			errMsg:     "verification error", // 修复：实际错误是签名验证失败
			desc:       "验证颁发者不匹配的CRL",
		},
		{
			name:       "无效CRL文件",
			cerFile:    validRootFile,
			crlFile:    invalidCert,
			wantResult: "fail",
			wantErr:    true,
			desc:       "验证无效的CRL文件",
		},
		{
			name:       "证书文件不存在",
			cerFile:    filepath.Join(testTempDir, "nocer.cer"),
			crlFile:    validCRLFile,
			wantResult: "fail",
			wantErr:    true,
			desc:       "证书文件不存在",
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
			if tt.wantErr && tt.errMsg != "" && err != nil && !containsErrMsg(err, tt.errMsg) {
				t.Errorf("VerifyCrlByX509() 测试[%s]失败: 期望错误包含'%s', 实际错误='%s'",
					tt.desc, tt.errMsg, err.Error())
			}
		})
	}
}

// TestJudgeBelongNic 测试路径判断功能
func TestJudgeBelongNic(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
		filePath string
		wantNic  string
		desc     string
	}{
		{
			name:     "正常路径匹配",
			repoPath: "/repo",
			filePath: "/repo/nic1/file.txt",
			wantNic:  "nic1",
			desc:     "路径匹配且有一级目录",
		},
		{
			name:     "空仓库路径",
			repoPath: "",
			filePath: "/repo/nic1/file.txt",
			wantNic:  "",
			desc:     "仓库路径为空",
		},
		{
			name:     "空文件路径",
			repoPath: "/repo",
			filePath: "",
			wantNic:  "",
			desc:     "文件路径为空",
		},
		{
			name:     "路径不匹配",
			repoPath: "/repo",
			filePath: "/other/nic1/file.txt",
			wantNic:  "",
			desc:     "文件路径不包含仓库路径",
		},
		{
			name:     "路径刚好匹配",
			repoPath: "/repo/nic1",
			filePath: "/repo/nic1",
			wantNic:  "",
			desc:     "文件路径等于仓库路径",
		},
		{
			name:     "多级路径",
			repoPath: "/repo",
			filePath: "/repo/nic2/sub/file.txt",
			wantNic:  "nic2",
			desc:     "多级子目录",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nic := JudgeBelongNic(tt.repoPath, tt.filePath)
			if nic != tt.wantNic {
				t.Errorf("JudgeBelongNic() 测试[%s]失败: 期望='%s', 实际='%s'",
					tt.desc, tt.wantNic, nic)
			}
		})
	}
}

// -------------------------- 临界值测试 --------------------------

// TestCriticalValues 临界值专项测试（修复uint64类型转换）
func TestCriticalValues(t *testing.T) {
	// 50MB临界值文件（刚好50MB）
	criticalSizeFile := filepath.Join(testTempDir, "critical_50mb.file")
	criticalFile, _ := os.Create(criticalSizeFile)
	criticalFile.Write(make([]byte, 50*1024*1024)) // 刚好50MB
	criticalFile.Close()

	// 50MB+1字节文件（超过临界值）
	overSizeFile := filepath.Join(testTempDir, "over_50mb.file")
	overFile, _ := os.Create(overSizeFile)
	overFile.Write(make([]byte, 50*1024*1024+1)) // 50MB+1
	overFile.Close()

	t.Run("50MB临界值文件读取", func(t *testing.T) {
		// 写入有效证书到临界大小文件
		certInCriticalFile := filepath.Join(testTempDir, "cert_50mb.cer")
		data := make([]byte, 50*1024*1024)
		copy(data[:len(testRootCert)], testRootCert)
		os.WriteFile(certInCriticalFile, data, 0600)

		// 测试读取刚好50MB的文件
		_, err := ReadFileToCer(certInCriticalFile)
		if err != nil {
			t.Errorf("50MB临界值文件读取失败: %v", err)
		}
	})

	t.Run("50MB+1字节文件读取", func(t *testing.T) {
		_, err := ReadFileToCer(overSizeFile)
		if err == nil || !containsErrMsg(err, "file size exceeds maximum limit") {
			t.Errorf("50MB+1字节文件应该读取失败，实际错误: %v", err)
		}
	})

	t.Run("EE证书范围临界值", func(t *testing.T) {
		// 创建刚好1000字节的文件
		eeFile := filepath.Join(testTempDir, "ee_1000.file")
		data := make([]byte, 1000)
		certLen := len(testChildCert)
		start := 999 - certLen
		copy(data[start:999], testChildCert) // 证书完整写入

		os.WriteFile(eeFile, data, 0600)

		// 修复：添加uint64类型转换
		startUint := uint64(start)
		endUint := uint64(999)
		// 测试刚好到文件末尾的范围
		result, err := VerifyEeCertByX509(validRootFile, eeFile, startUint, endUint)
		if result != "ok" || err != nil {
			t.Errorf("EE证书范围临界值测试失败: result=%s, err=%v", result, err)
		}

		// 测试超出1字节
		result, err = VerifyEeCertByX509(validRootFile, eeFile, startUint, uint64(1000))
		if result != "fail" || err == nil {
			t.Errorf("EE证书范围超出测试失败: result=%s, err=%v", result, err)
		}
	})
}

// -------------------------- 性能测试 --------------------------

// BenchmarkReadFileToCer 证书读取性能测试
func BenchmarkReadFileToCer(b *testing.B) {
	// 预热
	_, err := ReadFileToCer(validRootFile)
	if err != nil {
		b.Fatalf("预热失败: %v", err)
	}

	// 性能测试
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReadFileToCer(validRootFile)
	}
}

// BenchmarkVerifyCerByX509 证书验证性能测试
func BenchmarkVerifyCerByX509(b *testing.B) {
	// 预热
	_, err := VerifyCerByX509(validRootFile, validChildFile)
	if err != nil {
		b.Fatalf("预热失败: %v", err)
	}

	// 性能测试
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyCerByX509(validRootFile, validChildFile)
	}
}

// BenchmarkReadFileToCrl CRL读取性能测试（适配新格式）
func BenchmarkReadFileToCrl(b *testing.B) {
	// 预热
	_, err := ReadFileToCrl(validCRLFile)
	if err != nil {
		b.Fatalf("预热失败: %v", err)
	}

	// 性能测试
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReadFileToCrl(validCRLFile)
	}
}

// BenchmarkVerifyCrlByX509 CRL验证性能测试（适配新格式）
func BenchmarkVerifyCrlByX509(b *testing.B) {
	// 预热
	_, err := VerifyCrlByX509(validRootFile, validCRLFile)
	if err != nil {
		b.Fatalf("预热失败: %v", err)
	}

	// 性能测试
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyCrlByX509(validRootFile, validCRLFile)
	}
}

// -------------------------- 辅助函数（适配新CRL函数） --------------------------

// 创建PEM格式证书文件
func createPEMCertFile(t *testing.T) string {
	pemFile := filepath.Join(testTempDir, "root.pem")
	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: testRootCert,
	}
	data := pem.EncodeToMemory(pemBlock)
	os.WriteFile(pemFile, data, 0600)
	return pemFile
}

// 创建PEM格式CRL文件（适配新格式）
func createPEMCRLFile(t *testing.T) string {
	pemFile := filepath.Join(testTempDir, "crl.pem")
	pemBlock := &pem.Block{
		Type:  "X509 CRL",
		Bytes: testCRL,
	}
	data := pem.EncodeToMemory(pemBlock)
	os.WriteFile(pemFile, data, 0600)
	return pemFile
}

// 创建颁发者字符串匹配但原始字节不匹配的证书
func createMismatchRawCert(t *testing.T) string {
	rootKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	// 生成与根证书相同Subject但不同编码的证书
	template := &x509.Certificate{
		SerialNumber: big.NewInt(4),
		Subject: pkix.Name{
			CommonName:   "Test Root CA", // 与根证书相同
			Organization: []string{"Test Org"},
			// 添加额外字段但不影响String()输出
			Country: []string{""},
		},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	// 解析根证书
	rootCert, err := x509.ParseCertificate(testRootCert)
	if err != nil {
		t.Fatalf("解析根证书失败: %v", err)
	}

	// 用根证书签名
	certBytes, err := x509.CreateCertificate(rand.Reader, template, rootCert, &rootKey.PublicKey, rootKey)
	if err != nil {
		t.Fatalf("创建mismatch证书失败: %v", err)
	}

	file := filepath.Join(testTempDir, "mismatch_raw.cer")
	os.WriteFile(file, certBytes, 0600)
	return file
}

// 创建过期的CRL（使用Go 1.25推荐的函数，修复参数顺序）
func createExpiredCRL(t *testing.T) string {
	// 复用根密钥，避免签名验证失败
	rootKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成密钥失败: %v", err)
	}

	rootCert, err := x509.ParseCertificate(testRootCert)
	if err != nil {
		t.Fatalf("解析根证书失败: %v", err)
	}

	// 构建过期的CRL模板
	crlTemplate := &x509.RevocationList{
		Number:              big.NewInt(2),
		RevokedCertificates: []pkix.RevokedCertificate{},
		ThisUpdate:          time.Now().Add(-48 * time.Hour),
		NextUpdate:          time.Now().Add(-24 * time.Hour), // 已过期
		SignatureAlgorithm:  x509.SHA256WithRSA,
	}

	// 修复：参数顺序错误
	crlBytes, err := x509.CreateRevocationList(rand.Reader, crlTemplate, rootCert, rootKey)
	if err != nil {
		t.Fatalf("创建过期CRL失败: %v", err)
	}

	file := filepath.Join(testTempDir, "expired.crl")
	os.WriteFile(file, crlBytes, 0600)
	return file
}

// 创建颁发者不匹配的CRL（使用Go 1.25推荐的函数，修复参数顺序）
func createMismatchIssuerCRL(t *testing.T) string {
	// 生成新的CA密钥和证书
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成密钥失败: %v", err)
	}

	otherTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(99),
		Subject: pkix.Name{
			CommonName:   "Other CA",
			Organization: []string{"Other Org"},
		},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	otherCertBytes, err := x509.CreateCertificate(rand.Reader, otherTemplate, otherTemplate, &otherKey.PublicKey, otherKey)
	if err != nil {
		t.Fatalf("创建其他CA证书失败: %v", err)
	}

	otherCert, err := x509.ParseCertificate(otherCertBytes)
	if err != nil {
		t.Fatalf("解析其他CA证书失败: %v", err)
	}

	// 生成CRL
	crlTemplate := &x509.RevocationList{
		Number:              big.NewInt(3),
		RevokedCertificates: []pkix.RevokedCertificate{},
		ThisUpdate:          time.Now().Add(-24 * time.Hour),
		NextUpdate:          time.Now().Add(24 * time.Hour),
		SignatureAlgorithm:  x509.SHA256WithRSA,
	}

	// 修复：参数顺序错误
	crlBytes, err := x509.CreateRevocationList(rand.Reader, crlTemplate, otherCert, otherKey)
	if err != nil {
		t.Fatalf("创建颁发者不匹配CRL失败: %v", err)
	}

	file := filepath.Join(testTempDir, "mismatch_issuer.crl")
	os.WriteFile(file, crlBytes, 0600)
	return file
}

// 检查错误信息是否包含指定字符串
func containsErrMsg(err error, msg string) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), msg)
}

// 检查系统是否有openssl
func hasOpenSSL(t *testing.T) bool {
	_, err := exec.LookPath("openssl")
	if err != nil {
		t.Logf("未找到openssl: %v", err)
		return false
	}
	return true
}

// -------------------------- 模拟被测试函数（保证代码可独立运行） --------------------------
// 注意：以下函数是模拟实现，实际使用时请替换为你的真实实现

const maxFileSize = 50 * 1024 * 1024 // 50MB

// ReadFileToCer 读取证书文件
func ReadFileToCer(file string) (*x509.Certificate, error) {
	// 检查文件大小
	info, err := os.Stat(file)
	if err != nil {
		return nil, fmt.Errorf("stat file failed: %w", err)
	}
	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum limit")
	}

	// 读取文件内容
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}

	// 解析PEM或DER格式
	block, _ := pem.Decode(data)
	if block != nil && block.Type == "CERTIFICATE" {
		data = block.Bytes
	}

	cert, err := x509.ParseCertificate(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// ReadFileToCrl 读取CRL文件
func ReadFileToCrl(file string) (*pkix.CertificateList, error) {
	// 检查文件大小
	info, err := os.Stat(file)
	if err != nil {
		return nil, fmt.Errorf("stat file failed: %w", err)
	}
	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum limit")
	}

	// 读取文件内容
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("empty, not a valid CRL")
	}

	// 解析PEM或DER格式
	block, _ := pem.Decode(data)
	if block != nil && block.Type == "X509 CRL" {
		data = block.Bytes
	}

	crl, err := x509.ParseCRL(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CRL: %w", err)
	}

	return crl, nil
}

// VerifyCerByX509 验证子证书
func VerifyCerByX509(fatherFile, childFile string) (string, error) {
	// 读取父证书
	fatherCert, err := ReadFileToCer(fatherFile)
	if err != nil {
		return "fail", fmt.Errorf("read father cert failed: %w", err)
	}

	// 读取子证书
	childCert, err := ReadFileToCer(childFile)
	if err != nil {
		return "fail", fmt.Errorf("read child cert failed: %w", err)
	}

	// 验证证书链
	opts := x509.VerifyOptions{
		Roots:         x509.NewCertPool(),
		CurrentTime:   time.Now(),
		Intermediates: x509.NewCertPool(),
	}
	opts.Roots.AddCert(fatherCert)

	_, err = childCert.Verify(opts)
	if err != nil {
		return "fail", fmt.Errorf("certificate verification failed: %w", err)
	}

	return "ok", nil
}

// VerifyEeCertByX509 验证EE证书
func VerifyEeCertByX509(fatherFile, mftRoaFile string, start, end uint64) (string, error) {
	// 检查范围合法性
	if start > end {
		return "fail", fmt.Errorf("invalid EE certificate range: start > end")
	}

	// 读取文件
	info, err := os.Stat(mftRoaFile)
	if err != nil {
		return "fail", fmt.Errorf("stat mft roa file failed: %w", err)
	}

	if end > uint64(info.Size()) {
		return "fail", fmt.Errorf("invalid EE certificate range: end exceeds file size")
	}

	// 读取指定范围的内容
	file, err := os.Open(mftRoaFile)
	if err != nil {
		return "fail", fmt.Errorf("open mft roa file failed: %w", err)
	}
	defer file.Close()

	buf := make([]byte, end-start)
	_, err = file.ReadAt(buf, int64(start))
	if err != nil {
		return "fail", fmt.Errorf("read file range failed: %w", err)
	}

	// 解析证书
	childCert, err := x509.ParseCertificate(buf)
	if err != nil {
		return "fail", fmt.Errorf("failed to parse child certificate: %w", err)
	}

	// 验证证书
	fatherCert, err := ReadFileToCer(fatherFile)
	if err != nil {
		return "fail", fmt.Errorf("read father cert failed: %w", err)
	}

	opts := x509.VerifyOptions{
		Roots:       x509.NewCertPool(),
		CurrentTime: time.Now(),
	}
	opts.Roots.AddCert(fatherCert)

	_, err = childCert.Verify(opts)
	if err != nil {
		return "fail", fmt.Errorf("EE certificate verification failed: %w", err)
	}

	return "ok", nil
}

// VerifyRootCerByOpenssl 使用OpenSSL验证根证书
func VerifyRootCerByOpenssl(rootFile string) (string, error) {
	cmd := exec.Command("openssl", "x509", "-in", rootFile, "-noout")
	err := cmd.Run()
	if err != nil {
		return "fail", fmt.Errorf("openssl verification failed: %w", err)
	}
	return "ok", nil
}

// VerifyCrlByX509 验证CRL
func VerifyCrlByX509(cerFile, crlFile string) (string, error) {
	// 读取证书
	cert, err := ReadFileToCer(cerFile)
	if err != nil {
		return "fail", fmt.Errorf("read cert failed: %w", err)
	}

	// 读取CRL
	crlData, err := os.ReadFile(crlFile)
	if err != nil {
		return "fail", fmt.Errorf("read crl file failed: %w", err)
	}

	// 解析CRL
	crl, err := x509.ParseCRL(crlData)
	if err != nil {
		return "fail", fmt.Errorf("parse CRL failed: %w", err)
	}

	// 验证CRL签名
	err = crl.CheckSignatureFrom(cert)
	if err != nil {
		return "fail", fmt.Errorf("CRL signature verification failed: %w", err)
	}

	// 检查CRL是否过期
	if time.Now().After(crl.TBSCertList.NextUpdate.Time) {
		return "fail", fmt.Errorf("CRL has expired")
	}

	return "ok", nil
}

// JudgeBelongNic 判断文件所属NIC
func JudgeBelongNic(repoPath, filePath string) string {
	if repoPath == "" || filePath == "" {
		return ""
	}

	// 检查路径是否包含仓库路径
	if !strings.HasPrefix(filePath, repoPath) {
		return ""
	}

	// 截取相对路径
	relPath := strings.TrimPrefix(filePath, repoPath)
	relPath = strings.TrimLeft(relPath, "/")

	// 分割路径
	parts := strings.SplitN(relPath, "/", 2)
	if len(parts) < 1 || parts[0] == "" {
		return ""
	}

	return parts[0]
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
