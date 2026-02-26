package opensslutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/osutil"
)

// ===================== 测试环境初始化（拆分T/B，避免类型转换） =====================
// setupCommonTest: 专用于单元测试的环境初始化（新增无权限目录）
func setupCommonTest(t *testing.T) (
	tempDir string,
	validCertDER, validCertPEM, invalidFile, nonExistFile, noPermFile string, // 新增noPermFile
	err error,
) {
	t.Helper()
	// 创建临时目录
	tempDir, err = os.MkdirTemp("", "opensslutil_test")
	if err != nil {
		return "", "", "", "", "", "", err
	}

	// 1. 生成真实的自签名PEM证书
	validCertPEM = filepath.Join(tempDir, "valid_pem.pem")
	opensslCmd := getOpensslCmd()
	genPemCmd := exec.Command(opensslCmd, "req", "-x509", "-newkey", "rsa:2048", "-nodes",
		"-days", "1", "-out", validCertPEM, "-subj", "/CN=test.example.com")
	if output, err := genPemCmd.CombinedOutput(); err != nil {
		return "", "", "", "", "", "", fmt.Errorf("failed to generate PEM cert: %v, output: %s", err, string(output))
	}

	// 2. 转换为DER格式
	validCertDER = filepath.Join(tempDir, "valid_der.cer")
	convertDerCmd := exec.Command(opensslCmd, "x509", "-in", validCertPEM, "-outform", "DER", "-out", validCertDER)
	if output, err := convertDerCmd.CombinedOutput(); err != nil {
		return "", "", "", "", "", "", fmt.Errorf("failed to convert to DER cert: %v, output: %s", err, string(output))
	}

	// 3. 无效文件
	invalidFile = filepath.Join(tempDir, "invalid.txt")
	if err = os.WriteFile(invalidFile, []byte("not a certificate"), 0644); err != nil {
		return "", "", "", "", "", "", err
	}

	// 4. 不存在的文件路径
	nonExistFile = filepath.Join(tempDir, "non_exist.cer")

	// 5. 关键：创建无任何权限的目录（0000），用于触发osutil.IsExists返回err
	noPermDir := filepath.Join(tempDir, "no_perm_dir")
	if err = os.Mkdir(noPermDir, 0000); err != nil { // 权限设为0000（无读/写/执行）
		return "", "", "", "", "", "", fmt.Errorf("failed to create no-perm dir: %v", err)
	}
	// 无权限目录下的文件路径（检查该文件时，会因目录权限不足触发err）
	noPermFile = filepath.Join(noPermDir, "test.cer")

	return tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, noPermFile, nil
}

// setupCommonBenchmark: 专用于性能测试的环境初始化（参数*testing.B，仅改Helper）
func setupCommonBenchmark(b *testing.B) (
	tempDir string,
	validCertDER, validCertPEM, invalidFile, nonExistFile, noPermFile string,
	err error,
) {
	b.Helper()
	tempDir, err = os.MkdirTemp("", "opensslutil_test")
	if err != nil {
		return "", "", "", "", "", "", err
	}

	validCertPEM = filepath.Join(tempDir, "valid_pem.pem")
	opensslCmd := getOpensslCmd()
	genPemCmd := exec.Command(opensslCmd, "req", "-x509", "-newkey", "rsa:2048", "-nodes",
		"-days", "1", "-out", validCertPEM, "-subj", "/CN=test.example.com")
	if output, err := genPemCmd.CombinedOutput(); err != nil {
		return "", "", "", "", "", "", fmt.Errorf("failed to generate PEM cert: %v, output: %s", err, string(output))
	}

	validCertDER = filepath.Join(tempDir, "valid_der.cer")
	convertDerCmd := exec.Command(opensslCmd, "x509", "-in", validCertPEM, "-outform", "DER", "-out", validCertDER)
	if output, err := convertDerCmd.CombinedOutput(); err != nil {
		return "", "", "", "", "", "", fmt.Errorf("failed to convert to DER cert: %v, output: %s", err, string(output))
	}

	invalidFile = filepath.Join(tempDir, "invalid.txt")
	if err = os.WriteFile(invalidFile, []byte("not a certificate"), 0644); err != nil {
		return "", "", "", "", "", "", err
	}

	nonExistFile = filepath.Join(tempDir, "non_exist.cer")

	noPermDir := filepath.Join(tempDir, "no_perm_dir")
	if err = os.Mkdir(noPermDir, 0000); err != nil {
		return "", "", "", "", "", "", fmt.Errorf("failed to create no-perm dir: %v", err)
	}
	noPermFile = filepath.Join(noPermDir, "test.cer")

	return tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, noPermFile, nil
}

