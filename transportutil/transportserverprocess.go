package transportutil

// tcp server extension interface: need to expand yourself process
type TransportServerProcess interface {
	OnConnectProcess(transportConn *TransportConn)
	OnReceiveAndSendProcess(transportConn *TransportConn, receiveData []byte) (nextConnectPolicy int, leftData []byte, err error)
	OnCloseProcess(transportConn *TransportConn)
}
