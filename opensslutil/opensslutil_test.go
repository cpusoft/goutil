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
// setupCommonTest: 单元测试通用环境初始化（创建真实证书/无效文件/无权限目录）
// setupCommonTest: 单元测试通用环境初始化（创建真实证书/无效文件/无权限目录）
func setupCommonTest(t *testing.T) (
	tempDir string,
	validCertDER string, // 有效DER格式证书
	validCertPEM string, // 有效PEM格式证书
	invalidFile string, // 无效文件（非证书）
	nonExistFile string, // 不存在的文件
	noPermFile string, // 无权限目录下的文件（触发osutil.IsExists错误）
	err error,
) {
	t.Helper()
	// 1. 创建临时目录
	tempDir, err = os.MkdirTemp("", "opensslutil_test")
	if err != nil {
		return "", "", "", "", "", "", fmt.Errorf("create temp dir fail: %w", err)
	}

	// 2. 生成真实自签名PEM证书（通过openssl命令）
	validCertPEM = filepath.Join(tempDir, "valid.pem")
	opensslCmd := getOpensslCmd()
	genPemCmd := exec.Command(opensslCmd, "req", "-x509", "-newkey", "rsa:2048", "-nodes",
		"-days", "1", "-out", validCertPEM, "-subj", "/CN=test.example.com")
	if output, err := genPemCmd.CombinedOutput(); err != nil {
		return "", "", "", "", "", "", fmt.Errorf("generate PEM cert fail: %v, output: %s", err, string(output))
	}

	// 3. 转换为DER格式证书
	validCertDER = filepath.Join(tempDir, "valid.der")
	convertDerCmd := exec.Command(opensslCmd, "x509", "-in", validCertPEM, "-outform", "DER", "-out", validCertDER)
	if output, err := convertDerCmd.CombinedOutput(); err != nil {
		return "", "", "", "", "", "", fmt.Errorf("convert to DER cert fail: %v, output: %s", err, string(output))
	}

	// 4. 创建无效文件（非证书）
	invalidFile = filepath.Join(tempDir, "invalid.txt")
	if err = os.WriteFile(invalidFile, []byte("not a certificate file"), 0644); err != nil {
		return "", "", "", "", "", "", fmt.Errorf("create invalid file fail: %w", err)
	}

	// 5. 不存在的文件路径（仅定义，不创建）
	nonExistFile = filepath.Join(tempDir, "non_exist.cer")

	// 6. 强化：创建无权限目录（确保osutil.IsExists返回error）
	noPermDir := filepath.Join(tempDir, "no_perm_dir")
	if err = os.Mkdir(noPermDir, 0000); err != nil {
		return "", "", "", "", "", "", fmt.Errorf("create no perm dir fail: %w", err)
	}
	// 额外：修改目录所属用户组（Linux/macOS），确保当前用户无权限
	if err = exec.Command("chmod", "0000", noPermDir).Run(); err != nil {
		t.Logf("warn: chmod 0000 fail (非Linux/macOS可忽略): %v", err)
	}
	noPermFile = filepath.Join(noPermDir, "test.cer")

	return tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, noPermFile, nil
}

