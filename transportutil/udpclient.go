package transportutil

import (
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
	TransportMsg chan TransportMsg
}

// server: 0.0.0.0:port
func NewUdpClient(udpClientProcess UdpClientProcess, transportMsg chan TransportMsg) (tc *UdpClient) {

	belogs.Debug("NewUdpClient():udpClientProcess:", udpClientProcess)
	tc = &UdpClient{}
	tc.connType = "udp"
	tc.udpClientProcess = udpClientProcess
	tc.TransportMsg = transportMsg
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
	belogs.Debug("UdpClient.StartUdpClient(): NewFromUdpConn ok, server:", server, "   udpConn:", tc.udpConn.serverUdpAddr)
	go tc.waitTransportMsg()

	// onReceive
	go tc.onReceive()

	belogs.Info("UdpClient.StartUdpClient(): onReceive, server is  ", server, "  udpConn:", tc.udpConn.serverUdpAddr)
	return nil
}

func (tc *UdpClient) onReceive() (err error) {
	belogs.Debug("UdpClient.onReceive(): wait for onReceive, udpConn:", tc.udpConn.serverUdpAddr)

	// one packet
	buffer := make([]byte, 2048)
	// wait for new packet to read

	// when end onReceive, will onClose
	defer tc.onClose()
	for {
		start := time.Now()
		n, err := tc.udpConn.ReadFromServer(buffer)
		//	if n == 0 {
		//		continue
		//	}
		if err != nil {
			if err == io.EOF {
				// is not error, just client close
				belogs.Debug("UdpClient.onReceive(): io.EOF, client close, udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr, err)
				return nil
			}
			belogs.Error("UdpClient.onReceive(): Read fail or connect is closing, udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr, err)
			return err
		}

		belogs.Debug("UdpClient.onReceive(): Read n :", n, " from udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr,
			"  time(s):", time.Now().Sub(start))
		err = tc.udpClientProcess.OnReceiveProcess(tc.udpConn, append(buffer[:n]))

		if err != nil {
			belogs.Error("UdpClient.onReceive(): udpClientProcess.OnReceiveProcess  fail ,will close this udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr, err)
			return err
		}
		belogs.Info("UdpClient.onReceive(): udpClientProcess.OnReceiveProcess, udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr, " receive n: ", n,
			"  time(s):", time.Now().Sub(start))

		// reset buffer
		buffer = make([]byte, 2048)
		belogs.Debug("UdpClient.onReceive(): will reset buffer and wait for Read from udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr,
			"  time(s):", time.Now().Sub(start))

	}

}

func (tc *UdpClient) onClose() {
	// close in the end
	belogs.Info("UdpClient.onClose(): udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr)
	tc.udpConn.Close()

}

func (tc *UdpClient) SendMsg(transportMsg *TransportMsg) {

	belogs.Debug("UdpClient.SendMsg(): transportMsg:", jsonutil.MarshalJson(*transportMsg))
	tc.TransportMsg <- *transportMsg
}

func (tc *UdpClient) waitTransportMsg() (err error) {
	belogs.Debug("UdpClient.waitTransportMsg(): udpConn.serverUdpAddr:", tc.udpConn.serverUdpAddr)
	for {
		// wait next transportMsg: only error or NEXT_CONNECT_POLICY_CLOSE_** will end loop
		select {
		case transportMsg := <-tc.TransportMsg:
			belogs.Info("UdpClient.waitTransportMsg(): transportMsg:", jsonutil.MarshalJson(transportMsg),
				"  udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr)

			switch transportMsg.MsgType {
			case MSG_TYPE_CLIENT_CLOSE_CONNECT:
				belogs.Info("UdpClient.waitTransportMsg(): msgType is MSG_TYPE_CLIENT_CLOSE_CONNECT,",
					" will close for udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr, " will return, close waitTransportMsg")
				tc.onClose()
				// end for/select
				// will return, close waitTransportMsg
				return nil
			case MSG_TYPE_COMMON_SEND_DATA:
				belogs.Info("UdpClient.waitTransportMsg(): msgType is MSG_TYPE_COMMON_SEND_DATA,",
					" will send to udpConn.serverUdpAddr: ", tc.udpConn.serverUdpAddr)
				sendData := transportMsg.SendData
				belogs.Debug("UdpClient.waitTransportMsg(): send to server:", tc.udpConn.serverUdpAddr,
					"   sendData:", convert.PrintBytesOneLine(sendData))

				// send data
				start := time.Now()
				n, err := tc.udpConn.WriteToServer(sendData)
				if err != nil {
					belogs.Error("UdpClient.waitTransportMsg(): Write fail, will close  udpConn.serverUdpAddr:", tc.udpConn.serverUdpAddr, err)
					tc.onClose()
					return err
				}
				belogs.Info("UdpClient.waitTransportMsg(): Write to udpConn.serverUdpAddr:", tc.udpConn.serverUdpAddr,
					"  len(sendData):", len(sendData), "  write n:", n,
					"  time(s):", time.Since(start))

			}
		}
	}

}
