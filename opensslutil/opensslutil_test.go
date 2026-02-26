package opensslutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cpusoft/goutil/conf"
)

// ===================== 测试环境初始化（拆分T/B，避免类型转换） =====================
// setupCommonTest: 专用于单元测试的环境初始化（参数*testing.T）
func setupCommonTest(t *testing.T) (tempDir string, validCertDER, validCertPEM, invalidFile, nonExistFile string, err error) {
	t.Helper()
	// 创建临时目录
	tempDir, err = os.MkdirTemp("", "opensslutil_test")
	if err != nil {
		return "", "", "", "", "", err
	}

	// 1. 生成真实的自签名PEM证书
	validCertPEM = filepath.Join(tempDir, "valid_pem.pem")
	opensslCmd := getOpensslCmd()
	genPemCmd := exec.Command(opensslCmd, "req", "-x509", "-newkey", "rsa:2048", "-nodes",
		"-days", "1", "-out", validCertPEM, "-subj", "/CN=test.example.com")
	if output, err := genPemCmd.CombinedOutput(); err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to generate PEM cert: %v, output: %s", err, string(output))
	}

	// 2. 转换为DER格式
	validCertDER = filepath.Join(tempDir, "valid_der.cer")
	convertDerCmd := exec.Command(opensslCmd, "x509", "-in", validCertPEM, "-outform", "DER", "-out", validCertDER)
	if output, err := convertDerCmd.CombinedOutput(); err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to convert to DER cert: %v, output: %s", err, string(output))
	}

	// 3. 无效文件
	invalidFile = filepath.Join(tempDir, "invalid.txt")
	if err = os.WriteFile(invalidFile, []byte("not a certificate"), 0644); err != nil {
		return "", "", "", "", "", err
	}

	// 4. 不存在的文件路径
	nonExistFile = filepath.Join(tempDir, "non_exist.cer")

	return tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, nil
}

// setupCommonBenchmark: 专用于性能测试的环境初始化（参数*testing.B，仅改Helper，逻辑完全复用）
func setupCommonBenchmark(b *testing.B) (tempDir string, validCertDER, validCertPEM, invalidFile, nonExistFile string, err error) {
	b.Helper() // 仅把t.Helper()改成b.Helper()，其他逻辑和setupCommonTest完全一致
	tempDir, err = os.MkdirTemp("", "opensslutil_test")
	if err != nil {
		return "", "", "", "", "", err
	}

	validCertPEM = filepath.Join(tempDir, "valid_pem.pem")
	opensslCmd := getOpensslCmd()
	genPemCmd := exec.Command(opensslCmd, "req", "-x509", "-newkey", "rsa:2048", "-nodes",
		"-days", "1", "-out", validCertPEM, "-subj", "/CN=test.example.com")
	if output, err := genPemCmd.CombinedOutput(); err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to generate PEM cert: %v, output: %s", err, string(output))
	}

	validCertDER = filepath.Join(tempDir, "valid_der.cer")
	convertDerCmd := exec.Command(opensslCmd, "x509", "-in", validCertPEM, "-outform", "DER", "-out", validCertDER)
	if output, err := convertDerCmd.CombinedOutput(); err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to convert to DER cert: %v, output: %s", err, string(output))
	}

	invalidFile = filepath.Join(tempDir, "invalid.txt")
	if err = os.WriteFile(invalidFile, []byte("not a certificate"), 0644); err != nil {
		return "", "", "", "", "", err
	}

	nonExistFile = filepath.Join(tempDir, "non_exist.cer")

	return tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, nil
}

// setupTest: 单元测试初始化（调用setupCommonTest）
func setupTest(t *testing.T) (tempDir string, validCertDER, validCertPEM, invalidFile, nonExistFile string) {
	t.Helper()
	tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, err := setupCommonTest(t)
	if err != nil {
		t.Fatalf("Failed to setup test env: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})
	return tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile
}

// setupBenchmark: 性能测试初始化（调用setupCommonBenchmark，无类型转换）
func setupBenchmark(b *testing.B) (tempDir string, validCertDER, validCertPEM, invalidFile, nonExistFile string) {
	b.Helper()
	// 直接调用为*testing.B定制的setupCommonBenchmark，无类型断言，彻底解决错误
	tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, err := setupCommonBenchmark(b)
	if err != nil {
		b.Fatalf("Failed to setup benchmark env: %v", err)
	}
	b.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})
	return tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile
}

// ===================== 单元测试 - getOpensslCmd =====================
func TestGetOpensslCmd(t *testing.T) {
	_ = conf.SetString("openssl::path", "")
	cmd := getOpensslCmd()
	if cmd != "openssl" {
		t.Errorf("getOpensslCmd() = %s, want openssl", cmd)
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
			wantErr:  true,
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
			name:     "osutil.IsExists返回error（模拟不可访问路径）",
			certFile: string(os.PathSeparator) + "proc" + string(os.PathSeparator) + "1" + string(os.PathSeparator) + "fd" + string(os.PathSeparator) + "9999",
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
func BenchmarkGetResultsByOpensslX509(b *testing.B) {
	_, validCert, _, _, _ := setupBenchmark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetResultsByOpensslX509(validCert)
	}
}

func BenchmarkGetResultsByOpensslAns1(b *testing.B) {
	_, validCert, _, _, _ := setupBenchmark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetResultsByOpensslAns1(validCert)
	}
}

func BenchmarkProcessOpensslOutput(b *testing.B) {
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