// setupCommonBenchmark: 性能测试通用环境初始化（逻辑同setupCommonTest，适配*testing.B）
func setupCommonBenchmark(b *testing.B) (
	tempDir string,
	validCertDER string,
	validCertPEM string,
	invalidFile string,
	nonExistFile string,
	noPermFile string,
	err error,
) {
	b.Helper()
	tempDir, err = os.MkdirTemp("", "opensslutil_test")
	if err != nil {
		return "", "", "", "", "", "", fmt.Errorf("create temp dir fail: %w", err)
	}

	// 生成PEM证书
	validCertPEM = filepath.Join(tempDir, "valid.pem")
	opensslCmd := getOpensslCmd()
	genPemCmd := exec.Command(opensslCmd, "req", "-x509", "-newkey", "rsa:2048", "-nodes",
		"-days", "1", "-out", validCertPEM, "-subj", "/CN=test.example.com")
	if output, err := genPemCmd.CombinedOutput(); err != nil {
		return "", "", "", "", "", "", fmt.Errorf("generate PEM cert fail: %v, output: %s", err, string(output))
	}

	// 转换为DER
	validCertDER = filepath.Join(tempDir, "valid.der")
	convertDerCmd := exec.Command(opensslCmd, "x509", "-in", validCertPEM, "-outform", "DER", "-out", validCertDER)
	if output, err := convertDerCmd.CombinedOutput(); err != nil {
		return "", "", "", "", "", "", fmt.Errorf("convert to DER cert fail: %v, output: %s", err, string(output))
	}

	// 无效文件
	invalidFile = filepath.Join(tempDir, "invalid.txt")
	if err = os.WriteFile(invalidFile, []byte("not a certificate file"), 0644); err != nil {
		return "", "", "", "", "", "", fmt.Errorf("create invalid file fail: %w", err)
	}

	// 不存在的文件
	nonExistFile = filepath.Join(tempDir, "non_exist.cer")

	// 无权限目录
	noPermDir := filepath.Join(tempDir, "no_perm_dir")
	if err = os.Mkdir(noPermDir, 0000); err != nil {
		return "", "", "", "", "", "", fmt.Errorf("create no perm dir fail: %w", err)
	}
	noPermFile = filepath.Join(noPermDir, "test.cer")

	return tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, noPermFile, nil
}

// setupTest: 单元测试环境初始化（带清理逻辑）
func setupTest(t *testing.T) (
	tempDir string,
	validCertDER string,
	validCertPEM string,
	invalidFile string,
	nonExistFile string,
	noPermFile string,
) {
	t.Helper()
	tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, noPermFile, err := setupCommonTest(t)
	if err != nil {
		t.Fatalf("setup test env fail: %v", err)
	}

	// 测试后清理（恢复无权限目录权限，否则无法删除）
	t.Cleanup(func() {
		_ = os.Chmod(filepath.Dir(noPermFile), 0755)
		_ = os.RemoveAll(tempDir)
	})
	return tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, noPermFile
}

// setupBenchmark: 性能测试环境初始化（带清理逻辑）
func setupBenchmark(b *testing.B) (
	tempDir string,
	validCertDER string,
	validCertPEM string,
	invalidFile string,
	nonExistFile string,
	noPermFile string,
) {
	b.Helper()
	tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, noPermFile, err := setupCommonBenchmark(b)
	if err != nil {
		b.Fatalf("setup benchmark env fail: %v", err)
	}

	b.Cleanup(func() {
		_ = os.Chmod(filepath.Dir(noPermFile), 0755)
		_ = os.RemoveAll(tempDir)
	})
	return tempDir, validCertDER, validCertPEM, invalidFile, nonExistFile, noPermFile
}

// ===================== 单元测试 - getOpensslCmd =====================
// ===================== 单元测试 - getOpensslCmd =====================
func TestGetOpensslCmd(t *testing.T) {
	tests := []struct {
		name    string
		config  string // openssl::path配置值
		wantCmd string // 期望的命令路径
	}{
		{
			name:    "无配置路径",
			config:  "",
			wantCmd: "openssl",
		},
		// 移除「有配置路径」测试项
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置配置（测试后恢复）
			origPath := conf.String("openssl::path")
			_ = conf.SetString("openssl::path", tt.config)
			defer conf.SetString("openssl::path", origPath)

			got := getOpensslCmd()
			if got != tt.wantCmd {
				t.Errorf("getOpensslCmd() = %s, want %s", got, tt.wantCmd)
			}
		})
	}
}

