package tcpudputil

import (
	"net"

	belogs "github.com/astaxie/beego/logs"
)

// shoud add defer conn.Close()
type clientProcess func(conn net.Conn) error

// server: **.**.**.**:port
func CreateTcpClient(server string, clientProcess clientProcess) (err error) {
	belogs.Debug("CreateTcpClient():create client, server is  ", server, "   clientProcess:", clientProcess)
	conn, err := net.Dial("tcp4", server)
	if err != nil {
		belogs.Error("CreateTcpClient(): Dial fail: ", server, err)
		return err
	}

	belogs.Debug("CreateTcpClient():create client ok, server is  ", server, "   clientProcess:", clientProcess)
	return clientProcess(conn)
}

// shoud add defer conn.Close()
type serverProcess func(conn net.Conn)

// server: 0.0.0.0:port
func CreateTcpServer(server string, serverProcess serverProcess) (err error) {

	belogs.Debug("CreateTcpServer():create server  ", server, "   serverProcess:", serverProcess)
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

	belogs.Debug("CreateTcpServer(): create server ok, server is ", server, "  will accept client,  serverProcess:", serverProcess)
	for {
		conn, err := listen.AcceptTCP()
		belogs.Info("CreateTcpServer(): Accept remote: ", conn.RemoteAddr().String())
		if err != nil {
			belogs.Error("CreateTcpServer(): Accept remote fail: ", server, conn.RemoteAddr().String(), err)
			continue
		}
		// 每次建立一个连接就放到单独的协程内做处理
		go serverProcess(conn)
	}
}
