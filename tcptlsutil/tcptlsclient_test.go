package tcptlsutil

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	_ "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	_ "github.com/cpusoft/goutil/logs"
	"github.com/gogo/protobuf/test"
)

func CreateTcpClient(t *test.T) {
	clientProcessFunc := new(ClientProcessFunc)
	belogs.Debug("CreateTcpClient():", "192.168.83.139:9999")
	//CreateTcpClient("127.0.0.1:9999", ClientProcess1)
	tc := NewTcpClient(clientProcessFunc)
	err := tc.StartTcpClient("192.168.83.139:9999")
	if err != nil {
		belogs.Error("CreateTcpClient(): StartTcpClient tc fail: ", &tc, err)
		return
	}
	belogs.Debug("CreateTcpClient(): tcpclient will SendData")
	tcpClientSendMsg := &TcpTlsClientSendMsg{NextConnectClosePolicy: NEXT_CONNECT_POLICE_KEEP,
		NextRwPolice: NEXT_RW_POLICE_WAIT_READ,
		SendData:     GetTcpClientData(),
	}
	tc.SendMsg(tcpClientSendMsg)
	time.Sleep(60 * time.Second)

	belogs.Debug("CreateTcpClient(): tcpclient will stop")
	tc.CloseGraceful()

}

func CreateTlsClient(t *test.T) {
	clientProcessFunc := new(ClientProcessFunc)
	tlsRootCrtFileName := `ca.cer`
	tlsPublicCrtFileName := `client.cer`
	tlsPrivateKeyFileName := `clientkey.pem`
	belogs.Debug("CreateTlsClient(): tlsRootCrtFileName:", tlsRootCrtFileName,
		"tlsPublicCrtFileName:", tlsPublicCrtFileName,
		"tlsPrivateKeyFileName:", tlsPrivateKeyFileName)
	//CreateTcpClient("192.168.83.139:9999", ClientProcess1)
	tc, err := NewTlsClient(tlsRootCrtFileName, tlsPublicCrtFileName, tlsPrivateKeyFileName, clientProcessFunc)
	if err != nil {
		belogs.Error("CreateTcpClient(): NewTlsClient tc fail: ", &tc, err)
		return
	}
	err = tc.StartTlsClient("192.168.83.139:9999")
	if err != nil {
		belogs.Error("CreateTcpClient(): StartTlsClient tc fail: ", &tc, err)
		return
	}
	belogs.Debug("CreateTcpClient(): tcpclient will SendData")
	tcpClientSendMsg := &TcpTlsClientSendMsg{NextConnectClosePolicy: NEXT_CONNECT_POLICE_KEEP,
		NextRwPolice: NEXT_RW_POLICE_WAIT_READ,
		SendData:     GetTcpClientData(),
	}
	tc.SendMsg(tcpClientSendMsg)
	time.Sleep(60 * time.Second)

	belogs.Debug("CreateTcpClient(): tcpclient will stop")
	tcpClientSendMsg.NextConnectClosePolicy = NEXT_CONNECT_POLICE_CLOSE_GRACEFUL
	tcpClientSendMsg.SendData = nil
	tc.SendMsg(tcpClientSendMsg)

}

type ClientProcessFunc struct {
}

func (cp *ClientProcessFunc) OnConnectProcess(tcpTlsConn *TcpTlsConn) {

	belogs.Info("OnConnectProcess(): tcpclient tcpTlsConn:", tcpTlsConn.RemoteAddr().String())

}
func (cp *ClientProcessFunc) OnCloseProcess(tcpTlsConn *TcpTlsConn) {
	if tcpTlsConn != nil {
		belogs.Info("OnCloseProcess(): tcpclient tcpTlsConn:", tcpTlsConn.RemoteAddr().String())
	}
}

func (sq *ClientProcessFunc) OnReceiveProcess(tcpTlsConn *TcpTlsConn, receiveData []byte) (nextRwPolicy int, leftData []byte, err error) {

	belogs.Debug("OnReceiveProcess() tcpclient  :", tcpTlsConn, convert.Bytes2String(receiveData))
	return NEXT_RW_POLICE_END_READ, make([]byte, 0), nil
}

func GetTcpClientData() (buffer []byte) {

	return []byte{0x00, 0x0b, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x0b, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
}
