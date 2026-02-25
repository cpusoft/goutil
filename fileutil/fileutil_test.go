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

const testTempDir = "./test_temp_fileutil_linux"

func TestMain(m *testing.M) {
	if err := os.MkdirAll(testTempDir, 0755); err != nil {
		panic("Failed to create test temp dir (Linux): " + err.Error())
	}

	code := m.Run()

	if err := os.RemoveAll(testTempDir); err != nil {
		panic("Failed to clean test temp dir (Linux): " + err.Error())
	}

	os.Exit(code)
}

func getTestFile(name string) string {
	return filepath.Join(testTempDir, name)
}

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

	_, err = ReadFileToLines("")
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("ReadFileToLines empty param should return error (Linux)")
	}

	_, err = ReadFileToLines(getTestFile("not_exists_lines.txt"))
	if err == nil {
		t.Error("ReadFileToLines not exists file should return error (Linux)")
	}
}

func TestReadFileToBytes(t *testing.T) {
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

	_, err = ReadFileToBytes("")
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("ReadFileToBytes empty param should return error (Linux)")
	}

	_, err = ReadFileToBytes(getTestFile("not_exists_bytes.txt"))
	if err == nil {
		t.Error("ReadFileToBytes not exists file should return error (Linux)")
	}

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

	emptyFile := getTestFile("empty_cert.txt")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write empty test file (Linux): %v", err)
	}
	_, _, err = ReadFileAndDecodeCertBase64(emptyFile)
	if err == nil || !strings.Contains(err.Error(), "file is empty") {
		t.Error("ReadFileAndDecodeCertBase64 empty file should return error (Linux)")
	}

	invalidFile := getTestFile("invalid_cert_base64.txt")
	if err := os.WriteFile(invalidFile, []byte("invalid base64 !@#$%^ (Linux)"), 0644); err != nil {
		t.Fatalf("Failed to write invalid test file (Linux): %v", err)
	}
	_, _, err = ReadFileAndDecodeCertBase64(invalidFile)
	if err == nil {
		t.Error("ReadFileAndDecodeCertBase64 invalid base64 should return error (Linux)")
	}

	_, _, err = ReadFileAndDecodeCertBase64("")
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("ReadFileAndDecodeCertBase64 empty param should return error (Linux)")
	}
}

func TestWriteBytesToFile(t *testing.T) {
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

	overwriteContent := []byte("overwrite content (Linux)")
	err = WriteBytesToFile(testFile, overwriteContent)
	if err != nil {
		t.Fatalf("WriteBytesToFile overwrite failed (Linux): %v", err)
	}
	readBytes, err = os.ReadFile(testFile)
	if string(readBytes) != string(overwriteContent) {
		t.Errorf("WriteBytesToFile overwrite error (Linux), got: %s", string(readBytes))
	}

	err = WriteBytesToFile("", content)
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("WriteBytesToFile empty path should return error (Linux)")
	}

	err = WriteBytesToFile(getTestFile("empty_bytes.txt"), []byte(""))
	if err == nil || !strings.Contains(err.Error(), "bytes to write is empty") {
		t.Error("WriteBytesToFile empty bytes should return error (Linux)")
	}

	// 修复：超长路径改为多层目录+合法长度文件名
	longDirPath := filepath.Join(testTempDir, strings.Repeat("subdir_", 20)) // 20层目录
	if err := os.MkdirAll(longDirPath, 0755); err != nil {
		t.Fatalf("Failed to create long dir path (Linux): %v", err)
	}
	// 精确计算：240个a + .txt = 240+4=244 ≤255
	longFileName := strings.Repeat("a", 240) + ".txt"
	longPath := filepath.Join(longDirPath, longFileName)
	err = WriteBytesToFile(longPath, []byte("long path test (Linux)"))
	if err != nil {
		t.Fatalf("WriteBytesToFile long path failed (Linux): %v", err)
	}

	// 超过单个文件名长度（预期失败）
	tooLongFileName := strings.Repeat("b", 252) + ".txt" // 256字符
	tooLongPath := filepath.Join(longDirPath, tooLongFileName)
	err = WriteBytesToFile(tooLongPath, []byte("too long path (Linux)"))
	if err == nil {
		t.Error("WriteBytesToFile too long filename should return error (Linux)")
	}
}

