package tcpclient

import (
	"net"
	"testing"
	"time"

	belogs "github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/tcpserverclient/util"
)

var TcpTestClient *TcpClient

func TestTcpServer(t *testing.T) {
	CreateTcpClient()
	select {}
}

func CreateTcpClient() {
	clientProcessFunc := new(ClientProcessFunc)

	//CreateTcpClient("127.0.0.1:9999", ClientProcess1)
	TcpTestClient = NewTcpClient("stop", clientProcessFunc)
	err := TcpTestClient.Start("192.168.83.139:9999")
	belogs.Debug("CreateTcpClient(): tcpclient: ", TcpTestClient, err)
	if err != nil {
		return
	}
	belogs.Debug("CreateTcpClient(): tcpclient will SendData")
	tcpClientMsg := &TcpClientMsg{NextConnectClosePolicy: util.NEXT_CONNECT_POLICE_KEEP,
		NextRwPolice: util.NEXT_RW_POLICE_WAIT_READ,
		SendData:     GetClientData(),
	}
	TcpTestClient.SendMsg(tcpClientMsg)
	time.Sleep(60 * time.Second)

	belogs.Debug("CreateTcpClient(): tcpclient will stop")
	tcpClientMsg.NextConnectClosePolicy = util.NEXT_CONNECT_POLICE_CLOSE_GRACEFUL
	tcpClientMsg.SendData = nil
	TcpTestClient.SendMsg(tcpClientMsg)

}

type ClientProcessFunc struct {
}

func (cp *ClientProcessFunc) OnConnectProcess(tcpConn *net.TCPConn) {

	belogs.Info("OnConnectProcess(): tcpclient tcpConn:", tcpConn.RemoteAddr().String())

}
func (cp *ClientProcessFunc) OnCloseProcess(tcpConn *net.TCPConn) {
	if tcpConn != nil {
		belogs.Info("OnCloseProcess(): tcpclient tcpConn:", tcpConn.RemoteAddr().String())
	}
}

func (sq *ClientProcessFunc) OnReceiveProcess(tcpConn *net.TCPConn, receiveData []byte) (nextRwPolicy int, leftData []byte, err error) {

	belogs.Debug("OnReceiveProcess() tcpclient  :", tcpConn, convert.Bytes2String(receiveData))
	return util.NEXT_RW_POLICE_END_READ, make([]byte, 0), nil
}

func GetClientData() (buffer []byte) {

	return []byte{0x00, 0x0a, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x0a, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
}
