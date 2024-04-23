package transportutil

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/cpusoft/goutil/belogs"
	_ "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
)

var dnsUdpServer *DnsUdpServer

type DnsUdpServer struct {
	// tcp/tls server and callback Func
	udpServer         *UdpServer
	businessToConnMsg chan BusinessToConnMsg
}

func StartDnsUdpServer(serverProtocol string, serverPort string) (err error) {
	belogs.Debug("StartDnsUdpServer(): serverProtocol:", serverProtocol, "   serverPort:", serverPort)

	// no :=
	dnsUdpServer = &DnsUdpServer{}
	// msg
	dnsUdpServer.businessToConnMsg = make(chan BusinessToConnMsg, 15)
	belogs.Debug("StartDnsUdpServer(): businessToConnMsg:", dnsUdpServer.businessToConnMsg)

	// process
	dnsServerProcess := NewServerProcess(dnsUdpServer.businessToConnMsg)
	belogs.Debug("StartDnsUdpServer(): dnsServerProcess:", dnsServerProcess)

	// tclTlsServer
	dnsUdpServer.udpServer = NewUdpServer(dnsServerProcess, dnsUdpServer.businessToConnMsg, 4096)
	belogs.Debug("StartDnsUdpServer(): dnsUdpServer:", dnsUdpServer)
	if serverProtocol == "udp" {
		err = dnsUdpServer.udpServer.StartUdpServer(serverPort)
	}
	if err != nil {
		belogs.Error("StartDnsUdpServer(): StartTlsServer or StartTcpServer fail,  serverProtocol:", serverProtocol,
			err)
		return err
	}
	belogs.Info("StartDnsUdpServer(): start serverProtocol:", serverProtocol, "   serverPort:", serverPort)
	return nil
}

type ServerProcess struct {
	businessToConnMsg chan BusinessToConnMsg
}

func NewServerProcess(businessToConnMsg chan BusinessToConnMsg) *ServerProcess {
	c := &ServerProcess{}
	c.businessToConnMsg = businessToConnMsg
	return c
}

func (c *ServerProcess) OnReceiveAndSendProcess(udpConn *UdpConn, clientUdpAddr *net.UDPAddr, receiveData []byte) (err error) {
	fmt.Println("OnReceiveAndSendProcess():", convert.PrintBytesOneLine(receiveData))
	// 发送数据
	sendStr := string(receiveData) + " from server"

	//len, err := udpConn.WriteToClient([]byte(sendStr))
	serverConnKey := GetUdpAddrKey(clientUdpAddr)
	sendBytes := []byte(sendStr)
	businessToConnMsg := &BusinessToConnMsg{
		BusinessToConnMsgType: BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_DATA,
		SendData:              sendBytes,
		ServerConnKey:         serverConnKey,
	}
	dnsUdpServer.udpServer.SendBusinessToConnMsg(businessToConnMsg)

	return
}

func TestDnsUdpServer(t *testing.T) {

	StartDnsUdpServer("udp", "9998")
	time.Sleep(50000 * time.Second)

}
