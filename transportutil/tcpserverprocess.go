package transportutil

// tcp server extension interface: need to expand yourself process
type TcpServerProcess interface {
	OnConnectProcess(tcpConn *TcpConn)
	OnReceiveAndSendProcess(tcpConn *TcpConn, receiveData []byte) (nextConnectPolicy int, leftData []byte, err error)
	OnCloseProcess(tcpConn *TcpConn)
}
