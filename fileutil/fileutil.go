package fileutil

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cpusoft/goutil/base64util"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/urlutil"
)

// 统一常量（仅适配Linux）
const (
	FileNameMaxLength = 256
	PathNameMaxLength = 4096
	FileModeReadWrite = 0600
	FileModeAppend    = 0600
)

// ------------------------------ 基础文件操作函数 ------------------------------
// ReadFileToLines 读取文件内容到字符串切片（按行分割）
func ReadFileToLines(file string) (lines []string, err error) {
	if file == "" {
		return nil, errors.New("file path is empty")
	}

	fi, err := os.Open(file)
	if err != nil {
		belogs.Error("ReadFileToLines(): open file fail:", file, err)
		return nil, err
	}
	defer fi.Close()

	buf := bufio.NewReader(fi)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if line != "" {
					line = strings.TrimSpace(line)
					lines = append(lines, line)
				}
				break
			} else {
				belogs.Error("ReadFileToLines(): ReadString file fail:", file, err)
				return nil, err
			}
		}
		line = strings.TrimSpace(line)
		lines = append(lines, line)
	}
	return lines, nil
}

// ReadFileToBytes 读取文件内容到字节切片（完整读取）
func ReadFileToBytes(file string) (bytes []byte, err error) {
	if file == "" {
		return nil, errors.New("file path is empty")
	}
	fi, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fi.Close()
	return io.ReadAll(fi)
}

// ReadFileAndDecodeCertBase64 读取文件并解码证书格式的Base64内容
func ReadFileAndDecodeCertBase64(file string) (fileByte []byte, fileDecodeBase64Byte []byte, err error) {
	if file == "" {
		return nil, nil, errors.New("file path is empty")
	}

	belogs.Debug("ReadFileAndDecodeCertBase64(): file:", file)
	fileByte, err = os.ReadFile(file)
	if err != nil {
		belogs.Error("ReadFileAndDecodeCertBase64(): ReadFile err:", file, err)
		return nil, nil, err
	}
	if len(fileByte) == 0 {
		belogs.Error("ReadFileAndDecodeCertBase64(): fileByte is empty:", file)
		return nil, nil, errors.New("file is empty: " + file)
	}
	fileDecodeBase64Byte, err = base64util.DecodeCertBase64(fileByte)
	if err != nil {
		belogs.Error("ReadFileAndDecodeCertBase64(): DecodeCertBase64 err:", file, err)
		return nil, nil, err
	}
	return fileByte, fileDecodeBase64Byte, nil
}

// WriteBytesToFile 将字节切片写入文件（覆盖写入）
func WriteBytesToFile(file string, bytes []byte) (err error) {
	if file == "" {
		return errors.New("file path is empty")
	}
	if len(bytes) == 0 {
		return errors.New("bytes to write is empty")
	}

	if _, err := os.Stat(file); err == nil {
		belogs.Warn("WriteBytesToFile(): file already exists, will overwrite:", file)
	}

	return os.WriteFile(file, bytes, FileModeReadWrite)
}

// WriteBytesAppendFile 将字节切片追加写入文件
func WriteBytesAppendFile(file string, bytes []byte) (err error) {
	if file == "" {
		return errors.New("file path is empty")
	}
	if len(bytes) == 0 {
		return errors.New("bytes to write is empty")
	}

	fd, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, FileModeAppend)
	if err != nil {
		belogs.Error("WriteBytesAppendFile(): OpenFile fail, file:", file, err)
		return err
	}
	defer fd.Close()
	_, err = fd.Write(bytes)
	return err
}

// GetFileLength 获取文件长度（字节数）
func GetFileLength(file string) (length int64, err error) {
	if file == "" {
		return 0, errors.New("file path is empty")
	}
	f, err := os.Stat(file)
	if err != nil {
		return 0, err
	}
	return f.Size(), nil
}

// ------------------------------ 长度校验函数 ------------------------------
func CheckFileNameMaxLength(fileName string) bool {
	if len(fileName) > 0 && len(fileName) <= FileNameMaxLength {
		return true
	}
	return false
}

