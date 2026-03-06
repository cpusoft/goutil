package tcpserver

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/cpusoft/goutil/convert"
)

type Server1ProcessFunc struct {
}

func (spf *Server1ProcessFunc) OnConnect(conn *net.TCPConn) (err error) {
	fmt.Println("Client connected:", conn.RemoteAddr())
	return nil
}

func (spf *Server1ProcessFunc) OnReceiveAndSend(conn *net.TCPConn, receiveData []byte) (err error) {
	fmt.Println("server read:", convert.Bytes2String(receiveData))

	buffer1 := []byte{0x11, 0x12, 0x00, 0x00, 0x13, 0x14, 0x00, 0x00, 0x00}
	// 设置写超时
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Write(buffer1)
	fmt.Println("server Write :", n)
	if err != nil {
		fmt.Println("server Write failed:", n, err)
		return err
	}
	return nil
}

func (spf *Server1ProcessFunc) OnClose(conn *net.TCPConn) {
	fmt.Println("Client disconnected:", conn.RemoteAddr())
}

func (spf *Server1ProcessFunc) ActiveSend(conn *net.TCPConn, sendData []byte) (err error) {
	n, err := conn.Write(sendData)
	fmt.Printf("ActiveSend to %s: %d bytes sent\n", conn.RemoteAddr(), n)
	return err
}

func TestCreateTcpServer(t *testing.T) {
	serverProcessFunc := new(Server1ProcessFunc)
	// 创建服务端，设置读写超时
	ts := NewTcpServer(serverProcessFunc, WithReadWriteTimeout(30*time.Second, 10*time.Second))

	// 监听退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived stop signal, shutting down server...")
		ts.Stop()
	}()

	// 启动服务
	fmt.Println("Starting TCP server on 0.0.0.0:9999")
	if err := ts.Start("0.0.0.0:9999"); err != nil {
		t.Fatal("Server start failed:", err)
	}
	fmt.Println("Server stopped")
}
