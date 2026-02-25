package fileutil

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试临时目录（仅Linux）
const testTempDir = "./test_temp_fileutil_linux"

// TestMain 测试入口，初始化/清理测试环境（仅Linux）
func TestMain(m *testing.M) {
	// 创建Linux测试临时目录
	if err := os.MkdirAll(testTempDir, 0755); err != nil {
		panic("Failed to create test temp dir (Linux): " + err.Error())
	}

	// 运行所有测试
	code := m.Run()

	// 清理测试临时目录
	if err := os.RemoveAll(testTempDir); err != nil {
		panic("Failed to clean test temp dir (Linux): " + err.Error())
	}

	os.Exit(code)
}

// 获取测试文件路径（仅Linux）
func getTestFile(name string) string {
	return filepath.Join(testTempDir, name)
}

// 生成随机字节（用于Linux大文件测试）
func generateRandomBytes(size int) []byte {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		panic("Failed to generate random bytes (Linux): " + err.Error())
	}
	return b
}

// ------------------------------ 基础文件操作函数测试 ------------------------------
func TestReadFileToLines(t *testing.T) {
	// 场景1：正常文件
	testFile := getTestFile("read_lines.txt")
	content := "line1\nline2\nline3\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file (Linux): %v", err)
	}

	lines, err := ReadFileToLines(testFile)
	if err != nil {
		t.Fatalf("ReadFileToLines failed (Linux): %v", err)
	}
	if len(lines) != 3 || lines[0] != "line1" || lines[2] != "line3" {
		t.Errorf("ReadFileToLines result error (Linux), got: %v", lines)
	}

	// 场景2：空文件
	emptyFile := getTestFile("empty_lines.txt")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write empty test file (Linux): %v", err)
	}
	lines, err = ReadFileToLines(emptyFile)
	if err != nil {
		t.Fatalf("ReadFileToLines empty file failed (Linux): %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("ReadFileToLines empty file should return empty slice (Linux), got: %v", lines)
	}

	// 场景3：空参数
	_, err = ReadFileToLines("")
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("ReadFileToLines empty param should return error (Linux)")
	}

	// 场景4：文件不存在
	_, err = ReadFileToLines(getTestFile("not_exists_lines.txt"))
	if err == nil {
		t.Error("ReadFileToLines not exists file should return error (Linux)")
	}
}

func TestReadFileToBytes(t *testing.T) {
	// 场景1：正常文件
	testFile := getTestFile("read_bytes.txt")
	content := []byte("test read bytes 123456 (Linux)")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to write test file (Linux): %v", err)
	}

	bytes, err := ReadFileToBytes(testFile)
	if err != nil {
		t.Fatalf("ReadFileToBytes failed (Linux): %v", err)
	}
	if string(bytes) != string(content) {
		t.Errorf("ReadFileToBytes result error (Linux), got: %s, expected: %s", string(bytes), string(content))
	}

	// 场景2：空参数
	_, err = ReadFileToBytes("")
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("ReadFileToBytes empty param should return error (Linux)")
	}

	// 场景3：文件不存在
	_, err = ReadFileToBytes(getTestFile("not_exists_bytes.txt"))
	if err == nil {
		t.Error("ReadFileToBytes not exists file should return error (Linux)")
	}

	// 临界值：Linux大文件（10MB）
	bigFile := getTestFile("big_read_bytes.txt")
	bigContent := generateRandomBytes(10 * 1024 * 1024)
	if err := os.WriteFile(bigFile, bigContent, 0644); err != nil {
		t.Fatalf("Failed to write big test file (Linux): %v", err)
	}
	bigBytes, err := ReadFileToBytes(bigFile)
	if err != nil {
		t.Fatalf("ReadFileToBytes big file failed (Linux): %v", err)
	}
	if len(bigBytes) != len(bigContent) {
		t.Errorf("ReadFileToBytes big file length error (Linux), got: %d, expected: %d", len(bigBytes), len(bigContent))
	}
}

