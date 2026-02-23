package byteutil

import (
	"fmt"
	"testing"
)

/*
# 1. 运行所有单元测试（详细输出）
go test -v ./byteutil

# 2. 运行单元测试并生成覆盖率报告
go test -v -coverprofile=coverage.out ./byteutil
# 查看覆盖率详情
go tool cover -html=coverage.out

# 3. 运行所有基准测试（输出内存分配+耗时）
go test -bench=. ./byteutil -benchmem -benchtime=5s

# 4. 仅运行指定基准测试（如大数据量测试）
go test -bench=LargeData ./byteutil -benchmem
*/

func TestIndexStartAndEnd1(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	subData := []byte{0x06, 0x06}
	startIndex, endIndex, err := IndexStartAndEnd(data, subData)
	fmt.Println(startIndex, endIndex, err)
	if err != nil {
		return
	} else if startIndex < 0 {
		return
	}

	subData2 := data[int(startIndex):int(endIndex)]
	fmt.Println(subData2)
}

// TestIndexStartAndEnd 全覆盖单元测试（包含所有临界值场景）
func TestIndexStartAndEnd(t *testing.T) {
	// 定义测试用例，覆盖：正常匹配、临界值、异常场景
	tests := []struct {
		name       string
		data       []byte
		subData    []byte
		wantStart  int
		wantEnd    int
		wantErrMsg string // 预期错误信息（空表示无错误）
	}{
		// ========== 正常匹配场景 ==========
		{
			name:       "正常匹配-单字节匹配",
			data:       []byte{0xFF},
			subData:    []byte{0xFF},
			wantStart:  0,
			wantEnd:    1,
			wantErrMsg: "",
		},
		{
			name:       "正常匹配-多字节中间位置",
			data:       []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			subData:    []byte{0x02, 0x03},
			wantStart:  1,
			wantEnd:    3,
			wantErrMsg: "",
		},
		{
			name:       "正常匹配-全长度匹配（临界）",
			data:       []byte{0xAA, 0xBB, 0xCC},
			subData:    []byte{0xAA, 0xBB, 0xCC},
			wantStart:  0,
			wantEnd:    3,
			wantErrMsg: "",
		},
		{
			name:       "正常匹配-末尾位置（临界）",
			data:       []byte{0x11, 0x22, 0x33, 0x44},
			subData:    []byte{0x33, 0x44},
			wantStart:  2,
			wantEnd:    4,
			wantErrMsg: "",
		},

		// ========== 临界异常场景 ==========
		{
			name:       "临界-Data长度=SubData长度-1",
			data:       []byte{0x01, 0x02},
			subData:    []byte{0x01, 0x02, 0x03},
			wantStart:  0,
			wantEnd:    0,
			wantErrMsg: "length of data is smaller than subData",
		},
		{
			name:       "临界-Data非空SubData为空",
			data:       []byte{0x01},
			subData:    []byte{},
			wantStart:  0,
			wantEnd:    0,
			wantErrMsg: "data or subData is wrong",
		},
		{
			name:       "临界-Data为空SubData非空",
			data:       []byte{},
			subData:    []byte{0x01},
			wantStart:  0,
			wantEnd:    0,
			wantErrMsg: "data or subData is wrong",
		},
		{
			name:       "临界-Data和SubData都为空",
			data:       []byte{},
			subData:    []byte{},
			wantStart:  0,
			wantEnd:    0,
			wantErrMsg: "data or subData is wrong",
		},

		// ========== 其他异常场景 ==========
		{
			name:       "异常-未找到子串（单字节）",
			data:       []byte{0x01, 0x02, 0x03},
			subData:    []byte{0x04},
			wantStart:  0,
			wantEnd:    0,
			wantErrMsg: "subData not found in data",
		},
		{
			name:       "异常-未找到子串（多字节）",
			data:       []byte{0x01, 0x02, 0x03, 0x04},
			subData:    []byte{0x02, 0x04},
			wantStart:  0,
			wantEnd:    0,
			wantErrMsg: "subData not found in data",
		},
		{
			name:       "异常-Data长度远小于SubData",
			data:       []byte{0x01},
			subData:    []byte{0x01, 0x02, 0x03, 0x04},
			wantStart:  0,
			wantEnd:    0,
			wantErrMsg: "length of data is smaller than subData",
		},
	}

	// 执行每个测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用待测试函数
			start, end, err := IndexStartAndEnd(tt.data, tt.subData)

			// 校验起始索引
			if start != tt.wantStart {
				t.Errorf("startIndex 不匹配: 实际=%d, 预期=%d", start, tt.wantStart)
			}

			// 校验结束索引
			if end != tt.wantEnd {
				t.Errorf("endIndex 不匹配: 实际=%d, 预期=%d", end, tt.wantEnd)
			}

			// 校验错误信息
			if tt.wantErrMsg == "" {
				// 预期无错误
				if err != nil {
					t.Errorf("非预期错误: %v", err)
				}
			} else {
				// 预期有错误
				if err == nil {
					t.Error("预期错误但未返回错误")
				} else if err.Error() != tt.wantErrMsg {
					t.Errorf("错误信息不匹配: 实际=%q, 预期=%q", err.Error(), tt.wantErrMsg)
				}
			}
		})
	}
}

// ========== 高性能基准测试（多维度验证） ==========

// BenchmarkIndexStartAndEnd_SmallData 小数据量（100字节）性能测试
func BenchmarkIndexStartAndEnd_SmallData(b *testing.B) {
	// 构造100字节测试数据
	data := make([]byte, 100)
	for i := 0; i < 100; i++ {
		data[i] = byte(i % 256)
	}
	// 子串取中间位置（模拟常见场景）
	subData := data[45:55]

	// 重置计时器（排除数据构造耗时）
	b.ResetTimer()
	// 执行b.N次，统计性能
	for i := 0; i < b.N; i++ {
		IndexStartAndEnd(data, subData)
	}
}

// BenchmarkIndexStartAndEnd_LargeData 大数据量（10000字节）性能测试
func BenchmarkIndexStartAndEnd_LargeData(b *testing.B) {
	// 构造10000字节测试数据
	data := make([]byte, 10000)
	for i := 0; i < 10000; i++ {
		data[i] = byte(i % 256)
	}
	// 子串取中间位置
	subData := data[4995:5005]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IndexStartAndEnd(data, subData)
	}
}

// BenchmarkIndexStartAndEnd_WorstCase 最坏场景（未找到子串）性能测试
func BenchmarkIndexStartAndEnd_WorstCase(b *testing.B) {
	// 构造1000字节数据
	data := make([]byte, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = byte(i % 256)
	}
	// 构造不存在的子串
	subData := []byte{0xFF, 0xFE, 0xFD}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IndexStartAndEnd(data, subData)
	}
}

// BenchmarkIndexStartAndEnd_SingleByte 单字节匹配（高频场景）性能测试
func BenchmarkIndexStartAndEnd_SingleByte(b *testing.B) {
	// 构造10000字节数据
	data := make([]byte, 10000)
	for i := 0; i < 10000; i++ {
		data[i] = byte(i % 256)
	}
	// 单字节子串（末尾位置）
	subData := []byte{data[9999]}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IndexStartAndEnd(data, subData)
	}
}
