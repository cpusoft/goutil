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

// Linux专属常量（单个文件名最大255字符，路径总长度4096）
const (
	FileNameMaxLength = 255  // Linux单个文件名限制（包含后缀）
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

	// 新增：Linux单个文件名长度校验
	_, fileName := filepath.Split(file)
	if len(fileName) > FileNameMaxLength {
		return errors.New("file name too long (Linux): " + fileName)
	}

	if _, err := os.Stat(file); err == nil {
		belogs.Warn("WriteBytesToFile(): file already exists, will overwrite:", file)
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
	// 修正：Linux单个文件名最大255字符
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

	// 新增：Linux单个文件名长度校验
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

	// 新增：拆分目录和文件名，避免超长文件名
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

// ------------------------------ 路径拼接与拷贝函数 ------------------------------
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

	filePathName, err = urlutil.JoinPrefixPathAndUrlFileName(destPath, url)
	if err != nil {
		// 修复：统一错误提示为英文，兼容中文提示
		errMsg := err.Error()
		if strings.Contains(errMsg, "Path校验失败") || strings.Contains(errMsg, "path traversal") {
			return "", errors.New("invalid file path (path traversal detected or not in whitelist): " + filePathName)
		}
		belogs.Error("JoinPrefixAndUrlFileNameAndWriteBase64ToFile(): JoinPrefixPathAndUrlFileName fail, destPath:", destPath,
			"  url:", url, err)
		return "", err
	}

	// 修复：路径校验逻辑，正确处理/./
	inDir, err := isPathInDir(filePathName, destPath)
	if err != nil {
		return "", errors.New("invalid file path (path traversal detected): " + filePathName)
	}
	if !inDir {
		return "", errors.New("invalid file path (path traversal detected or not in whitelist): " + filePathName)
	}

	// 修复：特殊字符校验（包含Linux非法字符）
	cleanPath := filepath.Clean(filePathName)
	invalidChars := []rune{'\000', '/', ':', '*', '?', '"', '<', '>', '|'} // Linux非法字符
	for _, c := range cleanPath {
		for _, invalidChar := range invalidChars {
			if c == invalidChar && c != '/' { // 保留/作为路径分隔符
				return "", errors.New("invalid file path (contains special chars): " + filePathName)
			}
		}
	}

	// 新增：单个文件名长度校验
	_, fileName := filepath.Split(filePathName)
	if len(fileName) > FileNameMaxLength {
		return "", errors.New("file name too long (Linux): " + fileName)
	}

	err = CreateAndWriteBase64ToFile(filePathName, base64)
	if err != nil {
		belogs.Error("JoinPrefixAndUrlFileNameAndWriteBase64ToFile(): CreateAndWriteBase64ToFile fail, filePathName:", filePathName, err)
		return "", err
	}
	return filePathName, nil
}

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
func isPathInDir(filePath, dirPath string) (bool, error) {
	// 修复：正确处理/./路径
	absDir, err := filepath.Abs(dirPath)
	if err != nil {
		return false, err
	}
	absFile, err := filepath.Abs(filePath)
	if err != nil {
		return false, err
	}

	// 修复：清理路径中的/./
	absDir = filepath.Clean(absDir)
	absFile = filepath.Clean(absFile)

	// 修复：处理目录路径末尾无/的情况
	absDirWithSep := absDir
	if !strings.HasSuffix(absDirWithSep, "/") {
		absDirWithSep += "/"
	}

	// 修复：前缀匹配逻辑
	if strings.HasPrefix(absFile, absDirWithSep) {
		return true, nil
	}
	return absFile == absDir, nil
}
