package talutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ===================== 单元测试 - GetAllTalFile =====================
func TestGetAllTalFile(t *testing.T) {
	// 场景1：传入单个合法.tal文件
	t.Run("single valid tal file", func(t *testing.T) {
		tmpDir := t.TempDir()
		talFile := filepath.Join(tmpDir, "single.tal")
		if err := os.WriteFile(talFile, []byte("test"), 0644); err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}

		files, err := GetAllTalFile(talFile)
		if err != nil {
			t.Fatalf("GetAllTalFile执行失败: %v", err)
		}
		if len(files) != 1 || files[0] != talFile {
			t.Errorf("期望返回1个文件，实际返回%d个，内容=%v", len(files), files)
		}
	})

	// 场景2：传入包含多个.tal和非.tal文件的目录
	t.Run("dir with tal and non-tal files", func(t *testing.T) {
		tmpDir := t.TempDir()
		talFile1 := filepath.Join(tmpDir, "test1.tal")
		talFile2 := filepath.Join(tmpDir, "test2.tal")
		nonTalFile1 := filepath.Join(tmpDir, "test3.txt")
		nonTalFile2 := filepath.Join(tmpDir, "test4.log")
		_ = os.WriteFile(talFile1, []byte(""), 0644)
		_ = os.WriteFile(talFile2, []byte(""), 0644)
		_ = os.WriteFile(nonTalFile1, []byte(""), 0644)
		_ = os.WriteFile(nonTalFile2, []byte(""), 0644)

		files, err := GetAllTalFile(tmpDir)
		if err != nil {
			t.Fatalf("GetAllTalFile执行失败: %v", err)
		}
		if len(files) != 2 {
			t.Errorf("期望返回2个.tal文件，实际返回%d个，内容=%v", len(files), files)
		}
		fileSet := make(map[string]bool)
		for _, f := range files {
			fileSet[f] = true
		}
		if !fileSet[talFile1] || !fileSet[talFile2] {
			t.Error("返回的文件不是预期的.tal文件")
		}
	})

	// 场景3：传入空目录（无任何文件）
	t.Run("empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		files, err := GetAllTalFile(tmpDir)
		if err != nil {
			t.Fatalf("GetAllTalFile执行失败: %v", err)
		}
		if len(files) != 0 {
			t.Errorf("期望返回0个文件，实际返回%d个", len(files))
		}
	})

	// 场景4：传入不存在的路径（异常场景）
	t.Run("non-existent path", func(t *testing.T) {
		nonExistPath := filepath.Join(t.TempDir(), "non_exist_1234.tal")
		_, err := GetAllTalFile(nonExistPath)
		if err == nil {
			t.Error("期望返回错误，但实际未返回")
		}
	})

	// 场景5：传入单个非.tal文件（边界场景）
	t.Run("single non-tal file", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonTalFile := filepath.Join(tmpDir, "test.txt")
		_ = os.WriteFile(nonTalFile, []byte("test"), 0644)

		files, err := GetAllTalFile(nonTalFile)
		if err != nil {
			t.Fatalf("GetAllTalFile执行失败: %v", err)
		}
		if len(files) != 1 || files[0] != nonTalFile {
			t.Errorf("非.tal文件应被直接返回，实际返回=%v", files)
		}
	})
}

