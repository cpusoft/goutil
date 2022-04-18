package tlsclient

import "crypto/tls"

type TlsClientProcessFunc interface {
	OnConnectProcess(tlsConn *tls.Conn)
	OnCloseProcess(tlsConn *tls.Conn)
	OnReceiveProcess(tlsConn *tls.Conn, sendData []byte) (nextRwPolicy int, leftData []byte, err error)
}
