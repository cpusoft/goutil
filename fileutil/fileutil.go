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

// 常量定义（增加注释+权限常量）
const (
	// FILENAME_MAXLENGTH 仅适用于Linux系统（/usr/include/linux/limits.h）
	// 跨平台建议使用 filepath.MaxPathLength 或动态获取
	FILENAME_MAXLENGTH = 256
	PATHNAME_MAXLENGTH = 4096

	// 文件权限常量（按需调整，降低默认权限）
	FileModeReadWrite = 0600 // 仅当前用户可读写
	FileModeAppend    = 0600 // 追加写入权限
)

// 修复：ReadFileToLines 逻辑缺陷（错误时不添加不完整行）+ 参数校验
func ReadFileToLines(file string) (lines []string, err error) {
	// 增加参数校验
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
				// EOF时处理最后一行（如果非空）
				if line != "" {
					line = strings.TrimSpace(line)
					lines = append(lines, line)
				}
				break
			} else {
				belogs.Error("ReadFileToLines(): ReadString file fail:", file, err)
				return nil, err // 错误时直接返回，不添加当前行
			}
		}
		// 无错误时才添加行
		line = strings.TrimSpace(line)
		lines = append(lines, line)
	}
	return lines, nil
}

// 保留其他函数并补充参数校验
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

// 修复：ReadFileAndDecodeCertBase64 错误信息拼写 + 空文件判断 + 参数校验
func ReadFileAndDecodeCertBase64(file string) (fileByte []byte, fileDecodeBase64Byte []byte, err error) {
	// 参数校验
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

// 修复：文件权限 + 覆盖校验 + 参数校验
func WriteBytesToFile(file string, bytes []byte) (err error) {
	// 参数校验
	if file == "" {
		return errors.New("file path is empty")
	}
	if len(bytes) == 0 {
		return errors.New("bytes to write is empty")
	}

	// 检查文件是否存在，防止误覆盖（可通过参数控制是否覆盖）
	if _, err := os.Stat(file); err == nil {
		belogs.Warn("WriteBytesToFile(): file already exists, will overwrite:", file)
		// 如需禁止覆盖，可返回错误：
		// return errors.New("file already exists: " + file)
	}

	// 降低权限，仅当前用户可读写
	return os.WriteFile(file, bytes, FileModeReadWrite)
}

// 修复：拼写错误 + 参数校验（移除并发锁）
func WriteBytesAppendFile(file string, bytes []byte) (err error) {
	// 参数校验
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

func CheckFileNameMaxLength(fileName string) bool {
	if len(fileName) > 0 && len(fileName) <= FILENAME_MAXLENGTH {
		return true
	}
	return false
}

func CheckPathNameMaxLength(pathName string) bool {
	if len(pathName) > 0 && len(pathName) <= PATHNAME_MAXLENGTH {
		return true
	}
	return false
}

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

// 修复：IsFileDiffWithBase64 分块计算哈希（优化大文件性能）+ 参数校验
func IsFileDiffWithBase64(filePathName, base64 string) (bool, error) {
	// 参数校验
	if filePathName == "" {
		return false, errors.New("file path is empty")
	}
	if base64 == "" {
		return false, errors.New("base64 string is empty")
	}

	// 分块计算文件哈希（避免读取整个文件到内存）
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

// 分块计算文件SHA256哈希（优化大文件性能）
func calculateFileHashChunked(filePathName string) ([32]byte, error) {
	file, err := os.Open(filePathName)
	if err != nil {
		return [32]byte{}, err
	}
	defer file.Close()

	hash := sha256.New()
	buf := make([]byte, 4096) // 4KB分块
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

// 修复：JoinPrefixAndUrlFileNameAndWriteBase64ToFile 路径遍历防护 + 参数校验
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

	// 路径遍历防护：确保最终路径在destPath目录下
	absDestPath, err := filepath.Abs(destPath)
	if err != nil {
		return "", err
	}
	absFilePathName, err := filepath.Abs(filePathName)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(absFilePathName, absDestPath) {
		return "", errors.New("invalid file path (path traversal detected): " + filePathName)
	}

	err = CreateAndWriteBase64ToFile(filePathName, base64)
	if err != nil {
		belogs.Error("JoinPrefixAndUrlFileNameAndWriteBase64ToFile(): CreateAndWriteBase64ToFile fail, filePathName:", filePathName, err)
		return "", err
	}
	return filePathName, nil
}

// 修复：Copy 函数日志笔误 + 参数校验 + 关闭文件错误处理
func Copy(srcFilePathName, dstFilePathName string) error {
	// 参数校验
	if srcFilePathName == "" {
		return errors.New("source file path is empty")
	}
	if dstFilePathName == "" {
		return errors.New("destination file path is empty")
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
	// 确保dstFile正确关闭（即使Copy出错）
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
