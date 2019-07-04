package osutil

import (
	"fmt"
	"os"
	"testing"
)

func TestGetFilePathAndFileName(t *testing.T) {
	fileAllPath := `G:\Download\cert\cache\rpki.ripe.net\repository\DEFAULT\b0\3bfc31-dc32-4541-8460-c927b8c2c7c4\1\cF5Nt5Q1B6BFc5cD15QWWEw4qbw.mft`
	filePath, fileName := GetFilePathAndFileName(fileAllPath)
	fmt.Println(filePath, ":", fileName)
}

func TestGetNewLineSep(t *testing.T) {
	fmt.Println(GetNewLineSep())
}

func TestRemoveAll(t *testing.T) {
	err := os.RemoveAll(`G:\Download\tmp\root\`)
	if err != nil {
		fmt.Println(err)
	}
}

func TestGetFilesInDir(t *testing.T) {
	m := make(map[string]string, 0)
	m[".cer"] = ".cer"
	files, err := GetFilesInDir(`G:\Download\cert\`, m)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(files)
}
