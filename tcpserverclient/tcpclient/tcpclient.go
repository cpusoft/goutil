package tcpclient

import (
	"io"
	"net"
	"time"

	belogs "github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/tcpserverclient/util"
)

type TcpClientMsg struct {
	NextConnectClosePolicy int //NEXT_CONNECT_CLOSE_POLICE_NO  NEXT_CONNECT_CLOSE_POLICE_GRACEFUL  NEXT_CONNECT_CLOSE_POLICE_FORCIBLE
	NextRwPolice           int //NEXT_RW_POLICE_ALL,NEXT_RW_POLICE_WAIT_READ,NEXT_RW_POLICE_WAIT_WRITE
	SendData               []byte
}

type TcpClient struct {
	tcpClientMsgCh       chan TcpClientMsg
	tcpClientProcessFunc TcpClientProcessFunc
}

// server: 0.0.0.0:port
func NewTcpClient(tcpClientProcessFunc TcpClientProcessFunc) (tc *TcpClient) {

	belogs.Debug("NewTcpClient():tcpClientProcessFuncs:", tcpClientProcessFunc)
	tc = &TcpClient{}
	tc.tcpClientMsgCh = make(chan TcpClientMsg)
	tc.tcpClientProcessFunc = tcpClientProcessFunc
	belogs.Info("NewTcpClient():tc:", tc)
	return tc
}

// server: **.**.**.**:port
func (tc *TcpClient) Start(server string) (err error) {
	belogs.Debug("Start(): tcpclient  create client, server is  ", server)

	tcpServer, err := net.ResolveTCPAddr("tcp", server)
	if err != nil {
		belogs.Error("Start(): tcpclient   ResolveTCPAddr fail: ", server, err)
		return err
	}
	belogs.Debug("Start(): tcpclient  create client, server is  ", server, "  tcpServer:", tcpServer)

	tcpConn, err := net.DialTCP("tcp4", nil, tcpServer)
	if err != nil {
		belogs.Error("Start(): tcpclient   Dial fail, server:", server, "  tcpServer:", tcpServer, err)
		return err
	}

	tc.OnConnect(tcpConn)
	belogs.Info("Start(): tcpclient  OnConnect, server is  ", server, "  tcpConn:", tcpConn.RemoteAddr().String())

	//active send to server, and receive from server, loop
	go tc.SendAndReceive(tcpConn)
	belogs.Debug("Start(): tcpclient  waite SendAndReceive, server:", server, "   tcpConn:", tcpConn.RemoteAddr().String())
	return nil
}

func (tc *TcpClient) OnConnect(tcpConn *net.TCPConn) {
	// call process func OnConnect
	belogs.Debug("OnConnect(): tcpclient  tcpConn: ", tcpConn.RemoteAddr().String(), "   call process func: OnConnect ")
	tc.tcpClientProcessFunc.OnConnectProcess(tcpConn)
	belogs.Info("OnConnect(): tcpclient  after OnConnectProcess, tcpConn: ", tcpConn.RemoteAddr().String())
}

func (tc *TcpClient) OnClose(tcpConn *net.TCPConn) {
	// close in the end
	tcpConn.Close()
	close(tc.tcpClientMsgCh)
	belogs.Info("OnClose(): tcpclient  client, tcpConn: ", tcpConn.RemoteAddr().String())
	tcpConn = nil
}

func (tc *TcpClient) SendMsg(tcpClientMsg *TcpClientMsg) {

	belogs.Debug("SendMsg(): tcpclient, tcpClientMsg:", jsonutil.MarshalJson(*tcpClientMsg))
	tc.tcpClientMsgCh <- *tcpClientMsg
}