func TestWriteBytesAppendFile(t *testing.T) {
	testFile := getTestFile("append_test.txt")
	if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to init append test file (Linux): %v", err)
	}

	content1 := []byte("first line (Linux)\n")
	err := WriteBytesAppendFile(testFile, content1)
	if err != nil {
		t.Fatalf("WriteBytesAppendFile first write failed (Linux): %v", err)
	}

	content2 := []byte("second line (Linux)")
	err = WriteBytesAppendFile(testFile, content2)
	if err != nil {
		t.Fatalf("WriteBytesAppendFile append failed (Linux): %v", err)
	}

	readBytes, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read append file (Linux): %v", err)
	}
	expected := string(content1) + string(content2)
	if string(readBytes) != expected {
		t.Errorf("WriteBytesAppendFile result error (Linux), got: %s, expected: %s", string(readBytes), expected)
	}

	err = WriteBytesAppendFile("", content1)
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("WriteBytesAppendFile empty path should return error (Linux)")
	}

	err = WriteBytesAppendFile(getTestFile("empty_append.txt"), []byte(""))
	if err == nil || !strings.Contains(err.Error(), "bytes to write is empty") {
		t.Error("WriteBytesAppendFile empty bytes should return error (Linux)")
	}
}

func TestGetFileLength(t *testing.T) {
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

	_, err = GetFileLength("")
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("GetFileLength empty param should return error (Linux)")
	}

	_, err = GetFileLength(getTestFile("not_exists_length.txt"))
	if err == nil {
		t.Error("GetFileLength not exists file should return error (Linux)")
	}

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
	normalName := strings.Repeat("a", 100)
	if !CheckFileNameMaxLength(normalName) {
		t.Error("CheckFileNameMaxLength normal name should return true (Linux)")
	}

	// 严格遵循255字符限制
	criticalName := strings.Repeat("b", FileNameMaxLength)
	if !CheckFileNameMaxLength(criticalName) {
		t.Error("CheckFileNameMaxLength 255 chars should return true (Linux)")
	}

	tooLongName := strings.Repeat("c", FileNameMaxLength+1)
	if CheckFileNameMaxLength(tooLongName) {
		t.Error("CheckFileNameMaxLength 256 chars should return false (Linux)")
	}

	if CheckFileNameMaxLength("") {
		t.Error("CheckFileNameMaxLength empty name should return false (Linux)")
	}
}

func TestCheckPathNameMaxLength(t *testing.T) {
	normalPath := "/" + strings.Repeat("a", 999)
	if !CheckPathNameMaxLength(normalPath) {
		t.Error("CheckPathNameMaxLength normal path should return true (Linux)")
	}

	criticalPath := "/" + strings.Repeat("b", PathNameMaxLength-1)
	if !CheckPathNameMaxLength(criticalPath) {
		t.Error("CheckPathNameMaxLength 4096 chars should return true (Linux)")
	}

	tooLongPath := "/" + strings.Repeat("c", PathNameMaxLength)
	if CheckPathNameMaxLength(tooLongPath) {
		t.Error("CheckPathNameMaxLength 4097 chars should return false (Linux)")
	}

	if CheckPathNameMaxLength("") {
		t.Error("CheckPathNameMaxLength empty path should return false (Linux)")
	}
}

