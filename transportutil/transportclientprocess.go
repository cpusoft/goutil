package transportutil

type TransportClientProcess interface {
	OnConnectProcess(transportConn *TransportConn)
	OnCloseProcess(transportConn *TransportConn)
	OnReceiveProcess(transportConn *TransportConn, receiveData []byte) (nextRwPolicy int, leftData []byte, err error)
}
