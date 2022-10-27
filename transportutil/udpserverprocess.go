package transportutil

import "net"

// tcp server extension interface: need to expand yourself process
type UdpServerProcess interface {
	//	OnConnectProcess(tcpConn *TcpConn)
	OnReceiveAndSendProcess(udpConn *UdpConn, clientUdpAddr *net.UDPAddr, receiveData []byte) (err error)
	//	OnCloseProcess(tcpConn *TcpConn)
}
