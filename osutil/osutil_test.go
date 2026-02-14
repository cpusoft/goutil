package osutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// -------------- 基础文件/目录判断函数测试 --------------
func TestIsExists(t *testing.T) {
	// 空参数
	exists, err := IsExists("")
	assert.Error(t, err)
	assert.Equal(t, "file is empty", err.Error())
	assert.False(t, exists)

	// 存在的文件
	tmpDir := t.TempDir()
	existFile := filepath.Join(tmpDir, "exist.txt")
	err = os.WriteFile(existFile, []byte("test"), 0644)
	assert.NoError(t, err)
	exists, err = IsExists(existFile)
	assert.NoError(t, err)
	assert.True(t, exists)

	// 不存在的文件
	notExistFile := filepath.Join(tmpDir, "not_exist.txt")
	exists, err = IsExists(notExistFile)
	assert.NoError(t, err)
	assert.False(t, exists)

	// 权限不足场景（仅非Windows）
	if runtime.GOOS != "windows" {
		permFile := filepath.Join(tmpDir, "perm.txt")
		err = os.WriteFile(permFile, []byte("test"), 0000)
		assert.NoError(t, err)
		exists, err = IsExists(permFile)
		assert.Error(t, err)
		os.Chmod(permFile, 0644)
	}
}

func TestIsDirAndIsFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_file.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	assert.NoError(t, err)

	// IsDir测试
	isDir, err := IsDir(tmpDir)
	assert.NoError(t, err)
	assert.True(t, isDir)

	isDir, err = IsDir(tmpFile)
	assert.NoError(t, err)
	assert.False(t, isDir)

	isDir, err = IsDir("")
	assert.Error(t, err)
	assert.Equal(t, "file is empty", err.Error())
	assert.False(t, isDir)

	// IsFile测试
	isFile, err := IsFile(tmpFile)
	assert.NoError(t, err)
	assert.True(t, isFile)

	isFile, err = IsFile(tmpDir)
	assert.NoError(t, err)
	assert.False(t, isFile)

	isFile, err = IsFile("")
	assert.Error(t, err)
	assert.Equal(t, "file is empty", err.Error())
	//assert.False(t, isFile)
}

