package fileutil

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cpusoft/goutil/base64util"
	"github.com/stretchr/testify/assert"
)

// 测试用临时目录
const testTempDir = "./fileutil_test_temp"

// TestMain 初始化/清理测试环境
func TestMain(m *testing.M) {
	// 初始化：创建临时目录
	_ = os.MkdirAll(testTempDir, 0755)

	// 运行测试
	code := m.Run()

	// 清理：删除临时目录
	_ = os.RemoveAll(testTempDir)

	os.Exit(code)
}

// 获取临时文件路径
func getTestFile(name string) string {
	return filepath.Join(testTempDir, name)
}

// ------------------------------ 基础功能测试 ------------------------------

// TestReadFileToLines 测试读取文件到行切片
func TestReadFileToLines(t *testing.T) {
	// 测试场景1：正常文件
	testFile := getTestFile("test_lines.txt")
	content := "line1\nline2\nline3"
	_ = os.WriteFile(testFile, []byte(content), 0644)

	lines, err := ReadFileToLines(testFile)
	if err != nil {
		t.Fatalf("ReadFileToLines failed: %v", err)
	}
	if len(lines) != 3 || lines[0] != "line1" || lines[2] != "line3" {
		t.Errorf("ReadFileToLines result error, got: %v", lines)
	}

	// 测试场景2：空文件
	emptyFile := getTestFile("empty_lines.txt")
	_ = os.WriteFile(emptyFile, []byte(""), 0644)
	lines, err = ReadFileToLines(emptyFile)
	if err != nil {
		t.Fatalf("ReadFileToLines empty file failed: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("ReadFileToLines empty file should return empty slice, got: %v", lines)
	}

	// 测试场景3：最后一行无换行符
	noEolFile := getTestFile("no_eol.txt")
	_ = os.WriteFile(noEolFile, []byte("line1\nline2\nline3_no_eol"), 0644)
	lines, err = ReadFileToLines(noEolFile)
	if err != nil {
		t.Fatalf("ReadFileToLines no eol failed: %v", err)
	}
	if len(lines) != 3 || lines[2] != "line3_no_eol" {
		t.Errorf("ReadFileToLines no eol result error, got: %v", lines)
	}

	// 测试场景4：空参数
	_, err = ReadFileToLines("")
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Error("ReadFileToLines empty param should return error")
	}

	// 测试场景5：文件不存在
	_, err = ReadFileToLines(getTestFile("not_exists.txt"))
	if err == nil {
		t.Error("ReadFileToLines not exists file should return error")
	}
}

// TestReadFileToBytes 测试读取文件到字节切片
func TestReadFileToBytes(t *testing.T) {
	// 正常场景
	testFile := getTestFile("test_bytes.txt")
	content := []byte("test content")
	_ = os.WriteFile(testFile, content, 0644)

	bytes, err := ReadFileToBytes(testFile)
	if err != nil {
		t.Fatalf("ReadFileToBytes failed: %v", err)
	}
	if string(bytes) != string(content) {
		t.Errorf("ReadFileToBytes result error, got: %s", string(bytes))
	}

	// 空参数
	_, err = ReadFileToBytes("")
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Error("ReadFileToBytes empty param should return error")
	}
}

// TestReadFileAndDecodeCertBase64 测试读取并解码证书Base64
func TestReadFileAndDecodeCertBase64(t *testing.T) {
	// 正常场景
	testFile := getTestFile("test_cert_base64.txt")
	rawContent := []byte("test cert content")
	base64Content := base64.StdEncoding.EncodeToString(rawContent)
	_ = os.WriteFile(testFile, []byte(base64Content), 0644)

	fileByte, decodeByte, err := ReadFileAndDecodeCertBase64(testFile)
	if err != nil {
		t.Fatalf("ReadFileAndDecodeCertBase64 failed: %v", err)
	}
	if string(fileByte) != base64Content || string(decodeByte) != string(rawContent) {
		t.Errorf("ReadFileAndDecodeCertBase64 result error")
	}

	// 空文件
	emptyFile := getTestFile("empty_cert.txt")
	_ = os.WriteFile(emptyFile, []byte(""), 0644)
	_, _, err = ReadFileAndDecodeCertBase64(emptyFile)
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Error("ReadFileAndDecodeCertBase64 empty file should return error")
	}

	// 无效Base64
	invalidFile := getTestFile("invalid_base64.txt")
	_ = os.WriteFile(invalidFile, []byte("invalid base64 !@#"), 0644)
	_, _, err = ReadFileAndDecodeCertBase64(invalidFile)
	if err == nil {
		t.Error("ReadFileAndDecodeCertBase64 invalid base64 should return error")
	}
}

