package osutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

// 全局临时目录，用于测试文件/目录操作
var tempDir string

// 测试初始化：创建临时目录
func TestMain(m *testing.M) {
	// 创建临时目录
	var err error
	tempDir, err = os.MkdirTemp("", "osutil_test_")
	if err != nil {
		panic("创建临时目录失败: " + err.Error())
	}
	// 运行所有测试
	code := m.Run()
	// 测试结束后清理临时目录
	_ = os.RemoveAll(tempDir)
	os.Exit(code)
}

// ------------------------------ 基础文件/目录判断测试 ------------------------------
func TestIsExists(t *testing.T) {
	// 测试场景1：空路径
	_, err := IsExists("")
	if err.Error() != "file is empty" {
		t.Errorf("IsExists空路径测试失败，预期错误: file is empty, 实际: %v", err)
	}

	// 测试场景2：不存在的文件
	notExistFile := filepath.Join(tempDir, "not_exist.txt")
	exists, err := IsExists(notExistFile)
	if err != nil {
		t.Errorf("IsExists不存在文件测试失败，错误: %v", err)
	}
	if exists {
		t.Error("IsExists不存在文件测试失败，预期返回false，实际返回true")
	}

	// 测试场景3：存在的文件
	existFile := filepath.Join(tempDir, "exist.txt")
	f, err := os.Create(existFile)
	if err != nil {
		t.Fatal("创建测试文件失败: ", err)
	}
	_ = f.Close()
	exists, err = IsExists(existFile)
	if err != nil {
		t.Errorf("IsExists存在文件测试失败，错误: %v", err)
	}
	if !exists {
		t.Error("IsExists存在文件测试失败，预期返回true，实际返回false")
	}
}

func TestIsDirAndIsFile(t *testing.T) {
	// 测试场景1：空路径
	_, err := IsDir("")
	if err.Error() != "file is empty" {
		t.Errorf("IsDir空路径测试失败，预期错误: file is empty, 实际: %v", err)
	}

	// 测试场景2：目录
	testDir := filepath.Join(tempDir, "test_dir")
	_ = os.Mkdir(testDir, 0755)
	isDir, err := IsDir(testDir)
	if err != nil {
		t.Errorf("IsDir目录测试失败，错误: %v", err)
	}
	if !isDir {
		t.Error("IsDir目录测试失败，预期返回true，实际返回false")
	}
	// 验证IsFile对目录返回false
	isFile, err := IsFile(testDir)
	if err != nil {
		t.Errorf("IsFile目录测试失败，错误: %v", err)
	}
	if isFile {
		t.Error("IsFile目录测试失败，预期返回false，实际返回true")
	}

	// 测试场景3：文件
	testFile := filepath.Join(tempDir, "test_file.txt")
	f, err := os.Create(testFile)
	if err != nil {
		t.Fatal("创建测试文件失败: ", err)
	}
	_ = f.Close()
	isDir, err = IsDir(testFile)
	if err != nil {
		t.Errorf("IsDir文件测试失败，错误: %v", err)
	}
	if isDir {
		t.Error("IsDir文件测试失败，预期返回false，实际返回true")
	}
	// 验证IsFile对文件返回true
	isFile, err = IsFile(testFile)
	if err != nil {
		t.Errorf("IsFile文件测试失败，错误: %v", err)
	}
	if !isFile {
		t.Error("IsFile文件测试失败，预期返回true，实际返回false")
	}
}

