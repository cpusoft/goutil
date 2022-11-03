package transportutil

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

// core struct: Start/onConnect/receiveAndSend....
type UdpServer struct {
	// state
	state uint64
	// udp
	connType string

	// udp
	udpConnsMutext sync.RWMutex
	udpConn        *UdpConn

	// process
	udpServerProcess UdpServerProcess
	// for channel
	businessToConnMsg chan BusinessToConnMsg
}

//
func NewUdpServer(udpServerProcess UdpServerProcess, businessToConnMsg chan BusinessToConnMsg) (us *UdpServer) {

	belogs.Debug("NewUdpServer():udpServerProcess:", udpServerProcess)
	us = &UdpServer{}
	us.state = SERVER_STATE_INIT
	us.connType = "udp"
	us.udpServerProcess = udpServerProcess
	us.businessToConnMsg = businessToConnMsg
	belogs.Debug("NewUdpServer():us:", us)
	return us
}

// port: `8888` --> `0.0.0.0:8888`
func (us *UdpServer) StartUdpServer(port string) (err error) {
	// resolve
	serverUdpAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:"+port)
	if err != nil {
		belogs.Error("StartUdpServer(): ResolveUDPAddr fail, port:", port, err)
		return err
	}
	belogs.Debug("StartUdpServer(): ResolveUDPAddr ok,  serverUdpAddr:", serverUdpAddr)

	// get udpConn --> udpConn
	conn, err := net.ListenUDP("udp", serverUdpAddr)
	if err != nil {
		belogs.Error("StartUdpServer(): ListenUDP fail, serverUdpAddr:", serverUdpAddr, err)
		return err
	}
	us.udpConn = NewFromUdpConn(conn)
	us.udpConn.SetServerUdpAddr(serverUdpAddr)
	belogs.Debug("StartUdpServer(): ListenUDP ok,  udpConn:", us.udpConn)

	go us.waitBusinessToConnMsg()

	// wait new conn
	go us.receiveAndSend()
	return nil
}

func (us *UdpServer) receiveAndSend() {

	belogs.Debug("UdpServer.acceptNewReadFromClient(): will read from client")
	us.state = SERVER_STATE_RUNNING
	for {
		buffer := make([]byte, 1024)
		len, clientUdpAddr, err := us.udpConn.ReadFromClient(buffer)
		if err != nil {
			if err == io.EOF {
				// is not error, just client close
				belogs.Info("UdpServer.acceptNewReadFromClient(): Read io.EOF, client close,  clientUdpAddr:", clientUdpAddr, err)
				return
			}
			belogs.Error("UdpServer.acceptNewReadFromClient(): Read remote fail: ", err)
			continue
		}
		belogs.Info("UdpServer.acceptNewReadFromClient():  Accept remote, clientAddrKey: ", clientUdpAddr, "  len:", len)
		// no onConnect
		go func() {
			err := us.udpServerProcess.OnReceiveAndSendProcess(us.udpConn, clientUdpAddr, buffer[:len])
			if err != nil {
				belogs.Error("UdpServer.receiveAndSend(): OnReceiveAndSendProcess fail ,will remove this udpConn : ", clientUdpAddr, err)
				us.udpConn.DelClientUdpAddr(clientUdpAddr)
				return
			}
		}()
	}
}

func (us *UdpServer) onClose() {
	// close in the end
	if us.udpConn == nil {
		return
	}
	us.udpConn.Close()
	us.udpConn = nil
}

func (us *UdpServer) SendBusinessToConnMsg(businessToConnMsg *BusinessToConnMsg) {

	belogs.Debug("UdpServer.SendBusinessToConnMsg():, businessToConnMsg:", jsonutil.MarshalJson(*businessToConnMsg))
	us.businessToConnMsg <- *businessToConnMsg
}

func (us *UdpServer) waitBusinessToConnMsg() {
	belogs.Debug("UdpServer.waitBusinessToConnMsg(): will waitBusinessToConnMsg")
	for {
		select {
		case businessToConnMsg := <-us.businessToConnMsg:
			belogs.Info("UdpServer.waitBusinessToConnMsg(): businessToConnMsg:", jsonutil.MarshalJson(businessToConnMsg))

			switch businessToConnMsg.BusinessToConnMsgType {
			case BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_FORCIBLE:
				// ignore conns's writing/reading, just close
				belogs.Info("UdpServer.waitBusinessToConnMsg(): businessToConnMsgType is BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_FORCIBLE")
				fallthrough
			case BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_GRACEFUL:
				// close and wait connect.Read and Accept
				us.state = SERVER_STATE_CLOSING
				us.onClose()
				belogs.Info("UdpServer.waitBusinessToConnMsg(): will close server graceful, will return waitBusinessToConnMsg:")
				// end for/select
				us.state = SERVER_STATE_CLOSED
				// will return, close waitBusinessToConnMsg
				return

			case BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_DATA:

				serverConnKey := businessToConnMsg.ServerConnKey
				sendData := businessToConnMsg.SendData
				belogs.Info("UdpServer.waitBusinessToConnMsg(): businessToConnMsgType is BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_DATA, serverConnKey:", serverConnKey,
					"  sendData:", convert.PrintBytesOneLine(sendData))
				start := time.Now()
				n, err := us.udpConn.WriteToClient(sendData, serverConnKey)
				if err != nil {
					belogs.Error("UdpServer.waitBusinessToConnMsg(): activeSend fail, serverConnKey:", serverConnKey,
						"  sendData:", convert.PrintBytesOneLine(sendData), err)
					// err, no return
					// return
				} else {
					belogs.Info("UdpServer.waitBusinessToConnMsg(): activeSend ok, serverConnKey:", serverConnKey,
						"  sendData:", convert.PrintBytesOneLine(sendData), " write n:", n,
						"  time(s):", time.Since(start))
				}
			}
		}
	}
}
