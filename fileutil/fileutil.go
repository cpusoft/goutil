package fileutil

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/cpusoft/goutil/base64util"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/urlutil"
)

const (
	// /usr/include/linux/limits.h
	FILENAME_MAXLENGTH = 256
	PATHNAME_MAXLENGTH = 4096
)

func ReadFileToLines(file string) (lines []string, err error) {
	fi, err := os.Open(file)
	if err != nil {
		belogs.Error("ReadFileToLines(): open file fail:", file, err)
		return nil, err
	}
	defer fi.Close()

	buf := bufio.NewReader(fi)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		lines = append(lines, line)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				belogs.Error("ReadFileToLines(): ReadString file fail:", file, err)
				return nil, err
			}
		}

	}
	return lines, nil
}

func ReadFileToBytes(file string) (bytes []byte, err error) {
	fi, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fi.Close()
	return ioutil.ReadAll(fi)
}

// -rw-rw--r--
func WriteBytesToFile(file string, bytes []byte) (err error) {
	return ioutil.WriteFile(file, bytes, 0664)
}

func GetFileLength(file string) (length int64, err error) {
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
	bytes, err := base64util.DecodeBase64(strings.TrimSpace(base64))
	if err != nil {
		belogs.Error("WriteBase64ToFile(): DecodeBase64 fail, base64:", base64, err)
		return err
	}

	err = WriteBytesToFile(filePathName, bytes)
	if err != nil {
		belogs.Error("WriteBase64ToFile(): WriteBytesToFile fail:", filePathName, "  len(bytes):", len(bytes), err)
		return err
	}
	belogs.Debug("WriteBase64ToFile(): save filePathName ", filePathName, "  ok")
	return nil
}

func CreateAndWriteBase64ToFile(filePathName, base64 string) (err error) {
	filePath, _ := osutil.Split(filePathName)
	belogs.Debug("CreateAndWriteBase64ToFile(): Split filePathName:", filePathName)
	err = os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		belogs.Error("CreateAndWriteBase64ToFile(): MkdirAll fail, filePathName:", filePathName, err)
		return err
	}
	return WriteBase64ToFile(filePathName, base64)
}

func JoinPrefixAndUrlFileNameAndWriteBase64ToFile(destPath, url, base64 string) (filePathName string, err error) {
	belogs.Debug("JoinPrefixAndUrlFileNameAndWriteBase64ToFile(): destPath:", destPath, "  url:", url)

	filePathName, err = urlutil.JoinPrefixPathAndUrlFileName(destPath, url)
	if err != nil {
		belogs.Error("JoinPrefixAndUrlFileNameAndWriteBase64ToFile(): JoinPrefixPathAndUrlFileName fail, destPath:", destPath,
			"  url:", url, err)
		return "", err
	}
	err = CreateAndWriteBase64ToFile(filePathName, base64)
	if err != nil {
		belogs.Error("JoinPrefixAndUrlFileNameAndWriteBase64ToFile(): CreateAndWriteBase64ToFile fail, filePathName:", filePathName, err)
		return "", err
	}
	return filePathName, nil
}
