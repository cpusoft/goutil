package datetime

import (
	"fmt"
	"testing"
	"time"

	timeconv "github.com/Andrew-M-C/go.timeconv"
)

/*
go test -v ./datetime -bench=. -benchmem
# 生成覆盖率文件
go test ./datetime -coverprofile=coverage.out

# 查看覆盖率报告
go tool cover -func=coverage.out

# 生成HTML可视化报告（更直观）
go tool cover -html=coverage.out -o coverage.html

*/
// ========== 功能测试 ==========

// TestNow 测试Now函数
func TestNow(t *testing.T) {
	// 验证返回格式是否正确
	nowStr := Now()
	if len(nowStr) != len("2006-01-02 15:04:05") {
		t.Errorf("Now()返回长度错误，期望19，实际%d", len(nowStr))
	}
	// 验证格式是否符合YYYY-MM-DD HH:MM:SS
	_, err := time.Parse(TimeLayoutDefault, nowStr)
	if err != nil {
		t.Errorf("Now()返回格式错误: %v, 结果: %s", err, nowStr)
	}
}

// TestToString 测试ToString函数
func TestToString(t *testing.T) {
	tests := []struct {
		name  string
		input time.Time
		want  string
	}{
		{
			name:  "正常时间",
			input: time.Date(2024, 10, 1, 12, 34, 56, 0, time.UTC),
			want:  "2024-10-01 12:34:56",
		},
		{
			name:  "零值时间",
			input: time.Time{},
			want:  "",
		},
		{
			name:  "带时区时间",
			input: time.Date(2024, 2, 29, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
			want:  "2024-02-29 08:00:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToString(tt.input); got != tt.want {
				t.Errorf("ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseTime 测试ParseTime函数
func TestParseTime(t *testing.T) {
	validTime := "2024-01-01 00:00:00"
	validLayout := TimeLayoutDefault
	expectedTime, _ := time.Parse(validLayout, validTime)

	tests := []struct {
		name        string
		inputStr    string
		inputLayout string
		wantErr     bool
		wantTime    time.Time
	}{
		{
			name:        "正常解析",
			inputStr:    validTime,
			inputLayout: validLayout,
			wantErr:     false,
			wantTime:    expectedTime,
		},
		{
			name:        "空字符串输入",
			inputStr:    "",
			inputLayout: validLayout,
			wantErr:     true,
			wantTime:    time.Time{},
		},
		{
			name:        "空格式",
			inputStr:    validTime,
			inputLayout: "",
			wantErr:     true,
			wantTime:    time.Time{},
		},
		{
			name:        "格式不匹配",
			inputStr:    "2024/01/01 00:00:00",
			inputLayout: validLayout,
			wantErr:     true,
			wantTime:    time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTime(tt.inputStr, tt.inputLayout)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.wantTime) {
				t.Errorf("ParseTime() = %v, want %v", got, tt.wantTime)
			}
		})
	}
}

// TestAddDateByDuration 测试AddDateByDuration函数
func TestAddDateByDuration(t *testing.T) {
	baseTime := time.Date(2024, 2, 29, 10, 0, 0, 0, time.UTC) // 闰年2月29日（边界场景）

	tests := []struct {
		name      string
		inputTime time.Time
		duration  string
		wantErr   bool
		wantTime  time.Time // 期望结果
	}{
		{
			name:      "加1年",
			inputTime: baseTime,
			duration:  "1y",
			wantErr:   false,
			wantTime:  timeconv.AddDate(baseTime, 1, 0, 0), // 2025-02-28
		},
		{
			name:      "减2月",
			inputTime: baseTime,
			duration:  "-2m",
			wantErr:   false,
			wantTime:  timeconv.AddDate(baseTime, 0, -2, 0), // 2023-12-29
		},
		{
			name:      "加3天",
			inputTime: baseTime,
			duration:  "3d",
			wantErr:   false,
			wantTime:  timeconv.AddDate(baseTime, 0, 0, 3), // 2024-03-03
		},
		{
			name:      "空duration",
			inputTime: baseTime,
			duration:  "",
			wantErr:   true,
			wantTime:  baseTime, // 错误时返回原时间
		},
		{
			name:      "零值时间输入",
			inputTime: time.Time{},
			duration:  "1d",
			wantErr:   true,
			wantTime:  time.Time{},
		},
		{
			name:      "非法单位",
			inputTime: baseTime,
			duration:  "1h", // 仅支持y/m/d
			wantErr:   true,
			wantTime:  baseTime,
		},
		{
			name:      "无单位",
			inputTime: baseTime,
			duration:  "123",
			wantErr:   true,
			wantTime:  baseTime,
		},
		{
			name:      "非数字前缀",
			inputTime: baseTime,
			duration:  "abc1y",
			wantErr:   true,
			wantTime:  baseTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddDateByDuration(tt.inputTime, tt.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddDateByDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.wantTime) {
				t.Errorf("AddDateByDuration() = %v, want %v", got, tt.wantTime)
			}
			if tt.wantErr && !got.Equal(tt.wantTime) {
				t.Errorf("AddDateByDuration() 错误时返回值错误 = %v, want %v", got, tt.wantTime)
			}
		})
	}
}

// ========== 性能基准测试 ==========

// BenchmarkNow 基准测试Now函数
func BenchmarkNow(b *testing.B) {
	// 重置计时器（忽略初始化耗时）
	b.ResetTimer()
	// 执行b.N次，评估性能
	for i := 0; i < b.N; i++ {
		Now()
	}
}

// BenchmarkToString 基准测试ToString函数
func BenchmarkToString(b *testing.B) {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToString(testTime)
	}
}

// BenchmarkParseTime 基准测试ParseTime函数
func BenchmarkParseTime(b *testing.B) {
	testStr := "2024-01-01 00:00:00"
	testLayout := TimeLayoutDefault
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseTime(testStr, testLayout)
	}
}

// BenchmarkAddDateByDuration 基准测试AddDateByDuration函数
func BenchmarkAddDateByDuration(b *testing.B) {
	testTime := time.Date(2024, 2, 29, 10, 0, 0, 0, time.UTC)
	testDuration := "1y"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AddDateByDuration(testTime, testDuration)
	}
}
func TestXYZ(t *testing.T) {
	TIME_LAYOUT := "060102150405Z"
	tm, e := ParseTime("190601095044Z", TIME_LAYOUT)
	fmt.Println(tm, e)

}

func TestAddDataByDuration(t *testing.T) {
	newT, err := AddDateByDuration(time.Now(), "12m")
	fmt.Println(newT, err)

	newT, err = AddDateByDuration(time.Now(), "30d")
	fmt.Println(newT, err)

	newT, err = AddDateByDuration(time.Now(), "1y")
	fmt.Println(newT, err)
}
