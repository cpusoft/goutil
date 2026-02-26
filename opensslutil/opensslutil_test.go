package opensslutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cpusoft/goutil/conf"
)

// ===================== 重构测试前置函数（兼容Test和Benchmark） =====================
// setupCommon: 通用的测试环境初始化（不依赖*testing.T/*testing.B）
// 返回：临时目录、有效DER证书、有效PEM证书、无效文件、不存在的文件路径，以及错误
func setupCommon() (tempDir string, validCertDER, validCertPEM, invalidFile, nonExistFile string, err error) {
	// 创建临时目录
	tempDir, err = os.MkdirTemp("", "opensslutil_test")
	if err != nil {
		return "", "", "", "", "", err
	}

	// 1. 有效 DER 格式证书（简化的DER头部，仅用于测试）
	validCertDER = filepath.Join(tempDir, "valid_der.cer")
	derContent := []byte{0x30, 0x82, 0x01, 0x22, 0x30, 0x82, 0x01, 0x0a}
	if err = os.WriteFile(validCertDER, derContent, 0644); err != nil {
		return "", "", "", "", "", err
	}

	// 2. 有效 PEM 格式证书（测试用示例内容）
	validCertPEM = filepath.Join(tempDir, "valid_pem.pem")
	pemContent := []byte("-----BEGIN CERTIFICATE-----\nMIICUTCCAfugAwIBAgIBADANBgkqhkiG9w0BAQQFADBXMQswCQYDVQQGEwJDTjEL\nMAkGA1UECBMCUE4xCzAJBgNVBAcTAkNOMQswCQYDVQQKEwJPTjELMAkGA1UECxMC\nVU4xCzAJBgNVBAMTAkRLMA0GCSqGSIb3DQEBBQUAA4GBADsw5rM5JnO4l2vXlLzZJ\n8h6nHf+6zK7L5fQVQ8bR7XZ8eX9y5t7e8vX7+6t5Z8e7X9y5t7e8vX7+6t5Z8e7X9\ny5t7e8vX7+6t5Z8e7X9y5t7e8vX7+6t5Z8e7X9\n-----END CERTIFICATE-----")
	if err = os.WriteFile(validCertPEM, pemContent, 0644); err != nil {
		return "", "", "", "", "", err
	}

	// 3. 无效文件（非证书）
	invalidFile = filepath.Join(tempDir, "invalid.txt")
	if err = os.WriteFile(invalidFile, []byte("not a certificate"), 0644); err != nil {
		return "", "", "", "", "", err
	}

	// 4. 不存在的文件路径
	nonExistFile = filepath.Join(tempDir, "non_exist.cer")

	return tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, nil
}

// setupTest: 专用于单元测试的环境初始化（带T的清理逻辑）
func setupTest(t *testing.T) (tempDir string, validCertDER, validCertPEM, invalidFile, nonExistFile string) {
	t.Helper() // 标记为测试辅助函数，错误定位更准确
	tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, err := setupCommon()
	if err != nil {
		t.Fatalf("Failed to setup test env: %v", err)
	}
	// 测试结束清理临时目录
	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})
	return tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile
}

// setupBenchmark: 专用于性能测试的环境初始化（带B的清理逻辑）
func setupBenchmark(b *testing.B) (tempDir string, validCertDER, validCertPEM, invalidFile, nonExistFile string) {
	b.Helper()
	tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, err := setupCommon()
	if err != nil {
		b.Fatalf("Failed to setup benchmark env: %v", err)
	}
	// 性能测试结束清理临时目录
	b.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})
	return tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile
}

// ===================== 单元测试 - getOpensslCmd =====================
func TestGetOpensslCmd(t *testing.T) {
	// 场景1：未配置openssl路径，返回默认"openssl"
	_ = conf.SetString("openssl::path", "")
	cmd := getOpensslCmd()
	if cmd != "openssl" {
		t.Errorf("getOpensslCmd() = %s, want openssl", cmd)
	}

	// 场景2：配置了openssl路径，拼接路径
	tempDir := t.TempDir()
	_ = conf.SetString("openssl::path", tempDir)
	expected := filepath.Join(tempDir, "openssl")
	cmd = getOpensslCmd()
	if cmd != expected {
		t.Errorf("getOpensslCmd() = %s, want %s", cmd, expected)
	}
}

// ===================== 单元测试 - validateCertFile =====================
func TestValidateCertFile(t *testing.T) {
	_, validCert, _, _, nonExistFile := setupTest(t)

	tests := []struct {
		name     string
		certFile string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "空路径",
			certFile: "",
			wantErr:  true,
			errMsg:   "certificate file path is empty",
		},
		{
			name:     "路径遍历（../../）",
			certFile: "../../etc/passwd",
			wantErr:  true, // 清理后文件不存在，最终返回不存在错误
			errMsg:   "certificate file not found",
		},
		{
			name:     "相对路径转绝对路径",
			certFile: "./test.cer",
			wantErr:  true,
			errMsg:   "certificate file not found",
		},
		{
			name:     "文件不存在",
			certFile: nonExistFile,
			wantErr:  true,
			errMsg:   "certificate file not found",
		},
		{
			name:     "文件存在（合法路径）",
			certFile: validCert,
			wantErr:  false,
		},
		{
			name:     "osutil.IsExists返回error（模拟权限不足）",
			certFile: "/root/non_exist.cer", // 无权限访问的路径
			wantErr:  true,
			errMsg:   "failed to check certificate file existence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCertFile(tt.certFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCertFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 替换osutil.Contains为标准库strings.Contains
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateCertFile() error msg = %v, want contains %s", err, tt.errMsg)
			}
		})
	}
}