// TestWriteBytesToFile 测试写入字节到文件
func TestWriteBytesToFile(t *testing.T) {
	// 正常场景
	testFile := getTestFile("write_bytes.txt")
	content := []byte("test write bytes")
	err := WriteBytesToFile(testFile, content)
	if err != nil {
		t.Fatalf("WriteBytesToFile failed: %v", err)
	}

	// 验证写入结果
	readBytes, _ := os.ReadFile(testFile)
	if string(readBytes) != string(content) {
		t.Errorf("WriteBytesToFile result error, got: %s", string(readBytes))
	}

	// 覆盖写入
	newContent := []byte("new content")
	err = WriteBytesToFile(testFile, newContent)
	if err != nil {
		t.Fatalf("WriteBytesToFile overwrite failed: %v", err)
	}
	readBytes, _ = os.ReadFile(testFile)
	if string(readBytes) != string(newContent) {
		t.Errorf("WriteBytesToFile overwrite error")
	}

	// 空参数
	err = WriteBytesToFile("", content)
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Error("WriteBytesToFile empty path should return error")
	}

	// 空字节
	err = WriteBytesToFile(getTestFile("empty_bytes.txt"), []byte(""))
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Error("WriteBytesToFile empty bytes should return error")
	}
}

// TestWriteBytesAppendFile 测试追加写入字节
func TestWriteBytesAppendFile(t *testing.T) {
	testFile := getTestFile("append_bytes.txt")
	// 清空文件
	_ = os.WriteFile(testFile, []byte(""), 0644)

	// 第一次写入
	content1 := []byte("first part\n")
	err := WriteBytesAppendFile(testFile, content1)
	if err != nil {
		t.Fatalf("WriteBytesAppendFile first write failed: %v", err)
	}

	// 第二次追加
	content2 := []byte("second part")
	err = WriteBytesAppendFile(testFile, content2)
	if err != nil {
		t.Fatalf("WriteBytesAppendFile append failed: %v", err)
	}

	// 验证结果
	readBytes, _ := os.ReadFile(testFile)
	expected := string(content1) + string(content2)
	if string(readBytes) != expected {
		t.Errorf("WriteBytesAppendFile result error, got: %s, expected: %s", string(readBytes), expected)
	}
}

// TestGetFileLength 测试获取文件长度
func TestGetFileLength(t *testing.T) {
	testFile := getTestFile("file_length.txt")
	content := []byte("1234567890")
	_ = os.WriteFile(testFile, content, 0644)

	length, err := GetFileLength(testFile)
	if err != nil {
		t.Fatalf("GetFileLength failed: %v", err)
	}
	if length != int64(len(content)) {
		t.Errorf("GetFileLength error, got: %d, expected: %d", length, len(content))
	}

	// 空参数
	_, err = GetFileLength("")
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Error("GetFileLength empty param should return error")
	}

	// 文件不存在
	_, err = GetFileLength(getTestFile("not_exists_length.txt"))
	if err == nil {
		t.Error("GetFileLength not exists file should return error")
	}
}

// TestCheckFileNameMaxLength 测试文件名长度校验
func TestCheckFileNameMaxLength(t *testing.T) {
	// 正常长度
	if !CheckFileNameMaxLength("normal_name.txt") {
		t.Error("CheckFileNameMaxLength normal name should return true")
	}

	// 临界值：256个字符
	longName := strings.Repeat("a", FILENAME_MAXLENGTH)
	if !CheckFileNameMaxLength(longName) {
		t.Error("CheckFileNameMaxLength 256 chars should return true")
	}

	// 超过临界值：257个字符
	tooLongName := strings.Repeat("a", FILENAME_MAXLENGTH+1)
	if CheckFileNameMaxLength(tooLongName) {
		t.Error("CheckFileNameMaxLength 257 chars should return false")
	}

	// 空文件名
	if CheckFileNameMaxLength("") {
		t.Error("CheckFileNameMaxLength empty name should return false")
	}
}

// TestCheckPathNameMaxLength 测试路径长度校验
func TestCheckPathNameMaxLength(t *testing.T) {
	// 正常长度
	if !CheckPathNameMaxLength("/normal/path/file.txt") {
		t.Error("CheckPathNameMaxLength normal path should return true")
	}

	// 临界值：4096个字符
	longPath := "/" + strings.Repeat("a", PATHNAME_MAXLENGTH-1)
	if !CheckPathNameMaxLength(longPath) {
		t.Error("CheckPathNameMaxLength 4096 chars should return true")
	}

	// 超过临界值：4097个字符
	tooLongPath := "/" + strings.Repeat("a", PATHNAME_MAXLENGTH)
	if CheckPathNameMaxLength(tooLongPath) {
		t.Error("CheckPathNameMaxLength 4097 chars should return false")
	}

	// 空路径
	if CheckPathNameMaxLength("") {
		t.Error("CheckPathNameMaxLength empty path should return false")
	}
}

