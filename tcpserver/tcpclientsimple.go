package tcpserver

import (
	"errors"
	"net"
	"time"

	belogs "github.com/astaxie/beego/logs"
)

// server: **.**.**.**:port
// sendData: will send
// receiveData: should smaller than 2048
func ClientSendAndReceive(server string, sendData []byte) (receiveData []byte, err error) {
	start := time.Now()
	belogs.Debug("ClientSendAndReceive(): client will send to server : ", server, "   len(sendData):", len(sendData))
	if server == "" || sendData == nil || len(sendData) == 0 {
		return nil, errors.New("server or sendData or receiveData is error")
	}
	// resolve tcp addr
	tcpServer, err := net.ResolveTCPAddr("tcp", server)
	if err != nil {
		belogs.Error("ClientSendAndReceive(): ResolveTCPAddr fail: ", server, err)
		return nil, err
	}

	// connect
	conn, err := net.DialTCP("tcp4", nil, tcpServer)
	if err != nil {
		belogs.Error("ClientSendAndReceive(): Dial fail: ", server, tcpServer, err)
		return nil, err
	}
	defer conn.Close()

	// write
	_, err = conn.Write(sendData)
	if err != nil {
		belogs.Error("ClientSendAndReceive(): Write fail: ", server, tcpServer, err)
		return nil, err
	}

	// read
	receiveData = make([]byte, 2048)
	n, err := conn.Read(receiveData)
	if err != nil {
		belogs.Error("ClientSendAndReceive(): Read fail: ", server, tcpServer, err)
		return nil, err
	}
	belogs.Info("ClientSendAndReceive():send to server ", server, "  sendData:", len(sendData), " len(receiveData):", n,
		"   time(s):", time.Now().Sub(start).Seconds())
	return receiveData[:n], nil

}