func TestReadFileAndDecodeCertBase64(t *testing.T) {
	// 场景1：正常Base64文件
	testFile := getTestFile("cert_base64.txt")
	rawCert := []byte("test cert content for decode 123 (Linux)")
	base64Cert := base64.StdEncoding.EncodeToString(rawCert)
	if err := os.WriteFile(testFile, []byte(base64Cert), 0644); err != nil {
		t.Fatalf("Failed to write test file (Linux): %v", err)
	}

	fileByte, decodeByte, err := ReadFileAndDecodeCertBase64(testFile)
	if err != nil {
		t.Fatalf("ReadFileAndDecodeCertBase64 failed (Linux): %v", err)
	}
	if string(fileByte) != base64Cert || string(decodeByte) != string(rawCert) {
		t.Error("ReadFileAndDecodeCertBase64 content error (Linux)")
	}

	// 场景2：空文件
	emptyFile := getTestFile("empty_cert.txt")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write empty test file (Linux): %v", err)
	}
	_, _, err = ReadFileAndDecodeCertBase64(emptyFile)
	if err == nil || !strings.Contains(err.Error(), "file is empty") {
		t.Error("ReadFileAndDecodeCertBase64 empty file should return error (Linux)")
	}

	// 场景3：无效Base64
	invalidFile := getTestFile("invalid_cert_base64.txt")
	if err := os.WriteFile(invalidFile, []byte("invalid base64 !@#$%^ (Linux)"), 0644); err != nil {
		t.Fatalf("Failed to write invalid test file (Linux): %v", err)
	}
	_, _, err = ReadFileAndDecodeCertBase64(invalidFile)
	if err == nil {
		t.Error("ReadFileAndDecodeCertBase64 invalid base64 should return error (Linux)")
	}

	// 场景4：空参数
	_, _, err = ReadFileAndDecodeCertBase64("")
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("ReadFileAndDecodeCertBase64 empty param should return error (Linux)")
	}
}

func TestWriteBytesToFile(t *testing.T) {
	// 场景1：正常写入
	testFile := getTestFile("write_bytes.txt")
	content := []byte("test write bytes 789 (Linux)")
	err := WriteBytesToFile(testFile, content)
	if err != nil {
		t.Fatalf("WriteBytesToFile failed (Linux): %v", err)
	}

	readBytes, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read written file (Linux): %v", err)
	}
	if string(readBytes) != string(content) {
		t.Errorf("WriteBytesToFile result error (Linux), got: %s, expected: %s", string(readBytes), string(content))
	}

	// 场景2：覆盖写入
	overwriteContent := []byte("overwrite content (Linux)")
	err = WriteBytesToFile(testFile, overwriteContent)
	if err != nil {
		t.Fatalf("WriteBytesToFile overwrite failed (Linux): %v", err)
	}
	readBytes, err = os.ReadFile(testFile)
	if string(readBytes) != string(overwriteContent) {
		t.Errorf("WriteBytesToFile overwrite error (Linux), got: %s", string(readBytes))
	}

	// 场景3：空参数
	err = WriteBytesToFile("", content)
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("WriteBytesToFile empty path should return error (Linux)")
	}

	// 场景4：空字节
	err = WriteBytesToFile(getTestFile("empty_bytes.txt"), []byte(""))
	if err == nil || !strings.Contains(err.Error(), "bytes to write is empty") {
		t.Error("WriteBytesToFile empty bytes should return error (Linux)")
	}

	// 临界值：Linux最大路径长度
	pathPrefix := testTempDir + "/"
	remainingLen := PathNameMaxLength - len(pathPrefix) - 4 // 预留.txt后缀
	longFileName := strings.Repeat("a", remainingLen)
	longPath := pathPrefix + longFileName + ".txt"
	err = WriteBytesToFile(longPath, []byte("long path test (Linux)"))
	if err != nil {
		t.Fatalf("WriteBytesToFile long path failed (Linux): %v", err)
	}

	// 超过最大路径长度（预期失败）
	tooLongFileName := strings.Repeat("b", remainingLen+2)
	tooLongPath := pathPrefix + tooLongFileName + ".txt"
	err = WriteBytesToFile(tooLongPath, []byte("too long path (Linux)"))
	if err == nil {
		t.Error("WriteBytesToFile too long path should return error (Linux)")
	}
}

