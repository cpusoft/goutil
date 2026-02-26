package netutil

import (
	"fmt"
	"net"
	"testing"
)

// -------------------------- 单元测试 - ResolveIp --------------------------
// TestResolveIp 覆盖ResolveIp的所有分支：正常IP/域名、空字符串、无效host、全空格host
func TestResolveIp(t *testing.T) {
	tests := []struct {
		name    string // 测试用例名称
		host    string // 输入host
		wantNil bool   // 期望返回nil（true=返回nil，false=非nil）
		checkIp bool   // 是否校验IP合法性（仅wantNil=false时生效）
	}{
		{
			name:    "正常用例-IPv4",
			host:    "127.0.0.1",
			wantNil: false,
			checkIp: true,
		},
		{
			name:    "正常用例-IPv6",
			host:    "::1",
			wantNil: false,
			checkIp: true,
		},
		{
			name:    "正常用例-本地域名",
			host:    "localhost",
			wantNil: false,
			checkIp: true,
		},
		{
			name:    "临界值-空字符串",
			host:    "",
			wantNil: true,
			checkIp: false,
		},
		{
			name:    "临界值-全空格字符串",
			host:    "   ",
			wantNil: true,
			checkIp: false,
		},
		{
			name:    "异常用例-无效host",
			host:    "invalid.host.12345678",
			wantNil: true,
			checkIp: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveIp(tt.host)

			if (got == nil) != tt.wantNil {
				t.Errorf("ResolveIp() = %v, wantNil %v", got, tt.wantNil)
				return
			}

			if !tt.wantNil && tt.checkIp {
				if got.IP == nil || (len(got.IP) != net.IPv4len && len(got.IP) != net.IPv6len) {
					t.Errorf("ResolveIp() return invalid IP: %v", got)
				}
			}
		})
	}
}

// -------------------------- 单元测试 - LookupIp --------------------------
// TestLookupIp 覆盖LookupIp的所有分支：正常IP/域名、空字符串、无效host、全空格host
func TestLookupIp(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantNil bool // true=返回nil，false=返回非空切片
	}{
		{
			name:    "正常用例-IPv4",
			host:    "127.0.0.1",
			wantNil: false,
		},
		{
			name:    "正常用例-本地域名",
			host:    "localhost",
			wantNil: false,
		},
		{
			name:    "临界值-空字符串",
			host:    "",
			wantNil: true,
		},
		{
			name:    "临界值-全空格字符串",
			host:    "   ",
			wantNil: true,
		},
		{
			name:    "异常用例-无效host",
			host:    "invalid.host.12345678",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LookupIp(tt.host)

			if (got == nil) != tt.wantNil {
				t.Errorf("LookupIp() = %v, wantNil %v", got, tt.wantNil)
				return
			}

			if !tt.wantNil {
				if len(got) == 0 {
					t.Error("LookupIp() return empty IP slice")
					return
				}
				for _, ip := range got {
					if len(ip) != net.IPv4len && len(ip) != net.IPv6len {
						t.Errorf("LookupIp() return invalid IP: %v", ip)
					}
				}
			}
		})
	}
}

// -------------------------- 单元测试 - LookupIpByUrl --------------------------
// TestLookupIpByUrl 覆盖LookupIpByUrl的所有分支：正常URL、带端口URL、空URL、无效URL、无host URL
func TestLookupIpByUrl(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantNil bool
	}{
		{
			name:    "正常用例-HTTP URL（本地IP）",
			url:     "http://127.0.0.1",
			wantNil: false,
		},
		{
			name:    "正常用例-HTTPS URL（本地IP）",
			url:     "https://127.0.0.1",
			wantNil: false,
		},
		{
			name:    "临界值-带端口的URL（localhost）",
			url:     "http://localhost:8080",
			wantNil: false,
		},
		{
			name:    "临界值-带端口的URL（本地IP）",
			url:     "https://127.0.0.1:443",
			wantNil: false,
		},
		{
			name:    "临界值-空URL",
			url:     "",
			wantNil: true,
		},
		{
			name:    "临界值-全空格URL",
			url:     "   ",
			wantNil: true,
		},
		{
			name:    "异常用例-无效URL",
			url:     "invalid://url",
			wantNil: true,
		},
		{
			name:    "异常用例-无host的URL",
			url:     "http://",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LookupIpByUrl(tt.url)

			// 修复点1：区分nil和空切片（核心！）
			// 原代码错误：将空切片([])和nil等同，实际代码返回的是nil而非空切片
			gotIsEmpty := got == nil || len(got) == 0
			if gotIsEmpty != tt.wantNil {
				t.Errorf("LookupIpByUrl(%q) = %v, wantNil %v", tt.url, got, tt.wantNil)
				return
			}

			if !tt.wantNil {
				if len(got) == 0 {
					t.Error("LookupIpByUrl() return empty IP slice")
					return
				}
				for _, ip := range got {
					if len(ip) != net.IPv4len && len(ip) != net.IPv6len {
						t.Errorf("LookupIpByUrl() return invalid IP: %v", ip)
					}
				}
			}
		})
	}
}

// -------------------------- 性能测试 - Benchmark --------------------------
// BenchmarkResolveIp 测试ResolveIp的性能（本地IP避免网络波动）
func BenchmarkResolveIp(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ResolveIp("127.0.0.1")
	}
}

// BenchmarkLookupIp 测试LookupIp的性能
func BenchmarkLookupIp(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LookupIp("127.0.0.1")
	}
}

// BenchmarkLookupIpByUrl 测试LookupIpByUrl的性能
func BenchmarkLookupIpByUrl(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LookupIpByUrl("http://127.0.0.1:8080")
	}
}

// -------------------------- 额外：覆盖率辅助测试（可选） --------------------------
// TestAllCoverage 确保所有分支被执行（可省略，仅用于强制覆盖）
func TestAllCoverage(t *testing.T) {
	// 调用所有异常分支
	ResolveIp("")
	ResolveIp("   ")
	ResolveIp("invalid.host.12345")

	LookupIp("")
	LookupIp("   ")
	LookupIp("invalid.host.12345")

	LookupIpByUrl("")
	LookupIpByUrl("   ")
	LookupIpByUrl("invalid://url")
	LookupIpByUrl("http://")

	// 调用正常分支
	ResolveIp("::1")
	LookupIp("localhost")
	LookupIpByUrl("https://127.0.0.1:443")
}

///////////////////////////////////////////////
////////////////////////////////////////////////////////////

func TestResolveIP(t *testing.T) {
	host := "baidu.com"
	fmt.Println("resolveIp:", ResolveIp(host))
}

func TestLookupIp1(t *testing.T) {
	host := "baidu.com"
	fmt.Println("resolveIp:", LookupIp(host))
}
func TestLookupIpByUrl1(t *testing.T) {
	url := "https://rrdp.apnic.net/rrdp/26afc8dd-3e3b-4b31-8ff1-710d8e944320/22733/delta.xml"
	fmt.Println("resolveIpByUrl:", LookupIpByUrl(url))
}
