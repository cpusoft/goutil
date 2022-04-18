package tcpserver

import "net"

// tcp server extension interface: need to expand yourself process
type TcpServerProcessFunc interface {
	OnConnectProcess(tcpConn *net.TCPConn)
	ReceiveAndSendProcess(tcpConn *net.TCPConn, receiveData []byte) (nextConnectPolicy int, leftData []byte, err error)
	OnCloseProcess(tcpConn *net.TCPConn)
	ActiveSendProcess(tcpConn *net.TCPConn, sendData []byte) (err error)
}
