package tcpclient

import "net"

type TcpClientProcessFunc interface {
	OnConnectProcess(tcpConn *net.TCPConn)
	OnCloseProcess(tcpConn *net.TCPConn)
	OnReceiveProcess(tcpConn *net.TCPConn, sendData []byte) (nextRwPolicy int, leftData []byte, err error)
}
