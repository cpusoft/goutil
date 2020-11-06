package tcpserver

import (
	"io"
	"net"
	"sync"
	"time"

	belogs "github.com/astaxie/beego/logs"
)

type TcpServer struct {
	tcpConns      []*net.TCPConn
	tcpConnsMutex sync.RWMutex

	tcpServerProcessFunc TcpServerProcessFunc
}

// server: 0.0.0.0:port
func NewTcpServer(tcpServerProcessFunc TcpServerProcessFunc) (ts *TcpServer) {

	belogs.Debug("NewTcpServer():tcpProcessFunc:", tcpServerProcessFunc)
	ts = &TcpServer{}
	ts.tcpConns = make([]*net.TCPConn, 0, 16)
	ts.tcpServerProcessFunc = tcpServerProcessFunc
	belogs.Debug("NewTcpServer():ts:", ts, "   ts.tcpConnsMutex:", ts.tcpConnsMutex)
	return ts
}

// server: 0.0.0.0:port
func (ts *TcpServer) Start(server string) (err error) {
	tcpServer, err := net.ResolveTCPAddr("tcp", server)
	if err != nil {
		belogs.Error("Start(): ResolveTCPAddr fail: ", server, err)
		return err
	}

	listen, err := net.ListenTCP("tcp", tcpServer)
	if err != nil {
		belogs.Error("Start(): ListenTCP fail: ", server, err)
		return err
	}
	defer listen.Close()
	belogs.Debug("Start(): create server ok, server is ", server, "  will accept client")

	for {
		conn, err := listen.AcceptTCP()
		belogs.Info("Start(): Accept remote: ", conn.RemoteAddr().String())
		if err != nil {
			belogs.Error("Start(): Accept remote fail: ", server, conn.RemoteAddr().String(), err)
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

type TcpServerProcessFunc interface {
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
	ts.tcpServerProcessFunc.OnConnect(conn)
	belogs.Info("OnConnect():add conn: ", conn, "   len(tcpConns): ", len(ts.tcpConns), "   time(s):", time.Now().Sub(start).Seconds())
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
				belogs.Debug("ReceiveAndSend():server  Read io.EOF, client close: ", conn, err)
				return
			}
			belogs.Error("ReceiveAndSend():server  Read fail, err ", conn, err)
			return
		}
		if n == 0 {
			continue
		}

		// call process func OnReceiveAndSend
		// copy to new []byte
		receiveData := make([]byte, n)
		copy(receiveData, buffer[0:n])
		belogs.Debug("ReceiveAndSend():server conn: ", conn, "  len(receiveData):", len(receiveData),
			" , will call process func: OnReceiveAndSend,  time(s):", time.Now().Sub(start).Seconds())
		err = ts.tcpServerProcessFunc.OnReceiveAndSend(conn, receiveData)
		belogs.Debug("ReceiveAndSend():server conn: ", conn, "  len(receiveData): ", len(receiveData), "  time(s):", time.Now().Sub(start).Seconds())
		if err != nil {
			belogs.Error("OnReceiveAndSend():server fail ,will remove this conn : ", conn, err)
			break
		}
	}
}

func (ts *TcpServer) OnClose(conn *net.TCPConn) {
	// close in the end
	defer conn.Close()
	start := time.Now()

	// call process func OnClose
	belogs.Debug("OnClose():server,conn: ", conn, "   call process func: OnClose ")
	ts.tcpServerProcessFunc.OnClose(conn)

	// remove conn from tcpConns
	ts.tcpConnsMutex.Lock()
	defer ts.tcpConnsMutex.Unlock()
	belogs.Debug("OnClose():server,conn: ", conn, "   old len(tcpConns): ", len(ts.tcpConns))
	newTcpConns := make([]*net.TCPConn, 0, len(ts.tcpConns))
	for i := range ts.tcpConns {
		if ts.tcpConns[i] != conn {
			newTcpConns = append(newTcpConns, ts.tcpConns[i])
		}
	}
	ts.tcpConns = newTcpConns
	belogs.Info("OnClose():server,new len(tcpConns): ", len(ts.tcpConns), "  time(s):", time.Now().Sub(start).Seconds())
}

func (ts *TcpServer) ActiveSend(sendData []byte) (err error) {
	ts.tcpConnsMutex.RLock()
	defer ts.tcpConnsMutex.RUnlock()
	start := time.Now()

	belogs.Debug("ActiveSend():server,len(sendData):", len(sendData), "   len(tcpConns): ", len(ts.tcpConns))
	for i := range ts.tcpConns {
		belogs.Debug("ActiveSend(): client: ", i, "    ts.tcpConns[i]:", ts.tcpConns[i], "   call process func: ActiveSend ")
		err = ts.tcpServerProcessFunc.ActiveSend(ts.tcpConns[i], sendData)
		if err != nil {
			// just logs, not return or break
			belogs.Error("ActiveSend(): fail, client: ", i, "    ts.tcpConns[i]:", ts.tcpConns[i], err)
		}
	}
	belogs.Info("ActiveSend(): send to all clients ok,  len(sendData):", len(sendData), "   len(tcpConns): ", len(ts.tcpConns),
		"  time(s):", time.Now().Sub(start).Seconds())
	return
}