func TestWriteBytesAppendFile(t *testing.T) {
	testFile := getTestFile("append_test.txt")
	// 清空文件
	if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to init append test file (Linux): %v", err)
	}

	// 场景1：第一次写入
	content1 := []byte("first line (Linux)\n")
	err := WriteBytesAppendFile(testFile, content1)
	if err != nil {
		t.Fatalf("WriteBytesAppendFile first write failed (Linux): %v", err)
	}

	// 场景2：第二次追加
	content2 := []byte("second line (Linux)")
	err = WriteBytesAppendFile(testFile, content2)
	if err != nil {
		t.Fatalf("WriteBytesAppendFile append failed (Linux): %v", err)
	}

	// 验证结果
	readBytes, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read append file (Linux): %v", err)
	}
	expected := string(content1) + string(content2)
	if string(readBytes) != expected {
		t.Errorf("WriteBytesAppendFile result error (Linux), got: %s, expected: %s", string(readBytes), expected)
	}

	// 场景3：空参数
	err = WriteBytesAppendFile("", content1)
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("WriteBytesAppendFile empty path should return error (Linux)")
	}

	// 场景4：空字节
	err = WriteBytesAppendFile(getTestFile("empty_append.txt"), []byte(""))
	if err == nil || !strings.Contains(err.Error(), "bytes to write is empty") {
		t.Error("WriteBytesAppendFile empty bytes should return error (Linux)")
	}
}

func TestGetFileLength(t *testing.T) {
	// 场景1：正常文件
	testFile := getTestFile("file_length.txt")
	content := []byte("1234567890 (Linux)")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to write test file (Linux): %v", err)
	}

	length, err := GetFileLength(testFile)
	if err != nil {
		t.Fatalf("GetFileLength failed (Linux): %v", err)
	}
	if length != int64(len(content)) {
		t.Errorf("GetFileLength error (Linux), got: %d, expected: %d", length, len(content))
	}

	// 场景2：空参数
	_, err = GetFileLength("")
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("GetFileLength empty param should return error (Linux)")
	}

	// 场景3：文件不存在
	_, err = GetFileLength(getTestFile("not_exists_length.txt"))
	if err == nil {
		t.Error("GetFileLength not exists file should return error (Linux)")
	}

	// 临界值：Linux超大文件（100MB）
	bigFile := getTestFile("big_length.txt")
	bigContent := generateRandomBytes(100 * 1024 * 1024)
	if err := os.WriteFile(bigFile, bigContent, 0644); err != nil {
		t.Fatalf("Failed to write big test file (Linux): %v", err)
	}
	bigLength, err := GetFileLength(bigFile)
	if err != nil {
		t.Fatalf("GetFileLength big file failed (Linux): %v", err)
	}
	if bigLength != int64(len(bigContent)) {
		t.Errorf("GetFileLength big file error (Linux), got: %d, expected: %d", bigLength, len(bigContent))
	}
}

// ------------------------------ 长度校验函数测试 ------------------------------
func TestCheckFileNameMaxLength(t *testing.T) {
	// 场景1：正常长度（100字符）
	normalName := strings.Repeat("a", 100)
	if !CheckFileNameMaxLength(normalName) {
		t.Error("CheckFileNameMaxLength normal name should return true (Linux)")
	}

	// 场景2：临界值（256字符）
	criticalName := strings.Repeat("b", FileNameMaxLength)
	if !CheckFileNameMaxLength(criticalName) {
		t.Error("CheckFileNameMaxLength 256 chars should return true (Linux)")
	}

	// 场景3：超过临界值（257字符）
	tooLongName := strings.Repeat("c", FileNameMaxLength+1)
	if CheckFileNameMaxLength(tooLongName) {
		t.Error("CheckFileNameMaxLength 257 chars should return false (Linux)")
	}

	// 场景4：空文件名
	if CheckFileNameMaxLength("") {
		t.Error("CheckFileNameMaxLength empty name should return false (Linux)")
	}
}

