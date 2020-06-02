package tcpserver

import (
	"net"
	"time"

	belogs "github.com/astaxie/beego/logs"
)

type TcpClient struct {
	stopChan chan string

	tcpClientProcessChan  chan string
	tcpClientProcessFuncs map[string]TcpClientProcessFunc
}

type TcpClientProcessFunc interface {
	ActiveSendAndReceive(conn *net.TCPConn) (err error)
}

// server: 0.0.0.0:port
func NewTcpClient(tcpClientProcessFuncs map[string]TcpClientProcessFunc) (tc *TcpClient) {

	belogs.Debug("NewTcpClient():tcpClientProcessFuncs:", tcpClientProcessFuncs)
	tc = &TcpClient{}
	tc.stopChan = make(chan string)
	tc.tcpClientProcessChan = make(chan string)
	belogs.Debug("NewTcpClient():tc:%p, %p:", tc.stopChan, tc.tcpClientProcessChan)

	tc.tcpClientProcessFuncs = tcpClientProcessFuncs
	belogs.Debug("NewTcpClient():tc:", tc)
	return tc
}

// server: **.**.**.**:port
func (tc *TcpClient) Start(server string) (err error) {
	belogs.Debug("Start():create client, server is  ", server)

	tcpServer, err := net.ResolveTCPAddr("tcp", server)
	if err != nil {
		belogs.Error("Start(): ResolveTCPAddr fail: ", server, err)
		return err
	}
	conn, err := net.DialTCP("tcp4", nil, tcpServer)
	if err != nil {
		belogs.Error("Start(): Dial fail: ", server, tcpServer, err)
		return err
	}
	defer conn.Close()
	belogs.Debug("Start():create client ok, server is  ", server, "   conn:", conn)

	// receive process func
	go func() {
		belogs.Debug("Start():wait for, conn:", conn, "  tcpClientProcessChan:", tc.tcpClientProcessChan)
		for {
			select {
			case tcpClientProcess := <-tc.tcpClientProcessChan:
				belogs.Debug("Start():  tcpClientProcess:", tcpClientProcess)
				if tcpClientProcessFunc, ok := tc.tcpClientProcessFuncs[tcpClientProcess]; ok {
					start := time.Now()
					belogs.Debug("Start(): tcpClientProcessFunc:", tcpClientProcessFunc, "  conn:", conn)

					err = tcpClientProcessFunc.ActiveSendAndReceive(conn)
					if err != nil {
						belogs.Error("Start(): tcpClientProcessFunc.ActiveSendAndReceive fail: ", server, tcpServer, err)
						return
					}
					belogs.Debug("Start(): tcpClientProcess:", tcpClientProcess, "  time(s):", time.Now().Sub(start).Seconds())
				}
			}
		}
	}()

	// wait for exit
	for {
		select {
		case stop := <-tc.stopChan:
			if stop == "stop" {
				close(tc.stopChan)
				close(tc.tcpClientProcessChan)
				belogs.Info("Start(): end client: ", server)
				return nil
			}
		}
	}

}

// exit: to quit
func (tc *TcpClient) CallProcessFunc(clientProcessFunc string) {
	belogs.Debug("CallProcessFunc():  clientProcessFunc:", clientProcessFunc)
	tc.tcpClientProcessChan <- clientProcessFunc
}

func (tc *TcpClient) CallStop() {
	belogs.Debug("CallStop():")
	tc.stopChan <- "stop"
}
