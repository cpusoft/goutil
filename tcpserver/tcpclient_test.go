package tcpserver

import (
	"fmt"
	"net"
	"testing"

	"github.com/cpusoft/goutil/convert"
)

type ClientProcessFunc struct {
}

func (cp *ClientProcessFunc) ActiveSend(conn *net.TCPConn, tcpClientProcessChan string) (err error) {

	fmt.Println("ActiveSend ActiveSendAndReceive will write tcpClientProcessChan:", tcpClientProcessChan)
	buffer := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x00, 0x00, 0x07}
	_, err = conn.Write(buffer)
	if err != nil {
		fmt.Println("ActiveSend():  conn.Write() fail,  ", err)
		return err
	}
	return nil
}

func (sq *ClientProcessFunc) OnReceive(conn *net.TCPConn, receiveData []byte) (err error) {

	fmt.Println("OnReceive :", conn, convert.Bytes2String(receiveData))
	return nil
}

func TestCreateTcpClient(t *testing.T) {
	clientProcessFunc := new(ClientProcessFunc)

	//CreateTcpClient("127.0.0.1:9999", ClientProcess1)
	tc := NewTcpClient(clientProcessFunc)
	err := tc.Start("192.168.83.139:9999")
	fmt.Println("tc:", tc, err)
	if err != nil {
		return
	}
	fmt.Println("CallProcessFunc: resetquery")
	tc.CallProcessFunc("resetquery")
	fmt.Println("CallProcessFunc: serialquery")
	tc.CallProcessFunc("serialquery")
	fmt.Println("CallStop:")
	tc.CallStop()
}
