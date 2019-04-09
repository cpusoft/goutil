package osutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
