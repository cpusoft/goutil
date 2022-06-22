package tcptlsutil

import (
	"crypto/tls"
	"errors"
	"net"
	"time"
)

type TcpTlsConn struct {
	// tcp: true
	// tls: false
	isTcpConn         bool
	tcpConn           *net.TCPConn
	tlsConn           *tls.Conn
	isConnected       bool
	nextConnectPolicy int
}

func NewFromTcpConn(tcpConn *net.TCPConn) (c *TcpTlsConn) {
	c = &TcpTlsConn{}
	c.tcpConn = tcpConn
	c.isTcpConn = true
	c.isConnected = true
	c.nextConnectPolicy = NEXT_CONNECT_POLICY_KEEP
	return c
}

func NewFromTlsConn(tlsConn *tls.Conn) (c *TcpTlsConn) {
	c = &TcpTlsConn{}
	c.tlsConn = tlsConn
	c.isTcpConn = false
	c.isConnected = true
	c.nextConnectPolicy = NEXT_CONNECT_POLICY_KEEP
	return c
}

func (c *TcpTlsConn) RemoteAddr() net.Addr {
	if c.isTcpConn && c.tcpConn != nil {
		return c.tcpConn.RemoteAddr()
	}
	if !c.isTcpConn && c.tlsConn != nil {
		return c.tlsConn.RemoteAddr()
	}
	return nil
}

func (c *TcpTlsConn) LocalAddr() net.Addr {
	if c.isTcpConn && c.tcpConn != nil {
		return c.tcpConn.LocalAddr()
	}
	if !c.isTcpConn && c.tlsConn != nil {
		return c.tlsConn.LocalAddr()
	}
	return nil
}

func (c *TcpTlsConn) Write(b []byte) (n int, err error) {
	if c.isTcpConn && c.tcpConn != nil && c.isConnected {
		return c.tcpConn.Write(b)
	}
	if !c.isTcpConn && c.tlsConn != nil && c.isConnected {
		return c.tlsConn.Write(b)
	}
	return -1, errors.New("is not conn")
}

func (c *TcpTlsConn) Read(b []byte) (n int, err error) {
	if c.isTcpConn && c.tcpConn != nil && c.isConnected {
		return c.tcpConn.Read(b)
	}
	if !c.isTcpConn && c.tlsConn != nil && c.isConnected {
		return c.tlsConn.Read(b)
	}
	return -1, errors.New("is not conn")
}

func (c *TcpTlsConn) Close() (err error) {
	if c.isTcpConn && c.tcpConn != nil {
		c.isConnected = false
		return c.tcpConn.Close()
	}
	if !c.isTcpConn && c.tlsConn != nil {
		c.isConnected = false
		return c.tlsConn.Close()
	}
	return errors.New("is not conn")
}

func (c *TcpTlsConn) SetDeadline(t time.Time) error {
	if c.isTcpConn && c.tcpConn != nil {
		return c.tcpConn.SetDeadline(t)
	}
	if !c.isTcpConn && c.tlsConn != nil {
		return c.tlsConn.SetDeadline(t)
	}
	return errors.New("is not conn")
}

func (c *TcpTlsConn) IsConnected() bool {
	return c.isConnected
}
