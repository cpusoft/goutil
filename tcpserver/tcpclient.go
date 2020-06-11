package tcpserver

import (
	"io"
	"net"
	"time"

	belogs "github.com/astaxie/beego/logs"
)

type TcpClient struct {
	stopChan chan string

	tcpClientProcessChan chan string
	tcpClientProcessFunc TcpClientProcessFunc
}

type TcpClientProcessFunc interface {
	ActiveSend(conn *net.TCPConn, tcpClientProcessChan string) (err error)
	OnReceive(conn *net.TCPConn, receiveData []byte) (err error)
}

// server: 0.0.0.0:port
func NewTcpClient(tcpClientProcessFunc TcpClientProcessFunc) (tc *TcpClient) {

	belogs.Debug("NewTcpClient():tcpClientProcessFuncs:", tcpClientProcessFunc)
	tc = &TcpClient{}
	tc.stopChan = make(chan string)
	tc.tcpClientProcessChan = make(chan string)
	belogs.Debug("NewTcpClient():tc:%p, %p:", tc.stopChan, tc.tcpClientProcessChan)

	tc.tcpClientProcessFunc = tcpClientProcessFunc
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

	// get process func to active send to server
	go tc.waitActiveSend(conn)

	// wait for receive bytes from server
	go tc.waitReceive(conn)

	// wait for exit
	for {
		belogs.Debug("Start():wait for stop  ", server, "   conn:", conn)
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

func (tc *TcpClient) waitActiveSend(conn *net.TCPConn) {
	belogs.Debug("waitActiveSend():wait for ActiveSend, conn:", conn)
	for {
		select {
		case tcpClientProcessChan := <-tc.tcpClientProcessChan:
			belogs.Debug("waitActiveSend():  tcpClientProcess:", tcpClientProcessChan)
			start := time.Now()

			err := tc.tcpClientProcessFunc.ActiveSend(conn, tcpClientProcessChan)
			if err != nil {
				belogs.Error("waitActiveSend(): tcpClientProcessFunc.ActiveSendAndReceive fail:  conn:", conn, err)
				return
			}
			belogs.Info("waitActiveSend(): tcpClientProcessChan:", tcpClientProcessChan, "  time(s):", time.Now().Sub(start).Seconds())
		}
	}
}

func (tc *TcpClient) waitReceive(conn *net.TCPConn) {
	belogs.Debug("waitReceive():wait for OnReceive, conn:", conn)
	// one packet
	buffer := make([]byte, 2048)

	// wait for new packet to read
	for {
		n, err := conn.Read(buffer)
		start := time.Now()
		belogs.Debug("waitReceive():server read: Read n: ", conn, n)
		if err != nil {
			if err == io.EOF {
				// is not error, just client close
				belogs.Debug("waitReceive(): io.EOF, client close: ", conn, err)
				return
			}
			belogs.Error("waitReceive(): Read fail, err ", conn, err)
			return
		}
		if n == 0 {
			continue
		}

		// copy to new []byte
		receiveData := make([]byte, n)
		copy(receiveData, buffer[0:n])
		belogs.Info("waitReceive():conn: ", conn, "  len(receiveData): ", len(receiveData),
			" , will call client tcpClientProcessFunc,  time(s):", time.Now().Sub(start).Seconds())
		err = tc.tcpClientProcessFunc.OnReceive(conn, receiveData)
		belogs.Debug("waitReceive():conn: ", conn, "  len(receiveData): ", len(receiveData), "  time(s):", time.Now().Sub(start).Seconds())
		if err != nil {
			belogs.Error("waitReceive(): fail ,will remove this conn : ", conn, err)
			break
		}
	}
}
