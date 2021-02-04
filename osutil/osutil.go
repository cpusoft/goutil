package osutil

import (
	"container/list"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	path "path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	belogs "github.com/astaxie/beego/logs"
	hashutil "github.com/cpusoft/goutil/hashutil"
)

// judge file is or not exists.
func IsExists(file string) (bool, error) {
	_, err := os.Stat(file)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// judge file is dir or not.
func IsDir(file string) (bool, error) {
	s, err := os.Stat(file)
	if err != nil {
		return false, err
	}
	return s.IsDir(), nil
}

func IsFile(file string) (bool, error) {
	s, err := IsDir(file)
	return !s, err
}

// make path.Base() using in windows,
func Base(p string) string {
	p = strings.Replace(p, "\\", "/", -1)
	return path.Base(p)
}

// make path.Split using in win
func Split(p string) (dir, file string) {
	p = strings.Replace(p, "\\", "/", -1)
	return path.Split(p)
}

// path.Ext() using in windows,
// get filname suffix(include dot)
func Ext(p string) string {
	p = strings.Replace(p, "\\", "/", -1)
	return path.Ext(p)
}

// path.Ext() using in windows,
// get filname suffix(not include dot)
func ExtNoDot(p string) string {
	return strings.Replace(Ext(p), ".", "", -1)
}

// get executable file's parent path: /root/abc/zzz/zz.sh --> /root/abc
// if go run, will be temporary program's parent path
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

// get executable file path: /root/abc/zzz/zz.sh --> /root/abc/zzz
// if go run, will be temporary program path
func GetCurPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	return path
}

// get current working directory: /root/abc/zzz/zz.exe --> /root/abc/zzz
// if go run, will current path
func GetPwd() string {
	pwd, _ := os.Getwd()
	return pwd
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
			suffix := Ext(filename)
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
			suffix := Ext(fileName)
			if _, ok := suffixs[suffix]; ok {
				files = append(files, fileName)
			}
		}
		return nil
	})
	return files, nil
}

func GetAllFileCountBySuffixs(directory string, suffixs map[string]string) (suffixCount map[string]uint64, err error) {

	suffixCount = make(map[string]uint64, 0)
	absolutePath, _ := filepath.Abs(directory)
	filepath.Walk(absolutePath, func(fileName string, fi os.FileInfo, err error) error {
		if err != nil || len(fileName) == 0 || nil == fi {
			belogs.Debug("GetAllFilesBySuffixs():filepath.Walk(): err:", err)
			return err
		}
		if !fi.IsDir() {
			suffix := Ext(fileName)
			if _, ok := suffixs[suffix]; ok {
				suffixNotDot := ExtNoDot(fileName)
				if c, ok := suffixCount[suffixNotDot]; ok {
					suffixCount[suffixNotDot] = c + 1
				} else {
					suffixCount[suffixNotDot] = 1
				}
			}
		}
		return nil
	})
	return suffixCount, nil
}

func GetFilesInDir(directory string, suffixs map[string]string) ([]string, error) {
	files := make([]string, 0, 10)
	dir, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, file := range dir {
		if file.IsDir() { // 忽略目录
			continue
		}
		suffix := Ext(file.Name())
		if _, ok := suffixs[suffix]; ok {
			files = append(files, file.Name())
		}
	}
	return files, nil
}

type FileStat struct {
	FilePath string    `json:"filePath"`
	FileName string    `json:"fileName"`
	ModeTime time.Time `json:"modeTime"`
	Size     int64     `json:"size"`
	Hash256  string    `json:"hash256"`
}

func GetAllFileStatsBySuffixs(directory string, suffixs map[string]string) ([]FileStat, error) {

	absolutePath, _ := filepath.Abs(directory)
	fileStats := make([]FileStat, 0)
	filepath.Walk(absolutePath, func(path string, fi os.FileInfo, err error) error {
		if err != nil || len(path) == 0 || nil == fi {
			belogs.Debug("GetAllFileStatsBySuffixs():filepath.Walk(): err:", err)
			return err
		}
		if !fi.IsDir() {

			suffix := Ext(path)
			if _, ok := suffixs[suffix]; ok {
				fileStat := FileStat{}
				fileStat.FilePath, _ = Split(path)
				fileStat.FileName = fi.Name()
				fileStat.ModeTime = fi.ModTime()
				fileStat.Size = fi.Size()
				fileStat.Hash256, _ = hashutil.Sha256File(JoinPathFile(fileStat.FilePath, fileStat.FileName))
				fileStats = append(fileStats, fileStat)
			}
		}
		return nil
	})
	return fileStats, nil

}

func GetFilePathAndFileName(fileAllPath string) (filePath string, fileName string) {
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

func GetPathSeparator() string {
	return string(os.PathSeparator)
}

func JoinPathFile(pathName, fileName string) string {
	fileName = strings.Replace(fileName, `/`, string(os.PathSeparator), -1)
	fileName = strings.Replace(fileName, `\`, string(os.PathSeparator), -1)
	pathName = strings.Replace(pathName, `/`, string(os.PathSeparator), -1)
	pathName = strings.Replace(pathName, `\`, string(os.PathSeparator), -1)
	if !strings.HasSuffix(pathName, string(os.PathSeparator)) && !strings.HasPrefix(fileName, string(os.PathSeparator)) {
		pathName = pathName + string(os.PathSeparator)
	}
	return pathName + fileName
}

func CloseAndRemoveFile(file *os.File) error {
	if file == nil {
		return nil
	}
	s, err := IsExists(file.Name())
	if err != nil {
		belogs.Debug("CloseAndRemoveFile():IsExists:err: ", file.Name(), err)
		return err
	}
	if !s {
		return nil
	}

	err = file.Close()
	if err != nil {
		belogs.Debug("CloseAndRemoveFile():file.Close():err: ", file.Name(), err)
		return err
	}
	err = os.Remove(file.Name())
	if err != nil {
		belogs.Error("CloseAndRemoveFile():os.Remove:err:", file.Name(), err)
		return nil
	}
	return nil
}

// only use in goutil/conf and goutil/log.    .
// relativePath: "conf" or "log"
// dont use in others.
func GetCurrentOrParentAbsolutePath(relativePath string) (absolutePath string, err error) {
	path := GetPwd()
	absolutePath = path + GetPathSeparator() + relativePath
	ok, err := IsDir(absolutePath)
	if err == nil && ok {
		return absolutePath, nil
	}
	pos := strings.LastIndex(path, GetPathSeparator())
	path = string([]byte(path)[:pos])
	absolutePath = path + GetPathSeparator() + relativePath
	ok, err = IsDir(absolutePath)
	if err == nil && ok {
		return absolutePath, nil
	}
	return "", errors.New("cannot found absolutePath of relativePath " + relativePath)
}
