package tcptlsutil

// tcp server extension interface: need to expand yourself process
type TcpTlsServerProcess interface {
	OnConnectProcess(tcpTlsConn *TcpTlsConn)
	OnReceiveAndSendProcess(tcpTlsConn *TcpTlsConn, receiveData []byte) (nextConnectPolicy int, leftData []byte, err error)
	OnCloseProcess(tcpTlsConn *TcpTlsConn)
}
