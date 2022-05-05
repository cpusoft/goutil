package tcptlsutil

type TcpTlsClientProcess interface {
	OnConnectProcess(tcpTlsConn *TcpTlsConn)
	OnCloseProcess(tcpTlsConn *TcpTlsConn)
	OnReceiveProcess(tcpTlsConn *TcpTlsConn, receiveData []byte) (nextRwPolicy int, leftData []byte, err error)
}
