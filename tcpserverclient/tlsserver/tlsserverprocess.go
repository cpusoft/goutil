package tlsserver

import "crypto/tls"

type TlsServerProcessFunc interface {
	OnConnectProcess(tlsConn *tls.Conn) (err error)
	ReceiveAndSendProcess(tlsConn *tls.Conn, receiveData []byte) (nextConnectPolicy int, leftData []byte, err error)
	OnCloseProcess(tlsConn *tls.Conn)
	ActiveSendProcess(tlsConn *tls.Conn, sendData []byte) (err error)
}