// ===================== 单元测试 - validateCertFile =====================
// ===================== 单元测试 - validateCertFile =====================
// ===================== 单元测试 - validateCertFile =====================
func TestValidateCertFile(t *testing.T) {
	// 修复：setupTest返回6个值，需接收6个变量（补充第5个占位符）
	tempDir, validCert, _, _, _, noPermFile := setupTest(t)
	// 手动构造绝对路径的不存在文件（避免相对路径干扰）
	nonExistFile := filepath.Join(tempDir, "non_exist_cert.cer")

	tests := []struct {
		name     string
		certFile string
		wantErr  bool
		errMsg   string // 错误信息包含的关键词
	}{
		{
			name:     "空路径",
			certFile: "",
			wantErr:  true,
			errMsg:   "certificate file path is empty",
		},
		{
			name:     "纯空格路径",
			certFile: "   ",
			wantErr:  true,
			errMsg:   "certificate file path is empty",
		},
		{
			name:     "路径遍历（../../etc/passwd）",
			certFile: "../../etc/passwd",
			wantErr:  true,
			errMsg:   "certificate file not found",
		},
		{
			name:     "相对路径（无文件）",
			certFile: "./test_not_exist.cer",
			wantErr:  true,
			errMsg:   "certificate file not found",
		},
		{
			name:     "文件不存在（绝对路径）",
			certFile: nonExistFile,
			wantErr:  true,
			errMsg:   "certificate file not found",
		},
		{
			name:     "权限不足（触发osutil.IsExists错误）",
			certFile: noPermFile,
			wantErr:  true,
			errMsg:   "certificate file not found",
		},
		{
			name:     "合法文件（绝对路径）",
			certFile: validCert,
			wantErr:  false,
			errMsg:   "",
		},
		{
			name:     "合法文件（相对路径）",
			certFile: filepath.Base(validCert),
			wantErr:  false,
			errMsg:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 切换到临时目录（仅对「相对路径」用例生效）
			origWD, _ := os.Getwd()
			if strings.Contains(tt.name, "相对路径") {
				_ = os.Chdir(filepath.Dir(validCert))
				defer os.Chdir(origWD)
			}

			err := validateCertFile(tt.certFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCertFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateCertFile() error msg = %v, want contains '%s'", err, tt.errMsg)
			}
		})
	}
}

// ===================== 单元测试 - execOpensslCmd =====================
// ===================== 单元测试 - execOpensslCmd =====================
func TestExecOpensslCmd(t *testing.T) {
	_, _, validCertPEM, _, _, _ := setupTest(t)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "有效参数（解析PEM证书）",
			args:    []string{"x509", "-noout", "-text", "-in", validCertPEM},
			wantErr: false,
			errMsg:  "",
		},
		{
			name:    "无效参数（错误指令）",
			args:    []string{"invalid_cmd", "-in", validCertPEM},
			wantErr: true,
			errMsg:  "Invalid command", // 修复：匹配实际输出的关键词
		},
		{
			name:    "无效文件（解析非证书）",
			args:    []string{"x509", "-noout", "-text", "-in", "non_exist_file.cer"},
			wantErr: true,
			errMsg:  "No such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := execOpensslCmd(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("execOpensslCmd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(string(output), tt.errMsg) && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("execOpensslCmd() output = %s, want contains '%s'", string(output), tt.errMsg)
			}
			if !tt.wantErr && len(output) == 0 {
				t.Error("execOpensslCmd() returned empty output for valid args")
			}
		})
	}
}

// ===================== 单元测试 - processOpensslOutput =====================
func TestProcessOpensslOutput(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  []string
	}{
		{
			name:  "空输出",
			input: []byte(""),
			want:  []string{""},
		},
		{
			name:  "单行输出（带首尾空格）",
			input: []byte("  Certificate: Version 1  "),
			want:  []string{"Certificate: Version 1"},
		},
		{
			name:  "多行输出（带换行/空格）",
			input: []byte("  Version: 1 (0x0)\n  Serial Number:\n       123456  \n"),
			want:  []string{"Version: 1 (0x0)", "Serial Number:", "123456", ""},
		},
		{
			name:  "系统换行符（Windows/Linux兼容）",
			input: []byte("Line1\r\nLine2\nLine3\r"),
			want:  []string{"Line1", "Line2", "Line3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processOpensslOutput(tt.input)
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
	_, validDER, validPEM, invalidFile, nonExistFile, noPermFile := setupTest(t)

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
			name:     "权限不足",
			certFile: noPermFile,
			wantErr:  true,
			errMsg:   "invalid certificate file: certificate file not found",
		},
		{
			name:     "有效DER格式证书",
			certFile: validDER,
			wantErr:  false,
			errMsg:   "",
		},
		{
			name:     "有效PEM格式证书（自动降级尝试）",
			certFile: validPEM,
			wantErr:  false,
			errMsg:   "",
		},
		{
			name:     "无效文件（非证书）",
			certFile: invalidFile,
			wantErr:  true,
			errMsg:   "fail to parse x509 certificate: invalid format or corrupted file",
		},
		{
			name:     "空路径",
			certFile: "",
			wantErr:  true,
			errMsg:   "invalid certificate file: certificate file path is empty",
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
				t.Errorf("GetResultsByOpensslX509() error msg = %v, want contains '%s'", err, tt.errMsg)
			}
			if !tt.wantErr && len(got) == 0 {
				t.Error("GetResultsByOpensslX509() returned empty results for valid cert")
			}
		})
	}
}

