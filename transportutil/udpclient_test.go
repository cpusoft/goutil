package transportutil

import (
	"testing"
	"time"

	"github.com/cpusoft/goutil/belogs"
	_ "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/tcptlsutil"
)

var dnsUdpClient *DnsUdpClient

//////////////////////////////////
//

type DnsUdpClient struct {
	// tcp/tls server and callback Func
	udpClient         *UdpClient
	businessToConnMsg chan BusinessToConnMsg
}

func StartDnsUdpClient(serverProtocol string, serverHost string, serverPort string) (err error) {
	belogs.Debug("StartDnsUdpClient(): serverProtocol:", serverProtocol,
		"  serverHost:", serverHost,
		"  serverPort:", serverPort)

	// no :=
	dnsUdpClient = &DnsUdpClient{}
	dnsUdpClient.businessToConnMsg = make(chan BusinessToConnMsg, 15)
	belogs.Debug("StartDnsUdpClient(): businessToConnMsg:", dnsUdpClient.businessToConnMsg)

	// process
	dnsClientProcess := NewDnsClientProcess(dnsUdpClient.businessToConnMsg)
	belogs.Debug("StartDnsUdpClient(): NewDnsClientProcess:", dnsClientProcess)

	// tclTlsClient
	dnsUdpClient.udpClient = NewUdpClient(dnsClientProcess, dnsUdpClient.businessToConnMsg)
	belogs.Debug("StartDnsUdpClient(): dnsUdpClient:", dnsUdpClient)

	// set to global dnsClient
	if serverProtocol == "udp" {
		err = dnsUdpClient.udpClient.StartUdpClient(serverHost + ":" + serverPort)
	}
	belogs.Info("StartDnsUdpClient(): start serverHost:", serverHost, "   serverPort:", serverPort, " serverProtocol:", serverProtocol)

	return nil

}

type DnsClientProcess struct {
	businessToConnMsg chan BusinessToConnMsg
}

func NewDnsClientProcess(businessToConnMsg chan BusinessToConnMsg) *DnsClientProcess {
	c := &DnsClientProcess{}
	c.businessToConnMsg = businessToConnMsg

	return c
}

func (c *DnsClientProcess) OnReceiveProcess(udpConn *UdpConn, receiveData []byte) (err error) {
	belogs.Debug("OnReceiveProcess(): client len(receiveData):", len(receiveData), "   receiveData:", convert.PrintBytesOneLine(receiveData))

	receiveStr := string(receiveData)
	belogs.Debug("OnReceiveProcess():receiveStr:", receiveStr)

	// continue to receive next receiveData
	return nil
}

func TestDnsUdpClient(t *testing.T) {

	err := StartDnsUdpClient("udp", "127.0.0.1", "9998")
	if err != nil {
		belogs.Error("TestDnsUdpClient(): StartDnsUdpClient fail:", err)
	}
	sendBytes := []byte("test udp")
	businessToConnMsg := &BusinessToConnMsg{
		BusinessToConnMsgType: tcptlsutil.BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_DATA,
		SendData:              sendBytes,
	}
	dnsUdpClient.udpClient.SendBusinessToConnMsg(businessToConnMsg)

	time.Sleep(5 * time.Second)

}
