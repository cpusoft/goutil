package tcpserver

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
