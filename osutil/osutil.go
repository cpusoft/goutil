package osutil

import (
	"container/list"
	"errors"
	"os"
	"os/exec"
	path "path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/hashutil"
)

// judge file is or not exists.
func IsExists(file string) (bool, error) {
	if len(file) == 0 {
		return false, errors.New("file is empty")
	}
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
	if len(file) == 0 {
		return false, errors.New("file is empty")
	}
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
	//	p = strings.Replace(p, "\\", "/", -1)
	return path.Base(p)
}

// make path.Split using in win
func Split(p string) (dir, file string) {
	//	p = strings.Replace(p, "\\", "/", -1)
	return path.Split(p)
}

// path.Ext() using in windows,
// get filname suffix(include dot)
func Ext(p string) string {
	//p = strings.Replace(p, "\\", "/", -1)
	return path.Ext(p)
}

// path.Ext() using in windows,
// get filname suffix(not include dot)
func ExtNoDot(p string) string {
	return strings.TrimPrefix(Ext(p), ".")
}

// get executable file path: /root/abc/zzz/zz.sh --> /root/abc/zzz
// if go run, will be temporary program path
func GetCurPath() (string, error) {
	if len(os.Args) == 0 {
		return "", errors.New("os.Args is empty")
	}
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	absPath, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	return filepath.Dir(absPath), nil
}

// get executable file's parent path: /root/abc/zzz/zz.sh --> /root/abc
// if go run, will be temporary program's parent path
func GetParentPath() (string, error) {
	execDir, err := GetCurPath()
	if err != nil {
		return "", err
	}
	// 再获取目录的父目录（/root/abc）
	parentDir := filepath.Dir(execDir)
	return parentDir, nil
}

// get current working directory: /root/abc/zzz/zz.exe --> /root/abc/zzz
// if go run, will current path
func GetPwd() (string, error) {
	return os.Getwd()
}

// Deprecated, will use GetAllFilesBySuffixs()
func GetAllFilesInDirectoryBySuffixs(directory string, suffixs map[string]string) *list.List {

	absolutePath, err := filepath.Abs(directory)
	if err != nil {
		belogs.Error("GetAllFilesInDirectoryBySuffixs(): abs fail, directory:", directory, err)
		return nil
	}
	listStr := list.New()
	filepath.Walk(absolutePath, func(filename string, fi os.FileInfo, err error) error {
		if err != nil || len(filename) == 0 || nil == fi {
			belogs.Error("GetAllFilesInDirectoryBySuffixs(): Walk fail, filename:", filename, "  fi:", fi, err)
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

	absolutePath, err := filepath.Abs(directory)
	if err != nil {
		belogs.Error("GetAllFilesBySuffixs(): abs fail, directory:", directory, err)
		return nil, err
	}
	files := make([]string, 0)
	filepath.Walk(absolutePath, func(fileName string, fi os.FileInfo, err error) error {
		if err != nil || len(fileName) == 0 || nil == fi {
			belogs.Error("GetAllFilesBySuffixs():filepath.Walk(): err:", err)
			return err
		}
		// 关键修复：跳过符号链接（软链接），避免循环递归
		if fi.Mode()&os.ModeSymlink != 0 {
			belogs.Debug("GetAllFilesBySuffixs(): skip symlink:", fileName)
			return filepath.SkipDir // 跳过链接指向的目录/文件
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
	absolutePath, err := filepath.Abs(directory)
	if err != nil {
		belogs.Error("GetAllFileCountBySuffixs(): abs fail, directory:", directory, err)
		return nil, err
	}
	filepath.Walk(absolutePath, func(fileName string, fi os.FileInfo, err error) error {
		if err != nil || len(fileName) == 0 || nil == fi {
			belogs.Error("GetAllFilesBySuffixs():filepath.Walk(): err:", err)
			return err
		}
		// 关键修复：跳过符号链接（软链接），避免循环递归
		if fi.Mode()&os.ModeSymlink != 0 {
			belogs.Debug("GetAllFileCountBySuffixs(): skip symlink:", fileName)
			return filepath.SkipDir // 跳过链接指向的目录/文件
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
	if len(directory) == 0 {
		return nil, errors.New("directory is empty")
	}
	files := make([]string, 0, 10)
	dir, err := os.ReadDir(directory)
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

	absolutePath, err := filepath.Abs(directory)
	if err != nil {
		belogs.Error("GetAllFileStatsBySuffixs(): abs fail, directory:", directory, err)
		return nil, err
	}

	fileStats := make([]FileStat, 0)
	filepath.Walk(absolutePath, func(path string, fi os.FileInfo, err error) error {
		if err != nil || len(path) == 0 || nil == fi {
			belogs.Debug("GetAllFileStatsBySuffixs():filepath.Walk(): err:", err)
			return err
		}

		// 关键修复：跳过符号链接（软链接），避免循环递归
		if fi.Mode()&os.ModeSymlink != 0 {
			belogs.Debug("GetAllFileStatsBySuffixs(): skip symlink:", path)
			return filepath.SkipDir // 跳过链接指向的目录/文件
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
	case "linux", "darwin":
		return "\n"
	default:
		return "\n"

	}
}

func GetPathSeparator() string {
	return string(os.PathSeparator)
}

func JoinPathFile(pathName, fileName string) string {
	//	fileName = strings.Replace(fileName, `/`, string(os.PathSeparator), -1)
	//	fileName = strings.Replace(fileName, `\`, string(os.PathSeparator), -1)
	//	pathName = strings.Replace(pathName, `/`, string(os.PathSeparator), -1)
	//	pathName = strings.Replace(pathName, `\`, string(os.PathSeparator), -1)
	//	if !strings.HasSuffix(pathName, string(os.PathSeparator)) && !strings.HasPrefix(fileName, string(os.PathSeparator)) {
	//		pathName = pathName + string(os.PathSeparator)
	//	}
	//	return pathName + fileName
	return filepath.Join(pathName, fileName)
}

func CloseAndRemoveFile(file *os.File) error {
	if file == nil {
		return nil
	}

	err := file.Close()
	if err != nil {
		belogs.Debug("CloseAndRemoveFile():file.Close():err: ", file.Name(), err)
		return err
	}
	err = os.Remove(file.Name())
	if err != nil {
		belogs.Error("CloseAndRemoveFile():os.Remove:err:", file.Name(), err)
		return err
	}
	return nil
}

// relativePath: "conf" or "log"
// if not exist conf or log, confOrLogPath is "", can use currentPath
// only use in goutil/conf and goutil/log.  dont use in others.
func GetConfOrLogPath(relativePath string) (confOrLogPath string, currentPath string, err error) {
	currentPath, err = GetPwd()
	if err != nil {
		return "", relativePath, err
	}
	currentPath = currentPath + GetPathSeparator()
	confOrLogPath = JoinPathFile(currentPath, relativePath) + GetPathSeparator()
	ok, err := IsDir(confOrLogPath)
	if err == nil && ok {
		return confOrLogPath, currentPath, nil
	}
	return "", currentPath, nil
}