// -------------- 路径处理工具函数测试 --------------
func TestBase(t *testing.T) {
	if runtime.GOOS == "windows" {
		assert.Equal(t, "demo.txt", Base(`C:\Users\test\demo.txt`))
		t.Log(Base(`C:\Users\test\demo.txt`))
		t.Log(Base(`C:\`))
		assert.Equal(t, `\`, Base(`C:\`))

		t.Log(Base(""))
		t.Log(Base("testfile"))
		assert.Equal(t, ".", Base(""))
		assert.Equal(t, "testfile", Base("testfile"))
	} else {
		assert.Equal(t, "demo.txt", Base("/root/test/demo.txt"))
		assert.Equal(t, "", Base("/"))

		t.Log(Base(""))
		t.Log(Base("testfile"))
		assert.Equal(t, ".", Base(""))
		assert.Equal(t, "testfile", Base("testfile"))
	}

}

func TestSplit(t *testing.T) {
	if runtime.GOOS == "windows" {
		dir, file := Split(`C:\Users\test\demo.txt`)
		assert.Equal(t, `C:\Users\test\`, dir)
		assert.Equal(t, "demo.txt", file)

		dir, file = Split(`C:\`)
		assert.Equal(t, `C:\`, dir)
		assert.Equal(t, "", file)
	} else {
		dir, file := Split("/root/test/demo.txt")
		assert.Equal(t, "/root/test/", dir)
		assert.Equal(t, "demo.txt", file)

		dir, file = Split("/")
		assert.Equal(t, "/", dir)
		assert.Equal(t, "", file)
	}
	dir, file := Split("")
	assert.Equal(t, "", dir)
	assert.Equal(t, "", file)
}

func TestExtAndExtNoDot(t *testing.T) {
	// Ext测试
	assert.Equal(t, ".txt", Ext(`C:\Users\test\demo.txt`))
	assert.Equal(t, ".txt", Ext("/root/test/demo.txt"))
	assert.Equal(t, "", Ext("testfile"))
	assert.Equal(t, "", Ext(""))

	// ExtNoDot测试
	assert.Equal(t, "txt", ExtNoDot(`C:\Users\test\demo.txt`))
	assert.Equal(t, "", ExtNoDot("testfile"))
	assert.Equal(t, "", ExtNoDot(""))
}

func TestGetFilePathAndFileName(t *testing.T) {
	// 正常路径
	if runtime.GOOS == "windows" {
		path := `C:\Users\test\demo.txt`
		filePath, fileName := GetFilePathAndFileName(path)
		assert.Equal(t, `C:\Users\test\`, filePath)
		assert.Equal(t, "demo.txt", fileName)
	} else {
		path := "/root/test/demo.txt"
		filePath, fileName := GetFilePathAndFileName(path)
		assert.Equal(t, "/root/test/", filePath)
		assert.Equal(t, "demo.txt", fileName)
	}

	// 无分隔符
	filePath, fileName := GetFilePathAndFileName("demo.txt")
	assert.Equal(t, "", filePath)
	assert.Equal(t, "demo.txt", fileName)

	// 根路径
	if runtime.GOOS == "windows" {
		filePath, fileName := GetFilePathAndFileName(`C:\`)
		assert.Equal(t, `C:\`, filePath)
		assert.Equal(t, "", fileName)
	} else {
		filePath, fileName := GetFilePathAndFileName("/")
		assert.Equal(t, "/", filePath)
		assert.Equal(t, "", fileName)
	}

	// 空参数
	filePath, fileName = GetFilePathAndFileName("")
	assert.Equal(t, "", filePath)
	assert.Equal(t, "", fileName)
}

func TestGetNewLineSepAndGetPathSeparator(t *testing.T) {
	// 换行符
	switch runtime.GOOS {
	case "windows":
		assert.Equal(t, "\r\n", GetNewLineSep())
	case "linux", "darwin":
		assert.Equal(t, "\n", GetNewLineSep())
	default:
		assert.Equal(t, "\n", GetNewLineSep())
	}

	// 路径分隔符
	if runtime.GOOS == "windows" {
		assert.Equal(t, "\\", GetPathSeparator())
	} else {
		assert.Equal(t, "/", GetPathSeparator())
	}
}

func TestJoinPathFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		assert.Equal(t, `C:\test\demo.txt`, JoinPathFile(`C:\test`, "demo.txt"))
	} else {
		assert.Equal(t, "/root/test/demo.txt", JoinPathFile("/root/test", "demo.txt"))
	}

	// 冗余分隔符自动处理
	assert.Equal(t, filepath.Join("test", "demo.txt"), JoinPathFile("test/", "/demo.txt"))

	// 空值场景
	assert.Equal(t, "demo.txt", JoinPathFile("", "demo.txt"))
	assert.Equal(t, "test", JoinPathFile("test", ""))
	assert.Equal(t, "", JoinPathFile("", ""))
}

// -------------- 路径获取函数测试 --------------
func TestGetCurPath(t *testing.T) {
	curPath, err := GetCurPath()
	assert.NoError(t, err)
	assert.NotEmpty(t, curPath)
}

func TestGetParentPath(t *testing.T) {
	parentPath, err := GetParentPath()
	assert.NoError(t, err)
	assert.NotEmpty(t, parentPath)
}

func TestGetPwd(t *testing.T) {
	pwd, err := GetPwd()
	assert.NoError(t, err)
	assert.NotEmpty(t, pwd)
}

func TestGetConfOrLogPath(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(originalWD)
	os.Chdir(tmpDir)

	// 存在的相对路径
	confDir := filepath.Join(tmpDir, "conf")
	os.Mkdir(confDir, 0755)
	confPath, currentPath, err := GetConfOrLogPath("conf")
	assert.NoError(t, err)
	assert.NotEmpty(t, confPath)
	assert.NotEmpty(t, currentPath)

	// 不存在的相对路径
	nonExistPath, currentPath, err := GetConfOrLogPath("non_exist")
	assert.NoError(t, err)
	assert.Empty(t, nonExistPath)
	assert.NotEmpty(t, currentPath)
}

// -------------- 文件操作函数测试 --------------
func TestCloseAndRemoveFile(t *testing.T) {
	tmpDir := t.TempDir()

	// 正常文件
	tmpFile := filepath.Join(tmpDir, "test_remove.txt")
	f, err := os.Create(tmpFile)
	assert.NoError(t, err)
	err = CloseAndRemoveFile(f)
	assert.NoError(t, err)
	exists, _ := IsExists(tmpFile)
	assert.False(t, exists)

	// 空指针
	err = CloseAndRemoveFile(nil)
	assert.NoError(t, err)

	// 关闭失败场景（重复关闭）
	tmpFile2 := filepath.Join(tmpDir, "test_close_fail.txt")
	f2, err := os.Create(tmpFile2)
	assert.NoError(t, err)
	f2.Close()
	err = CloseAndRemoveFile(f2)
	assert.NoError(t, err)
	exists, _ = IsExists(tmpFile2)
	assert.False(t, exists)
}

func TestCheckDirectoryAndSuffixs(t *testing.T) {
	tmpDir := t.TempDir()
	suffixs := map[string]string{".txt": ".txt"}

	// 空目录参数
	err := checkDirectoryAndSuffixs("", suffixs)
	assert.Error(t, err)
	assert.Equal(t, "directory is empty", err.Error())

	// 空后缀参数
	err = checkDirectoryAndSuffixs(tmpDir, map[string]string{})
	assert.Error(t, err)
	assert.Equal(t, "suffixs is empty", err.Error())

	// 非目录参数
	nonDirFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(nonDirFile, []byte("test"), 0644)
	err = checkDirectoryAndSuffixs(nonDirFile, suffixs)
	assert.Error(t, err)
	assert.Equal(t, "directory is not a directory", err.Error())

	// 正常参数
	err = checkDirectoryAndSuffixs(tmpDir, suffixs)
	assert.NoError(t, err)
}

func TestGetAllFilesBySuffixs(t *testing.T) {
	tmpDir := t.TempDir()
	// 创建测试文件
	os.WriteFile(filepath.Join(tmpDir, "test1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test2.json"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test3.tal"), []byte("test"), 0644)
	// 子目录文件
	subDir := filepath.Join(tmpDir, "sub")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "test4.txt"), []byte("test"), 0644)

	// 过滤.txt和.tal
	suffixs := map[string]string{".txt": ".txt", ".tal": ".tal"}
	files, err := GetAllFilesBySuffixs(tmpDir, suffixs)
	assert.NoError(t, err)
	assert.Len(t, files, 3)
	assert.Contains(t, files, filepath.Join(tmpDir, "test1.txt"))
	assert.Contains(t, files, filepath.Join(tmpDir, "test3.tal"))
	assert.Contains(t, files, filepath.Join(subDir, "test4.txt"))

	// 空目录
	emptyDir := t.TempDir()
	files, err = GetAllFilesBySuffixs(emptyDir, suffixs)
	assert.NoError(t, err)
	assert.Len(t, files, 0)
}

func TestGetAllFileCountBySuffixs(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "c.tal"), []byte("test"), 0644)

	suffixs := map[string]string{".txt": ".txt", ".tal": ".tal"}
	countMap, err := GetAllFileCountBySuffixs(tmpDir, suffixs)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), countMap["txt"])
	assert.Equal(t, uint64(1), countMap["tal"])
}

func TestGetFilesInDir(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test2.json"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	suffixs := map[string]string{".txt": ".txt"}
	files, err := GetFilesInDir(tmpDir, suffixs)
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "test1.txt", files[0])
}

func TestGetAllFileStatsBySuffixs(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	suffixs := map[string]string{".txt": ".txt"}
	fileStats, err := GetAllFileStatsBySuffixs(tmpDir, suffixs)
	assert.NoError(t, err)
	assert.Len(t, fileStats, 1)

	stat := fileStats[0]
	assert.Equal(t, filepath.Dir(testFile)+string(os.PathSeparator), stat.FilePath)
	assert.Equal(t, "test.txt", stat.FileName)
	assert.Equal(t, int64(4), stat.Size)
	assert.NotNil(t, stat.ModeTime)
}

// -------------- 废弃函数测试（基础验证） --------------
func TestGetAllFilesInDirectoryBySuffixs(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644)
	suffixs := map[string]string{".txt": ".txt"}

	listResult := GetAllFilesInDirectoryBySuffixs(tmpDir, suffixs)
	assert.NotNil(t, listResult)
	assert.Equal(t, 1, listResult.Len())
}

func TestGetNewLineSep(t *testing.T) {
	fmt.Println(GetNewLineSep())
}