// TestWriteBase64ToFile 测试Base64写入文件
func TestWriteBase64ToFile(t *testing.T) {
	testFile := getTestFile("base64_file.txt")
	rawContent := []byte("test base64 content")
	base64Str := base64.StdEncoding.EncodeToString(rawContent)

	// 写入Base64
	err := WriteBase64ToFile(testFile, base64Str)
	if err != nil {
		t.Fatalf("WriteBase64ToFile failed: %v", err)
	}

	// 验证结果
	readBytes, _ := os.ReadFile(testFile)
	if string(readBytes) != string(rawContent) {
		t.Errorf("WriteBase64ToFile result error, got: %s", string(readBytes))
	}

	// 无效Base64
	err = WriteBase64ToFile(getTestFile("invalid_base64_write.txt"), "invalid!@#")
	if err == nil {
		t.Error("WriteBase64ToFile invalid base64 should return error")
	}
}

// TestCreateAndWriteBase64ToFile 测试创建目录并写入Base64
func TestCreateAndWriteBase64ToFile(t *testing.T) {
	testFile := getTestFile("subdir1/subdir2/base64_create.txt")
	rawContent := []byte("create dir and write")
	base64Str := base64.StdEncoding.EncodeToString(rawContent)

	// 写入（自动创建目录）
	err := CreateAndWriteBase64ToFile(testFile, base64Str)
	if err != nil {
		t.Fatalf("CreateAndWriteBase64ToFile failed: %v", err)
	}

	// 验证文件存在且内容正确
	readBytes, _ := os.ReadFile(testFile)
	if string(readBytes) != string(rawContent) {
		t.Errorf("CreateAndWriteBase64ToFile result error")
	}
}

// TestIsFileDiffWithBase64 测试文件与Base64内容对比
func TestIsFileDiffWithBase64(t *testing.T) {
	testFile := getTestFile("diff_test.txt")
	rawContent := []byte("test diff content")
	_ = os.WriteFile(testFile, rawContent, 0644)

	// 相同内容
	base64Same := base64.StdEncoding.EncodeToString(rawContent)
	diff, err := IsFileDiffWithBase64(testFile, base64Same)
	if err != nil {
		t.Fatalf("IsFileDiffWithBase64 failed: %v", err)
	}
	if diff {
		t.Error("IsFileDiffWithBase64 same content should return false")
	}

	// 不同内容
	base64Diff := base64.StdEncoding.EncodeToString([]byte("different content"))
	diff, err = IsFileDiffWithBase64(testFile, base64Diff)
	if err != nil {
		t.Fatalf("IsFileDiffWithBase64 failed: %v", err)
	}
	if !diff {
		t.Error("IsFileDiffWithBase64 different content should return true")
	}
}

// TestJoinPrefixAndUrlFileNameAndWriteBase64ToFile 测试路径拼接+Base64写入
func TestJoinPrefixAndUrlFileNameAndWriteBase64ToFile(t *testing.T) {
	// 正常场景
	destPath := testTempDir
	url := "http://example.com/test_file.txt"
	rawContent := []byte("test join path")
	base64Str := base64.StdEncoding.EncodeToString(rawContent)

	filePathName, err := JoinPrefixAndUrlFileNameAndWriteBase64ToFile(destPath, url, base64Str)
	if err != nil {
		t.Fatalf("JoinPrefixAndUrlFileNameAndWriteBase64ToFile failed: %v", err)
	}

	// 验证文件存在且内容正确
	readBytes, _ := os.ReadFile(filePathName)
	if string(readBytes) != string(rawContent) {
		t.Errorf("JoinPrefixAndUrlFileNameAndWriteBase64ToFile content error")
	}

	// 路径遍历攻击测试
	urlAttack := "http://example.com/../../etc/passwd"
	_, err = JoinPrefixAndUrlFileNameAndWriteBase64ToFile(destPath, urlAttack, base64Str)
	if err == nil || !strings.Contains(err.Error(), "path traversal") {
		t.Error("JoinPrefixAndUrlFileNameAndWriteBase64ToFile should block path traversal attack")
	}
}

