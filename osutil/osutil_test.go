package osutil

import (
	"fmt"
	"os"
	"testing"

	jsonutil "github.com/cpusoft/goutil/jsonutil"
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
	m[".txt"] = ".txt"
	files, err := GetFilesInDir(`G:\Download\`, m)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(files)
}

func TestGetAllFileCountBySuffixs(t *testing.T) {
	m := make(map[string]string, 0)
	m[".cer"] = ".cer"
	m[".crl"] = ".crl"
	m[".roa"] = ".roa"
	m[".mft"] = ".mft"
	files, err := GetAllFileCountBySuffixs(`G:\Download\cert\`, m)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(files)
}

func TestIsDir(t *testing.T) {
	f := `E:\Go\rpstir2\source\rpstir2\.project`
	s, err := IsDir(f)
	fmt.Println(s, err)

	s, err = IsExists(f)
	fmt.Println(s, err)

	s, err = IsFile(f)
	fmt.Println(s, err)

}
func TestCloseAndRemoveFile(t *testing.T) {
	userFile := `G:\Download\test.txt`
	f, err := os.Create(userFile)
	fmt.Println(err)

	err = CloseAndRemoveFile(f)
	fmt.Println(err)
}

func TestJoinPathFile(t *testing.T) {
	url := "/rrdp.apnic.net/4ea5d894-c6fc-4892-8494-cfd580a414e3/44302/snapshot.xml"
	dst := `G:\Download\rrdp`
	filePath := JoinPathFile(dst, url)
	fmt.Println(filePath)
}

func TestGetAllFileStatsBySuffixs(t *testing.T) {
	m := make(map[string]string, 0)
	m[".cer"] = ".cer"
	f, err := GetAllFileStatsBySuffixs(`G:\Download\cert\cache\rpki.afrinic.net\repository`, m)

	fmt.Println(jsonutil.MarshalJson(f), err)
}
