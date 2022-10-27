package transportutil

type TcpClientProcess interface {
	OnConnectProcess(tcpConn *TcpConn)
	OnCloseProcess(tcpConn *TcpConn)
	OnReceiveProcess(tcpConn *TcpConn, receiveData []byte) (nextRwPolicy int, leftData []byte, err error)
}
