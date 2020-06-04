package tcpserver

import (
	"fmt"
	"net"
	"testing"

	"github.com/cpusoft/goutil/convert"
)

type ServerProcessFunc struct {
}

func (spf *ServerProcessFunc) OnConnect(conn *net.TCPConn) {

}
func (spf *ServerProcessFunc) OnReceiveAndSend(conn *net.TCPConn, receiveData []byte) (err error) {
	fmt.Println("server read:", convert.Bytes2String(receiveData))

	buffer1 := []byte{0x11, 0x12, 0x00, 0x00, 0x13, 0x14, 0x00, 0x00, 0x00}
	n, err := conn.Write(buffer1)
	fmt.Println("server Write :", n)
	if err != nil {
		fmt.Println("server Write failed:", n, err)
		return err
	}
	return nil
}
func (spf *ServerProcessFunc) OnClose(conn *net.TCPConn) {

}
func (spf *ServerProcessFunc) ActiveSend(conn *net.TCPConn, sendData []byte) (err error) {
	return
}

func TestCreateTcpServer(t *testing.T) {
	serverProcessFunc := new(ServerProcessFunc)
	ts := NewTcpServer(serverProcessFunc)
	ts.Start("0.0.0.0:9999")
}