// ------------------------------ Base64 相关函数测试 ------------------------------
func TestWriteBase64ToFile(t *testing.T) {
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

	err = WriteBase64ToFile("", base64Str)
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("WriteBase64ToFile empty path should return error (Linux)")
	}

	err = WriteBase64ToFile(getTestFile("empty_base64.txt"), "")
	if err == nil || !strings.Contains(err.Error(), "base64 string is empty") {
		t.Error("WriteBase64ToFile empty base64 should return error (Linux)")
	}

	err = WriteBase64ToFile(getTestFile("invalid_base64.txt"), "invalid!@#$% (Linux)")
	if err == nil {
		t.Error("WriteBase64ToFile invalid base64 should return error (Linux)")
	}

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
	testFile := getTestFile("subdir1/subdir2/create_base64.txt")
	rawContent := []byte("test create dir and write base64 (Linux)")
	base64Str := base64.StdEncoding.EncodeToString(rawContent)

	err := CreateAndWriteBase64ToFile(testFile, base64Str)
	if err != nil {
		t.Fatalf("CreateAndWriteBase64ToFile failed (Linux): %v", err)
	}

	if _, err := os.Stat(filepath.Dir(testFile)); err != nil {
		t.Error("CreateAndWriteBase64ToFile dir create failed (Linux)")
	}

	readBytes, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read create base64 file (Linux): %v", err)
	}
	if string(readBytes) != string(rawContent) {
		t.Errorf("CreateAndWriteBase64ToFile content error (Linux), got: %s", string(readBytes))
	}

	err = CreateAndWriteBase64ToFile("", base64Str)
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("CreateAndWriteBase64ToFile empty path should return error (Linux)")
	}

	err = CreateAndWriteBase64ToFile(getTestFile("empty_create_base64.txt"), "")
	if err == nil || !strings.Contains(err.Error(), "base64 string is empty") {
		t.Error("CreateAndWriteBase64ToFile empty base64 should return error (Linux)")
	}

	// 最终修复：精确计算文件名长度（240 + 14 = 254 ≤255）
	longDir := filepath.Join(testTempDir, strings.Repeat("subdir_", 20))
	// 240个a + "_long.txt" = 240+9=249 ≤255
	longFileName := strings.Repeat("a", 240) + "_long.txt"
	longPath := filepath.Join(longDir, longFileName)
	err = CreateAndWriteBase64ToFile(longPath, base64Str)
	if err != nil {
		t.Fatalf("CreateAndWriteBase64ToFile long path failed (Linux): %v", err)
	}
}

func TestIsFileDiffWithBase64(t *testing.T) {
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

	base64Diff := base64.StdEncoding.EncodeToString([]byte("different content (Linux)"))
	diff, err = IsFileDiffWithBase64(testFile, base64Diff)
	if err != nil {
		t.Fatalf("IsFileDiffWithBase64 failed (Linux): %v", err)
	}
	if !diff {
		t.Error("IsFileDiffWithBase64 different content should return true (Linux)")
	}

	_, err = IsFileDiffWithBase64("", base64Same)
	if err == nil || !strings.Contains(err.Error(), "file path is empty") {
		t.Error("IsFileDiffWithBase64 empty file path should return error (Linux)")
	}

	_, err = IsFileDiffWithBase64(testFile, "")
	if err == nil || !strings.Contains(err.Error(), "base64 string is empty") {
		t.Error("IsFileDiffWithBase64 empty base64 should return error (Linux)")
	}

	_, err = IsFileDiffWithBase64(getTestFile("not_exists_diff.txt"), base64Same)
	if err == nil {
		t.Error("IsFileDiffWithBase64 not exists file should return error (Linux)")
	}

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
	// 最终修复：正常路径测试（彻底清理/./）
	t.Run("normal path with ./ (Linux)", func(t *testing.T) {
		destPath := testTempDir
		url := "http://example.com/test_normal.txt"
		rawContent := []byte("test join path with ./ symbol 123 (Linux)")
		base64Str := base64.StdEncoding.EncodeToString(rawContent)

		filePathName, err := JoinPrefixAndUrlFileNameAndWriteBase64ToFile(destPath, url, base64Str)
		if err != nil {
			t.Fatalf("JoinPrefixAndUrlFileNameAndWriteBase64ToFile normal path failed (Linux): %v", err)
		}

		readBytes, err := ReadFileToBytes(filePathName)
		if err != nil {
			t.Fatalf("Failed to read written file (Linux): %v", err)
		}
		if string(readBytes) != string(rawContent) {
			t.Errorf("File content mismatch (Linux): got %s, want %s", string(readBytes), string(rawContent))
		}

		inDir, err := isPathInDir(filePathName, destPath)
		if err != nil || !inDir {
			t.Error("File path should be in dest dir (Linux)")
		}
	})

	// 修复：路径遍历测试（兼容中文错误提示）
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
		} else {
			if !strings.Contains(err.Error(), "path traversal detected") &&
				!strings.Contains(err.Error(), "Path校验失败") {
				t.Errorf("Error message should contain 'path traversal' or 'Path校验失败' (Linux), but got: %v", err)
			}
		}
	})

	t.Run("empty params (Linux)", func(t *testing.T) {
		rawContent := []byte("test empty params (Linux)")
		base64Str := base64.StdEncoding.EncodeToString(rawContent)

		_, err := JoinPrefixAndUrlFileNameAndWriteBase64ToFile("", "http://example.com/test.txt", base64Str)
		if err == nil || !strings.Contains(err.Error(), "destination path is empty") {
			t.Error("Should return error for empty dest path (Linux)")
		}

		_, err = JoinPrefixAndUrlFileNameAndWriteBase64ToFile(testTempDir, "", base64Str)
		if err == nil || !strings.Contains(err.Error(), "url is empty") {
			t.Error("Should return error for empty url (Linux)")
		}

		_, err = JoinPrefixAndUrlFileNameAndWriteBase64ToFile(testTempDir, "http://example.com/test.txt", "")
		if err == nil || !strings.Contains(err.Error(), "base64 string is empty") {
			t.Error("Should return error for empty base64 (Linux)")
		}
	})

	// 最终修复：特殊字符测试（使用*，不会被URL编码）
	t.Run("path with special chars (Linux)", func(t *testing.T) {
		destPath := testTempDir
		// 使用*作为非法字符（Linux禁止，且URL解析不会编码）
		url := "http://example.com/test*file.txt"
		rawContent := []byte("test special chars (Linux)")
		base64Str := base64.StdEncoding.EncodeToString(rawContent)

		_, err := JoinPrefixAndUrlFileNameAndWriteBase64ToFile(destPath, url, base64Str)
		if err == nil || !strings.Contains(err.Error(), "contains special chars") {
			t.Error("Should return error for path with special chars (Linux)")
		}
	})

	// 最终修复：超长文件名测试（240+4=244 ≤255）
	t.Run("long filename (Linux)", func(t *testing.T) {
		destPath := testTempDir
		// 240个a + .txt = 244 ≤255
		longFileName := strings.Repeat("a", 240) + ".txt"
		url := "http://example.com/" + longFileName
		rawContent := []byte("test long filename (Linux)")
		base64Str := base64.StdEncoding.EncodeToString(rawContent)

		filePathName, err := JoinPrefixAndUrlFileNameAndWriteBase64ToFile(destPath, url, base64Str)
		if err != nil {
			t.Fatalf("JoinPrefixAndUrlFileNameAndWriteBase64ToFile long filename failed (Linux): %v", err)
		}

		_, fileName := filepath.Split(filePathName)
		if len(fileName) > FileNameMaxLength {
			t.Errorf("Filename exceeds max length (Linux): %d > %d", len(fileName), FileNameMaxLength)
		}
	})
}