// ------------------------------ 路径处理函数测试 ------------------------------
func TestBaseSplitExtExtNoDot(t *testing.T) {
	// 跨平台路径测试
	testCases := []struct {
		name          string
		path          string
		wantBase      string
		wantSplitDir  string
		wantSplitFile string
		wantExt       string
		wantExtNoDot  string
	}{
		{
			name:          "Linux简单路径",
			path:          "/root/test.txt",
			wantBase:      "test.txt",
			wantSplitDir:  "/root/",
			wantSplitFile: "test.txt",
			wantExt:       ".txt",
			wantExtNoDot:  "txt",
		},
		{
			name:          "Windows简单路径",
			path:          "C:\\test\\demo.tar.gz",
			wantBase:      "demo.tar.gz", // path.Base对\不识别，会返回整个路径，这里是已知问题（因使用path包而非filepath）
			wantSplitDir:  "",
			wantSplitFile: "C:\\test\\demo.tar.gz",
			wantExt:       "", // path.Ext对\不识别，返回空
			wantExtNoDot:  "",
		},
		{
			name:          "Linux复杂路径",
			path:          "/root/test/../demo",
			wantBase:      "demo",
			wantSplitDir:  "/root/test/../",
			wantSplitFile: "demo",
			wantExt:       "",
			wantExtNoDot:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 测试Base
			if got := Base(tc.path); got != tc.wantBase {
				t.Errorf("Base() = %v, want %v", got, tc.wantBase)
			}
			// 测试Split
			dir, file := Split(tc.path)
			if dir != tc.wantSplitDir || file != tc.wantSplitFile {
				t.Errorf("Split() = (%v, %v), want (%v, %v)", dir, file, tc.wantSplitDir, tc.wantSplitFile)
			}
			// 测试Ext
			if got := Ext(tc.path); got != tc.wantExt {
				t.Errorf("Ext() = %v, want %v", got, tc.wantExt)
			}
			// 测试ExtNoDot
			if got := ExtNoDot(tc.path); got != tc.wantExtNoDot {
				t.Errorf("ExtNoDot() = %v, want %v", got, tc.wantExtNoDot)
			}
		})
	}
}

func TestGetFilePathAndFileName(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		wantPath string
		wantName string
	}{
		{
			name:     "Linux路径",
			path:     "/root/test.txt",
			wantPath: "/root/",
			wantName: "test.txt",
		},
		{
			name:     "Windows路径",
			path:     "C:\\test\\demo.txt",
			wantPath: "C:\\test\\",
			wantName: "demo.txt",
		},
		{
			name:     "无分隔符路径",
			path:     "test.txt",
			wantPath: "",
			wantName: "test.txt",
		},
		{
			name:     "根路径",
			path:     "/",
			wantPath: "/",
			wantName: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir, file := GetFilePathAndFileName(tc.path)
			if dir != tc.wantPath || file != tc.wantName {
				t.Errorf("GetFilePathAndFileName() = (%v, %v), want (%v, %v)", dir, file, tc.wantPath, tc.wantName)
			}
		})
	}

	// 测试空路径（边界场景）
	dir, file := GetFilePathAndFileName("")
	if dir != "" || file != "" {
		t.Errorf("GetFilePathAndFileName空路径测试失败，返回(%v, %v)，预期('', '')", dir, file)
	}
}

