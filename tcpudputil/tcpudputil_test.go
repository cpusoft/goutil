package tcpudputil

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/cpusoft/goutil/convert"
)

func ServerProcess1(conn net.Conn) {
	defer conn.Close()

	fmt.Println("server get new client: ", conn)
	buffer := make([]byte, 1024)
	//状态机处理数据
	for {
		n, err := conn.Read(buffer)
		fmt.Println("server read: Read n: ", n)
		if err != nil {
			if err != io.EOF {
				fmt.Println("serverProcess(): Read fail: ", err)
			} else {
				fmt.Println("serverProcess(): Read client close: ", err)
			}
			return
		}
		if n == 0 {
			continue
		}

		recvByte := buffer[0:n]
		fmt.Println("server read:", convert.Bytes2String(recvByte))

		buffer1 := []byte{0x11, 0x12, 0x00, 0x00, 0x13, 0x14, 0x00, 0x00, 0x00}
		n, err = conn.Write(buffer1)
		fmt.Println("server Write:", convert.Bytes2String(buffer1), n, err)
		if err != nil {
			fmt.Println("server Write failed:", n, err)
			return
		}
		return
	}
}

func ClientProcess1(conn net.Conn) error {
	defer conn.Close()

	buffer := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x00, 0x00, 0x07}
	fmt.Println("client will write:", convert.Bytes2String(buffer))
	conn.Write(buffer)

	buffer1 := make([]byte, 1024)

	n, err := conn.Read(buffer1)
	fmt.Println("client read:", n, convert.Bytes2String(buffer1))
	if err != nil {
		fmt.Println("client read  Read failed:", err)
		return err
	}
	recvByte := buffer1[0:n]
	fmt.Println("client read:", n, convert.Bytes2String(recvByte))
	return nil
}
func TestCreateTcpClient(t *testing.T) {
	CreateTcpClient("127.0.0.1:9999", ClientProcess1)

	time.Sleep(time.Duration(10) * time.Second)
}

func TestCreateTcpServer(t *testing.T) {
	CreateTcpServer("0.0.0.0:9999", ServerProcess1)
}
