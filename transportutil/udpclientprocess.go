package transportutil

type UdpClientProcess interface {
	OnReceiveProcess(udpConn *UdpConn, receiveData []byte) (connToBusinessMsg *ConnToBusinessMsg, err error)
}
