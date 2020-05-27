package tcpserver

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/cpusoft/goutil/convert"
)

func ClientProcess1(conn net.Conn) error {
	defer conn.Close()
	for i := 0; i < 10; i++ {
		buffer := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x00, 0x00, 0x07}
		fmt.Println("client will write i:", i, convert.Bytes2String(buffer))
		conn.Write(buffer)

		buffer1 := make([]byte, 1024)

		n, err := conn.Read(buffer1)
		fmt.Println("client read i:", i, n, convert.Bytes2String(buffer1))
		if err != nil {
			fmt.Println("client read  Read failed:", err)
			return err
		}
		recvByte := buffer1[0:n]
		fmt.Println("client read i:", i, n, convert.Bytes2String(recvByte))
		time.Sleep(time.Duration(10) * time.Second)
	}
	return nil
}

func TestCreateTcpClient(t *testing.T) {

	CreateTcpClient("127.0.0.1:9999", ClientProcess1)

}
