package passwordutil

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpusoft/goutil/hashutil"
)

// ===================== 基础功能测试 =====================
// TestGetHashPasswordAndSalt 测试哈希密码和盐生成功能
func TestGetHashPasswordAndSalt(t *testing.T) {
	password := "test123456"
	hashPwd, salt := GetHashPasswordAndSalt(password)

	// 验证盐非空
	if salt == "" {
		t.Error("生成的盐值为空，不符合预期")
	}

	// 验证哈希密码可通过VerifyHashPassword验证
	if !VerifyHashPassword(password, salt, hashPwd) {
		t.Error("生成的哈希密码无法通过验证，功能异常")
	}
}

// TestVerifyHashPassword 测试密码验证功能（覆盖匹配/不匹配/空密码场景）
func TestVerifyHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		salt     string
		hashPwd  string
		want     bool
	}{
		{
			name:     "密码匹配",
			password: "test123456",
			salt:     "test-salt-123",
			hashPwd:  GetHashPassword("test123456", "test-salt-123"),
			want:     true,
		},
		{
			name:     "密码不匹配",
			password: "wrong123",
			salt:     "test-salt-123",
			hashPwd:  GetHashPassword("test123456", "test-salt-123"),
			want:     false,
		},
		{
			name:     "空密码匹配",
			password: "",
			salt:     "empty-salt",
			hashPwd:  GetHashPassword("", "empty-salt"),
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VerifyHashPassword(tt.password, tt.salt, tt.hashPwd); got != tt.want {
				t.Errorf("VerifyHashPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ===================== 临界值&异常场景测试 =====================
// TestForceTestHashPassword_FoundPassword 测试能找到匹配密码的场景
func TestForceTestHashPassword_FoundPassword(t *testing.T) {
	// 1. 创建临时字典文件
	tempFile, err := os.CreateTemp("", "dict-*.txt")
	if err != nil {
		t.Fatalf("创建临时字典文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name()) // 测试后删除文件

	// 2. 写入测试内容（包含匹配密码和随机密码）
	content := `random123
test-target-password
wrong456
`
	if _, err := tempFile.WriteString(content); err != nil {
		t.Fatalf("写入临时字典文件失败: %v", err)
	}
	tempFile.Close()

	// 3. 生成目标密码的哈希和盐
	targetPwd := "test-target-password"
	hashPwd, salt := GetHashPasswordAndSalt(targetPwd)

	// 4. 执行暴力破解
	foundPwd, err := ForceTestHashPassword(hashPwd, salt, tempFile.Name())
	if err != nil {
		t.Fatalf("ForceTestHashPassword执行失败: %v", err)
	}

	// 5. 验证结果
	if foundPwd != targetPwd {
		t.Errorf("找到的密码不匹配，预期: %s, 实际: %s", targetPwd, foundPwd)
	}
}

// TestForceTestHashPassword_NotFoundPassword 测试字典中无匹配密码的场景
func TestForceTestHashPassword_NotFoundPassword(t *testing.T) {
	// 1. 创建临时字典文件
	tempFile, err := os.CreateTemp("", "dict-*.txt")
	if err != nil {
		t.Fatalf("创建临时字典文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// 2. 写入无匹配的密码
	content := `wrong123
wrong456
wrong789
`
	if _, err := tempFile.WriteString(content); err != nil {
		t.Fatalf("写入临时字典文件失败: %v", err)
	}
	tempFile.Close()

	// 3. 生成目标密码的哈希和盐
	targetPwd := "test-target-password"
	hashPwd, salt := GetHashPasswordAndSalt(targetPwd)

	// 4. 执行暴力破解
	foundPwd, err := ForceTestHashPassword(hashPwd, salt, tempFile.Name())
	if err != nil {
		t.Fatalf("ForceTestHashPassword执行失败: %v", err)
	}

	// 5. 验证结果为空
	if foundPwd != "" {
		t.Errorf("预期返回空密码，实际返回: %s", foundPwd)
	}
}

// TestForceTestHashPassword_EmptyDictFile 测试空字典文件场景
func TestForceTestHashPassword_EmptyDictFile(t *testing.T) {
	// 1. 创建空临时文件
	tempFile, err := os.CreateTemp("", "dict-empty-*.txt")
	if err != nil {
		t.Fatalf("创建临时字典文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// 2. 生成目标密码的哈希和盐
	hashPwd, salt := GetHashPasswordAndSalt("test123")

	// 3. 执行暴力破解
	foundPwd, err := ForceTestHashPassword(hashPwd, salt, tempFile.Name())
	if err != nil {
		t.Fatalf("ForceTestHashPassword执行失败: %v", err)
	}

	// 4. 验证结果为空
	if foundPwd != "" {
		t.Errorf("空字典文件应返回空密码，实际返回: %s", foundPwd)
	}
}

// TestForceTestHashPassword_FileNotExist 测试文件不存在的异常场景
func TestForceTestHashPassword_FileNotExist(t *testing.T) {
	// 1. 构造不存在的文件路径
	nonExistFile := filepath.Join(os.TempDir(), "non-exist-dict-"+GetHashPassword("test", "salt")+".txt")

	// 2. 执行暴力破解
	foundPwd, err := ForceTestHashPassword("dummy-hash", "dummy-salt", nonExistFile)

	// 3. 验证异常处理
	if err == nil {
		t.Error("文件不存在时应返回错误，实际未返回")
	}
	if foundPwd != "" {
		t.Errorf("文件不存在时应返回空密码，实际返回: %s", foundPwd)
	}
}

// TestForceTestHashPassword_EmptyLine 测试字典包含空行的场景
func TestForceTestHashPassword_EmptyLine(t *testing.T) {
	// 1. 创建包含空行的临时字典
	tempFile, err := os.CreateTemp("", "dict-emptyline-*.txt")
	if err != nil {
		t.Fatalf("创建临时字典文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// 2. 写入包含空行的内容
	content := `

test-target-password

wrong456
`
	if _, err := tempFile.WriteString(content); err != nil {
		t.Fatalf("写入临时字典文件失败: %v", err)
	}
	tempFile.Close()

	// 3. 生成目标密码的哈希和盐
	targetPwd := "test-target-password"
	hashPwd, salt := GetHashPasswordAndSalt(targetPwd)

	// 4. 执行暴力破解
	foundPwd, err := ForceTestHashPassword(hashPwd, salt, tempFile.Name())
	if err != nil {
		t.Fatalf("ForceTestHashPassword执行失败: %v", err)
	}

	// 5. 验证结果
	if foundPwd != targetPwd {
		t.Errorf("找到的密码不匹配，预期: %s, 实际: %s", targetPwd, foundPwd)
	}
}

// ===================== 性能测试 =====================
// BenchmarkForceTestHashPassword_FoundEarly 性能测试：密码出现在字典开头（快速终止）
func BenchmarkForceTestHashPassword_FoundEarly(b *testing.B) {
	// 1. 创建包含1000行的临时字典，目标密码在第一行
	tempFile, err := os.CreateTemp("", "bench-dict-*.txt")
	if err != nil {
		b.Fatalf("创建临时字典文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// 2. 写入测试内容
	content := "bench-target-password\n"
	for i := 0; i < 999; i++ {
		content += "wrong-password-" + string(rune(i)) + "\n"
	}
	if _, err := tempFile.WriteString(content); err != nil {
		b.Fatalf("写入临时字典文件失败: %v", err)
	}
	tempFile.Close()

	// 3. 生成目标密码的哈希和盐
	hashPwd, salt := GetHashPasswordAndSalt("bench-target-password")

	// 4. 执行性能测试
	b.ResetTimer() // 重置计时器，排除文件创建耗时
	for i := 0; i < b.N; i++ {
		_, err := ForceTestHashPassword(hashPwd, salt, tempFile.Name())
		if err != nil {
			b.Fatalf("ForceTestHashPassword执行失败: %v", err)
		}
	}
}

// BenchmarkForceTestHashPassword_NotFound 性能测试：遍历整个字典无匹配
func BenchmarkForceTestHashPassword_NotFound(b *testing.B) {
	// 1. 创建包含1000行的临时字典，无匹配密码
	tempFile, err := os.CreateTemp("", "bench-dict-*.txt")
	if err != nil {
		b.Fatalf("创建临时字典文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// 2. 写入测试内容
	content := ""
	for i := 0; i < 1000; i++ {
		content += "wrong-password-" + string(rune(i)) + "\n"
	}
	if _, err := tempFile.WriteString(content); err != nil {
		b.Fatalf("写入临时字典文件失败: %v", err)
	}
	tempFile.Close()

	// 3. 生成目标密码的哈希和盐（字典中无此密码）
	hashPwd, salt := GetHashPasswordAndSalt("bench-target-password")

	// 4. 执行性能测试
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ForceTestHashPassword(hashPwd, salt, tempFile.Name())
		if err != nil {
			b.Fatalf("ForceTestHashPassword执行失败: %v", err)
		}
	}
}

// ////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////

func TestGetHashPasswordAndSalt1(t *testing.T) {
	p := `a`
	h := GetHashPassword(p, "")
	fmt.Println(h)

	b := VerifyHashPassword(p, "", h)
	fmt.Println(b)
}

func TestForceTestHashPassword1(t *testing.T) {
	file := `./all.txt`
	p, err := ForceTestHashPassword("1", "", file)
	fmt.Println(p, err)
}

func TestHash111(t *testing.T) {
	password := `a`
	salt := ""
	hashPassword1 := hashutil.Sha256([]byte(password + salt))
	fmt.Println(hashPassword1)
	if hashPassword1 == "1" {
		fmt.Println("equal")
	} else {
		fmt.Println("not equal")
	}
}