func (tc *TcpClient) SendAndReceive(tcpConn *net.TCPConn) (err error) {
	belogs.Debug("SendAndReceive(): tcpclient , tcpConn:", tcpConn.RemoteAddr().String())
	for {
		// wait next tcpClientMsgCh: only error or NEXT_CONNECT_POLICE_CLOSE_** will end loop
		select {
		case tcpClientMsg := <-tc.tcpClientMsgCh:
			nextConnectClosePolicy := tcpClientMsg.NextConnectClosePolicy
			nextRwPolice := tcpClientMsg.NextRwPolice
			sendData := tcpClientMsg.SendData
			belogs.Debug("SendAndReceive(): tcpclient , tcpConn:", tcpConn.RemoteAddr().String(),
				"  tcpClientMsg: ", jsonutil.MarshalJson(tcpClientMsg))

			// if close
			if nextConnectClosePolicy == util.NEXT_CONNECT_POLICE_CLOSE_GRACEFUL ||
				nextConnectClosePolicy == util.NEXT_CONNECT_POLICE_CLOSE_FORCIBLE {
				belogs.Info("SendAndReceive(): tcpclient   nextConnectClosePolicy close end client, will end tcpConn: ", tcpConn.RemoteAddr().String(),
					"   nextConnectClosePolicy:", nextConnectClosePolicy)
				tc.OnClose(tcpConn)
				return nil
			}

			// send data
			start := time.Now()
			n, err := tcpConn.Write(sendData)
			if err != nil {
				belogs.Error("SendAndReceive(): tcpclient   Write fail:  tcpConn:", tcpConn.RemoteAddr().String(), err)
				return err
			}
			belogs.Debug("SendAndReceive(): tcpclient   Write to tcpConn:", tcpConn.RemoteAddr().String(),
				"  len(sendData):", len(sendData), "  write n:", n, "   nextRwPolice:", nextRwPolice,
				"  time(s):", time.Now().Sub(start).Seconds())

			// if wait receive, then wait next tcpClientMsgCh
			if nextRwPolice == util.NEXT_RW_POLICE_WAIT_READ {
				// if server tell client: end this loop, or end conn
				err := tc.OnReceive(tcpConn)
				if err != nil {
					belogs.Error("SendAndReceive(): tcpclient   Write fail:  tcpConn:", tcpConn.RemoteAddr().String(), err)
					return err
				}
				belogs.Info("SendAndReceive(): tcpclient  shouldWaitReceive yes, tcpConn:", tcpConn.RemoteAddr().String(),
					"  len(sendData):", len(sendData), "  write n:", n,
					"  time(s):", time.Now().Sub(start).Seconds())
				continue
			} else {
				belogs.Info("SendAndReceive(): tcpclient  OnReceive, shouldWaitReceive no, will return: ", tcpConn.RemoteAddr().String())
				continue
			}
		}
	}

}

func (tc *TcpClient) OnReceive(tcpConn *net.TCPConn) (err error) {
	belogs.Debug("OnReceive(): tcpclient  wait for OnReceive, tcpConn:", tcpConn.RemoteAddr().String())
	var leftData []byte
	// one packet
	buffer := make([]byte, 2048)
	// wait for new packet to read

	for {
		n, err := tcpConn.Read(buffer)
		start := time.Now()
		belogs.Debug("OnReceive(): tcpclient  client read: Read n: ", tcpConn.RemoteAddr().String(), n)
		if err != nil {
			if err == io.EOF {
				// is not error, just client close
				belogs.Debug("OnReceive(): tcpclient   io.EOF, client close: ", tcpConn.RemoteAddr().String(), err)
				return nil
			}
			belogs.Error("OnReceive(): tcpclient   Read fail, err ", tcpConn.RemoteAddr().String(), err)
			return err
		}
		if n == 0 {
			continue
		}

		belogs.Debug("OnReceive(): tcpclient  client tcpConn: ", tcpConn.RemoteAddr().String(), "  n:", n,
			" , will call process func: OnReceiveAndSend,  time(s):", time.Now().Sub(start))
		nextRwPolicy, leftData, err := tc.tcpClientProcessFunc.OnReceiveProcess(tcpConn, append(leftData, buffer[:n]...))
		belogs.Info("OnReceive(): tcpclient  tcpClientProcessFunc.OnReceiveProcess, tcpConn: ", tcpConn.RemoteAddr().String(), " receive n: ", n,
			"  len(leftData):", len(leftData), "  nextRwPolicy:", nextRwPolicy, "  time(s):", time.Now().Sub(start))
		if err != nil {
			belogs.Error("OnReceive(): tcpclient  tcpClientProcessFunc.OnReceiveProcess  fail ,will close this tcpConn : ", tcpConn.RemoteAddr().String(), err)
			return err
		}
		if nextRwPolicy == util.NEXT_RW_POLICE_END_READ {
			belogs.Debug("OnReceive(): tcpclient  nextRwPolicy, will end this write/read loop: ", tcpConn.RemoteAddr().String())
			return nil
		}
	}

}