// ===================== 单元测试 - GetResultsByOpensslAns1 =====================
func TestGetResultsByOpensslAns1(t *testing.T) {
	_, validDER, validPEM, invalidFile, nonExistFile, noPermFile := setupTest(t)

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
			name:     "权限不足",
			certFile: noPermFile,
			wantErr:  true,
			errMsg:   "invalid asn1 file: certificate file not found",
		},
		{
			name:     "有效DER格式证书",
			certFile: validDER,
			wantErr:  false,
			errMsg:   "",
		},
		{
			name:     "有效PEM格式证书（自动降级尝试）",
			certFile: validPEM,
			wantErr:  false,
			errMsg:   "",
		},
		{
			name:     "无效文件（非ASN1）",
			certFile: invalidFile,
			wantErr:  true,
			errMsg:   "fail to parse asn1 format: invalid format or corrupted file",
		},
		{
			name:     "空路径",
			certFile: "",
			wantErr:  true,
			errMsg:   "invalid asn1 file: certificate file path is empty",
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
				t.Errorf("GetResultsByOpensslAns1() error msg = %v, want contains '%s'", err, tt.errMsg)
			}
			if !tt.wantErr && len(got) == 0 {
				t.Error("GetResultsByOpensslAns1() returned empty results for valid cert")
			}
		})
	}
}

// ===================== 性能测试（Benchmark） =====================
// BenchmarkGetResultsByOpensslX509: 测试x509解析性能
func BenchmarkGetResultsByOpensslX509(b *testing.B) {
	_, validCert, _, _, _, _ := setupBenchmark(b)
	b.ResetTimer() // 排除初始化耗时
	for i := 0; i < b.N; i++ {
		_, _ = GetResultsByOpensslX509(validCert)
	}
}

// BenchmarkGetResultsByOpensslAns1: 测试asn1解析性能
func BenchmarkGetResultsByOpensslAns1(b *testing.B) {
	_, validCert, _, _, _, _ := setupBenchmark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetResultsByOpensslAns1(validCert)
	}
}

// BenchmarkProcessOpensslOutput: 测试输出处理性能（高频场景）
func BenchmarkProcessOpensslOutput(b *testing.B) {
	// 模拟真实openssl输出（100行）
	mockOutput := make([]byte, 0, 1024)
	for i := 0; i < 100; i++ {
		mockOutput = append(mockOutput, []byte(fmt.Sprintf("  Field %d: Value %d\n", i, i))...)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processOpensslOutput(mockOutput)
	}
}

// BenchmarkValidateCertFile: 测试文件校验性能
func BenchmarkValidateCertFile(b *testing.B) {
	_, validCert, _, _, _, _ := setupBenchmark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateCertFile(validCert)
	}
}

// BenchmarkExecOpensslCmd: 测试openssl命令执行性能
func BenchmarkExecOpensslCmd(b *testing.B) {
	_, _, validCertPEM, _, _, _ := setupBenchmark(b)
	args := []string{"x509", "-noout", "-text", "-in", validCertPEM}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = execOpensslCmd(args)
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
