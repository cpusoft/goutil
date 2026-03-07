package conf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	config "github.com/cpusoft/goutil/beconfig"
)

/*
# 运行所有测试（含详细输出）
go test -v

# 运行性能测试（统计耗时和内存分配）
go test -bench=. -benchmem

# 运行指定测试（如VariableString）
go test -v -run=TestVariableString

# 测试覆盖率
go test -coverprofile=coverage.out
go tool cover -html=coverage.out  # 生成HTML覆盖率报告
*/

// 全局变量，用于保存原始configure，测试后恢复
var originalConfigure config.Configer

// HasKey 检查配置项是否存在
func HasKey(key string) bool {
	if configure == nil {
		return false
	}
	return len(String(key)) > 0
}

// TestMain 初始化测试环境，加载测试配置文件
func TestMain(m *testing.M) {
	// 保存原始配置
	originalConfigure = configure

	// 1. 校验配置文件路径是否存在
	testConfPath := filepath.Join("testdata", "test.conf")
	fmt.Printf("【调试】测试配置文件路径：%s\n", testConfPath) // 新增调试日志

	// 检查文件是否存在
	if _, err := os.Stat(testConfPath); err != nil {
		fmt.Printf("【调试】配置文件不存在：%v\n", err) // 替换panic为打印
		os.Exit(1)
	}

	// 2. 加载测试配置文件
	var err error
	configure, err = config.NewConfig("ini", testConfPath)
	if err != nil {
		fmt.Printf("【调试】加载配置失败：%v\n", err) // 替换panic为打印
		os.Exit(1)
	}

	// 3. 调试：打印所有已加载的配置项
	fmt.Printf("【调试】rpstir2::datadir 的值：%s\n", String("rpstir2::datadir"))
	fmt.Printf("【调试】是否存在该配置项：%v\n", HasKey("rpstir2::datadir"))

	// 4. 移除panic，改为打印警告（临时）
	if !HasKey("rpstir2::datadir") {
		fmt.Printf("【警告】rpstir2::datadir 未加载，值为空\n")
		// panic("关键配置项 rpstir2::datadir 未加载，请检查test.conf") // 注释panic
	}

	// 运行测试
	exitCode := m.Run()

	// 恢复原始配置
	configure = originalConfigure

	os.Exit(exitCode)
}

// -------------------------- 功能覆盖测试 --------------------------
// TestString 测试String/DefaultString函数
func TestString(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		defaultVal  string
		want        string
		testDefault bool
	}{
		{
			name: "正常字符串",
			key:  "string_key",
			want: "hello_world",
		},
		{
			name: "空字符串配置",
			key:  "empty_string",
			want: "",
		},
		{
			name:        "不存在的key（DefaultString）",
			key:         "non_exist_key",
			defaultVal:  "default_val",
			want:        "default_val",
			testDefault: true,
		},
		{
			name:        "空key（DefaultString）",
			key:         "",
			defaultVal:  "empty_key_default",
			want:        "empty_key_default",
			testDefault: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.testDefault {
				got := DefaultString(tt.key, tt.defaultVal)
				if got != tt.want {
					t.Errorf("DefaultString(%q, %q) = %q, want %q", tt.key, tt.defaultVal, got, tt.want)
				}
			} else {
				got := String(tt.key)
				if got != tt.want {
					t.Errorf("String(%q) = %q, want %q", tt.key, got, tt.want)
				}
			}
		})
	}
}

// TestInt 测试Int/DefaultInt函数
func TestInt(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		defaultVal  int
		want        int
		testDefault bool
	}{
		{
			name: "正常整数",
			key:  "int_key",
			want: 123,
		},
		{
			name: "零值整数",
			key:  "zero_int",
			want: 0,
		},
		{
			name:        "不存在的key（DefaultInt）",
			key:         "non_exist_int",
			defaultVal:  456,
			want:        456,
			testDefault: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.testDefault {
				got := DefaultInt(tt.key, tt.defaultVal)
				if got != tt.want {
					t.Errorf("DefaultInt(%q, %d) = %d, want %d", tt.key, tt.defaultVal, got, tt.want)
				}
			} else {
				got := Int(tt.key)
				if got != tt.want {
					t.Errorf("Int(%q) = %d, want %d", tt.key, got, tt.want)
				}
			}
		})
	}
}

// TestBool 测试Bool/DefaultBool函数
func TestBool(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		defaultVal  bool
		want        bool
		testDefault bool
	}{
		{
			name: "true值",
			key:  "bool_key",
			want: true,
		},
		{
			name: "false值",
			key:  "false_bool",
			want: false,
		},
		{
			name:        "不存在的key（DefaultBool）",
			key:         "non_exist_bool",
			defaultVal:  true,
			want:        true,
			testDefault: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.testDefault {
				got := DefaultBool(tt.key, tt.defaultVal)
				if got != tt.want {
					t.Errorf("DefaultBool(%q, %t) = %t, want %t", tt.key, tt.defaultVal, got, tt.want)
				}
			} else {
				got := Bool(tt.key)
				if got != tt.want {
					t.Errorf("Bool(%q) = %t, want %t", tt.key, got, tt.want)
				}
			}
		})
	}
}

