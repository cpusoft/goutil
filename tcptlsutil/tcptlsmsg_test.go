package tcptlsutil

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestTcptlsMsg(t *testing.T) {
	sendData := []byte{0x01, 0x02, 0x03}
	tcpTlsMsg := &TcpTlsMsg{
		MsgType:                1,
		NextConnectClosePolicy: 2,
		NextRwPolicy:           3,
		SendData:               sendData,
	}
	fmt.Println("sendMessageModel(): tcpTlsMsg, will send tcpTlsMsg:",
		jsonutil.MarshalJson(*tcpTlsMsg))

}