// ===================== 单元测试 - processOpensslOutput =====================
func TestProcessOpensslOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name:   "空输出",
			output: []byte(""),
			want:   []string{""},
		},
		{
			name:   "单行输出（带空格）",
			output: []byte("  Certificate:  "),
			want:   []string{"Certificate:"},
		},
		{
			name:   "多行输出（带换行和空格）",
			output: []byte("  Version: 1 (0x0)\n  Serial Number:\n       123456  "),
			want:   []string{"Version: 1 (0x0)", "Serial Number:", "123456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processOpensslOutput(tt.output)
			if len(got) != len(tt.want) {
				t.Errorf("processOpensslOutput() len = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("processOpensslOutput()[%d] = %s, want %s", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// ===================== 单元测试 - GetResultsByOpensslX509 =====================
func TestGetResultsByOpensslX509(t *testing.T) {
	_, validDER, validPEM, invalidFile, nonExistFile := setupTest(t)

	tests := []struct {
		name     string
		certFile string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "文件不存在",
			certFile: nonExistFile,
			wantErr:  true,
			errMsg:   "invalid certificate file: certificate file not found",
		},
		{
			name:     "有效DER格式证书",
			certFile: validDER,
			wantErr:  false,
		},
		{
			name:     "有效PEM格式证书（先试DER失败，再试PEM成功）",
			certFile: validPEM,
			wantErr:  false,
		},
		{
			name:     "无效文件（非证书）",
			certFile: invalidFile,
			wantErr:  true,
			errMsg:   "fail to parse x509 certificate: invalid format or corrupted file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetResultsByOpensslX509(tt.certFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetResultsByOpensslX509() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 替换osutil.Contains为标准库strings.Contains
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("GetResultsByOpensslX509() error msg = %v, want contains %s", err, tt.errMsg)
			}
			if !tt.wantErr && len(got) == 0 {
				t.Error("GetResultsByOpensslX509() returned empty results for valid cert")
			}
		})
	}
}

// ===================== 单元测试 - GetResultsByOpensslAns1 =====================
func TestGetResultsByOpensslAns1(t *testing.T) {
	_, validDER, validPEM, invalidFile, nonExistFile := setupTest(t)

	tests := []struct {
		name     string
		certFile string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "文件不存在",
			certFile: nonExistFile,
			wantErr:  true,
			errMsg:   "invalid asn1 file: certificate file not found",
		},
		{
			name:     "有效DER格式证书",
			certFile: validDER,
			wantErr:  false,
		},
		{
			name:     "有效PEM格式证书（先试DER失败，再试PEM成功）",
			certFile: validPEM,
			wantErr:  false,
		},
		{
			name:     "无效文件（非ASN1）",
			certFile: invalidFile,
			wantErr:  true,
			errMsg:   "fail to parse asn1 format: invalid format or corrupted file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetResultsByOpensslAns1(tt.certFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetResultsByOpensslAns1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 替换osutil.Contains为标准库strings.Contains
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("GetResultsByOpensslAns1() error msg = %v, want contains %s", err, tt.errMsg)
			}
			if !tt.wantErr && len(got) == 0 {
				t.Error("GetResultsByOpensslAns1() returned empty results for valid cert")
			}
		})
	}
}

// ===================== 性能测试（Benchmark） =====================
// 测试GetResultsByOpensslX509性能
func BenchmarkGetResultsByOpensslX509(b *testing.B) {
	_, validCert, _, _, _ := setupBenchmark(b)
	b.ResetTimer() // 重置计时器，排除setup耗时
	for i := 0; i < b.N; i++ {
		_, _ = GetResultsByOpensslX509(validCert)
	}
}

// 测试GetResultsByOpensslAns1性能
func BenchmarkGetResultsByOpensslAns1(b *testing.B) {
	_, validCert, _, _, _ := setupBenchmark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetResultsByOpensslAns1(validCert)
	}
}

// 测试processOpensslOutput性能（高频处理场景）
func BenchmarkProcessOpensslOutput(b *testing.B) {
	// 模拟真实的OpenSSL输出（100行）
	var output []byte
	for i := 0; i < 100; i++ {
		output = append(output, []byte("  Version: 1 (0x0)\n")...)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processOpensslOutput(output)
	}
}

////////////////////////////////////////////////////////
/*
func TestOpenssl(t *testing.T) {
	cmd := exec.Command("openssl", "version")
	ldLibraryPath := `/home/openssl/openssl/lib64`
	path := `/home/openssl/openssl/bin:$PATH`
	if len(ldLibraryPath) > 0 && len(path) > 0 {
		cmd.Env = append(os.Environ(), "LD_LIBRARY_PATH="+ldLibraryPath)
		cmd.Env = append(os.Environ(), "PATH="+path)
		belogs.Debug("GetResultsByOpensslX509(): ldLibraryPath:", ldLibraryPath, "  path:", path)
	}
	output, err := cmd.CombinedOutput()
	fmt.Println(output, err)
}
*/