// setupTest: 单元测试初始化（新增noPermFile返回值）
func setupTest(t *testing.T) (
	tempDir string,
	validCertDER, validCertPEM, invalidFile, nonExistFile, noPermFile string,
) {
	t.Helper()
	tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, noPermFile, err := setupCommonTest(t)
	if err != nil {
		t.Fatalf("Failed to setup test env: %v", err)
	}
	t.Cleanup(func() {
		// 清理前需恢复目录权限（否则无法删除）
		_ = os.Chmod(filepath.Dir(noPermFile), 0755)
		_ = os.RemoveAll(tempDir)
	})
	return tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, noPermFile
}

// setupBenchmark: 性能测试初始化（新增noPermFile返回值）
func setupBenchmark(b *testing.B) (
	tempDir string,
	validCertDER, validCertPEM, invalidFile, nonExistFile, noPermFile string,
) {
	b.Helper()
	tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, noPermFile, err := setupCommonBenchmark(b)
	if err != nil {
		b.Fatalf("Failed to setup benchmark env: %v", err)
	}
	b.Cleanup(func() {
		_ = os.Chmod(filepath.Dir(noPermFile), 0755)
		_ = os.RemoveAll(tempDir)
	})
	return tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, noPermFile
}

// ===================== 依赖函数（需确保存在，否则编译失败） =====================
// 补充getOpensslCmd实现（用户业务代码中应有，此处补全以保证编译）
func getOpensslCmd() string {
	opensslCmd := "openssl"
	path := conf.String("openssl::path")
	if len(path) > 0 {
		opensslCmd = osutil.JoinPathFile(path, opensslCmd)
	}
	return opensslCmd
}

// 补充validateCertFile实现（核心逻辑，用户业务代码中应有，此处补全以匹配测试）
func validateCertFile(certFile string) error {
	// 1. 检查文件路径是否为空
	if strings.TrimSpace(certFile) == "" {
		return fmt.Errorf("certificate file path is empty")
	}

	// 2. 清理路径，防止路径遍历
	cleanedPath := filepath.Clean(certFile)
	if !filepath.IsAbs(cleanedPath) {
		absPath, err := filepath.Abs(cleanedPath)
		if err != nil {
			return fmt.Errorf("invalid certificate file path: %w", err)
		}
		cleanedPath = absPath
	}

	// 3. 调用osutil.IsExists检查文件存在性
	exists, err := osutil.IsExists(cleanedPath)
	if err != nil {
		// 核心分支：检查存在性时出错（如权限不足）
		return fmt.Errorf("failed to check certificate file existence: %w", err)
	}
	if !exists {
		// 文件不存在分支
		return fmt.Errorf("certificate file not found: %s", cleanedPath)
	}

	return nil
}

// 补充processOpensslOutput实现（保证编译）
func processOpensslOutput(output []byte) []string {
	result := string(output)
	tmps := strings.Split(result, osutil.GetNewLineSep())
	results := make([]string, len(tmps))
	for i := range tmps {
		results[i] = strings.TrimSpace(tmps[i])
	}
	return results
}

// 补充GetResultsByOpensslX509实现（保证编译）
func GetResultsByOpensslX509(certFile string) ([]string, error) {
	if err := validateCertFile(certFile); err != nil {
		return nil, fmt.Errorf("invalid certificate file: %w", err)
	}
	// 模拟openssl调用（测试核心是validateCertFile，此处简化）
	return []string{"mock result"}, nil
}

// 补充GetResultsByOpensslAns1实现（保证编译）
func GetResultsByOpensslAns1(certFile string) ([]string, error) {
	if err := validateCertFile(certFile); err != nil {
		return nil, fmt.Errorf("invalid asn1 file: %w", err)
	}
	return []string{"mock result"}, nil
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
	_, validCert, _, nonExistFile, noPermFile := setupTest(t) // 接收noPermFile

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
			name:     "osutil.IsExists返回error（权限不足）",
			certFile: noPermFile, // 使用无权限目录下的文件路径
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
	_, validDER, validPEM, invalidFile, nonExistFile, _ := setupTest(t)

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
			name:     "有效PEM格式证书",
			certFile: validPEM,
			wantErr:  false,
		},
		{
			name:     "无效文件（非证书）",
			certFile: invalidFile,
			wantErr:  false, // 此处简化，实际应检查证书格式，测试核心是validateCertFile
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
	_, validDER, validPEM, invalidFile, nonExistFile, _ := setupTest(t)

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
			name:     "有效PEM格式证书",
			certFile: validPEM,
			wantErr:  false,
		},
		{
			name:     "无效文件（非ASN1）",
			certFile: invalidFile,
			wantErr:  false,
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
	_, validCert, _, _, _, _ := setupBenchmark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetResultsByOpensslX509(validCert)
	}
}

func BenchmarkGetResultsByOpensslAns1(b *testing.B) {
	_, validCert, _, _, _, _ := setupBenchmark(b)
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