// ===================== 单元测试 - ParseTalInfos/parseTalInfo =====================
func TestParseTalInfos(t *testing.T) {
	// 场景1：正常文件（含空行，验证跳过空行逻辑）
	t.Run("normal file with empty lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		talFile := filepath.Join(tmpDir, "normal_with_empty.tal")
		content := `https://sync.example.com

			-----BEGIN PUBLIC KEY-----
			
			MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxZ8+
			-----END PUBLIC KEY-----
		`
		if err := os.WriteFile(talFile, []byte(content), 0644); err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}

		files := []string{talFile}
		talInfos, err := ParseTalInfos(files)
		if err != nil {
			t.Fatalf("ParseTalInfos执行失败: %v", err)
		}

		if talInfos[0].SyncUrl != "https://sync.example.com" {
			t.Errorf("SyncUrl解析错误，期望=https://sync.example.com，实际=%s", talInfos[0].SyncUrl)
		}
		expectedPubKey := "-----BEGINPUBLICKEY-----MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxZ8+-----ENDPUBLICKEY-----"
		actualPubKey := strings.ReplaceAll(talInfos[0].PubKey, " ", "")
		if actualPubKey != expectedPubKey {
			t.Errorf("PubKey解析错误，期望=%s，实际=%s", expectedPubKey, actualPubKey)
		}
	})

	// 场景2：空文件（临界值）
	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		talFile := filepath.Join(tmpDir, "empty.tal")
		_ = os.WriteFile(talFile, []byte(""), 0644)

		files := []string{talFile}
		talInfos, err := ParseTalInfos(files)
		if err != nil {
			t.Fatalf("ParseTalInfos执行失败: %v", err)
		}

		if len(talInfos) != 1 || talInfos[0].SyncUrl != "" || talInfos[0].PubKey != "" {
			t.Errorf("空文件解析错误，talInfo=%+v", talInfos[0])
		}
	})

	// 场景3：只有一行的文件（临界值）
	t.Run("single line file", func(t *testing.T) {
		tmpDir := t.TempDir()
		talFile := filepath.Join(tmpDir, "single_line.tal")
		_ = os.WriteFile(talFile, []byte("https://single.example.com"), 0644)

		files := []string{talFile}
		talInfos, err := ParseTalInfos(files)
		if err != nil {
			t.Fatalf("ParseTalInfos执行失败: %v", err)
		}

		if talInfos[0].SyncUrl != "https://single.example.com" || talInfos[0].PubKey != "" {
			t.Errorf("单行文件解析错误，talInfo=%+v", talInfos[0])
		}
	})

	// 场景4：超大行文件（临界值，支持2MB单行）
	t.Run("extra large line file", func(t *testing.T) {
		tmpDir := t.TempDir()
		talFile := filepath.Join(tmpDir, "large_line.tal")

		// 生成100KB的SyncUrl、200KB的PubKey（均在2MB限制内）
		largeSyncUrl := strings.Repeat("a", 1024*100)
		largePubKey := strings.Repeat("b", 1024*200)
		content := largeSyncUrl + "\n" + largePubKey

		_ = os.WriteFile(talFile, []byte(content), 0644)

		files := []string{talFile}
		talInfos, err := ParseTalInfos(files)
		if err != nil {
			t.Fatalf("ParseTalInfos执行失败: %v", err)
		}

		if len(talInfos[0].SyncUrl) != 1024*100 || len(talInfos[0].PubKey) != 1024*200 {
			t.Errorf("超大行解析错误，SyncUrl长度=%d, PubKey长度=%d", len(talInfos[0].SyncUrl), len(talInfos[0].PubKey))
		}
	})

	// 场景5：无读取权限文件（异常场景，兼容Unix/Linux权限）
	t.Run("no read permission file", func(t *testing.T) {
		tmpDir := t.TempDir()
		talFile := filepath.Join(tmpDir, "no_perm.tal")
		// 1. 创建文件并写入内容
		if err := os.WriteFile(talFile, []byte("test"), 0644); err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}
		// 2. 修改文件权限为0200（仅属主可写，无读权限）
		if err := os.Chmod(talFile, 0200); err != nil {
			t.Skipf("系统不支持修改文件权限，跳过该测试: %v", err)
		}
		// 3. 尝试打开文件验证权限（提前验证，确保测试有效）
		_, err := os.Open(talFile)
		if err == nil {
			t.Skip("当前用户仍有读权限，跳过该测试（可能是root/管理员权限）")
		}

		files := []string{talFile}
		_, err = ParseTalInfos(files)
		if err == nil {
			t.Error("期望返回权限错误，但实际未返回")
		}
	})

	// 场景6：多个文件解析（正常+异常）
	t.Run("multiple files parse", func(t *testing.T) {
		tmpDir := t.TempDir()
		validFile := filepath.Join(tmpDir, "valid.tal")
		invalidFile := filepath.Join(tmpDir, "invalid.tal")
		_ = os.WriteFile(validFile, []byte("https://test.com\npubkey"), 0644)

		files := []string{validFile, invalidFile}
		_, err := ParseTalInfos(files)
		if err == nil {
			t.Error("解析不存在的文件时，期望返回错误，但实际未返回")
		}
	})
}

// ===================== 性能测试 =====================
func BenchmarkParseTalInfos_SingleLargeFile(b *testing.B) {
	tmpDir := b.TempDir()
	talFile := filepath.Join(tmpDir, "bench_large.tal")

	syncUrl := strings.Repeat("sync_", 1024*128)
	pubKey := strings.Repeat("pubkey_", 1024*1152)
	content := syncUrl + "\n" + pubKey
	if err := os.WriteFile(talFile, []byte(content), 0644); err != nil {
		b.Fatalf("创建基准测试文件失败: %v", err)
	}
	files := []string{talFile}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseTalInfos(files)
		if err != nil {
			b.Fatalf("基准测试执行失败: %v", err)
		}
	}
}

func BenchmarkParseTalInfos_ManySmallFiles(b *testing.B) {
	tmpDir := b.TempDir()
	var files []string
	for i := 0; i < 100; i++ {
		fileName := filepath.Join(tmpDir, string(rune(i))+".tal")
		_ = os.WriteFile(fileName, []byte("https://test.com\npubkey"+string(rune(i))), 0644)
		files = append(files, fileName)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseTalInfos(files)
		if err != nil {
			b.Fatalf("基准测试执行失败: %v", err)
		}
	}
}

func BenchmarkGetAllTalFile_LargeDir(b *testing.B) {
	tmpDir := b.TempDir()
	for i := 0; i < 500; i++ {
		talFile := filepath.Join(tmpDir, "tal_"+string(rune(i))+".tal")
		nonTalFile := filepath.Join(tmpDir, "non_tal_"+string(rune(i))+".txt")
		_ = os.WriteFile(talFile, []byte(""), 0644)
		_ = os.WriteFile(nonTalFile, []byte(""), 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetAllTalFile(tmpDir)
		if err != nil {
			b.Fatalf("基准测试执行失败: %v", err)
		}
	}
}
