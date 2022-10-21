package transportutil

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestTransportMsg(t *testing.T) {
	sendData := []byte{0x01, 0x02, 0x03}
	transportMsg := &TransportMsg{
		MsgType:  1,
		SendData: sendData,
	}
	fmt.Println("sendMessageModel(): transportMsg, will send transportMsg:",
		jsonutil.MarshalJson(*transportMsg))

}
