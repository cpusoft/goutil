package osutil

import (
	"container/list"
	"os"
	"os/exec"
	path "path"
	"path/filepath"
	"runtime"
	"strings"

	belogs "github.com/astaxie/beego/logs"
)

func IsDir(file string) (bool, error) {

	f, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return false, err
		}
		return false, err
	}
	return f.IsDir(), nil
}

func GetParentPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	dirs := strings.Split(path, string(os.PathSeparator))
	index := len(dirs)
	if len(dirs) > 2 {
		index = len(dirs) - 2
	}
	ret := strings.Join(dirs[:index], string(os.PathSeparator))
	return ret
}

// will deprecated, will use GetAllFilesBySuffixs()
func GetAllFilesInDirectoryBySuffixs(directory string, suffixs map[string]string) *list.List {

	absolutePath, _ := filepath.Abs(directory)
	listStr := list.New()
	filepath.Walk(absolutePath, func(filename string, fi os.FileInfo, err error) error {
		if err != nil || len(filename) == 0 || nil == fi {
			return err
		}
		if !fi.IsDir() {
			suffix := path.Ext(filename)
			//fmt.Println(suffix)
			if _, ok := suffixs[suffix]; ok {
				listStr.PushBack(filename)
			}
		}
		return nil
	})
	return listStr
}
func GetAllFilesBySuffixs(directory string, suffixs map[string]string) ([]string, error) {

	absolutePath, _ := filepath.Abs(directory)
	files := make([]string, 0)
	filepath.Walk(absolutePath, func(fileName string, fi os.FileInfo, err error) error {
		if err != nil || len(fileName) == 0 || nil == fi {
			belogs.Debug("GetAllFilesBySuffixs():filepath.Walk(): err:", err)
			return err
		}
		if !fi.IsDir() {
			suffix := path.Ext(fileName)
			if _, ok := suffixs[suffix]; ok {
				files = append(files, fileName)
			}
		}
		return nil
	})
	return files, nil
}

func GetFilePathAndFileName(fileAllPath string) (filePath string, fileName string) {
	belogs.Debug("GetFilePathAndFileName(): fileAllPath:", fileAllPath)
	i := strings.LastIndex(fileAllPath, string(os.PathSeparator))
	return fileAllPath[:i+1], fileAllPath[i+1:]

}

func GetNewLineSep() string {
	switch runtime.GOOS {
	case "windows":
		return "\r\n"
	case "linux":
		return "\n"
	default:
		return "\n"

	}
}