func TestCopy(t *testing.T) {
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

	nonPwdDst := "/tmp/copy_non_pwd_linux.txt"
	err = Copy(srcFile, nonPwdDst)
	if err != nil {
		t.Fatalf("Copy to non pwd path failed (Linux): %v", err)
	}
	defer os.Remove(nonPwdDst)

	err = Copy(srcFile, "../../copy_attack_linux.txt")
	if err == nil || !strings.Contains(err.Error(), "contains ../") {
		t.Error("Copy to invalid path should return error (Linux)")
	}

	err = Copy("", dstFile)
	if err == nil || !strings.Contains(err.Error(), "source file path is empty") {
		t.Error("Copy empty src should return error (Linux)")
	}

	err = Copy(srcFile, "")
	if err == nil || !strings.Contains(err.Error(), "destination file path is empty") {
		t.Error("Copy empty dst should return error (Linux)")
	}

	err = Copy(getTestFile("not_exists_src.txt"), dstFile)
	if err == nil {
		t.Error("Copy not exists src should return error (Linux)")
	}

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
	subDir := filepath.Join(testTempDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir (Linux): %v", err)
	}

	file := filepath.Join(subDir, "test.txt")
	os.WriteFile(file, []byte("test (Linux)"), 0644)
	inDir, err := isPathInDir(file, testTempDir)
	if err != nil {
		t.Fatalf("isPathInDir failed (Linux): %v", err)
	}
	if !inDir {
		t.Error("File in subdir should return true (Linux)")
	}

	inDir, err = isPathInDir(testTempDir, testTempDir)
	if err != nil {
		t.Fatalf("isPathInDir failed (Linux): %v", err)
	}
	if !inDir {
		t.Error("File equals dir should return true (Linux)")
	}

	file = filepath.Join(testTempDir, "../../etc/passwd")
	inDir, err = isPathInDir(file, testTempDir)
	if err != nil {
		t.Fatalf("isPathInDir failed (Linux): %v", err)
	}
	if inDir {
		t.Error("Path traversal should return false (Linux)")
	}

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