func TestCheckPathNameMaxLength(t *testing.T) {
	// 场景1：正常长度（1000字符）
	normalPath := "/" + strings.Repeat("a", 999)
	if !CheckPathNameMaxLength(normalPath) {
		t.Error("CheckPathNameMaxLength normal path should return true (Linux)")
	}

	// 场景2：临界值（4096字符）
	criticalPath := "/" + strings.Repeat("b", PathNameMaxLength-1)
	if !CheckPathNameMaxLength(criticalPath) {
		t.Error("CheckPathNameMaxLength 4096 chars should return true (Linux)")
	}

	// 场景3：超过临界值（4097字符）
	tooLongPath := "/" + strings.Repeat("c", PathNameMaxLength)
	if CheckPathNameMaxLength(tooLongPath) {
		t.Error("CheckPathNameMaxLength 4097 chars should return false (Linux)")
	}

	// 场景4：空路径
	if CheckPathNameMaxLength("") {
		t.Error("CheckPathNameMaxLength empty path should return false (Linux)")
	}
}

// ------------------------------ Base64 相关函数测试 ------------------------------
func TestWriteBase64ToFile(t *testing.T) {
	// 场景1：正常Base64写入
	testFile := getTestFile("write_base64.txt")
	rawContent := []byte("test base64 write 123456 (Linux)")
	base64Str := base64.StdEncoding.EncodeToString(rawContent)

	err := WriteBase64ToFile(testFile, base64Str)
	if err != nil {
		t.Fatalf("WriteBase64ToFile failed (Linux): %v", err)
	}

	readBytes, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read base64 file (Linux): %v", err)
	}
	if string(readBytes) != string(rawContent) {
		t.Errorf("WriteBase64ToFile result error (Linux), got: %s", string(readBytes))
	}

	// 场景2：空参数
	err = WriteBase64ToFile("", base64Str)
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("WriteBase64ToFile empty path should return error (Linux)")
	}

	err = WriteBase64ToFile(getTestFile("empty_base64.txt"), "")
	if err == nil || !strings.Contains(err.Error(), "base64 string is empty") {
		t.Error("WriteBase64ToFile empty base64 should return error (Linux)")
	}

	// 场景3：无效Base64
	err = WriteBase64ToFile(getTestFile("invalid_base64.txt"), "invalid!@#$% (Linux)")
	if err == nil {
		t.Error("WriteBase64ToFile invalid base64 should return error (Linux)")
	}

	// 临界值：Linux大Base64（5MB）
	bigFile := getTestFile("big_base64.txt")
	bigContent := generateRandomBytes(5 * 1024 * 1024)
	bigBase64 := base64.StdEncoding.EncodeToString(bigContent)
	err = WriteBase64ToFile(bigFile, bigBase64)
	if err != nil {
		t.Fatalf("WriteBase64ToFile big file failed (Linux): %v", err)
	}
	bigReadBytes, err := os.ReadFile(bigFile)
	if err != nil {
		t.Fatalf("Failed to read big base64 file (Linux): %v", err)
	}
	if len(bigReadBytes) != len(bigContent) {
		t.Errorf("WriteBase64ToFile big file length error (Linux), got: %d, expected: %d", len(bigReadBytes), len(bigContent))
	}
}

