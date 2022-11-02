package transportutil

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestBusinessToConnMsg(t *testing.T) {
	sendData := []byte{0x01, 0x02, 0x03}
	businessToConnMsg := &BusinessToConnMsg{
		MsgType:  1,
		SendData: sendData,
	}
	fmt.Println("TestBusinessToConnMsg(): businessToConnMsg, will send businessToConnMsg:",
		jsonutil.MarshalJson(*businessToConnMsg))

}
