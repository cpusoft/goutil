package tcptlsutil

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	_ "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	_ "github.com/cpusoft/goutil/logs"
)

func TestCreateTcpServer(t *testing.T) {
	serverProcessFunc := new(ServerProcessFunc)
	tcpTlsMsg := make(chan TcpTlsMsg, 16)
	ts := NewTcpServer(serverProcessFunc, tcpTlsMsg)
	fmt.Println("CreateTcpServer():", 9999)
	err := ts.StartTcpServer("9999")
	if err != nil {
		fmt.Println("CreateTcpServer(): StartTcpServer ts fail: ", &ts, err)
		return
	}
	time.Sleep(2 * time.Second)
	ts.ActiveSend("", GetData())

	time.Sleep(5 * time.Second)
	ts.CloseGraceful()
}
func TestCreateTlsServer(t *testing.T) {
	serverProcessFunc := new(ServerProcessFunc)
	tlsRootCrtFileName := `catlsroot.cer`       //`ca.cer`
	tlsPublicCrtFileName := `servertlscrt.cer`  //`server.cer`
	tlsPrivateKeyFileName := `servertlskey.pem` //`serverkey.pem`
	fmt.Println("CreateTlsServer(): tlsRootCrtFileName:", tlsRootCrtFileName,
		"tlsPublicCrtFileName:", tlsPublicCrtFileName,
		"tlsPrivateKeyFileName:", tlsPrivateKeyFileName)
	tcpTlsMsg := make(chan TcpTlsMsg, 16)
	ts, err := NewTlsServer(tlsRootCrtFileName, tlsPublicCrtFileName,
		tlsPrivateKeyFileName, true, serverProcessFunc, tcpTlsMsg)
	if err != nil {
		fmt.Println("CreateTlsServer(): NewTlsServer ts fail: ", &ts, err)
		return
	}
	go ts.StartTlsServer("9999")

	time.Sleep(5 * time.Second)
	ts.ActiveSend("", GetData())
	time.Sleep(8 * time.Second)
	ts.CloseGraceful()
}
func RtrProcess(receiveData []byte) (sendData []byte, err error) {
	buf := bytes.NewReader(receiveData)
	fmt.Println("RtrProcess(): buf:", buf)
	return nil, nil
}
func GetData() (buffer []byte) {

	return []byte{0x00, 0x0a, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x0a, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
}

type ServerProcessFunc struct {
}

func (spf *ServerProcessFunc) OnConnectProcess(tcpTlsConn *TcpTlsConn) {

}
func (spf *ServerProcessFunc) ReceiveAndSendProcess(tcpTlsConn *TcpTlsConn, receiveData []byte) (nextConnectPolicy int, leftData []byte, err error) {
	fmt.Println("ReceiveAndSendProcess(): len(receiveData):", len(receiveData), "   receiveData:", convert.Bytes2String(receiveData))
	// need recombine
	packets, leftData, err := RecombineReceiveData(receiveData, PDU_TYPE_MIN_LEN, PDU_TYPE_LENGTH_START, PDU_TYPE_LENGTH_END)
	if err != nil {
		fmt.Println("ReceiveAndSendProcess(): RecombineReceiveData fail:", err)
		return NEXT_CONNECT_POLICY_CLOSE_FORCIBLE, nil, err
	}
	fmt.Println("ReceiveAndSendProcess(): RecombineReceiveData packets.Len():", packets.Len())

	if packets == nil || packets.Len() == 0 {
		fmt.Println("ReceiveAndSendProcess(): RecombineReceiveData packets is empty:  len(leftData):", len(leftData))
		return NEXT_CONNECT_POLICY_CLOSE_GRACEFUL, leftData, nil
	}
	for e := packets.Front(); e != nil; e = e.Next() {
		packet, ok := e.Value.([]byte)
		if !ok || packet == nil || len(packet) == 0 {
			fmt.Println("ReceiveAndSendProcess(): for packets fail:", convert.ToString(e.Value))
			break
		}
		_, err := RtrProcess(packet)
		if err != nil {
			fmt.Println("ReceiveAndSendProcess(): RtrProcess fail:", err)
			return NEXT_CONNECT_POLICY_CLOSE_FORCIBLE, nil, err
		}

	}

	_, err = tcpTlsConn.Write(GetData())
	if err != nil {
		fmt.Println("ReceiveAndSendProcess(): tcp  Write fail:  tcpTlsConn:", tcpTlsConn.RemoteAddr().String(), err)
		return NEXT_CONNECT_POLICY_CLOSE_FORCIBLE, nil, err
	}
	// continue to receive next receiveData
	return NEXT_CONNECT_POLICY_KEEP, leftData, nil
}
func (spf *ServerProcessFunc) OnCloseProcess(tcpTlsConn *TcpTlsConn) {

}
func (spf *ServerProcessFunc) ActiveSendProcess(tcpTlsConn *TcpTlsConn, sendData []byte) (err error) {
	return
}

const (
	PDU_TYPE_MIN_LEN      = 8
	PDU_TYPE_LENGTH_START = 4
	PDU_TYPE_LENGTH_END   = 8
)
