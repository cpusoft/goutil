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
	// udp receive bytes len
	receiveOnePacketLength int

	// for close
	udpConn *UdpConn

	// for channel
	businessToConnMsgCh chan BusinessToConnMsg

	// for onReceive to SendAndReceiveMsg
	connToBusinessMsgCh chan ConnToBusinessMsg
}

// server: 0.0.0.0:port
func NewUdpClient(udpClientProcess UdpClientProcess, businessToConnMsgCh chan BusinessToConnMsg, receiveOnePacketLength int) (uc *UdpClient) {

	belogs.Debug("NewUdpClient():udpClientProcess:", udpClientProcess, "  receiveOnePacketLength:", receiveOnePacketLength)
	uc = &UdpClient{}
	uc.connType = "udp"
	uc.udpClientProcess = udpClientProcess
	uc.businessToConnMsgCh = businessToConnMsgCh
	uc.connToBusinessMsgCh = make(chan ConnToBusinessMsg)
	uc.receiveOnePacketLength = receiveOnePacketLength
	belogs.Info("NewUdpClient():tc:", uc)
	return uc
}

// server: **.**.**.**:port
func (uc *UdpClient) StartUdpClient(server string) (err error) {
	belogs.Debug("UdpClient.StartUdpClient(): create client, server is  ", server)

	serverUdpAddr, _ := net.ResolveUDPAddr("udp", server)
	if err != nil {
		belogs.Error("UdpClient.StartUdpClient(): ResolveUDPAddr fail, server:", server, err)
		return err
	}

	//连接udpAddr，返回 udpConn
	udpConn, err := net.DialUDP("udp", nil, serverUdpAddr)
	uc.udpConn = NewFromUdpConn(udpConn)
	uc.udpConn.SetServerUdpAddr(serverUdpAddr)
	//active send to server, and receive from server, loop
	belogs.Info("UdpClient.StartUdpClient(): NewFromUdpConn ok, server:", server, "   udpConn:", uc.udpConn.serverUdpAddr)
	// onReceive
	go uc.onReceive()

	return nil
}

func (uc *UdpClient) onReceive() (err error) {
	belogs.Debug("UdpClient.onReceive(): wait for ReadFromServer, receiveOnePacketLength:", uc.receiveOnePacketLength)
	for {
		start := time.Now()
		buffer := make([]byte, uc.receiveOnePacketLength)
		n, err := uc.udpConn.ReadFromServer(buffer)
		if err != nil {
			if err == io.EOF {
				// is not error, just close
				belogs.Debug("UdpClient.onReceive(): io.EOF, client close, receiveOnePacketLength:", uc.receiveOnePacketLength,
					"   udpConn.serverUdpAddr:", uc.udpConn.serverUdpAddr, err)
				return nil
			}
			belogs.Error("UdpClient.onReceive(): Read fail or connect is closing, receiveOnePacketLength:", uc.receiveOnePacketLength,
				"  udpConn.serverUdpAddr: ", uc.udpConn.serverUdpAddr, err)
			return err
		}

		belogs.Debug("UdpClient.onReceive(): Read n :", n, " from udpConn.serverUdpAddr: ", uc.udpConn.serverUdpAddr,
			"  time(s):", time.Since(start))
		connToBusinessMsg, err := uc.udpClientProcess.OnReceiveProcess(uc.udpConn, append(buffer[:n]))
		if err != nil {
			belogs.Error("UdpClient.onReceive(): udpClientProcess.OnReceiveProcess  fail ,will close this udpConn.serverUdpAddr: ", uc.udpConn.serverUdpAddr, err)
			return err
		}
		belogs.Info("UdpClient.onReceive(): udpClientProcess.OnReceiveProcess, udpConn.serverUdpAddr: ", uc.udpConn.serverUdpAddr, " receive n: ", n,
			"  connToBusinessMsg:", jsonutil.MarshalJson(connToBusinessMsg), "  time(s):", time.Since(start))
		go func() {
			if !connToBusinessMsg.IsActiveSendFromServer {
				belogs.Debug("UdpClient.onReceive(): udpClientProcess.OnReceiveProcess, will send to uc.connToBusinessMsg:", jsonutil.MarshalJson(connToBusinessMsg))
				uc.connToBusinessMsgCh <- *connToBusinessMsg
				belogs.Debug("UdpClient.onReceive(): udpClientProcess.OnReceiveProcess, have send to uc.connToBusinessMsg:", jsonutil.MarshalJson(connToBusinessMsg))

			}
		}()
	}
}

