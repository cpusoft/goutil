package hashutil

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// -------------------------- 单元测试（覆盖正常场景 + 临界值） --------------------------

// TestMd5 测试Md5函数，覆盖空字符串、普通字符串、特殊字符、长字符串（临界值）
func TestMd5(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string", // 临界值：空输入
			input: "",
			want:  "d41d8cd98f00b204e9800998ecf8427e", // 空字符串的MD5固定值
		},
		{
			name:  "normal string",
			input: "hello world",
			want:  "5eb63bbbe01eeed093cb22bb8f5acdc3",
		},
		{
			name:  "special chars", // 临界值：特殊字符
			input: "!@#$%^&*()_+-=[]{}|;':\",./<>?",
			want:  "55f63ea4fdd78ef5e227f735e191afdc", // 正确值
		},
		{
			name:  "long string", // 临界值：长字符串（1024字节全0）
			input: string(make([]byte, 1024)),
			want:  "0f343b0931126a20f133d67c2b018a3b", // 正确值
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Md5(tt.input); got != tt.want {
				t.Errorf("Md5() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHmac 测试Hmac函数，覆盖空key/空data、普通key/data、特殊字符
func TestHmac(t *testing.T) {
	tests := []struct {
		name string
		key  string
		data string
		want string
	}{
		{
			name: "empty key and data", // 临界值：空key+空data
			key:  "",
			data: "",
			want: "74e6f7298a9c2d168935f58c001bad88",
		},
		{
			name: "normal key and data",
			key:  "secret",
			data: "hello world",
			want: "78d6997b1230f38e59b6d1642dfaa3a4", // 正确值
		},
		{
			name: "special key", // 临界值：特殊字符key
			key:  "!@#$%^&*()",
			data: "test",
			want: "5136fb2f34d13aff153b721cdb0e05a1", // 修正：辅助函数输出的正确值
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Hmac(tt.key, tt.data); got != tt.want {
				t.Errorf("Hmac(key=%v, data=%v) = %v, want %v", tt.key, tt.data, got, tt.want)
			}
		})
	}
}

// TestSha1 测试Sha1函数，覆盖空字节、普通字节、长字节（临界值）
func TestSha1(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "empty bytes", // 临界值：空输入
			input: []byte(""),
			want:  "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		},
		{
			name:  "normal bytes",
			input: []byte("hello world"),
			want:  "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed",
		},
		{
			name:  "long bytes", // 临界值：长字节（1024字节全0）
			input: make([]byte, 1024),
			want:  "60cacbf3d72e1e7834203da608037b1bf83b40e8", // 正确值
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sha1(tt.input); got != tt.want {
				t.Errorf("Sha1() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSha256 测试Sha256函数，覆盖空字节、普通字节、长字节（临界值）
func TestSha256(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "empty bytes", // 临界值：空输入
			input: []byte(""),
			want:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:  "normal bytes",
			input: []byte("hello world"),
			want:  "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name:  "long bytes", // 临界值：长字节（1024字节全0）
			input: make([]byte, 1024),
			want:  "5f70bf18a086007016e948b04aed3b82103a36bea41755b6cddfaf10ace3c6ef", // 正确值
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sha256(tt.input); got != tt.want {
				t.Errorf("Sha256() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSha256File 测试Sha256File函数，覆盖不存在文件、空文件、普通文件、大文件（临界值）
func TestSha256File(t *testing.T) {
	// 辅助函数：创建临时文件并写入内容，返回路径，测试结束后清理
	createTempFile := func(t *testing.T, content []byte) string {
		t.Helper()
		tmpFile, err := os.CreateTemp("", "hashutil_test_*.txt")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		// 测试结束后删除临时文件
		t.Cleanup(func() {
			_ = os.Remove(tmpFile.Name())
		})

		if len(content) > 0 {
			if _, err := tmpFile.Write(content); err != nil {
				t.Fatalf("failed to write to temp file: %v", err)
			}
		}
		_ = tmpFile.Close()
		return tmpFile.Name()
	}

	tests := []struct {
		name    string
		file    string
		want    string
		wantErr bool // 是否期望错误
	}{
		{
			name:    "non-existent file", // 临界值：不存在的文件
			file:    filepath.Join(t.TempDir(), "non_exist.txt"),
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty file", // 临界值：空文件
			file:    createTempFile(t, []byte("")),
			want:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantErr: false,
		},
		{
			name:    "normal file",
			file:    createTempFile(t, []byte("hello world")),
			want:    "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			wantErr: false,
		},
		{
			name:    "large file (1MB)", // 临界值：大文件（1MB全0）
			file:    createTempFile(t, make([]byte, 1024*1024)),
			want:    "30e14955ebf1352266dc2ff8067e68104607e750abb9d3b36582b8af909fcb58", // 正确值
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sha256File(tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sha256File() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Sha256File() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHashFileByte 测试HashFileByte函数，覆盖不存在文件、空文件、普通文件、大文件（临界值）
func TestHashFileByte(t *testing.T) {
	// 辅助函数：创建临时文件
	createTempFile := func(t *testing.T, content []byte) string {
		t.Helper()
		tmpFile, err := os.CreateTemp("", "hashutil_test_*.txt")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Remove(tmpFile.Name())
		})

		if len(content) > 0 {
			if _, err := tmpFile.Write(content); err != nil {
				t.Fatalf("failed to write to temp file: %v", err)
			}
		}
		_ = tmpFile.Close()
		return tmpFile.Name()
	}

	// 空文件的Sha256字节数组
	emptyFileHash := [32]byte{
		0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14,
		0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9, 0x24,
		0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c,
		0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55,
	}

	// 普通文件（hello world）的Sha256字节数组
	normalFileHash := [32]byte{
		0xb9, 0x4d, 0x27, 0xb9, 0x93, 0x4d, 0x3e, 0x08,
		0xa5, 0x2e, 0x52, 0xd7, 0xda, 0x7d, 0xab, 0xfa,
		0xc4, 0x84, 0xef, 0xe3, 0x7a, 0x53, 0x80, 0xee,
		0x90, 0x88, 0xf7, 0xac, 0xe2, 0xef, 0xcd, 0xe9,
	}

	// 1MB全0文件的Sha256字节数组（正确值）
	largeFileHash := [32]byte{
		0x30, 0xe1, 0x49, 0x55, 0xeb, 0xf1, 0x35, 0x22,
		0x66, 0xdc, 0x2f, 0xf8, 0x06, 0x7e, 0x68, 0x10,
		0x46, 0x07, 0xe7, 0x50, 0xab, 0xb9, 0xd3, 0xb3,
		0x65, 0x82, 0xb8, 0xaf, 0x90, 0x9f, 0xcb, 0x58,
	}

	tests := []struct {
		name    string
		file    string
		want    [32]byte
		wantErr bool
	}{
		{
			name:    "non-existent file", // 临界值：不存在的文件
			file:    filepath.Join(t.TempDir(), "non_exist.txt"),
			want:    [32]byte{},
			wantErr: true,
		},
		{
			name:    "empty file", // 临界值：空文件
			file:    createTempFile(t, []byte("")),
			want:    emptyFileHash,
			wantErr: false,
		},
		{
			name:    "normal file",
			file:    createTempFile(t, []byte("hello world")),
			want:    normalFileHash,
			wantErr: false,
		},
		{
			name:    "large file (1MB)", // 临界值：大文件（1MB全0）
			file:    createTempFile(t, make([]byte, 1024*1024)),
			want:    largeFileHash, // 正确值
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HashFileByte(tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashFileByte() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HashFileByte() = %v, want %v", got, tt.want)
			}
		})
	}
}

// -------------------------- 性能测试（Benchmark） --------------------------

// BenchmarkMd5_SmallData 测试Md5小数据（10字节）性能
func BenchmarkMd5_SmallData(b *testing.B) {
	data := []byte("test123456") // 10字节小数据
	b.ResetTimer()               // 重置计时器，排除数据准备耗时
	for i := 0; i < b.N; i++ {
		Md5(string(data))
	}
}

// BenchmarkMd5_LargeData 测试Md5大数据（1MB）性能
func BenchmarkMd5_LargeData(b *testing.B) {
	data := make([]byte, 1024*1024) // 1MB大数据
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Md5(string(data))
	}
}

// BenchmarkHmac_SmallData 测试Hmac小数据性能
func BenchmarkHmac_SmallData(b *testing.B) {
	key := "secret"
	data := []byte("test123456")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Hmac(key, string(data))
	}
}

// BenchmarkHmac_LargeData 测试Hmac大数据（1MB）性能
func BenchmarkHmac_LargeData(b *testing.B) {
	key := "secret"
	data := make([]byte, 1024*1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Hmac(key, string(data))
	}
}

// BenchmarkSha1_SmallData 测试Sha1小数据性能
func BenchmarkSha1_SmallData(b *testing.B) {
	data := []byte("test123456")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sha1(data)
	}
}

// BenchmarkSha1_LargeData 测试Sha1大数据（1MB）性能
func BenchmarkSha1_LargeData(b *testing.B) {
	data := make([]byte, 1024*1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sha1(data)
	}
}

// BenchmarkSha256_SmallData 测试Sha256小数据性能
func BenchmarkSha256_SmallData(b *testing.B) {
	data := []byte("test123456")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sha256(data)
	}
}

// BenchmarkSha256_LargeData 测试Sha256大数据（1MB）性能
func BenchmarkSha256_LargeData(b *testing.B) {
	data := make([]byte, 1024*1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sha256(data)
	}
}

// BenchmarkSha256File_SmallFile 测试Sha256File小文件（1KB）性能
func BenchmarkSha256File_SmallFile(b *testing.B) {
	// 创建1KB临时文件
	tmpFile, err := os.CreateTemp("", "bench_hashutil_*.txt")
	if err != nil {
		b.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write(make([]byte, 1024))
	_ = tmpFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Sha256File(tmpFile.Name())
		if err != nil {
			b.Fatalf("Sha256File failed: %v", err)
		}
	}
}

// BenchmarkSha256File_LargeFile 测试Sha256File大文件（1MB）性能
func BenchmarkSha256File_LargeFile(b *testing.B) {
	// 创建1MB临时文件
	tmpFile, err := os.CreateTemp("", "bench_hashutil_*.txt")
	if err != nil {
		b.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write(make([]byte, 1024*1024))
	_ = tmpFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Sha256File(tmpFile.Name())
		if err != nil {
			b.Fatalf("Sha256File failed: %v", err)
		}
	}
}

// BenchmarkHashFileByte_SmallFile 测试HashFileByte小文件（1KB）性能
func BenchmarkHashFileByte_SmallFile(b *testing.B) {
	// 创建1KB临时文件
	tmpFile, err := os.CreateTemp("", "bench_hashutil_*.txt")
	if err != nil {
		b.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write(make([]byte, 1024))
	_ = tmpFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HashFileByte(tmpFile.Name())
		if err != nil {
			b.Fatalf("HashFileByte failed: %v", err)
		}
	}
}

// BenchmarkHashFileByte_LargeFile 测试HashFileByte大文件（1MB）性能
func BenchmarkHashFileByte_LargeFile(b *testing.B) {
	// 创建1MB临时文件
	tmpFile, err := os.CreateTemp("", "bench_hashutil_*.txt")
	if err != nil {
		b.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write(make([]byte, 1024*1024))
	_ = tmpFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HashFileByte(tmpFile.Name())
		if err != nil {
			b.Fatalf("HashFileByte failed: %v", err)
		}
	}
}

// -------------------------- 辅助验证函数（可选） --------------------------
// 用于手动验证哈希值的正确性（可删除）
func TestVerifyHashValues(t *testing.T) {
	// 验证特殊key的HMAC值
	key := "!@#$%^&*()"
	data := "test"
	h := hmac.New(md5.New, []byte(key))
	h.Write([]byte(data))
	actual := hex.EncodeToString(h.Sum(nil))
	t.Logf("Hmac(key=%s, data=%s) = %s", key, data, actual)
	// 运行测试时，复制这个输出到TestHmac的special_key用例的want中即可
}

/////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////

func TestSha2561(t *testing.T) {
	s := []byte{0x01, 0x02, 0x02}
	sh := Sha256(s)
	fmt.Println(sh)
}

func TestSha256File1(t *testing.T) {
	s := `G:\Download\undefined.txt`
	sh, err := Sha256File(s)
	fmt.Println(sh, err)

}

func TestSha256Password(t *testing.T) {
	p := Sha256([]byte("2e869b49-50c8-487b-ab1a-67c87c77ccc0" + "abc123!@#"))
	fmt.Println(p)
}

func TestSha256String(t *testing.T) {
	s := `
        MIIDFjCCAf4CAQEwDQYJKoZIhvcNAQELBQAwRjERMA8GA1UEAxMIQTkxOUI2M0MxMTAvBgNV
BAUTKDI1ODVEQTBCOTgwQTQ3RkVCQTBFMjM1MjA1REVFRTQwMkYyMEIzQ0IXDTE5MTEyNTAzMDAxN1oX
DTE5MTEyNzAzMDAxN1owggFQMBMCAihzFw0xOTExMjEwOTAyMTVaMBMCAih0Fw0xOTExMjExNTAwMTBa
MBMCAih1Fw0xOTExMjEyMTAwMjFaMBMCAih2Fw0xOTExMjIwMzAwMDRaMBMCAih3Fw0xOTExMjIwOTAx
MjJaMBMCAih4Fw0xOTExMjIxNTAwMzRaMBMCAih5Fw0xOTExMjIyMDU5MjlaMBMCAih6Fw0xOTExMjMw
MzAwMzdaMBMCAih7Fw0xOTExMjMwOTAxMzhaMBMCAih8Fw0xOTExMjMxNTAxNTFaMBMCAih9Fw0xOTEx
MjMyMDU5NTJaMBMCAih+Fw0xOTExMjQwMzAxMDZaMBMCAih/Fw0xOTExMjQwOTAwNDRaMBMCAiiAFw0x
OTExMjQxNTAwMzhaMBMCAiiBFw0xOTExMjQyMDU5NDVaMBMCAiiCFw0xOTExMjUwMzAwMTZaoDAwLjAf
BgNVHSMEGDAWgBQlhdoLmApH/roOI1IF3u5ALyCzyzALBgNVHRQEBAICUP0wDQYJKoZIhvcNAQELBQAD
ggEBAFmUueWNFT9n56ZJlDGwbwDmgUMBuS87ypRi+xRk1+cuM4+nYE/pWpLejkB8+AObUhrAiiun1VQa
06oXIu14X+/YREkaquSxPh4K1oHJY/bBQRaUxOj6elvhSXiCaplc4TLV2voTBCYBW3SZR06U5exq9KIh
LiocMrCTZOWRvcKs0DfbZUoCx8fm0XGDTSwiLhEkcJmyT5BxkFbrZXcCwvminNgk/iPqNVDm/MOtISX6
KCOuSHDD6gScUamzCy2jCT5truL2iKrb8xk+Yp5SUAA2TnGV/c6ToLuGU4DRZ/vsTDY4eomfxH+yfRqI
MhT+jBXdOpyAl2OE6yWR15SqrRE=
      `
	sh := Sha256([]byte(s))
	fmt.Println(sh)
	s1 := strings.Replace(s, "\r", "", -1)
	s2 := strings.Replace(s1, "\n", "", -1)
	s3 := strings.TrimSpace(s2)
	sh = Sha256([]byte(s3))
	fmt.Println(sh)

}