func TestCreateAndWriteBase64ToFile(t *testing.T) {
	// 场景1：正常创建目录并写入
	testFile := getTestFile("subdir1/subdir2/create_base64.txt")
	rawContent := []byte("test create dir and write base64 (Linux)")
	base64Str := base64.StdEncoding.EncodeToString(rawContent)

	err := CreateAndWriteBase64ToFile(testFile, base64Str)
	if err != nil {
		t.Fatalf("CreateAndWriteBase64ToFile failed (Linux): %v", err)
	}

	// 验证目录创建成功
	if _, err := os.Stat(filepath.Dir(testFile)); err != nil {
		t.Error("CreateAndWriteBase64ToFile dir create failed (Linux)")
	}

	// 验证内容写入成功
	readBytes, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read create base64 file (Linux): %v", err)
	}
	if string(readBytes) != string(rawContent) {
		t.Errorf("CreateAndWriteBase64ToFile content error (Linux), got: %s", string(readBytes))
	}

	// 场景2：空参数
	err = CreateAndWriteBase64ToFile("", base64Str)
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("CreateAndWriteBase64ToFile empty path should return error (Linux)")
	}

	err = CreateAndWriteBase64ToFile(getTestFile("empty_create_base64.txt"), "")
	if err == nil || !strings.Contains(err.Error(), "base64 string is empty") {
		t.Error("CreateAndWriteBase64ToFile empty base64 should return error (Linux)")
	}

	// 临界值：Linux超长目录路径
	longPath := getTestFile(strings.Repeat("subdir_", 50) + "long_create_base64.txt")
	err = CreateAndWriteBase64ToFile(longPath, base64Str)
	if err != nil {
		t.Fatalf("CreateAndWriteBase64ToFile long path failed (Linux): %v", err)
	}
}

func TestIsFileDiffWithBase64(t *testing.T) {
	// 场景1：内容相同
	testFile := getTestFile("diff_same.txt")
	rawContent := []byte("test diff same content 1234567890 (Linux)")
	if err := os.WriteFile(testFile, rawContent, 0644); err != nil {
		t.Fatalf("Failed to write test file (Linux): %v", err)
	}
	base64Same := base64.StdEncoding.EncodeToString(rawContent)

	diff, err := IsFileDiffWithBase64(testFile, base64Same)
	if err != nil {
		t.Fatalf("IsFileDiffWithBase64 failed (Linux): %v", err)
	}
	if diff {
		t.Error("IsFileDiffWithBase64 same content should return false (Linux)")
	}

	// 场景2：内容不同
	base64Diff := base64.StdEncoding.EncodeToString([]byte("different content (Linux)"))
	diff, err = IsFileDiffWithBase64(testFile, base64Diff)
	if err != nil {
		t.Fatalf("IsFileDiffWithBase64 failed (Linux): %v", err)
	}
	if !diff {
		t.Error("IsFileDiffWithBase64 different content should return true (Linux)")
	}

	// 场景3：空参数
	_, err = IsFileDiffWithBase64("", base64Same)
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("IsFileDiffWithBase64 empty file path should return error (Linux)")
	}

	_, err = IsFileDiffWithBase64(testFile, "")
	if err == nil || !strings.Contains(err.Error(), "base64 string is empty") {
		t.Error("IsFileDiffWithBase64 empty base64 should return error (Linux)")
	}

	// 场景4：文件不存在
	_, err = IsFileDiffWithBase64(getTestFile("not_exists_diff.txt"), base64Same)
	if err == nil {
		t.Error("IsFileDiffWithBase64 not exists file should return error (Linux)")
	}

	// 临界值：Linux大文件（20MB）
	bigFile := getTestFile("big_diff.txt")
	bigContent := generateRandomBytes(20 * 1024 * 1024)
	if err := os.WriteFile(bigFile, bigContent, 0644); err != nil {
		t.Fatalf("Failed to write big test file (Linux): %v", err)
	}
	bigBase64 := base64.StdEncoding.EncodeToString(bigContent)
	diff, err = IsFileDiffWithBase64(bigFile, bigBase64)
	if err != nil {
		t.Fatalf("IsFileDiffWithBase64 big file failed (Linux): %v", err)
	}
	if diff {
		t.Error("IsFileDiffWithBase64 big file same content should return false (Linux)")
	}
}

