package byteutil

import (
	"errors"

	"github.com/cpusoft/goutil/belogs"
)

// IndexStartAndEnd 在data字节切片中查找subData的起始和结束索引
// 新返回值规则：
// 1. 参数为空（data/subData长度0）：返回 0, 0, error("data or subData is wrong")
// 2. data长度 < subData长度：返回 0, 0, error("length of data is smaller than subData")
// 3. 未找到subData：返回 0, 0, error("subData not found in data")
// 4. 找到subData：返回 起始索引, 结束索引(不包含), nil
func IndexStartAndEnd(data, subData []byte) (startIndex, endIndex int, err error) {
	// 基础参数校验：空参数返回错误
	if len(data) == 0 || len(subData) == 0 {
		belogs.Error("IndexStartAndEnd(): data or subData is empty, len(data):", len(data),
			"    len(subData):", len(subData))
		return 0, 0, errors.New("data or subData is wrong")
	}

	// data长度不足：返回错误
	if len(data) < len(subData) {
		belogs.Error("IndexStartAndEnd(): data length less than subData, len(data):", len(data),
			"    len(subData):", len(subData))
		return 0, 0, errors.New("length of data is smaller than subData")
	}

	// 原生字节匹配（无内存分配，高性能）
	subLen := len(subData)
	for i := 0; i <= len(data)-subLen; i++ {
		match := true
		for j := 0; j < subLen; j++ {
			if data[i+j] != subData[j] {
				match = false
				break
			}
		}
		if match {
			startIndex = i
			endIndex = i + subLen
			belogs.Debug("IndexStartAndEnd(): found subData, startIndex:", startIndex, " endIndex:", endIndex)
			return startIndex, endIndex, nil
		}
	}

	// 未找到子串：返回错误
	belogs.Error("IndexStartAndEnd(): not found subData in data, len(data):", len(data),
		"  len(subData):", len(subData))
	return 0, 0, errors.New("subData not found in data")
}
