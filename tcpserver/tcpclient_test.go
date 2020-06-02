package tcpserver

import (
	"fmt"
	"net"
	"testing"

	"github.com/cpusoft/goutil/convert"
)

type RtrClientResetQuery struct {
}

func (rq *RtrClientResetQuery) ActiveSendAndReceive(conn *net.TCPConn) (err error) {
	buffer := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x00, 0x00, 0x07}
	fmt.Println("RtrClientResetQuery ActiveSendAndReceive will write :", convert.Bytes2String(buffer))
	conn.Write(buffer)

	buffer1 := make([]byte, 1024)

	n, err := conn.Read(buffer1)
	fmt.Println("client read :", n, convert.Bytes2String(buffer1))
	if err != nil {
		fmt.Println("client read  Read failed:", err)
		return err
	}
	recvByte := buffer1[0:n]
	fmt.Println("client read :", n, convert.Bytes2String(recvByte))
	return nil
}

type RtrClientSerialQuery struct {
}

func (sq *RtrClientSerialQuery) ActiveSendAndReceive(conn *net.TCPConn) (err error) {
	buffer := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x00, 0x00, 0x07}
	fmt.Println("RtrClientSerialQuery ActiveSendAndReceive will write :", convert.Bytes2String(buffer))
	conn.Write(buffer)

	buffer1 := make([]byte, 1024)

	n, err := conn.Read(buffer1)
	fmt.Println("client read :", n, convert.Bytes2String(buffer1))
	if err != nil {
		fmt.Println("client read  Read failed:", err)
		return err
	}
	recvByte := buffer1[0:n]
	fmt.Println("client read :", n, convert.Bytes2String(recvByte))
	return nil
}

func TestCreateTcpClient(t *testing.T) {
	rtrClientResetQuery := new(RtrClientResetQuery)
	rtrClientSerialQuery := new(RtrClientSerialQuery)

	tcpClientProcessFuncs := make(map[string]TcpClientProcessFunc, 0)
	tcpClientProcessFuncs["resetquery"] = rtrClientResetQuery
	tcpClientProcessFuncs["serialquery"] = rtrClientSerialQuery
	fmt.Println("client read :", tcpClientProcessFuncs)

	//CreateTcpClient("127.0.0.1:9999", ClientProcess1)
	tc := NewTcpClient(tcpClientProcessFuncs)
	fmt.Println("tc:", tc)
	err := tc.Start("127.0.0.1:9999")
	fmt.Println("tc:", tc, err)

	tc.CallProcessFunc("resetquery")
	tc.CallProcessFunc("serialquery")
	tc.CallExit()
}