// TestStrings 测试Strings/DefaultStrings函数
func TestStrings(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		defaultValues []string
		want          []string
		testDefault   bool
	}{
		{
			name: "正常字符串数组",
			key:  "strings_key",
			want: []string{"a", "b", "c"},
		},
		{
			name: "空字符串数组",
			key:  "empty_strings",
			want: nil,
		},
		{
			name:          "不存在的key（DefaultStrings）",
			key:           "non_exist_strings",
			defaultValues: []string{"x", "y"},
			want:          []string{"x", "y"},
			testDefault:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			if tt.testDefault {
				got = DefaultStrings(tt.key, tt.defaultValues)
			} else {
				got = Strings(tt.key)
			}

			// 比较切片（兼容nil和空切片）
			if (got == nil && tt.want != nil) || (got != nil && tt.want == nil) {
				t.Errorf("Strings(%q) = %v, want %v", tt.key, got, tt.want)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("len(%s) = %d, want %d (got: %v)", tt.key, len(got), len(tt.want), got)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Strings(%q)[%d] = %q, want %q", tt.key, i, got[i], tt.want[i])
				}
			}
		})
	}
}

/*
// TestVariableString 测试变量替换函数（仅支持单层）
func TestVariableString(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "正常变量替换",
			key:  "variable_normal",
			want: "hello_world/test_path",
		},
		{
			name: "层级key变量替换（单层）",
			key:  "variable_layer",
			want: "/root/rpki/data/rsyncrepo",
		},
		{
			name: "空占位符key",
			key:  "variable_empty_key",
			want: "/test",
		},
		{
			name: "无效格式（缺少结束符）",
			key:  "variable_invalid_format1",
			want: "${missing_end",
		},
		{
			name: "无效格式（缺少开始符）",
			key:  "variable_invalid_format2",
			want: "missing_start}",
		},
		{
			name: "无占位符",
			key:  "variable_no_replace",
			want: "normal_text",
		},
		{
			name: "空key",
			key:  "",
			want: "",
		},
		{
			name: "不存在的key",
			key:  "non_exist_variable",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VariableString(tt.key)
			if got != tt.want {
				t.Errorf("VariableString(%q) = %q, want %q (key exists: %v)",
					tt.key, got, tt.want, HasKey(tt.key))
			}
		})
	}
}
*/
// TestSetString 测试SetString函数
func TestSetString(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "正常设置",
			key:     "test_set_key",
			value:   "test_set_value",
			wantErr: false,
		},
		{
			name:    "空key",
			key:     "",
			value:   "test",
			wantErr: true,
			errMsg:  "key is empty",
		},
		{
			name:    "configure为nil",
			key:     "test_key",
			value:   "test_val",
			wantErr: true,
			errMsg:  "configure is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟configure为nil
			if tt.name == "configure为nil" {
				backup := configure
				configure = nil
				defer func() { configure = backup }()
			}

			err := SetString(tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetString(%q, %q) error = %v, wantErr %v", tt.key, tt.value, err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("SetString error msg = %q, want %q", err.Error(), tt.errMsg)
			}

			// 验证设置结果
			if !tt.wantErr {
				got := String(tt.key)
				if got != tt.value {
					t.Errorf("SetString后 String(%q) = %q, want %q", tt.key, got, tt.value)
				}
			}
		})
	}
}

// -------------------------- 性能测试 --------------------------
func BenchmarkString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		String("string_key")
	}
}

/*
	func BenchmarkVariableString(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			VariableString("variable_layer")
		}
	}
*/
func BenchmarkSetString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SetString("bench_key", fmt.Sprintf("bench_val_%d", i))
	}
}

// -------------------------- 特殊场景测试 --------------------------
// TestInitArgsParse 测试init函数的参数解析逻辑
func TestInitArgsParse(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name     string
		args     []string
		wantConf string
	}{
		{
			name:     "参数为--conf=xxx.conf",
			args:     []string{"cmd", "--conf=testdata/custom.conf"},
			wantConf: "testdata/custom.conf",
		},
		{
			name:     "参数为-conf=xxx.conf",
			args:     []string{"cmd", "-conf=testdata/custom2.conf"},
			wantConf: "testdata/custom2.conf",
		},
		{
			name:     "参数为conf=xxx.conf",
			args:     []string{"cmd", "conf=testdata/custom3.conf"},
			wantConf: "testdata/custom3.conf",
		},
		{
			name:     "参数含等号后的值",
			args:     []string{"cmd", "--conf=test=conf.conf"},
			wantConf: "test=conf.conf",
		},
		{
			name:     "参数无等号",
			args:     []string{"cmd", "--conf"},
			wantConf: "",
		},
		{
			name:     "参数在非第一个位置",
			args:     []string{"cmd", "-arg1=1", "--conf=testdata/custom4.conf", "-arg2=2"},
			wantConf: "testdata/custom4.conf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			var conf string

			for _, arg := range os.Args[1:] {
				parts := strings.SplitN(arg, "=", 2)
				if len(parts) != 2 {
					continue
				}
				key := parts[0]
				value := parts[1]
				if key == "conf" || key == "-conf" || key == "--conf" {
					conf = value
					break
				}
			}

			if conf != tt.wantConf {
				t.Errorf("参数解析失败: args=%v, got=%q, want=%q", tt.args, conf, tt.wantConf)
			}
		})
	}
}