func CheckPathNameMaxLength(pathName string) bool {
	if len(pathName) > 0 && len(pathName) <= PathNameMaxLength {
		return true
	}
	return false
}

// ------------------------------ Base64 相关函数 ------------------------------
// WriteBase64ToFile 将Base64字符串解码后写入文件
func WriteBase64ToFile(filePathName, base64 string) (err error) {
	if filePathName == "" {
		return errors.New("file path is empty")
	}
	if base64 == "" {
		return errors.New("base64 string is empty")
	}

	belogs.Debug("WriteBase64ToFile(): filePathName:", filePathName, " len(base64):", len(base64))
	bytes, err := base64util.DecodeBase64(strings.TrimSpace(base64))
	if err != nil {
		belogs.Error("WriteBase64ToFile(): DecodeBase64 fail, base64:", base64, err)
		return err
	}

	belogs.Debug("WriteBase64ToFile(): DecodeBase64, filePathName:", filePathName, " len(bytes):", len(bytes))
	err = WriteBytesToFile(filePathName, bytes)
	if err != nil {
		belogs.Error("WriteBase64ToFile(): WriteBytesToFile fail:", filePathName, "  len(bytes):", len(bytes), err)
		return err
	}
	belogs.Debug("WriteBase64ToFile(): save filePathName ", filePathName, "  ok")
	return nil
}

// CreateAndWriteBase64ToFile 创建目录并将Base64内容写入文件
func CreateAndWriteBase64ToFile(filePathName, base64 string) (err error) {
	if filePathName == "" {
		return errors.New("file path is empty")
	}
	if base64 == "" {
		return errors.New("base64 string is empty")
	}

	filePath, _ := osutil.Split(filePathName)
	belogs.Debug("CreateAndWriteBase64ToFile(): Split filePathName:", filePathName, " len(base64):", len(base64))
	err = os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		belogs.Error("CreateAndWriteBase64ToFile(): MkdirAll fail, filePathName:", filePathName, err)
		return err
	}
	return WriteBase64ToFile(filePathName, base64)
}

// IsFileDiffWithBase64 对比文件内容与Base64解码后内容是否不同（分块哈希，优化大文件）
func IsFileDiffWithBase64(filePathName, base64 string) (bool, error) {
	// 参数校验
	if filePathName == "" {
		return false, errors.New("file path is empty")
	}
	if base64 == "" {
		return false, errors.New("base64 string is empty")
	}

	// 分块计算文件SHA256哈希（避免读取整个大文件到内存）
	fileHash, err := calculateFileHashChunked(filePathName)
	if err != nil {
		return false, err
	}

	// 解码Base64并计算哈希
	decodedBytes, err := base64util.DecodeBase64(strings.TrimSpace(base64))
	if err != nil {
		return false, err
	}
	dataHash := sha256.Sum256(decodedBytes)

	// 对比哈希值：不同返回true，相同返回false
	return !bytes.Equal(fileHash[:], dataHash[:]), nil
}

// calculateFileHashChunked 分块计算文件SHA256哈希（辅助函数）
func calculateFileHashChunked(filePathName string) ([32]byte, error) {
	file, err := os.Open(filePathName)
	if err != nil {
		return [32]byte{}, err
	}
	defer file.Close()

	hash := sha256.New()
	buf := make([]byte, 4096) // 4KB分块，平衡性能与内存
	for {
		n, err := file.Read(buf)
		if n > 0 {
			hash.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return [32]byte{}, err
		}
	}
	// 将哈希结果转为固定长度数组
	return *(*[32]byte)(hash.Sum(nil)), nil
}

