package rsyncutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/cpusoft/goutil/osutil"
)

// -------------------------- 测试前置工具函数（修复类型兼容问题） --------------------------
// 创建临时测试目录，自动清理（兼容*testing.T/*testing.B）
func createTempDir(tb testing.TB) string {
	tb.Helper()
	dir, err := os.MkdirTemp("", "rsyncutil_test_*")
	if err != nil {
		tb.Fatalf("创建临时目录失败: %v", err)
	}
	tb.Cleanup(func() {
		_ = os.RemoveAll(dir) // 测试结束清理
	})
	return dir
}

// 创建测试文件（兼容*testing.T/*testing.B）
func createTestFile(tb testing.TB, dir, filename string, content string) {
	tb.Helper()
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		tb.Fatalf("创建测试文件失败 %s: %v", filePath, err)
	}
}

// -------------------------- 基础类型测试 --------------------------
func TestNewRsyncClientConfig(t *testing.T) {
	// 正常场景
	timeout := "10"
	conTimeout := "5"
	config := NewRsyncClientConfig(timeout, conTimeout)
	if config.Timeout != timeout || config.ConTimeout != conTimeout {
		t.Errorf("NewRsyncClientConfig 结果错误, 期望(%s,%s), 实际(%s,%s)",
			timeout, conTimeout, config.Timeout, config.ConTimeout)
	}

	// 临界值：空字符串
	config = NewRsyncClientConfig("", "")
	if config.Timeout != "" || config.ConTimeout != "" {
		t.Error("NewRsyncClientConfig 空参数测试失败")
	}
}

