package tcpserver

import (
	"bytes"
	"net"
	"testing"
	"time"

	belogs "github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	util "github.com/cpusoft/goutil/tcpserverclient/util"
)

var tcpTestServer *TcpServer

func TestTcpServer(t *testing.T) {
	CreateTcpServer()
	select {}
}

const (
	PDU_TYPE_MIN_LEN      = 8
	PDU_TYPE_LENGTH_START = 4
	PDU_TYPE_LENGTH_END   = 8
)

func CreateTcpServer() {
	serverProcessFunc := new(ServerProcessFunc)
	tcpTestServer = NewTcpServer(serverProcessFunc)
	tcpTestServer.Start("9999")
	time.Sleep(2 * time.Second)
	tcpTestServer.ActiveSend(GetTcpServerData(), "")
}

type ServerProcessFunc struct {
}

func (spf *ServerProcessFunc) OnConnectProcess(tcpConn *net.TCPConn) {
	belogs.Info("OnConnectProcess(): tcpserver tcpConn:", tcpConn.RemoteAddr().String())
}

// recombine to packets or just one receiveData
func (spf *ServerProcessFunc) ReceiveAndSendProcess(tcpConn *net.TCPConn, receiveData []byte) (nextConnectPolicy int, leftData []byte, err error) {
	belogs.Debug("ReceiveAndSendProcess(): tcpserver len(receiveData):", len(receiveData), "   receiveData:", convert.Bytes2String(receiveData))
	// need recombine
	packets, leftData, err := util.RecombineReceiveData(receiveData, PDU_TYPE_MIN_LEN, PDU_TYPE_LENGTH_START, PDU_TYPE_LENGTH_END)
	if err != nil {
		belogs.Error("ReceiveAndSendProcess(): tcpserver RecombineReceiveData fail:", err)
		return util.NEXT_CONNECT_POLICE_CLOSE_FORCIBLE, nil, err
	}
	belogs.Debug("ReceiveAndSendProcess(): tcpserver RecombineReceiveData packets.Len():", packets.Len())

	if packets == nil || packets.Len() == 0 {
		belogs.Debug("ReceiveAndSendProcess(): tcpserver RecombineReceiveData packets is empty:  len(leftData):", len(leftData))
		return util.NEXT_CONNECT_POLICE_CLOSE_GRACEFUL, leftData, nil
	}
	sendData := make([]byte, 0)
	for e := packets.Front(); e != nil; e = e.Next() {
		packet, ok := e.Value.([]byte)
		if !ok || packet == nil || len(packet) == 0 {
			belogs.Debug("ReceiveAndSendProcess(): tcpserver for packets fail:", convert.ToString(e.Value))
			break
		}
		tmpData, err := RtrProcess(packet)
		if err != nil {
			belogs.Error("ReceiveAndSendProcess(): tcpserver RtrProcess fail:", err)
			return util.NEXT_CONNECT_POLICE_CLOSE_FORCIBLE, nil, err
		}
		sendData = append(sendData, tmpData...)

	}

	// may send or not send
	_, err = tcpConn.Write(sendData)
	if err != nil {
		belogs.Error("ReceiveAndSendProcess(): tcpserver,  Write fail:  tcpConn:", tcpConn.RemoteAddr().String(), err)
		return util.NEXT_CONNECT_POLICE_CLOSE_FORCIBLE, nil, err
	}
	// continue to receive next receiveData
	return util.NEXT_CONNECT_POLICE_KEEP, leftData, nil
}
func (spf *ServerProcessFunc) OnCloseProcess(tcpConn *net.TCPConn) {
	if tcpConn != nil {
		belogs.Info("OnCloseProcess(): tcpserver tcpConn:", tcpConn.RemoteAddr().String())
	}

}
func (spf *ServerProcessFunc) ActiveSendProcess(tcpConn *net.TCPConn, sendData []byte) (err error) {
	_, err = tcpConn.Write(sendData)
	if err != nil {
		belogs.Error("ActiveSendProcess(): tcpserver sendData:", convert.PrintBytes(sendData, 8))
		return err
	}
	return nil
}

func GetTcpServerData() (buffer []byte) {

	return []byte{0x00, 0x0a, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x0a, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
}

func RtrProcess(receiveData []byte) (sendData []byte, err error) {
	buf := bytes.NewReader(receiveData)
	belogs.Debug("RtrProcess(): tcpserver buf:", buf)
	return GetTcpServerData(), nil
}
