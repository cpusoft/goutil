package tcpserver

import (
	"io"
	"net"
	"sync"
	"time"

	belogs "github.com/astaxie/beego/logs"
)

var tcpServer = &TcpServer{}

type TcpServer struct {
	tcpConns      []*net.TCPConn
	tcpConnsMutex *sync.RWMutex

	tcpProcessFunc TcpProcessFunc
}

// server: 0.0.0.0:port
func NewTcpServer(tcpProcessFunc TcpProcessFunc) (ts *TcpServer) {

	belogs.Debug("NewTcpServer():tcpProcessFunc:", tcpProcessFunc)
	ts = new(TcpServer)
	ts.tcpConns = make([]*net.TCPConn, 0, 16)
	ts.tcpConnsMutex = new(sync.RWMutex)
	ts.tcpProcessFunc = tcpProcessFunc
	return ts
}

// server: 0.0.0.0:port
func (ts *TcpServer) Start(server string) (err error) {
	tcpServer, err := net.ResolveTCPAddr("tcp", server)
	if err != nil {
		belogs.Error("CreateTcpServer(): ResolveTCPAddr fail: ", server, err)
		return err
	}

	listen, err := net.ListenTCP("tcp", tcpServer)
	if err != nil {
		belogs.Error("CreateTcpServer(): ListenTCP fail: ", server, err)
		return err
	}
	defer listen.Close()

	belogs.Debug("CreateTcpServer(): create server ok, server is ", server, "  will accept client")

	for {
		conn, err := listen.AcceptTCP()
		belogs.Info("CreateTcpServer(): Accept remote: ", conn.RemoteAddr().String())
		if err != nil {
			belogs.Error("CreateTcpServer(): Accept remote fail: ", server, conn.RemoteAddr().String(), err)
			continue
		}
		if conn == nil {
			continue
		}

		ts.OnConnect(conn)
		// call func to process conn
		go ts.ReceiveAndSend(conn)

	}
	return nil
}

type TcpProcessFunc interface {
	OnConnect(conn *net.TCPConn)
	OnReceiveAndSend(conn *net.TCPConn, receiveData []byte) (err error)
	OnClose(conn *net.TCPConn)
	ActiveSend(conn *net.TCPConn, sendData []byte) (err error)
}

func (ts *TcpServer) OnConnect(conn *net.TCPConn) {
	start := time.Now()
	belogs.Debug("OnConnect():  new conn: ", conn)

	// add new conn to tcpconns
	ts.tcpConnsMutex.Lock()
	defer ts.tcpConnsMutex.Unlock()
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(time.Second * 300)
	ts.tcpConns = append(ts.tcpConns, conn)
	belogs.Debug("OnConnect():conn: ", conn, ", new len(tcpConns): ", len(ts.tcpConns))

	// call process func OnConnect
	belogs.Debug("OnConnect():conn: ", conn, "   call process func: OnConnect ")
	ts.tcpProcessFunc.OnConnect(conn)
	belogs.Debug("OnConnect():add conn: ", conn, "   time(s):", time.Now().Sub(start).Seconds())
}

func (ts *TcpServer) ReceiveAndSend(conn *net.TCPConn) {

	defer ts.OnClose(conn)

	// one packet
	buffer := make([]byte, 2048)
	// wait for new packet to read
	for {
		n, err := conn.Read(buffer)
		start := time.Now()
		belogs.Debug("ReceiveAndSend():server read: Read n: ", conn, n)
		if err != nil {
			if err == io.EOF {
				// is not error, just client close
				belogs.Debug("ReceiveAndSend(): io.EOF, client close: ", conn, err)
				return
			}
			belogs.Error("ReceiveAndSend(): Read fail, err ", conn, err)
			return
		}
		if n == 0 {
			continue
		}

		// call process func OnReceiveAndSend
		recvByte := buffer[0:n]
		belogs.Debug("ReceiveAndSend():conn: ", conn, "   call process func: ReceiveAndSend ")
		err = ts.tcpProcessFunc.OnReceiveAndSend(conn, recvByte)
		belogs.Debug("ReceiveAndSend():conn: ", conn, "  len(recvByte): ", len(recvByte), "  time(s):", time.Now().Sub(start).Seconds())
		if err != nil {
			belogs.Error("OnReceiveAndSend(): fail ,will remove this conn : ", conn, err)
			break
		}
	}
}

func (ts *TcpServer) OnClose(conn *net.TCPConn) {
	// close in the end
	defer conn.Close()
	start := time.Now()

	// call process func OnClose
	belogs.Debug("OnClose():conn: ", conn, "   call process func: OnClose ")
	ts.tcpProcessFunc.OnClose(conn)

	// remove conn from tcpConns
	ts.tcpConnsMutex.Lock()
	defer ts.tcpConnsMutex.Unlock()
	belogs.Debug("OnClose():conn: ", conn, "   old len(tcpConns): ", len(ts.tcpConns))
	newTcpConns := make([]*net.TCPConn, 0, len(ts.tcpConns))
	for i := range ts.tcpConns {
		if ts.tcpConns[i] != conn {
			newTcpConns = append(newTcpConns, ts.tcpConns[i])
		}
	}
	ts.tcpConns = newTcpConns
	belogs.Debug("OnClose():new len(tcpConns): ", len(ts.tcpConns), "  time(s):", time.Now().Sub(start).Seconds())
}

func (ts *TcpServer) ActiveSend(sendData []byte) (err error) {
	ts.tcpConnsMutex.RLock()
	defer ts.tcpConnsMutex.RUnlock()
	start := time.Now()

	belogs.Debug("ActiveSend():len(sendData):", len(sendData), "   len(tcpConns): ", len(ts.tcpConns))
	for i := range ts.tcpConns {
		belogs.Debug("ActiveSend():i: ", "    ts.tcpConns[i]:", i, ts.tcpConns[i], "   call process func: ActiveSend ")
		err = ts.tcpProcessFunc.ActiveSend(ts.tcpConns[i], sendData)
		if err != nil {
			// just logs, not return or break
			belogs.Error("ActiveSend():ActiveSend err: ", i, "    ts.tcpConns[i]:", ts.tcpConns[i], err)
		}
	}
	belogs.Debug("ActiveSend():end len(sendData):", len(sendData), "   len(tcpConns): ", len(ts.tcpConns),
		"  time(s):", time.Now().Sub(start).Seconds())
	return
}
