package main

// tcp server extension interface: need to expand yourself process
type TcpTlsServerProcessFunc interface {
	OnConnectProcess(tcpTlsConn *TcpTlsConn)
	ReceiveAndSendProcess(tcpTlsConn *TcpTlsConn, receiveData []byte) (nextConnectPolicy int, leftData []byte, err error)
	OnCloseProcess(tcpTlsConn *TcpTlsConn)
	ActiveSendProcess(tcpTlsConn *TcpTlsConn, sendData []byte) (err error)
}
