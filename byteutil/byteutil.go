package byteutil

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/cpusoft/goutil/belogs"
)

func IndexStartAndEnd(data, subData []byte) (startIndex, endIndex int, err error) {
	/*
		dataLen := len(data)
		subDataLen := len(subData)
		if subDataLen == 0 {
			return 0
		}
		if dataLen < subDataLen {
			return -1
		}
		for i := 0; i <= dataLen-subDataLen; i++ {
			j := 0
			for ; j < subDataLen; j++ {
				if data[i+j] == subData[j] {
					startIndex = i
				}
				if data[i+j] != subData[j] && startIndex > 0 {
					endIndex = i
					break
				}
			}
			if j == subDataLen {
				return i
			}
		}
	*/
	if len(data) == 0 || len(subData) == 0 || len(data) < len(subData) {
		belogs.Debug("IndexStartAndEnd(): data or subData is wrong, len(data):", len(data),
			"    len(subData):", len(subData))
		return 0, 0, errors.New("data or subData is wrong")
	}

	dataHex := hex.EncodeToString(data)
	subDataHex := hex.EncodeToString(subData)
	belogs.Debug("IndexStartAndEnd(): dataHex:", dataHex, "  subDataHex:", subDataHex)
	index := strings.Index(dataHex, subDataHex)
	if index < 0 {
		belogs.Debug("IndexStartAndEnd(): not found subDataHex in dataHex, dataHex:", dataHex,
			"  subDataHex:", subDataHex, " index:", index)
		return -1, -1, nil
	} else if index%2 != 0 {
		belogs.Debug("IndexStartAndEnd(): hex encoding is wrong, index:", index)
		return -1, -1, nil
	}
	startIndex = index / 2
	endIndex = startIndex + len(subData)
	belogs.Debug("IndexStartAndEnd(): startIndex:", subDataHex, " endIndex:", endIndex)
	return startIndex, endIndex, nil
}
