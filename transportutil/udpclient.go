package transportutil

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type UdpClient struct {
	// both tcp and tls
	connType         string
	udpClientProcess UdpClientProcess

	// for close
	udpConn *UdpConn

	// for channel
	businessToConnMsg chan BusinessToConnMsg
}

// server: 0.0.0.0:port
func NewUdpClient(udpClientProcess UdpClientProcess, businessToConnMsg chan BusinessToConnMsg) (tc *UdpClient) {

	belogs.Debug("NewUdpClient():udpClientProcess:", udpClientProcess)
	tc = &UdpClient{}
	tc.connType = "udp"
	tc.udpClientProcess = udpClientProcess
	tc.businessToConnMsg = businessToConnMsg
	belogs.Info("NewUdpClient():tc:", tc)
	return tc
}

// server: **.**.**.**:port
func (tc *UdpClient) StartUdpClient(server string) (err error) {
	belogs.Debug("UdpClient.StartUdpClient(): create client, server is  ", server)

	serverUdpAddr, _ := net.ResolveUDPAddr("udp", server)
	if err != nil {
		belogs.Error("UdpClient.StartUdpClient(): ResolveUDPAddr fail, server:", server, err)
		return err
	}

	//连接udpAddr，返回 udpConn
	udpConn, err := net.DialUDP("udp", nil, serverUdpAddr)
	tc.udpConn = NewFromUdpConn(udpConn)
	tc.udpConn.SetServerUdpAddr(serverUdpAddr)
	//active send to server, and receive from server, loop
	belogs.Info("UdpClient.StartUdpClient(): NewFromUdpConn ok, server:", server, "   udpConn:", tc.udpConn.serverUdpAddr)
	return nil
}

func (tc *UdpClient) onClose() {
	// close in the end
	belogs.Info("UdpClient.onClose(): udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr)
	tc.udpConn.Close()

}

func (tc *UdpClient) SendAndReceiveMsg(businessToConnMsg *BusinessToConnMsg) (connToBusinessMsg *ConnToBusinessMsg, err error) {

	belogs.Debug("UdpClient.SendAndReceiveMsg(): businessToConnMsg:", jsonutil.MarshalJson(*businessToConnMsg))
	//tc.businessToConnMsg <- *businessToConnMsg

	switch businessToConnMsg.BusinessToConnMsgType {
	case BUSINESS_TO_CONN_MSG_TYPE_CLIENT_CLOSE_CONNECT:
		belogs.Info("UdpClient.SendAndReceiveMsg(): businessToConnMsgType is BUSINESS_TO_CONN_MSG_TYPE_CLIENT_CLOSE_CONNECT,",
			" will close for udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr, " will return, close SendAndReceiveMsg")
		// end for/select
		// will return, close SendAndReceiveMsg
		tc.onClose()
		return nil, nil
	case BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_AND_RECEIVE_DATA:
		belogs.Info("UdpClient.SendAndReceiveMsg(): businessToConnMsgType is BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_DATA,",
			" will send to udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr)
		sendData := businessToConnMsg.SendData
		belogs.Debug("UdpClient.SendAndReceiveMsg(): send to server:", tc.udpConn.serverUdpAddr,
			"   sendData:", convert.PrintBytesOneLine(sendData))

		// send data
		start := time.Now()
		n, err := tc.udpConn.WriteToServer(sendData)
		if err != nil {
			belogs.Error("UdpClient.SendAndReceiveMsg(): Write fail, will close  udpConn.serverUdpAddr:", tc.udpConn.serverUdpAddr, err)
			return nil, err
		}
		belogs.Info("UdpClient.SendAndReceiveMsg(): Write to udpConn.serverUdpAddr:", tc.udpConn.serverUdpAddr,
			"  len(sendData):", len(sendData), "  write n:", n,
			"  time(s):", time.Since(start))
		// one packet
		buffer := make([]byte, 2048)
		n, err = tc.udpConn.ReadFromServer(buffer)
		//	if n == 0 {
		//		continue
		//	}
		if err != nil {
			if err == io.EOF {
				// is not error, just client close
				belogs.Debug("UdpClient.SendAndReceiveMsg(): io.EOF, client close, udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr, err)
				return nil, nil
			}
			belogs.Error("UdpClient.SendAndReceiveMsg(): Read fail or connect is closing, udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr, err)
			return nil, err
		}

		belogs.Debug("UdpClient.SendAndReceiveMsg(): Read n :", n, " from udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr,
			"  time(s):", time.Now().Sub(start))
		connToBusinessMsg, err = tc.udpClientProcess.OnReceiveProcess(tc.udpConn, append(buffer[:n]))
		if err != nil {
			belogs.Error("UdpClient.SendAndReceiveMsg(): udpClientProcess.OnReceiveProcess  fail ,will close this udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr, err)
			return nil, err
		}
		belogs.Info("UdpClient.SendAndReceiveMsg(): udpClientProcess.OnReceiveProcess, udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr, " receive n: ", n,
			"  time(s):", time.Now().Sub(start))
		return connToBusinessMsg, nil
	}
	return nil, errors.New("BusinessToConnMsgType is not supported")
}