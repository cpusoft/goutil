package tcpserver

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/cpusoft/goutil/convert"
)

type RtrServer struct {
}

func (rs *RtrServer) OnConnect(conn *net.TCPConn) {

}
func (rs *RtrServer) OnReceiveAndSend(conn *net.TCPConn, receiveData []byte) (err error) {
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
func (rs *RtrServer) OnClose(conn *net.TCPConn) {

}
func (rs *RtrServer) ActiveSend(conn *net.TCPConn, sendData []byte) (err error) {
	return
}

func TestCreateTcpServer(t *testing.T) {
	rtrServer := new(RtrServer)
	ts := NewTcpServer(rtrServer)
	ts.CreateTcpServer("0.0.0.0:9999")
}