// ------------------------------ 路径拼接与拷贝函数 ------------------------------
// JoinPrefixAndUrlFileNameAndWriteBase64ToFile 拼接路径并写入Base64内容（防路径遍历，仅Linux）
func JoinPrefixAndUrlFileNameAndWriteBase64ToFile(destPath, url, base64 string) (filePathName string, err error) {
	// 参数校验
	if destPath == "" {
		return "", errors.New("destination path is empty")
	}
	if url == "" {
		return "", errors.New("url is empty")
	}
	if base64 == "" {
		return "", errors.New("base64 string is empty")
	}

	belogs.Debug("JoinPrefixAndUrlFileNameAndWriteBase64ToFile(): destPath:", destPath, "  url:", url)

	filePathName, err = urlutil.JoinPrefixPathAndUrlFileName(destPath, url)
	if err != nil {
		belogs.Error("JoinPrefixAndUrlFileNameAndWriteBase64ToFile(): JoinPrefixPathAndUrlFileName fail, destPath:", destPath,
			"  url:", url, err)
		return "", err
	}

	// 核心：路径越界校验（仅Linux）
	inDir, err := isPathInDir(filePathName, destPath)
	if err != nil {
		return "", errors.New("invalid file path (path traversal detected): " + filePathName)
	}
	if !inDir {
		return "", errors.New("invalid file path (path traversal detected or not in whitelist): " + filePathName)
	}

	// 非法字符校验（仅Linux）
	cleanPath := filepath.Clean(filePathName)
	for _, c := range cleanPath {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '_' || c == '.' || c == '/' || c == '-') { // 仅Linux合法字符
			return "", errors.New("invalid file path (contains special chars): " + filePathName)
		}
	}

	err = CreateAndWriteBase64ToFile(filePathName, base64)
	if err != nil {
		belogs.Error("JoinPrefixAndUrlFileNameAndWriteBase64ToFile(): CreateAndWriteBase64ToFile fail, filePathName:", filePathName, err)
		return "", err
	}
	return filePathName, nil
}

// Copy 拷贝文件（防路径遍历，仅Linux）
func Copy(srcFilePathName, dstFilePathName string) error {
	// 参数校验
	if srcFilePathName == "" {
		return errors.New("source file path is empty")
	}
	if dstFilePathName == "" {
		return errors.New("destination file path is empty")
	}

	// 仅校验路径中是否含../（防路径遍历，Linux）
	if strings.Contains(srcFilePathName, ".."+string(filepath.Separator)) ||
		strings.Contains(dstFilePathName, ".."+string(filepath.Separator)) ||
		strings.Contains(srcFilePathName, "../") ||
		strings.Contains(dstFilePathName, "../") {
		return errors.New("invalid path (contains ../): " + dstFilePathName)
	}

	belogs.Debug("Copy(): srcFilePathName:", srcFilePathName, "  dstFilePathName:", dstFilePathName)
	srcFile, err := os.Open(srcFilePathName)
	if err != nil {
		belogs.Error("Copy(): Open fail, srcFilePathName:", srcFilePathName, err)
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstFilePathName)
	if err != nil {
		belogs.Error("Copy(): Create fail, dstFilePathName:", dstFilePathName, err)
		return err
	}
	defer func() {
		closeErr := dstFile.Close()
		if closeErr != nil {
			belogs.Error("Copy(): Close dstFile fail, dstFilePathName:", dstFilePathName, closeErr)
		}
	}()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		belogs.Error("Copy(): Copy fail, src:", srcFilePathName, "  dst:", dstFilePathName, err)
		return err
	}
	return nil
}

// ------------------------------ 辅助函数 ------------------------------
// isPathInDir 判断文件路径是否在指定目录内（仅Linux）
func isPathInDir(filePath, dirPath string) (bool, error) {
	// 1. 转换为绝对路径（Linux）
	absDir, err := filepath.Abs(dirPath)
	if err != nil {
		return false, err
	}
	absFile, err := filepath.Abs(filePath)
	if err != nil {
		return false, err
	}

	// 2. 清理路径（Linux）
	absDir = filepath.Clean(absDir)
	absFile = filepath.Clean(absFile)

	// 3. Linux固定用/拼接前缀
	absDirWithSep := absDir + "/"

	// 4. 前缀匹配
	if strings.HasPrefix(absFile, absDirWithSep) {
		return true, nil
	}
	// 兼容文件路径等于目录路径的情况
	return absFile == absDir, nil
}