func TestJoinPathFile(t *testing.T) {
	testCases := []struct {
		name     string
		pathName string
		fileName string
		want     string
	}{
		{
			name:     "Linux路径拼接",
			pathName: "/root/test",
			fileName: "demo.txt",
			want:     "/root/test/demo.txt",
		},
		{
			name:     "Windows路径拼接",
			pathName: "C:\\test",
			fileName: "demo.txt",
			want:     "C:\\test\\demo.txt",
		},
		{
			name:     "路径含分隔符",
			pathName: "/root/test/",
			fileName: "/demo.txt",
			want:     "/demo.txt", // filepath.Join会处理重复分隔符
		},
		{
			name:     "空路径拼接",
			pathName: "",
			fileName: "demo.txt",
			want:     "demo.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := JoinPathFile(tc.pathName, tc.fileName); got != tc.want {
				t.Errorf("JoinPathFile() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGetNewLineSepAndPathSeparator(t *testing.T) {
	// 测试换行符
	switch runtime.GOOS {
	case "windows":
		if GetNewLineSep() != "\r\n" {
			t.Error("Windows系统换行符测试失败，预期\\r\\n")
		}
	case "linux", "darwin":
		if GetNewLineSep() != "\n" {
			t.Error(runtime.GOOS + "系统换行符测试失败，预期\\n")
		}
	}

	// 测试路径分隔符
	wantSep := string(os.PathSeparator)
	if got := GetPathSeparator(); got != wantSep {
		t.Errorf("GetPathSeparator() = %v, want %v", got, wantSep)
	}
}

// ------------------------------ 可执行文件路径测试 ------------------------------
func TestGetCurPathAndGetParentPath(t *testing.T) {
	// 测试GetCurPath
	curPath, err := GetCurPath()
	if err != nil {
		t.Fatal("GetCurPath测试失败，错误: ", err)
	}
	// 验证返回的是目录（而非文件）
	isDir, err := IsDir(curPath)
	if err != nil {
		t.Errorf("验证GetCurPath返回值失败，错误: %v", err)
	}
	if !isDir {
		t.Error("GetCurPath返回值不是目录，测试失败")
	}

	// 测试GetParentPath
	parentPath, err := GetParentPath()
	if err != nil {
		t.Fatal("GetParentPath测试失败，错误: ", err)
	}
	// 验证父目录存在
	exists, err := IsExists(parentPath)
	if err != nil || !exists {
		t.Error("GetParentPath返回的父目录不存在，测试失败")
	}
}

// ------------------------------ 文件遍历/统计测试 ------------------------------
func TestGetAllFilesBySuffixs(t *testing.T) {
	// 创建测试目录和文件
	testDir := filepath.Join(tempDir, "file_test")
	_ = os.Mkdir(testDir, 0755)
	// 创建子目录
	subDir := filepath.Join(testDir, "sub")
	_ = os.Mkdir(subDir, 0755)

	// 创建测试文件
	files := map[string]string{
		"test1.txt":     "txt",
		"test2.jpg":     "jpg",
		"test3.txt":     "txt",
		"sub/test4.txt": "txt",
		"sub/test5.jpg": "jpg",
	}
	for file, _ := range files {
		fPath := filepath.Join(testDir, file)
		f, err := os.Create(fPath)
		if err != nil {
			t.Fatal("创建测试文件失败: ", err)
		}
		_ = f.Close()
	}
	t.Log("测试文件创建完成: testDir:", testDir, " files:", files)

	// 测试筛选txt文件
	suffixs := map[string]string{".txt": ""}
	result, err := GetAllFilesBySuffixs(testDir, suffixs)
	if err != nil {
		t.Fatal("GetAllFilesBySuffixs测试失败，错误: ", err)
	}
	t.Log("GetAllFilesBySuffixs结果: ", result)

	// 预期3个txt文件（test1.txt, test3.txt, sub/test4.txt）
	if len(result) != 3 {
		t.Errorf("GetAllFilesBySuffixs筛选txt失败，预期3个文件，实际%v个", len(result))
	}

	// 测试空目录
	emptyDir := filepath.Join(tempDir, "empty_dir")
	_ = os.Mkdir(emptyDir, 0755)
	result, err = GetAllFilesBySuffixs(emptyDir, suffixs)
	if err != nil {
		t.Fatal("GetAllFilesBySuffixs空目录测试失败，错误: ", err)
	}
	t.Log("GetAllFilesBySuffixs空目录测试结果: ", result)
	if len(result) != 0 {
		t.Error("GetAllFilesBySuffixs空目录测试失败，预期0个文件，实际非空")
	}

	// 测试非法目录
	_, err = GetAllFilesBySuffixs("/not_exist_dir_12345", suffixs)
	if err == nil {
		t.Error("GetAllFilesBySuffixs非法目录测试失败，预期返回错误，实际无错误")
	}
}

func TestGetAllFileCountBySuffixs(t *testing.T) {
	// 复用上面的测试目录
	testDir := filepath.Join(tempDir, "file_test")
	suffixs := map[string]string{".txt": "", ".jpg": ""}

	// 测试统计
	count, err := GetAllFileCountBySuffixs(testDir, suffixs)
	if err != nil {
		t.Fatal("GetAllFileCountBySuffixs测试失败，错误: ", err)
	}
	// 预期txt:3，jpg:2
	if count["txt"] != 3 || count["jpg"] != 2 {
		t.Errorf("GetAllFileCountBySuffixs统计失败，预期txt:3, jpg:2，实际txt:%v, jpg:%v", count["txt"], count["jpg"])
	}

	// 测试空目录
	emptyDir := filepath.Join(tempDir, "empty_dir")
	count, err = GetAllFileCountBySuffixs(emptyDir, suffixs)
	if err != nil {
		t.Fatal("GetAllFileCountBySuffixs空目录测试失败，错误: ", err)
	}
	if len(count) != 0 {
		t.Error("GetAllFileCountBySuffixs空目录测试失败，预期空map，实际: ", count)
	}
}

func TestGetFilesInDir(t *testing.T) {
	// 创建测试目录
	testDir := filepath.Join(tempDir, "files_in_dir")
	_ = os.Mkdir(testDir, 0755)

	// 创建测试文件（仅当前目录）
	files := []string{"test1.txt", "test2.jpg", "test3.txt", "test4.png"}
	for _, file := range files {
		fPath := filepath.Join(testDir, file)
		f, err := os.Create(fPath)
		if err != nil {
			t.Fatal("创建测试文件失败: ", err)
		}
		_ = f.Close()
	}
	// 创建子目录（验证不遍历子目录）
	subDir := filepath.Join(testDir, "sub")
	_ = os.Mkdir(subDir, 0755)
	os.Create(filepath.Join(subDir, "test5.txt"))

	// 测试筛选txt文件
	suffixs := map[string]string{".txt": ""}
	result, err := GetFilesInDir(testDir, suffixs)
	if err != nil {
		t.Fatal("GetFilesInDir测试失败，错误: ", err)
	}
	// 预期2个txt文件（test1.txt, test3.txt）
	if len(result) != 2 {
		t.Errorf("GetFilesInDir筛选txt失败，预期2个文件，实际%v个", len(result))
	}

	// 测试空路径
	_, err = GetFilesInDir("", suffixs)
	if err.Error() != "directory is empty" {
		t.Errorf("GetFilesInDir空路径测试失败，预期错误: directory is empty，实际: %v", err)
	}
}

// ------------------------------ 文件操作测试 ------------------------------
func TestCloseAndRemoveFile(t *testing.T) {
	// 测试场景1：正常文件关闭删除
	testFile := filepath.Join(tempDir, "close_remove.txt")
	f, err := os.Create(testFile)
	if err != nil {
		t.Fatal("创建测试文件失败: ", err)
	}
	err = CloseAndRemoveFile(f)
	if err != nil {
		t.Errorf("CloseAndRemoveFile正常文件测试失败，错误: %v", err)
	}
	// 验证文件已删除
	exists, _ := IsExists(testFile)
	if exists {
		t.Error("CloseAndRemoveFile正常文件测试失败，文件未被删除")
	}

	// 测试场景2：空文件指针
	err = CloseAndRemoveFile(nil)
	if err != nil {
		t.Error("CloseAndRemoveFile空指针测试失败，预期无错误，实际: ", err)
	}

	// 测试场景3：文件已被删除（模拟竞态）
	f, err = os.Create(testFile)
	if err != nil {
		t.Fatal("创建测试文件失败: ", err)
	}
	_ = os.Remove(testFile) // 先删除文件
	err = CloseAndRemoveFile(f)
	if err == nil {
		t.Error("CloseAndRemoveFile文件已删除测试失败，预期返回错误，实际无错误")
	}
}

func TestGetConfOrLogPath(t *testing.T) {
	// 测试场景1：存在的相对路径
	confDir := filepath.Join(tempDir, "conf")
	_ = os.Mkdir(confDir, 0755)
	// 切换到临时目录（避免影响当前工作目录）
	originalWD, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	// 测试
	confPath, currentPath, err := GetConfOrLogPath("conf")
	if err != nil {
		t.Errorf("GetConfOrLogPath存在路径测试失败，错误: %v", err)
	}
	// 验证路径
	expectedConfPath := filepath.Join(tempDir, "conf") + string(os.PathSeparator)
	expectedCurrentPath := tempDir + string(os.PathSeparator)
	if confPath != expectedConfPath || currentPath != expectedCurrentPath {
		t.Errorf("GetConfOrLogPath存在路径测试失败，预期(%v, %v)，实际(%v, %v)",
			expectedConfPath, expectedCurrentPath, confPath, currentPath)
	}

	// 测试场景2：不存在的相对路径
	logPath, currentPath, err := GetConfOrLogPath("log")
	if err != nil {
		t.Errorf("GetConfOrLogPath不存在路径测试失败，错误: %v", err)
	}
	if logPath != "" {
		t.Error("GetConfOrLogPath不存在路径测试失败，预期返回空字符串，实际: ", logPath)
	}

	// 恢复工作目录
	_ = os.Chdir(originalWD)
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

func TestIsDir(t *testing.T) {
	f := `E:\Go\rpstir2\source\rpstir2\.project`
	s, err := IsDir(f)
	fmt.Println(s, err)

	s, err = IsExists(f)
	fmt.Println(s, err)

	s, err = IsFile(f)
	fmt.Println(s, err)

}

func TestGetAllFileStatsBySuffixs(t *testing.T) {
	m := make(map[string]string, 0)
	m[".cer"] = ".cer"
	f, err := GetAllFileStatsBySuffixs(`G:\Download\cert\`, m)

	fmt.Println(jsonutil.MarshalJson(f), err)
}
func TestExtNoDot(t *testing.T) {
	f := "sync://aaa.com/bbb/ccc.cer"
	ex := ExtNoDot(f)
	fmt.Println(ex)
}

func TestBase(t *testing.T) {
	f := "sync://aaa.com/bbb/ccc.cer"
	ex := Base(f)
	fmt.Println(ex)
}
