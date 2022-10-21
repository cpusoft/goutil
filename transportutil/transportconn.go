package transportutil

import (
	"crypto/tls"
	"errors"
	"net"
	"time"
)

type TransportConn struct {
	// tcp/tls/udp
	connType string
	tcpConn  *net.TCPConn
	tlsConn  *tls.Conn
	udpConn  *net.UDPConn

	isConnected       bool
	nextConnectPolicy int

	// udp
	udpAddr *net.UDPAddr
}

func NewFromTcpConn(tcpConn *net.TCPConn) (c *TransportConn) {
	c = &TransportConn{}
	c.tcpConn = tcpConn
	c.connType = "tcp"
	c.isConnected = true
	c.nextConnectPolicy = NEXT_CONNECT_POLICY_KEEP
	return c
}

func NewFromTlsConn(tlsConn *tls.Conn) (c *TransportConn) {
	c = &TransportConn{}
	c.tlsConn = tlsConn
	c.connType = "tls"
	c.isConnected = true
	c.nextConnectPolicy = NEXT_CONNECT_POLICY_KEEP
	return c
}

// when server, should get udpAddr from client
// when client, udpAddr is nil
func NewFromUdpConn(udpConn *net.UDPConn, udpAddr *net.UDPAddr) (c *TransportConn) {
	c = &TransportConn{}
	c.udpConn = udpConn
	c.connType = "udp"
	c.isConnected = true
	c.nextConnectPolicy = NEXT_CONNECT_POLICY_KEEP
	c.udpAddr = udpAddr
	return c
}

func (c *TransportConn) RemoteAddr() net.Addr {
	if c.connType == "tcp" && c.tcpConn != nil {
		return c.tcpConn.RemoteAddr()
	}
	if c.connType == "tls" && c.tlsConn != nil {
		return c.tlsConn.RemoteAddr()
	}
	if c.connType == "udp" && c.udpConn != nil {
		return c.udpConn.RemoteAddr()
	}
	return nil
}

func (c *TransportConn) LocalAddr() net.Addr {
	if c.connType == "tcp" && c.tcpConn != nil {
		return c.tcpConn.LocalAddr()
	}
	if c.connType == "tls" && c.tlsConn != nil {
		return c.tlsConn.LocalAddr()
	}
	if c.connType == "udp" && c.udpConn != nil {
		return c.udpConn.LocalAddr()
	}
	return nil
}

func (c *TransportConn) Write(b []byte) (n int, err error) {
	if c.connType == "tcp" && c.tcpConn != nil && c.isConnected {
		return c.tcpConn.Write(b)
	}
	if c.connType == "tls" && c.tlsConn != nil && c.isConnected {
		return c.tlsConn.Write(b)
	}
	if c.connType == "udp" && c.udpConn != nil && c.isConnected {
		if c.udpAddr == nil {
			return c.udpConn.Write(b)
		} else {
			return c.udpConn.WriteToUDP(b, c.udpAddr)
		}
	}
	return -1, errors.New("is not conn")
}

func (c *TransportConn) Read(b []byte) (n int, err error) {
	if c.connType == "tcp" && c.tcpConn != nil && c.isConnected {
		return c.tcpConn.Read(b)
	}
	if c.connType == "tls" && c.tlsConn != nil && c.isConnected {
		return c.tlsConn.Read(b)
	}
	if c.connType == "udp" && c.udpConn != nil && c.isConnected {
		return c.udpConn.Read(b)
	}
	return -1, errors.New("is not conn")
}

func (c *TransportConn) Close() (err error) {
	if c.connType == "tcp" && c.tcpConn != nil {
		c.isConnected = false
		return c.tcpConn.Close()
	}
	if c.connType == "tls" && c.tlsConn != nil {
		c.isConnected = false
		return c.tlsConn.Close()
	}
	if c.connType == "udp" && c.udpConn != nil {
		c.isConnected = false
		return c.udpConn.Close()
	}
	return errors.New("is not conn")
}

func (c *TransportConn) SetDeadline(t time.Time) error {
	if c.connType == "tcp" && c.tcpConn != nil {
		return c.tcpConn.SetDeadline(t)
	}
	if c.connType == "tls" && c.tlsConn != nil {
		return c.tlsConn.SetDeadline(t)
	}
	if c.connType == "udp" && c.udpConn != nil {
		return c.udpConn.SetDeadline(t)
	}
	return errors.New("is not conn")
}

func (c *TransportConn) IsConnected() bool {
	return c.isConnected
}