// ------------------------------ 路径拼接与拷贝函数测试 ------------------------------
func TestJoinPrefixAndUrlFileNameAndWriteBase64ToFile(t *testing.T) {
	// 场景1：正常路径（含/./符号，Linux）
	t.Run("normal path with ./ (Linux)", func(t *testing.T) {
		destPath := testTempDir
		url := "http://example.com/test_normal.txt"
		rawContent := []byte("test join path with ./ symbol 123 (Linux)")
		base64Str := base64.StdEncoding.EncodeToString(rawContent)

		filePathName, err := JoinPrefixAndUrlFileNameAndWriteBase64ToFile(destPath, url, base64Str)
		if err != nil {
			t.Fatalf("JoinPrefixAndUrlFileNameAndWriteBase64ToFile normal path failed (Linux): %v", err)
		}

		// 验证内容
		readBytes, err := ReadFileToBytes(filePathName)
		if err != nil {
			t.Fatalf("Failed to read written file (Linux): %v", err)
		}
		if string(readBytes) != string(rawContent) {
			t.Errorf("File content mismatch (Linux): got %s, want %s", string(readBytes), string(rawContent))
		}

		// 验证路径在目标目录内
		inDir, err := isPathInDir(filePathName, destPath)
		if err != nil || !inDir {
			t.Error("File path should be in dest dir (Linux)")
		}
	})

	// 场景2：路径遍历攻击（含../，Linux）
	t.Run("path traversal with ../ (Linux)", func(t *testing.T) {
		destPath := testTempDir
		url := "http://example.com/../../etc/passwd"
		rawContent := []byte("attack content (Linux)")
		base64Str := base64.StdEncoding.EncodeToString(rawContent)

		filePathName, err := JoinPrefixAndUrlFileNameAndWriteBase64ToFile(destPath, url, base64Str)
		if err == nil {
			t.Error("Should return error for path traversal attack (Linux), but got nil")
			if filePathName != "" {
				_ = os.Remove(filePathName)
			}
		} else if !strings.Contains(err.Error(), "path traversal detected") {
			t.Errorf("Error message should contain 'path traversal' (Linux), but got: %v", err)
		}
	})

	// 场景3：空参数（Linux）
	t.Run("empty params (Linux)", func(t *testing.T) {
		rawContent := []byte("test empty params (Linux)")
		base64Str := base64.StdEncoding.EncodeToString(rawContent)

		// 空目标路径
		_, err := JoinPrefixAndUrlFileNameAndWriteBase64ToFile("", "http://example.com/test.txt", base64Str)
		if err == nil || !strings.Contains(err.Error(), "destination path is empty") {
			t.Error("Should return error for empty dest path (Linux)")
		}

		// 空URL
		_, err = JoinPrefixAndUrlFileNameAndWriteBase64ToFile(testTempDir, "", base64Str)
		if err == nil || !strings.Contains(err.Error(), "url is empty") {
			t.Error("Should return error for empty url (Linux)")
		}

		// 空Base64
		_, err = JoinPrefixAndUrlFileNameAndWriteBase64ToFile(testTempDir, "http://example.com/test.txt", "")
		if err == nil || !strings.Contains(err.Error(), "base64 string is empty") {
			t.Error("Should return error for empty base64 (Linux)")
		}
	})

	// 场景4：含特殊字符的路径（Linux非法字符）
	t.Run("path with special chars (Linux)", func(t *testing.T) {
		destPath := testTempDir
		// Linux非法字符：空字符\x00
		url := "http://example.com/test\x00file.txt"
		rawContent := []byte("test special chars (Linux)")
		base64Str := base64.StdEncoding.EncodeToString(rawContent)

		_, err := JoinPrefixAndUrlFileNameAndWriteBase64ToFile(destPath, url, base64Str)
		if err == nil || !strings.Contains(err.Error(), "contains special chars") {
			t.Error("Should return error for path with special chars (Linux)")
		}
	})

	// 临界值：Linux超长文件名
	t.Run("long filename (Linux)", func(t *testing.T) {
		destPath := testTempDir
		longFileName := strings.Repeat("a", FileNameMaxLength-4) // 预留.txt后缀
		url := "http://example.com/" + longFileName + ".txt"
		rawContent := []byte("test long filename (Linux)")
		base64Str := base64.StdEncoding.EncodeToString(rawContent)

		filePathName, err := JoinPrefixAndUrlFileNameAndWriteBase64ToFile(destPath, url, base64Str)
		if err != nil {
			t.Fatalf("JoinPrefixAndUrlFileNameAndWriteBase64ToFile long filename failed (Linux): %v", err)
		}

		// 验证文件名长度
		_, fileName := filepath.Split(filePathName)
		if len(fileName) > FileNameMaxLength {
			t.Errorf("Filename exceeds max length (Linux): %d > %d", len(fileName), FileNameMaxLength)
		}
	})
}

