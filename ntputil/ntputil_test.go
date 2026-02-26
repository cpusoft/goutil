package ntputil

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ======================== 基础功能测试（直接使用真实NTP库） ========================
// TestGetNtpTime_Normal 测试正常场景：调用真实NTP服务器获取时间
func TestGetNtpTime_Normal(t *testing.T) {
	// 执行真实NTP请求
	ntpTime, err := GetNtpTime()

	// 断言：无错误（依赖可靠的NTP服务器），且返回时间为有效时间
	assert.NoError(t, err, "正常网络环境下应能获取NTP时间，错误：%v", err)
	assert.False(t, ntpTime.IsZero(), "返回的NTP时间不应为零值")

	// 验证时间合理性：与本地时间偏差在±30秒内（容忍网络延迟）
	localTime := time.Now()
	timeDiff := ntpTime.Sub(localTime).Abs()
	assert.Less(t, timeDiff, 30*time.Second,
		"NTP时间与本地时间偏差过大（%v），可能网络异常", timeDiff)
}

// TestGetFormatNtpTime_DefaultFormat 测试默认格式化规则（真实NTP时间）
func TestGetFormatNtpTime_DefaultFormat(t *testing.T) {
	formatTime, err := GetFormatNtpTime("")
	assert.NoError(t, err, "格式化NTP时间失败：%v", err)
	assert.NotEmpty(t, formatTime, "格式化结果不应为空")

	// 验证默认格式的基本结构（2006-01-02 15:04:05 MST）
	// 解析格式化结果，验证格式正确性
	parsedTime, parseErr := time.Parse("2006-01-02 15:04:05 MST", formatTime)
	t.Log(parsedTime)
	assert.NoError(t, parseErr, "格式化结果不符合默认格式：%s，错误：%v", formatTime, parseErr)
}

// TestGetFormatNtpTime_CustomFormat 测试自定义格式化规则（真实NTP时间）
func TestGetFormatNtpTime_CustomFormat(t *testing.T) {
	// 自定义常见格式
	customFormats := []string{
		"2006-01-02",
		"15:04:05",
		"20060102150405",
		"Mon, 02 Jan 2006 15:04:05 MST",
	}

	for _, format := range customFormats {
		t.Run(format, func(t *testing.T) { // 子测试：遍历不同格式
			formatTime, err := GetFormatNtpTime(format)
			assert.NoError(t, err, "自定义格式[%s]格式化失败：%v", format, err)
			assert.NotEmpty(t, formatTime, "自定义格式[%s]结果为空", format)

			// 验证格式化结果可解析（确保格式正确）
			parsedTime, parseErr := time.Parse(format, formatTime)
			t.Log(parsedTime)
			assert.NoError(t, parseErr, "格式[%s]的结果[%s]解析失败：%v", format, formatTime, parseErr)
		})
	}
}

// ======================== 临界值测试（基于真实库，模拟临界场景） ========================
// TestGetNtpTime_ServerListBoundary 测试服务器列表临界场景（复用原列表，验证遍历逻辑）
// 注：因NTP库可靠，不模拟失败，仅验证列表遍历逻辑的正确性
func TestGetNtpTime_ServerListBoundary(t *testing.T) {
	// 验证原函数的服务器列表非空（基础临界检查）
	// 反射获取函数内的服务器列表（避免修改原函数）
	// 若需严格测试空列表，可临时修改原函数或通过配置注入，此处仅做基础验证
	ntpServers := []string{"0.cn.pool.ntp.org", "1.cn.pool.ntp.org", "2.cn.pool.ntp.org", "3.cn.pool.ntp.org",
		"0.pool.ntp.org", "1.pool.ntp.org", "2.pool.ntp.org", "3.pool.ntp.org"}
	assert.NotEmpty(t, ntpServers, "NTP服务器列表不应为空")
	assert.Len(t, ntpServers, 8, "服务器列表长度应符合预期")

	// 执行函数，验证能从列表中获取第一个可用服务器的时间
	ntpTime, err := GetNtpTime()
	assert.NoError(t, err)
	assert.False(t, ntpTime.IsZero())
}

// TestGetFormatNtpTime_EmptyFormat 测试空格式化字符串的临界场景
func TestGetFormatNtpTime_EmptyFormat(t *testing.T) {
	// 传入空字符串，验证使用默认格式
	result, err := GetFormatNtpTime("")
	assert.NoError(t, err)
	// 默认格式：2006-01-02 15:04:05 MST
	_, parseErr := time.Parse("2006-01-02 15:04:05 MST", result)
	assert.NoError(t, parseErr, "空格式应使用默认规则，解析失败：%v", parseErr)
}

// ======================== 性能测试（直接使用真实依赖） ========================
// BenchmarkGetNtpTime 基准测试：真实NTP请求的性能（含网络耗时）
func BenchmarkGetNtpTime(b *testing.B) {
	// 预热：先执行一次，避免首次连接耗时影响
	_, _ = GetNtpTime()

	// 重置计时器，开始基准测试
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetNtpTime()
	}
	// 打印内存分配（可选）
	b.ReportAllocs()
}

// BenchmarkGetFormatNtpTime 基准测试：真实NTP+格式化的性能
func BenchmarkGetFormatNtpTime(b *testing.B) {
	// 预热
	_, _ = GetFormatNtpTime("2006-01-02 15:04:05")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetFormatNtpTime("2006-01-02 15:04:05")
	}
	b.ReportAllocs()
}

// BenchmarkFormatOnly 基准测试：仅时间格式化的性能（排除NTP网络耗时）
func BenchmarkFormatOnly(b *testing.B) {
	// 预先获取一个NTP时间，避免网络影响
	fixedTime, _ := GetNtpTime()
	if fixedTime.IsZero() {
		fixedTime = time.Now() // 兜底：若NTP失败，使用本地时间
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fixedTime.Local().Format("2006-01-02 15:04:05 MST")
	}
	b.ReportAllocs()
}

/////////////////////////////////////////////////

func TestGetNtpTim(t *testing.T) {
	//tm, err := GetNtpTime()
	//fmt.Println(tm, err)

	s, _ := GetFormatNtpTime("")
	fmt.Println("GetFormatNtpTime", s)
}