// TestCopy 测试文件拷贝
func TestCopy(t *testing.T) {
	srcFile := getTestFile("copy_src.txt")
	dstFile := getTestFile("copy_dst.txt")
	content := []byte("test copy content")
	_ = os.WriteFile(srcFile, content, 0644)

	// 拷贝文件
	err := Copy(srcFile, dstFile)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	// 验证拷贝结果
	readBytes, _ := os.ReadFile(dstFile)
	if string(readBytes) != string(content) {
		t.Errorf("Copy result error, got: %s", string(readBytes))
	}

	// 源文件不存在
	err = Copy(getTestFile("not_exists_src.txt"), dstFile)
	if err == nil {
		t.Error("Copy not exists src should return error")
	}
}

// ------------------------------ 性能测试（基准测试） ------------------------------

// 生成指定大小的随机字节
func generateRandomBytes(size int) []byte {
	b := make([]byte, size)
	_, _ = rand.Read(b)
	return b
}

// BenchmarkReadFileToBytes 大文件读取性能测试
func BenchmarkReadFileToBytes(b *testing.B) {
	// 准备100MB测试文件
	testFile := getTestFile("bench_read_100mb.txt")
	content := generateRandomBytes(100 * 1024 * 1024) // 100MB
	_ = os.WriteFile(testFile, content, 0644)

	b.ResetTimer() // 重置计时器，排除准备时间
	for i := 0; i < b.N; i++ {
		_, err := ReadFileToBytes(testFile)
		if err != nil {
			b.Fatalf("BenchmarkReadFileToBytes failed: %v", err)
		}
	}
}

// BenchmarkWriteBytesToFile 大文件写入性能测试
func BenchmarkWriteBytesToFile(b *testing.B) {
	testFile := getTestFile("bench_write_10mb.txt")
	content := generateRandomBytes(10 * 1024 * 1024) // 10MB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := WriteBytesToFile(testFile, content)
		if err != nil {
			b.Fatalf("BenchmarkWriteBytesToFile failed: %v", err)
		}
	}
}

// BenchmarkIsFileDiffWithBase64 哈希对比性能测试
func BenchmarkIsFileDiffWithBase64(b *testing.B) {
	// 准备10MB测试文件
	testFile := getTestFile("bench_diff_10mb.txt")
	content := generateRandomBytes(10 * 1024 * 1024) // 10MB
	_ = os.WriteFile(testFile, content, 0644)
	base64Str, _ := base64util.DecodeBase64(base64.StdEncoding.EncodeToString(content))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := IsFileDiffWithBase64(testFile, string(base64Str))
		if err != nil {
			b.Fatalf("BenchmarkIsFileDiffWithBase64 failed: %v", err)
		}
	}
}

// BenchmarkCopy 拷贝性能测试
func BenchmarkCopy(b *testing.B) {
	srcFile := getTestFile("bench_copy_src_5mb.txt")
	dstFile := getTestFile("bench_copy_dst_5mb.txt")
	content := generateRandomBytes(5 * 1024 * 1024) // 5MB
	_ = os.WriteFile(srcFile, content, 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := Copy(srcFile, dstFile)
		if err != nil {
			b.Fatalf("BenchmarkCopy failed: %v", err)
		}
	}
}

func TestWriteBase64ToFile1(t *testing.T) {
	s := `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`

	f := "/tmp/1.asa"
	err := WriteBase64ToFile(f, s)
	fmt.Println(err)

	f = "/tmp/2.asa"
	err = CreateAndWriteBase64ToFile(f, s)
	fmt.Println(err)
}

func TestCopy1(t *testing.T) {
	src := `/tmp/1.txt`
	dst := `/tmp/1_copy.txt`
	err := Copy(src, dst)
	fmt.Println(err)
}

func TestIsFileDiffWithBase641(t *testing.T) {
	tmpFile := "/tmp/test_fileutil_diff.txt"
	defer os.Remove(tmpFile)

	testContent := "Hello, World! Tesrt  "
	testContentBase64 := base64.StdEncoding.EncodeToString([]byte(testContent))
	differentContentBase64 := base64.StdEncoding.EncodeToString([]byte("different conten"))

	// Write test content to file
	err := WriteBytesToFile(tmpFile, []byte(testContent))
	assert.NoError(t, err)

	// no diff
	hasDiff, err := IsFileDiffWithBase64(tmpFile, testContentBase64)
	assert.NoError(t, err)
	assert.False(t, hasDiff, " same")

	// has diff
	hasDiff, err = IsFileDiffWithBase64(tmpFile, differentContentBase64)
	assert.NoError(t, err)
	assert.True(t, hasDiff, " different content")

}
