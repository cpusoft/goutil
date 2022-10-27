package transportutil

type UdpClientProcess interface {
	OnReceiveProcess(udpConn *UdpConn, receiveData []byte) (err error)
}
