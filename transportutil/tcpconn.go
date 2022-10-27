package transportutil

import (
	"crypto/tls"
	"errors"
	"net"
	"time"
)

type TcpConn struct {
	// tcp/tls
	connType string
	tcpConn  *net.TCPConn
	tlsConn  *tls.Conn
	udpConn  *net.UDPConn

	isConnected       bool
	nextConnectPolicy int
}

func NewFromTcpConn(tcpConn *net.TCPConn) (c *TcpConn) {
	c = &TcpConn{}
	c.tcpConn = tcpConn
	c.connType = "tcp"
	c.isConnected = true
	c.nextConnectPolicy = NEXT_CONNECT_POLICY_KEEP
	return c
}

func NewFromTlsConn(tlsConn *tls.Conn) (c *TcpConn) {
	c = &TcpConn{}
	c.tlsConn = tlsConn
	c.connType = "tls"
	c.isConnected = true
	c.nextConnectPolicy = NEXT_CONNECT_POLICY_KEEP
	return c
}

func (c *TcpConn) RemoteAddr() net.Addr {
	if c.connType == "tcp" && c.tcpConn != nil {
		return c.tcpConn.RemoteAddr()
	}
	if c.connType == "tls" && c.tlsConn != nil {
		return c.tlsConn.RemoteAddr()
	}

	return nil
}

func (c *TcpConn) LocalAddr() net.Addr {
	if c.connType == "tcp" && c.tcpConn != nil {
		return c.tcpConn.LocalAddr()
	}
	if c.connType == "tls" && c.tlsConn != nil {
		return c.tlsConn.LocalAddr()
	}

	return nil
}

func (c *TcpConn) Write(b []byte) (n int, err error) {
	if c.connType == "tcp" && c.tcpConn != nil && c.isConnected {
		return c.tcpConn.Write(b)
	}
	if c.connType == "tls" && c.tlsConn != nil && c.isConnected {
		return c.tlsConn.Write(b)
	}

	return -1, errors.New("is not conn")
}

func (c *TcpConn) Read(b []byte) (n int, err error) {
	if c.connType == "tcp" && c.tcpConn != nil && c.isConnected {
		return c.tcpConn.Read(b)
	}
	if c.connType == "tls" && c.tlsConn != nil && c.isConnected {
		return c.tlsConn.Read(b)
	}

	return -1, errors.New("is not conn")
}

func (c *TcpConn) Close() (err error) {
	if c.connType == "tcp" && c.tcpConn != nil {
		c.isConnected = false
		return c.tcpConn.Close()
	}
	if c.connType == "tls" && c.tlsConn != nil {
		c.isConnected = false
		return c.tlsConn.Close()
	}

	return errors.New("is not conn")
}

func (c *TcpConn) SetDeadline(t time.Time) error {
	if c.connType == "tcp" && c.tcpConn != nil {
		return c.tcpConn.SetDeadline(t)
	}
	if c.connType == "tls" && c.tlsConn != nil {
		return c.tlsConn.SetDeadline(t)
	}

	return errors.New("is not conn")
}

func (c *TcpConn) IsConnected() bool {
	return c.isConnected
}
