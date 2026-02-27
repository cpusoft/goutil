package regexputil

import (
	"fmt"
	"testing"
)

// ------------------------------ IsHex 测试 ------------------------------
func TestIsHex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    bool
		wantErr bool // 预编译后恒为false，仅做兼容校验
	}{
		// 正常场景
		{"纯小写十六进制", "1a2b3c", true, false},
		{"纯大写十六进制", "1A2B3C", true, false},
		{"单字符十六进制", "f", true, false},
		{"混合大小写十六进制", "aBcD12", true, false},
		// 异常场景
		{"包含非十六进制字符g", "12g34", false, false},
		{"包含符号", "12-34", false, false},
		{"空字符串", "", false, false},
		{"全中文", "十六进制", false, false},
		// 临界值（边界）
		{"单个合法字符0", "0", true, false},
		{"单个非法字符G", "G", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsHex(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsHex(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// IsHex 性能测试（分匹配/不匹配场景）
func BenchmarkIsHex(b *testing.B) {
	// 匹配场景
	b.Run("match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsHex("1a2b3c4d5e6f")
		}
	})
	// 不匹配场景
	b.Run("not_match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsHex("1a2b3c4d5e6g")
		}
	})
}

// ------------------------------ CheckRpkiFileName 测试 ------------------------------
func TestCheckRpkiFileName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// 正常场景
		{"合法名称+cer后缀", "test-123.cer", true},
		{"合法名称+moa后缀（新增）", "abc_def.toa", true},
		{"合法名称+toa后缀（新增）", "xyz_789.moa", true},
		{"最短合法名称", "a.sig", true},
		// 异常场景
		{"非法字符（空格）", "test 123.cer", false},
		{"非法后缀", "test123.txt", false},
		{"后缀大写", "test123.CER", false},
		{"空字符串", "", false},
		{"无后缀", "test123", false},
		// 临界值
		{"仅后缀", ".cer", false}, // 名称部分为空（临界）
		{"特殊字符仅下划线", "_123.roa", true},
		{"特殊字符仅横杠", "-456.crl", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckRpkiFileName(tt.input); got != tt.want {
				t.Errorf("CheckRpkiFileName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// CheckRpkiFileName 性能测试
func BenchmarkCheckRpkiFileName(b *testing.B) {
	b.Run("match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			CheckRpkiFileName("rpki-123456.roa")
		}
	})
	b.Run("not_match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			CheckRpkiFileName("rpki-123456.txt")
		}
	})
}

// ------------------------------ CheckPhone 测试 ------------------------------
func TestCheckPhone(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// 正常场景
		{"11位纯数字", "13800138000", true},
		// 异常场景
		{"10位数字", "1380013800", false},
		{"12位数字", "138001380000", false},
		{"包含非数字", "1380013800a", false},
		{"空字符串", "", false},
		// 临界值
		{"临界10位", "0123456789", false},
		{"临界11位", "01234567890", true},
		{"临界12位", "012345678901", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPhone(tt.input); got != tt.want {
				t.Errorf("CheckPhone(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// CheckPhone 性能测试
func BenchmarkCheckPhone(b *testing.B) {
	b.Run("match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			CheckPhone("13912345678")
		}
	})
	b.Run("not_match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			CheckPhone("139123456789")
		}
	})
}

// ------------------------------ CheckMail 测试 ------------------------------
func TestCheckMail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// 正常场景
		{"基础合法邮箱", "test@163.com", true},
		{"含下划线+横杠", "a_b-c@x.y.z.cn", true},
		{"最短用户名", "a@b.cn", true},
		{"4层域名", "test@a.b.c.d.cn", true},
		// 异常场景
		{"无@符号", "test163.com", false},
		{"@开头", "@163.com", false},
		{"@结尾", "test@", false},
		{"域名含非法字符", "test@163.com!", false},
		{"用户名超长（32位）", "12345678901234567890123456789012@163.com", false},
		// 临界值
		{"用户名临界31位", "1234567890123456789012345678901@163.com", true},
		{"域名超4层", "test@a.b.c.d.e.cn", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckMail(tt.input); got != tt.want {
				t.Errorf("CheckMail(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// CheckMail 性能测试
func BenchmarkCheckMail(b *testing.B) {
	b.Run("match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			CheckMail("user_123@domain.com.cn")
		}
	})
	b.Run("not_match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			CheckMail("user_123@domain.com.cn!")
		}
	})
}

// ------------------------------ CheckPassword 测试 ------------------------------
func TestCheckPassword(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// 正常场景
		{"6位（数字+字母+特殊字符）", "A1@bcd", true},
		{"20位（全类型）", "Abc123!@#45678901234", true}, // 修正：20位（原21位）
		{"中间长度", "123Abc!@#", true},
		// 异常场景
		{"长度5", "A1@bc", false},
		{"长度21", "Abc123!@#4567890123456", false},
		{"缺少数字", "Abcdef!@#", false},
		{"缺少字母", "123456!@#", false},
		{"缺少特殊字符", "Abc123456", false},
		{"非法特殊字符", "Abc123￥", false}, // ￥不在指定集合
		// 临界值
		{"临界长度5", "a1@bc", false},
		{"临界长度6", "a1@bcd", true},
		{"临界长度20", "a1@bcdefghijklmnopqr", true}, // 修正：20位（原21位）
		{"临界长度21", "a1@bcdefghijklmnopqrst", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPassword(tt.input); got != tt.want {
				t.Errorf("CheckPassword(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// CheckPassword 性能测试
func BenchmarkCheckPassword(b *testing.B) {
	b.Run("match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			CheckPassword("Pass123!@#456")
		}
	})
	b.Run("not_match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			CheckPassword("Pass123456") // 缺少特殊字符
		}
	})
}

// ------------------------------ CheckCompany 测试 ------------------------------
// ------------------------------ CheckCompany 测试 ------------------------------
func TestCheckCompany(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// 正常场景
		{"2位中文", "测试", true},
		// 修正：严格32位符文的混合字符串（逐个数：2+2+1+3+3+1+4+8+8=32）
		{"32位混合", "测试公司_ABC123 有限公司1234567812345678", true},
		{"含空格+下划线", "XX_科技 有限公司", true},
		// 异常场景
		{"1位字符", "测", false},
		{"33位字符", "测试公司_ABC123 有限公司12345678123456789", false},
		{"含非法字符!", "测试!公司", false},
		{"含非法字符@", "测试@公司", false},
		// 临界值
		{"临界1位", "A", false},
		{"临界2位", "AB", true},
		{"临界32位", "12345678901234567890123456789012", true},
		{"临界33位", "123456789012345678901234567890123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckCompany(tt.input); got != tt.want {
				t.Errorf("CheckCompany(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// CheckCompany 性能测试
func BenchmarkCheckCompany(b *testing.B) {
	b.Run("match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			CheckCompany("北京测试科技有限公司_ABC123")
		}
	})
	b.Run("not_match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			CheckCompany("北京测试科技有限公司!ABC123") // 含非法字符
		}
	})
}

// /////////////////////////////////////////////////////////////
// ////////////////////////////////////////////////////////
func TestIsHex1(t *testing.T) {
	ssss := `10d0c9f4328576d51cc73c042cfc15e9b3d6378`
	b, err := IsHex(ssss)
	fmt.Println(b, err)
}

func TestCheckRpkiFileName1(t *testing.T) {
	ssss := `ZoN_KCuLgQ_XLZREqsT884kSssE.aaa`
	b := CheckRpkiFileName(ssss)
	fmt.Println(b)
}
