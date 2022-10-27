package transportutil

import (
	"crypto/tls"
	"errors"
	"net"
	"time"

	"github.com/cpusoft/goutil/belogs"
)

// https://stackoverflow.com/questions/66755407/cancelling-a-net-listener-via-context-in-golang

type TcpListener struct {
	// tcp/tls
	connType    string
	tcpLisenter *net.TCPListener
	tlsListener net.Listener
}

func NewFromTcpListener(tcpListener *net.TCPListener) (c *TcpListener, err error) {
	c = &TcpListener{}
	c.connType = "tcp"
	c.tcpLisenter = tcpListener
	return c, nil
}
func NewFromTlsListener(tlsListener net.Listener) (c *TcpListener, err error) {
	c = &TcpListener{}
	c.connType = "tls"
	c.tlsListener = tlsListener
	return c, nil
}

func (c *TcpListener) Close() error {
	if c.connType == "tcp" && c.tcpLisenter != nil {
		return c.tcpLisenter.Close()
	}
	if c.connType == "tls" && c.tlsListener != nil {
		return c.tlsListener.Close()
	}

	return errors.New("not found connType " + c.connType + " for Close")
}

func (c *TcpListener) Accept() (tcpConn *TcpConn, err error) {
	if c.connType == "tcp" && c.tcpLisenter != nil {
		conn, err := c.tcpLisenter.AcceptTCP()
		if err != nil {
			belogs.Error("Accept(): TcpListener Accept TCP tcpConn remote fail: ", err)
			return nil, err
		}
		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(time.Second * 300)
		tcpConn = NewFromTcpConn(conn)
		belogs.Info("Accept(): TcpListener Accept TCP tcpConn remote: ", tcpConn.RemoteAddr().String())
		return tcpConn, nil
	}
	if c.connType == "tls" && c.tlsListener != nil {
		conn, err := c.tlsListener.Accept()
		if err != nil {
			belogs.Error("Accept(): TcpListener  Accept Tls remote fail: ", err)
			return nil, err
		}

		tlsConn, ok := conn.(*tls.Conn)
		if !ok {
			belogs.Error("Accept(): TcpListener  Accept Tls remote , conn cannot conver to tlsConn: ", conn.RemoteAddr().String(), err)
			return nil, err
		}
		tcpConn = NewFromTlsConn(tlsConn)
		belogs.Info("Accept(): TcpListener Accept Tls tcpConn remote: ", tcpConn.RemoteAddr().String())
		return tcpConn, nil
	}

	return nil, errors.New("not support connType " + c.connType + " for Accept")
}