func TestCopy(t *testing.T) {
	// 场景1：正常拷贝（Linux）
	srcFile := getTestFile("copy_src.txt")
	dstFile := getTestFile("copy_dst.txt")
	content := []byte("test copy function 123456 (Linux)")
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("Failed to write src file (Linux): %v", err)
	}

	err := Copy(srcFile, dstFile)
	if err != nil {
		t.Fatalf("Copy failed (Linux): %v", err)
	}

	readBytes, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read dst file (Linux): %v", err)
	}
	if string(readBytes) != string(content) {
		t.Errorf("Copy content error (Linux): got %s", string(readBytes))
	}

	// 场景2：拷贝到非pwd路径（Linux）
	nonPwdDst := "/tmp/copy_non_pwd_linux.txt"
	err = Copy(srcFile, nonPwdDst)
	if err != nil {
		t.Fatalf("Copy to non pwd path failed (Linux): %v", err)
	}
	defer os.Remove(nonPwdDst) // 清理

	// 场景3：路径遍历攻击（含../，Linux）
	err = Copy(srcFile, "../../copy_attack_linux.txt")
	if err == nil || !strings.Contains(err.Error(), "contains ../") {
		t.Error("Copy to invalid path should return error (Linux)")
	}

	// 场景4：空参数（Linux）
	err = Copy("", dstFile)
	if err == nil || !strings.Contains(err.Error(), "source file path is empty") {
		t.Error("Copy empty src should return error (Linux)")
	}

	err = Copy(srcFile, "")
	if err == nil || !strings.Contains(err.Error(), "destination file path is empty") {
		t.Error("Copy empty dst should return error (Linux)")
	}

	// 场景5：源文件不存在（Linux）
	err = Copy(getTestFile("not_exists_src.txt"), dstFile)
	if err == nil {
		t.Error("Copy not exists src should return error (Linux)")
	}

	// 临界值：Linux大文件拷贝（50MB）
	bigSrc := getTestFile("big_copy_src.txt")
	bigDst := getTestFile("big_copy_dst.txt")
	bigContent := generateRandomBytes(50 * 1024 * 1024)
	if err := os.WriteFile(bigSrc, bigContent, 0644); err != nil {
		t.Fatalf("Failed to write big src file (Linux): %v", err)
	}
	err = Copy(bigSrc, bigDst)
	if err != nil {
		t.Fatalf("Copy big file failed (Linux): %v", err)
	}
	bigReadBytes, err := os.ReadFile(bigDst)
	if err != nil {
		t.Fatalf("Failed to read big dst file (Linux): %v", err)
	}
	if len(bigReadBytes) != len(bigContent) {
		t.Errorf("Copy big file length error (Linux), got: %d, expected: %d", len(bigReadBytes), len(bigContent))
	}
}

