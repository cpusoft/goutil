package byteutil

import (
	"fmt"
	"testing"
)

func TestIndexStartAndEnd(t *testing.T) {
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