func (uc *UdpClient) onClose() {
	// close in the end
	belogs.Info("UdpClient.onClose(): udpConn.serverUdpAddr: ", uc.udpConn.serverUdpAddr)
	uc.udpConn.Close()
}
func (uc *UdpClient) SendAndReceiveMsg(businessToConnMsg *BusinessToConnMsg) (connToBusinessMsg *ConnToBusinessMsg, err error) {

	belogs.Debug("UdpClient.SendAndReceiveMsg(): businessToConnMsg:", jsonutil.MarshalJson(*businessToConnMsg))
	//uc.businessToConnMsg <- *businessToConnMsg

	switch businessToConnMsg.BusinessToConnMsgType {
	case BUSINESS_TO_CONN_MSG_TYPE_CLIENT_CLOSE_CONNECT:
		belogs.Info("UdpClient.SendAndReceiveMsg(): businessToConnMsgType is BUSINESS_TO_CONN_MSG_TYPE_CLIENT_CLOSE_CONNECT,",
			" will close for udpConn.serverUdpAddr: ", uc.udpConn.serverUdpAddr, " will return, close SendAndReceiveMsg")
		// end for/select
		// will return, close SendAndReceiveMsg
		uc.onClose()
		return nil, nil
	case BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_AND_RECEIVE_DATA:
		belogs.Info("UdpClient.SendAndReceiveMsg(): businessToConnMsgType is BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_DATA,",
			" will send to udpConn.serverUdpAddr: ", uc.udpConn.serverUdpAddr)
		sendData := businessToConnMsg.SendData
		belogs.Debug("UdpClient.SendAndReceiveMsg(): send to server:", uc.udpConn.serverUdpAddr,
			"   sendData:", convert.PrintBytesOneLine(sendData))
		belogs.Info("UdpClient.SendAndReceiveMsg(): send to server:", uc.udpConn.serverUdpAddr,
			"   len(sendData):", len(sendData))

		// send data
		start := time.Now()
		n, err := uc.udpConn.WriteToServer(sendData)
		if err != nil {
			belogs.Error("UdpClient.SendAndReceiveMsg(): Write fail, will close  udpConn.serverUdpAddr:", uc.udpConn.serverUdpAddr, err)
			return nil, err
		}
		belogs.Info("UdpClient.SendAndReceiveMsg(): Write to udpConn.serverUdpAddr:", uc.udpConn.serverUdpAddr,
			"  len(sendData):", len(sendData), "  write n:", n,
			"  time(s):", time.Since(start))
		if !businessToConnMsg.NeedClientWaitForServerResponse {
			belogs.Debug("UdpClient.SendAndReceiveMsg(): isnot NeedClientWaitForServerResponse, just return, businessToConnMsg:", jsonutil.MarshalJson(businessToConnMsg))
			return nil, nil
		}
		// wait receive msg from "onReceive"
		belogs.Debug("UdpClient.SendAndReceiveMsg(): will receive from uc.connToBusinessMsg: ")
		//connToBusinessMsg := <-uc.connToBusinessMsg
		for {
			belogs.Debug("UdpClient.SendAndReceiveMsg(): for select,  uc.connToBusinessMsgCh:", uc.connToBusinessMsgCh)
			select {
			case connToBusinessMsg := <-uc.connToBusinessMsgCh:
				belogs.Debug("UdpClient.SendAndReceiveMsg(): receive from uc.connToBusinessMsg, connToBusinessMsg:", jsonutil.MarshalJson(connToBusinessMsg),
					"  time(s):", time.Since(start))
				return &connToBusinessMsg, nil

			case <-time.After(5 * time.Second):
				belogs.Debug("UdpClient.SendAndReceiveMsg(): receive fail, timeout")
				return nil, errors.New("server response is timeout")
			}
		}
	}
	return nil, errors.New("BusinessToConnMsgType is not supported")
}
func (uc *UdpClient) GetUdpServerAddrKey() string {
	return GetUdpAddrKey(uc.udpConn.serverUdpAddr)
}
