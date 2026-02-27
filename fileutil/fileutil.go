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
)

// Linux专属常量（严格遵循Linux限制）
const (
	FileNameMaxLength = 255  // Linux单个文件名硬限制
	PathNameMaxLength = 4096 // Linux路径总长度限制
	FileModeReadWrite = 0600
	FileModeAppend    = 0600
)

// ------------------------------ 基础文件操作函数 ------------------------------
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

func WriteBytesToFile(file string, bytes []byte) (err error) {
	if file == "" {
		return errors.New("file path is empty")
	}
	if len(bytes) == 0 {
		return errors.New("bytes to write is empty")
	}

	// 严格校验单个文件名长度
	_, fileName := filepath.Split(file)
	if len(fileName) > FileNameMaxLength {
		return errors.New("file name too long (Linux): " + fileName)
	}

	if _, err := os.Stat(file); err == nil {
		belogs.Info("WriteBytesToFile(): file already exists, will overwrite:", file)
	}

	return os.WriteFile(file, bytes, FileModeReadWrite)
}

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
func WriteBase64ToFile(filePathName, base64 string) (err error) {
	if filePathName == "" {
		return errors.New("file path is empty")
	}
	if base64 == "" {
		return errors.New("base64 string is empty")
	}

	// 严格校验单个文件名长度
	_, fileName := filepath.Split(filePathName)
	if len(fileName) > FileNameMaxLength {
		return errors.New("file name too long (Linux): " + fileName)
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

func CreateAndWriteBase64ToFile(filePathName, base64 string) (err error) {
	if filePathName == "" {
		return errors.New("file path is empty")
	}
	if base64 == "" {
		return errors.New("base64 string is empty")
	}

	// 严格拆分并校验文件名长度
	filePath, fileName := osutil.Split(filePathName)
	if len(fileName) > FileNameMaxLength {
		return errors.New("file name too long (Linux): " + fileName)
	}

	belogs.Debug("CreateAndWriteBase64ToFile(): Split filePathName:", filePathName, " len(base64):", len(base64))
	err = os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		belogs.Error("CreateAndWriteBase64ToFile(): MkdirAll fail, filePathName:", filePathName, err)
		return err
	}
	return WriteBase64ToFile(filePathName, base64)
}

func IsFileDiffWithBase64(filePathName, base64 string) (bool, error) {
	if filePathName == "" {
		return false, errors.New("file path is empty")
	}
	if base64 == "" {
		return false, errors.New("base64 string is empty")
	}

	fileHash, err := calculateFileHashChunked(filePathName)
	if err != nil {
		return false, err
	}

	decodedBytes, err := base64util.DecodeBase64(strings.TrimSpace(base64))
	if err != nil {
		return false, err
	}
	dataHash := sha256.Sum256(decodedBytes)

	return !bytes.Equal(fileHash[:], dataHash[:]), nil
}

func calculateFileHashChunked(filePathName string) ([32]byte, error) {
	file, err := os.Open(filePathName)
	if err != nil {
		return [32]byte{}, err
	}
	defer file.Close()

	hash := sha256.New()
	buf := make([]byte, 4096)
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
	return *(*[32]byte)(hash.Sum(nil)), nil
}

/*
// ------------------------------ 核心修复：JoinPrefixAndUrlFileNameAndWriteBase64ToFile ------------------------------
func JoinPrefixAndUrlFileNameAndWriteBase64ToFile(destPath, url, base64 string) (filePathName string, err error) {
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

	// 核心修复1：先将目标目录转为绝对路径（解决相对路径导致的前缀不匹配）
	absDestPath, err := filepath.Abs(destPath)
	if err != nil {
		return "", errors.New("failed to get absolute dest path: " + err.Error())
	}
	// 确保目标目录以路径分隔符结尾（统一匹配规则）
	absDestPath = filepath.Clean(absDestPath) + string(filepath.Separator)

	// 核心修复2：基于绝对路径拼接文件路径
	filePathName, err = urlutil.JoinPrefixPathAndUrlFileName(absDestPath, url)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "Path校验失败") || strings.Contains(errMsg, "path traversal") {
			return "", errors.New("invalid file path (path traversal detected or not in whitelist): " + filePathName)
		}
		belogs.Error("JoinPrefixAndUrlFileNameAndWriteBase64ToFile(): JoinPrefixPathAndUrlFileName fail, destPath:", absDestPath,
			"  url:", url, err)
		return "", err
	}

	// 核心修复3：强制转为绝对路径并清理（消除所有冗余符号）
	cleanedFilePath, err := filepath.Abs(filepath.Clean(filePathName))
	if err != nil {
		return "", errors.New("invalid file path (failed to get absolute path): " + filePathName)
	}

	// 核心修复4：严格前缀匹配（确保文件路径在目标目录内）
	if !strings.HasPrefix(cleanedFilePath, absDestPath) {
		return "", errors.New("invalid file path (path traversal detected or not in whitelist): " + cleanedFilePath)
	}

	// 特殊字符校验（保留原有逻辑）
	invalidChars := []rune{'\000', ':', '*', '?', '"', '<', '>', '|', ' ', '\t', '\n'}
	for _, c := range cleanedFilePath {
		if c == filepath.Separator {
			continue
		}
		for _, invalidChar := range invalidChars {
			if c == invalidChar {
				return "", errors.New("invalid file path (contains special chars): " + cleanedFilePath)
			}
		}
	}

	// 文件名长度校验（保留原有逻辑）
	_, fileName := filepath.Split(cleanedFilePath)
	if len(fileName) > FileNameMaxLength {
		return "", errors.New("file name too long (Linux): " + fileName)
	}

	// 使用清理后的绝对路径写入文件
	err = CreateAndWriteBase64ToFile(cleanedFilePath, base64)
	if err != nil {
		belogs.Error("JoinPrefixAndUrlFileNameAndWriteBase64ToFile(): CreateAndWriteBase64ToFile fail, filePathName:", cleanedFilePath, err)
		return "", err
	}
	return cleanedFilePath, nil
}
*/

func Copy(srcFilePathName, dstFilePathName string) error {
	if srcFilePathName == "" {
		return errors.New("source file path is empty")
	}
	if dstFilePathName == "" {
		return errors.New("destination file path is empty")
	}

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
// 最终修复：彻底解决/./路径匹配问题
// ------------------------------ 简化并修复 isPathInDir（备用，本次核心用直接前缀匹配） ------------------------------
func isPathInDir(filePath, dirPath string) (bool, error) {
	absDir, err := filepath.Abs(dirPath)
	if err != nil {
		return false, err
	}
	absFile, err := filepath.Abs(filePath)
	if err != nil {
		return false, err
	}

	absDir = filepath.Clean(absDir)
	absFile = filepath.Clean(absFile)

	absDirWithSep := absDir + string(filepath.Separator)
	return strings.HasPrefix(absFile, absDirWithSep) || absFile == absDir, nil
}
