package fileutil

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"strings"

	belogs "github.com/astaxie/beego/logs"
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