// ------------------------------ 辅助函数测试 ------------------------------
func TestIsPathInDir(t *testing.T) {
	// 提前创建测试子目录（Linux）
	subDir := filepath.Join(testTempDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir (Linux): %v", err)
	}

	// 场景1：正常子目录（Linux）
	file := filepath.Join(subDir, "test.txt")
	os.WriteFile(file, []byte("test (Linux)"), 0644)
	inDir, err := isPathInDir(file, testTempDir)
	if err != nil {
		t.Fatalf("isPathInDir failed (Linux): %v", err)
	}
	if !inDir {
		t.Error("File in subdir should return true (Linux)")
	}

	// 场景2：文件等于目录（Linux）
	inDir, err = isPathInDir(testTempDir, testTempDir)
	if err != nil {
		t.Fatalf("isPathInDir failed (Linux): %v", err)
	}
	if !inDir {
		t.Error("File equals dir should return true (Linux)")
	}

	// 场景3：路径遍历（../，Linux）
	file = filepath.Join(testTempDir, "../../etc/passwd")
	inDir, err = isPathInDir(file, testTempDir)
	if err != nil {
		t.Fatalf("isPathInDir failed (Linux): %v", err)
	}
	if inDir {
		t.Error("Path traversal should return false (Linux)")
	}

	// 场景4：含/./的路径（Linux）
	subDir2 := filepath.Join(testTempDir, "./subdir2")
	os.MkdirAll(subDir2, 0755)
	file = filepath.Join(subDir2, "test2.txt")
	os.WriteFile(file, []byte("test2 (Linux)"), 0644)
	inDir, err = isPathInDir(file, testTempDir)
	if err != nil {
		t.Fatalf("isPathInDir failed (Linux): %v", err)
	}
	if !inDir {
		t.Error("Path with ./ should return true (Linux)")
	}
}

// ------------------------------ Linux性能测试 ------------------------------
func BenchmarkReadFileToBytes(b *testing.B) {
	// Linux 1MB测试文件
	testFile := getTestFile("bench_read_bytes.txt")
	content := generateRandomBytes(1 * 1024 * 1024)
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		b.Fatalf("Failed to write bench file (Linux): %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ReadFileToBytes(testFile)
		if err != nil {
			b.Fatalf("BenchmarkReadFileToBytes failed (Linux): %v", err)
		}
	}
}

func BenchmarkWriteBytesToFile(b *testing.B) {
	// Linux 1MB测试文件
	testFile := getTestFile("bench_write_bytes.txt")
	content := generateRandomBytes(1 * 1024 * 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := WriteBytesToFile(testFile, content)
		if err != nil {
			b.Fatalf("BenchmarkWriteBytesToFile failed (Linux): %v", err)
		}
	}
}

func BenchmarkIsFileDiffWithBase64(b *testing.B) {
	// Linux 1MB测试文件
	testFile := getTestFile("bench_diff.txt")
	content := generateRandomBytes(1 * 1024 * 1024)
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		b.Fatalf("Failed to write bench file (Linux): %v", err)
	}
	base64Str := base64.StdEncoding.EncodeToString(content)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := IsFileDiffWithBase64(testFile, base64Str)
		if err != nil {
			b.Fatalf("BenchmarkIsFileDiffWithBase64 failed (Linux): %v", err)
		}
	}
}

func BenchmarkJoinPrefixAndUrlFileNameAndWriteBase64ToFile(b *testing.B) {
	// Linux 1MB测试内容
	destPath := testTempDir
	url := "http://example.com/bench_join.txt"
	content := generateRandomBytes(1 * 1024 * 1024)
	base64Str := base64.StdEncoding.EncodeToString(content)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := JoinPrefixAndUrlFileNameAndWriteBase64ToFile(destPath, url, base64Str)
		if err != nil {
			b.Fatalf("BenchmarkJoinPrefix failed (Linux): %v", err)
		}
	}
}

func BenchmarkCopy(b *testing.B) {
	// Linux 1MB测试文件
	srcFile := getTestFile("bench_copy_src.txt")
	dstFile := getTestFile("bench_copy_dst.txt")
	content := generateRandomBytes(1 * 1024 * 1024)
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		b.Fatalf("Failed to write bench src file (Linux): %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := Copy(srcFile, dstFile)
		if err != nil {
			b.Fatalf("BenchmarkCopy failed (Linux): %v", err)
		}
	}
}

//////////////////////////////////////////////////////////

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