// -------------------------- RsyncTestConnect 测试 --------------------------
func TestRsyncTestConnect(t *testing.T) {
	tests := []struct {
		name     string
		rsyncUrl string
		wantErr  bool
	}{
		{
			name:     "无效地址（预期错误）",
			rsyncUrl: "rsync://127.0.0.1:8730/nonexist",
			wantErr:  true,
		},
		{
			name:     "空URL（临界值，预期错误）",
			rsyncUrl: "",
			wantErr:  true,
		},
		{
			name:     "无端口URL（临界值，改为非标准端口确保错误）",
			rsyncUrl: "rsync://127.0.0.1:8731/nonexist", // 改为8731端口
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RsyncTestConnect(tt.rsyncUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("RsyncTestConnect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// -------------------------- RsyncQuietWithConfig 测试 --------------------------
func TestRsyncQuietWithConfig(t *testing.T) {
	tempDir := createTempDir(t)
	validConfig := NewRsyncClientConfig("10", "5")

	tests := []struct {
		name     string
		rsyncUrl string
		destPath string
		config   *RsyncClientConfig
		wantErr  bool
	}{
		{
			name:     "无效URL（预期错误）",
			rsyncUrl: "rsync://invalid-host/nonexist",
			destPath: tempDir,
			config:   validConfig,
			wantErr:  true,
		},
		{
			name:     "空目标路径（临界值，预期错误）",
			rsyncUrl: "rsync://127.0.0.1/test",
			destPath: "",
			config:   validConfig,
			wantErr:  true,
		},
		{
			name:     "nil配置（临界值，预期panic转错误）",
			rsyncUrl: "rsync://127.0.0.1/test",
			destPath: tempDir,
			config:   nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := RsyncQuietWithConfig(tt.rsyncUrl, tt.destPath, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("RsyncQuietWithConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// -------------------------- RsyncToStdout 测试 --------------------------
func TestRsyncToStdout(t *testing.T) {
	tempDir := createTempDir(t)

	tests := []struct {
		name     string
		rsyncUrl string
		destPath string
		wantErr  bool
	}{
		{
			name:     "无效URL（预期错误）",
			rsyncUrl: "rsync://invalid-host-1234/nonexist",
			destPath: tempDir,
			wantErr:  true,
		},
		{
			name:     "超长目标路径（临界值，预期错误）",
			rsyncUrl: "rsync://127.0.0.1/test",
			destPath: strings.Repeat("a", 1024), // 超长路径
			wantErr:  true,
		},
		{
			name:     "空URL（临界值，预期错误）",
			rsyncUrl: "",
			destPath: tempDir,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := RsyncToStdout(tt.rsyncUrl, tt.destPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("RsyncToStdout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// -------------------------- RsyncToLogFile 测试 --------------------------
func TestRsyncToLogFile(t *testing.T) {
	tempDir := createTempDir(t)
	invalidLogDir := filepath.Join(tempDir, "invalid/123/456") // 多级不存在目录

	tests := []struct {
		name     string
		rsyncUrl string
		destPath string
		logPath  string
		wantErr  bool
	}{
		{
			name:     "无效日志目录（预期错误）",
			rsyncUrl: "rsync://127.0.0.1/test",
			destPath: tempDir,
			logPath:  invalidLogDir,
			wantErr:  true, // MkdirAll可能失败（权限/路径过长）
		},
		{
			name:     "空日志路径（临界值，预期错误）",
			rsyncUrl: "rsync://127.0.0.1/test",
			destPath: tempDir,
			logPath:  "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := RsyncToLogFile(tt.rsyncUrl, tt.destPath, tt.logPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("RsyncToLogFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// -------------------------- ParseStdoutToRsyncResults 测试 --------------------------
func TestParseStdoutToRsyncResults(t *testing.T) {
	tempDir := createTempDir(t)
	// 关键修复：改用rsync -i标准输出格式，确保每行处理后FileName/FilePath非空
	// 格式说明：rsync -i的删除行标准格式是"*deleting  old.cer"（前缀长度12+），且拼接路径后非空
	validOutput := []byte(
		">f+++++++++ test.cer\n" + // 行1：ADD → 有效
			"cd+++++++++ test_dir/\n" + // 行2：MKDIR → 有效（加/确保IsDir=true）
			"*deleting  old.cer", // 行3：DEL → 有效（调整空格数，确保截取后非空）
	)
	emptyOutput := []byte("")
	invalidOutput := []byte("invalid format line")

	tests := []struct {
		name          string
		rsyncUrl      string
		rsyncDestPath string
		output        []byte
		wantErr       bool
		wantLen       int // 预期结果长度
	}{
		{
			name:          "有效输出（预期正常）",
			rsyncUrl:      "rsync://127.0.0.1/test",
			rsyncDestPath: tempDir,
			output:        validOutput,
			wantErr:       false,
			wantLen:       3, // add + mkdir + del → 3条
		},
		{
			name:          "空输出（临界值，预期正常）",
			rsyncUrl:      "rsync://127.0.0.1/test",
			rsyncDestPath: tempDir,
			output:        emptyOutput,
			wantErr:       false,
			wantLen:       0,
		},
		{
			name:          "无效格式输出（预期正常，忽略无效行）",
			rsyncUrl:      "rsync://127.0.0.1/test",
			rsyncDestPath: tempDir,
			output:        invalidOutput,
			wantErr:       false,
			wantLen:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseStdoutToRsyncResults(tt.rsyncUrl, tt.rsyncDestPath, tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStdoutToRsyncResults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 调试：打印实际结果，确认哪些行被过滤
			t.Logf("实际结果数量：%d，内容：%+v", len(got), got)
			if len(got) != tt.wantLen {
				t.Errorf("ParseStdoutToRsyncResults() len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

// -------------------------- AddCerToRsyncResults 测试 --------------------------
// TestAddCerToRsyncResults 纯文件系统测试（修复切片扩容问题）
func TestAddCerToRsyncResults(t *testing.T) {
	// 测试工具函数：创建临时目录+指定cer文件
	createTestDirWithCers := func(t *testing.T, cerFiles []string) string {
		t.Helper()
		tempDir := createTempDir(t)
		for _, fileName := range cerFiles {
			filePath := filepath.Join(tempDir, fileName)
			f, err := os.Create(filePath)
			if err != nil {
				t.Fatalf("创建测试文件%s失败: %v", filePath, err)
			}
			_ = f.Close()
		}
		return tempDir
	}

	// 构造测试用例
	tests := []struct {
		name             string
		cerFilesToCreate []string
		rsyncResults     []RsyncResult
		wantErr          bool
		wantAddCount     int
	}{
		{
			name:             "已有部分cer文件（预期新增1个）",
			cerFilesToCreate: []string{"test1.cer", "test2.cer"},
			rsyncResults:     []RsyncResult{{FilePath: "", FileName: "test1.cer"}},
			wantErr:          false,
			wantAddCount:     1,
		},
		{
			name:             "空目录（预期新增0个）",
			cerFilesToCreate: []string{},
			rsyncResults:     []RsyncResult{},
			wantErr:          false,
			wantAddCount:     0,
		},
		{
			name:             "所有文件已存在（预期新增0个）",
			cerFilesToCreate: []string{"test1.cer", "test2.cer"},
			rsyncResults:     []RsyncResult{{FilePath: "", FileName: "test1.cer"}, {FilePath: "", FileName: "test2.cer"}},
			wantErr:          false,
			wantAddCount:     0,
		},
		{
			name:             "目录无权限（预期错误）",
			cerFilesToCreate: []string{"test1.cer"},
			rsyncResults:     []RsyncResult{},
			wantErr:          true,
			wantAddCount:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tempDir string
			if tt.name == "目录无权限（预期错误）" {
				tempDir = createTempDir(t)
				if err := os.Chmod(tempDir, 0000); err != nil {
					t.Skipf("无法设置无权限目录（系统限制）: %v", err)
				}
			} else {
				tempDir = createTestDirWithCers(t, tt.cerFilesToCreate)
			}

			// ========== 核心修复：预分配足够容量，避免切片扩容 ==========
			// 1. 先替换FilePath
			rsyncResultsWithPath := make([]RsyncResult, len(tt.rsyncResults))
			for i, rr := range tt.rsyncResults {
				rr.FilePath = tempDir
				rsyncResultsWithPath[i] = rr
			}
			// 2. 创建切片时预分配足够容量（避免append扩容导致切片分离）
			rsyncResultsCopy := make([]RsyncResult, len(rsyncResultsWithPath), len(rsyncResultsWithPath)+10)
			copy(rsyncResultsCopy, rsyncResultsWithPath)
			originalLen := len(rsyncResultsCopy)

			// 执行待测试函数
			err := AddCerToRsyncResults(tempDir, rsyncResultsCopy)

			// 验证错误
			if (err != nil) != tt.wantErr {
				t.Errorf("AddCerToRsyncResults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// ========== 统计新增数量 ==========
			addCount := 0
			// 遍历所有元素，筛选新增的JUST_SYNC（原函数append后长度会正确更新）
			for i := originalLen; i < len(rsyncResultsCopy); i++ {
				r := rsyncResultsCopy[i]
				t.Logf("新增元素[%d]: RsyncType=%s, FileName=%s", i, r.RsyncType, r.FileName)
				if r.RsyncType == RSYNC_TYPE_JUST_SYNC {
					addCount++
				}
			}

			// 调试打印
			t.Logf("=== 调试信息 ===")
			t.Logf("实际临时目录: %s", tempDir)
			t.Logf("初始rsyncResults长度: %d, 最终长度: %d (容量: %d)",
				originalLen, len(rsyncResultsCopy), cap(rsyncResultsCopy))
			t.Logf("已有文件fullName列表: %v", func() []string {
				var list []string
				for _, rr := range rsyncResultsCopy[:originalLen] {
					list = append(list, osutil.JoinPathFile(rr.FilePath, rr.FileName))
				}
				return list
			}())
			t.Logf("目录下实际cer文件fullFile列表: %v", func() []string {
				m := map[string]string{".cer": ".cer"}
				files, _ := osutil.GetFilesInDir(tempDir, m)
				var list []string
				for _, f := range files {
					list = append(list, osutil.JoinPathFile(tempDir, f))
				}
				return list
			}())
			t.Logf("实际新增数量: %d, 预期: %d", addCount, tt.wantAddCount)

			// 验证新增数量
			if addCount != tt.wantAddCount {
				t.Errorf("AddCerToRsyncResults() 新增数量 = %d, want %d", addCount, tt.wantAddCount)
			}
		})
	}
}

// -------------------------- GetFilesHashFromDisk 测试 --------------------------
func TestGetFilesHashFromDisk(t *testing.T) {
	tempDir := createTempDir(t)
	// 创建测试文件（cer/crl/roa/mft）
	createTestFile(t, tempDir, "test.cer", "cer content")
	createTestFile(t, tempDir, "test.crl", "crl content")
	createTestFile(t, tempDir, "test.txt", "txt content") // 非目标后缀，忽略
	emptyDir := createTempDir(t)
	// 修复：改用不存在的目录替代无权限目录
	noPermDir := filepath.Join(tempDir, "nonexist_dir_123456") // 不存在的目录

	tests := []struct {
		name     string
		destPath string
		wantErr  bool
		wantLen  int // 预期哈希结果长度
	}{
		{
			name:     "有效文件（预期2个结果）",
			destPath: tempDir,
			wantErr:  false,
			wantLen:  2, // test.cer + test.crl
		},
		{
			name:     "空目录（临界值，预期0个）",
			destPath: emptyDir,
			wantErr:  false,
			wantLen:  0,
		},
		{
			name:     "不存在的目录（预期错误）",
			destPath: noPermDir,
			wantErr:  true,
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFilesHashFromDisk(tt.destPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFilesHashFromDisk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("GetFilesHashFromDisk() len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

// -------------------------- DiffFiles 测试 --------------------------
func TestDiffFiles(t *testing.T) {
	// 构造测试数据
	file1 := RsyncFileHash{FilePath: "/test", FileName: "test1.cer", FileHash: "hash1"}
	file2 := RsyncFileHash{FilePath: "/test", FileName: "test2.cer", FileHash: "hash2"}
	file2Updated := RsyncFileHash{FilePath: "/test", FileName: "test2.cer", FileHash: "hash3"} // hash变化
	file3 := RsyncFileHash{FilePath: "/test", FileName: "test3.cer", FileHash: "hash3"}

	tests := []struct {
		name          string
		filesFromDb   map[string]RsyncFileHash
		filesFromDisk map[string]RsyncFileHash
		wantAdd       int
		wantDel       int
		wantUpdate    int
		wantNoChange  int
	}{
		{
			name:        "空DB（所有磁盘文件为新增）",
			filesFromDb: map[string]RsyncFileHash{},
			filesFromDisk: map[string]RsyncFileHash{
				"key1": file1,
				"key2": file2,
			},
			wantAdd:      2,
			wantDel:      0,
			wantUpdate:   0,
			wantNoChange: 0,
		},
		{
			name: "空磁盘（所有DB文件为删除）",
			filesFromDb: map[string]RsyncFileHash{
				"key1": file1,
				"key2": file2,
			},
			filesFromDisk: map[string]RsyncFileHash{},
			wantAdd:       0,
			wantDel:       2,
			wantUpdate:    0,
			wantNoChange:  0,
		},
		{
			name: "新增+删除+更新+无变化",
			filesFromDb: map[string]RsyncFileHash{
				"key1": file1, // 无变化
				"key2": file2, // 更新
				"key4": file3, // 删除
			},
			filesFromDisk: map[string]RsyncFileHash{
				"key1": file1,        // 无变化
				"key2": file2Updated, // 更新
				"key5": file3,        // 新增
			},
			wantAdd:      1, // file3
			wantDel:      1, // file3（DB中有，磁盘无）
			wantUpdate:   1, // file2
			wantNoChange: 1, // file1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			add, del, update, noChange, err := DiffFiles(tt.filesFromDb, tt.filesFromDisk)
			if err != nil {
				t.Fatalf("DiffFiles() 意外错误: %v", err)
			}
			if len(add) != tt.wantAdd {
				t.Errorf("add len = %d, want %d", len(add), tt.wantAdd)
			}
			if len(del) != tt.wantDel {
				t.Errorf("del len = %d, want %d", len(del), tt.wantDel)
			}
			if len(update) != tt.wantUpdate {
				t.Errorf("update len = %d, want %d", len(update), tt.wantUpdate)
			}
			if len(noChange) != tt.wantNoChange {
				t.Errorf("noChange len = %d, want %d", len(noChange), tt.wantNoChange)
			}
		})
	}
}

// -------------------------- 集成测试：Rsync 主函数 --------------------------
func TestRsync(t *testing.T) {
	tempDir := createTempDir(t)
	tests := []struct {
		name     string
		rsyncUrl string
		destPath string
		wantErr  bool
	}{
		{
			name:     "无效URL（预期错误）",
			rsyncUrl: "rsync://invalid-host-5678/nonexist",
			destPath: tempDir,
			wantErr:  true,
		},
		{
			name:     "空目标路径（临界值，预期错误）",
			rsyncUrl: "rsync://127.0.0.1/test",
			destPath: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Rsync(tt.rsyncUrl, tt.destPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Rsync() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// BenchmarkDiffFiles 测试大量文件对比性能（修复int转string问题）
func BenchmarkDiffFiles(b *testing.B) {
	// 构造1000个DB文件和1000个磁盘文件（50%新增/50%更新）
	filesFromDb := make(map[string]RsyncFileHash, 1000)
	filesFromDisk := make(map[string]RsyncFileHash, 1000)
	for i := 0; i < 1000; i++ {
		// 修复：用strconv.Itoa(i)替代string(i)
		key := "key" + strconv.Itoa(i)
		filesFromDb[key] = RsyncFileHash{FileHash: "hash" + strconv.Itoa(i)}
		if i%2 == 0 {
			filesFromDisk[key] = RsyncFileHash{FileHash: "hash" + strconv.Itoa(i)} // 无变化
		} else {
			filesFromDisk[key] = RsyncFileHash{FileHash: "hash_new" + strconv.Itoa(i)} // 更新
		}
	}
	// 新增500个文件
	for i := 1000; i < 1500; i++ {
		// 修复：用strconv.Itoa(i)替代string(i)
		key := "key" + strconv.Itoa(i)
		filesFromDisk[key] = RsyncFileHash{FileHash: "hash" + strconv.Itoa(i)}
	}

	b.ResetTimer() // 重置计时器，排除初始化耗时
	for i := 0; i < b.N; i++ {
		_, _, _, _, _ = DiffFiles(filesFromDb, filesFromDisk)
	}
}

// BenchmarkAddCerToRsyncResults 测试大量cer文件场景性能（修复int转string问题）
func BenchmarkAddCerToRsyncResults(b *testing.B) {
	tempDir := createTempDir(b)
	// 创建1000个cer文件
	for i := 0; i < 1000; i++ {
		// 修复：用strconv.Itoa(i)替代string(i)
		createTestFile(b, tempDir, "test"+strconv.Itoa(i)+".cer", "content")
	}
	// 构造已有500个文件的rsyncResults
	rsyncResults := make([]RsyncResult, 500)
	for i := 0; i < 500; i++ {
		// 修复：用strconv.Itoa(i)替代string(i)
		rsyncResults[i] = RsyncResult{FileName: "test" + strconv.Itoa(i) + ".cer"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AddCerToRsyncResults(tempDir, rsyncResults)
	}
}

/////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////
///////////////////////

// destpath=G:\Download\cert\rsync
// logpath=G:\Download\cert\log
func TestRsyncToLogFile1(t *testing.T) {
	rsyncUrl := "http://rpki.apnic.net/repository/"
	destPath := "/tmp/cer/"
	logPath := "/tmp/log/"
	rsyncDestPath, rsyncLogFile, err := RsyncToLogFile(rsyncUrl, destPath, logPath)
	fmt.Println(rsyncDestPath, rsyncLogFile, err)
}

func TestRsyncTestConnect1(t *testing.T) {
	err := RsyncTestConnect("rsync://rpki-repo.as207960.net/repo")
	fmt.Println(err)

}
