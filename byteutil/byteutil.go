package byteutil

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/cpusoft/goutil/belogs"
)

func IndexStartAndEnd(data, subData []byte) (startIndex, endIndex int, err error) {
	if len(data) == 0 || len(subData) == 0 || len(data) < len(subData) {
		belogs.Debug("IndexStartAndEnd(): data or subData is wrong, len(data):", len(data),
			"    len(subData):", len(subData))
		return 0, 0, errors.New("data or subData is wrong")
	}

	dataHex := hex.EncodeToString(data)
	subDataHex := hex.EncodeToString(subData)
	belogs.Debug("IndexStartAndEnd(): len(dataHex):", len(dataHex), "  len(subDataHex):", len(subDataHex))
	index := strings.Index(dataHex, subDataHex)
	if index < 0 {
		belogs.Debug("IndexStartAndEnd(): not found subDataHex in dataHex, dataHex:", dataHex,
			"  subDataHex:", subDataHex, " index:", index)
		return -1, -1, nil
	} else if index%2 != 0 {
		belogs.Debug("IndexStartAndEnd(): hex encoding is wrong, dataHex:", dataHex,
			"  subDataHex:", subDataHex, "   index:", index)
		return -1, -1, nil
	}
	startIndex = index / 2
	endIndex = startIndex + len(subData)
	belogs.Debug("IndexStartAndEnd(): startIndex:", startIndex, " endIndex:", endIndex)
	return startIndex, endIndex, nil
}
