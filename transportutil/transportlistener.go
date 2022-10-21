package transportutil

import (
	"crypto/tls"
	"errors"
	"net"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/transportutil/udpmock"
)

// https://stackoverflow.com/questions/66755407/cancelling-a-net-listener-via-context-in-golang

type TransportListener struct {
	// tcp/tls
	connType    string
	tcpLisenter *net.TCPListener
	tlsListener net.Listener
	udpListener *udpmock.UdpMockListener
}

func NewFromTcpListener(transportListener *net.TCPListener) (c *TransportListener, err error) {
	c = &TransportListener{}
	c.connType = "tcp"
	c.tcpLisenter = transportListener
	return c, nil
}
func NewFromTlsListener(tlsListener net.Listener) (c *TransportListener, err error) {
	c = &TransportListener{}
	c.connType = "tls"
	c.tlsListener = tlsListener
	return c, nil
}

func NewFromUdpListener(udpListener *udpmock.UdpMockListener) (c *TransportListener, err error) {
	c = &TransportListener{}
	c.connType = "udp"
	c.udpListener = udpListener
	return c, nil
}

func (c *TransportListener) Close() error {
	if c.connType == "tcp" && c.tcpLisenter != nil {
		return c.tcpLisenter.Close()
	}
	if c.connType == "tls" && c.tlsListener != nil {
		return c.tlsListener.Close()
	}
	if c.connType == "udp" {
		return c.udpListener.Close()
	}
	return errors.New("not found connType " + c.connType + " for Close")
}

func (c *TransportListener) Accept() (transportConn *TransportConn, err error) {
	if c.connType == "tcp" && c.tcpLisenter != nil {
		tcpConn, err := c.tcpLisenter.AcceptTCP()
		if err != nil {
			belogs.Error("Accept(): TransportListener Accept TCP tcpConn remote fail: ", err)
			return nil, err
		}
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(time.Second * 300)
		transportConn = NewFromTcpConn(tcpConn)
		belogs.Info("Accept(): TransportListener Accept TCP transportConn remote: ", transportConn.RemoteAddr().String())
		return transportConn, nil
	}
	if c.connType == "tls" && c.tlsListener != nil {
		conn, err := c.tlsListener.Accept()
		if err != nil {
			belogs.Error("Accept(): TransportListener  Accept Tls remote fail: ", err)
			return nil, err
		}

		tlsConn, ok := conn.(*tls.Conn)
		if !ok {
			belogs.Error("Accept(): TransportListener  Accept Tls remote , conn cannot conver to tlsConn: ", conn.RemoteAddr().String(), err)
			return nil, err
		}
		transportConn = NewFromTlsConn(tlsConn)
		belogs.Info("Accept(): TransportListener Accept Tls transportConn remote: ", transportConn.RemoteAddr().String())
		return transportConn, nil
	}
	if c.connType == "udp" && c.udpListener != nil {
		udpConn, udpAddr, err := c.udpListener.Accept()
		if err != nil {
			belogs.Error("Accept(): TransportListener Accept Udp remote fail: ", err)
			return nil, err
		}

		transportConn = NewFromUdpConn(udpConn, udpAddr)
		belogs.Info("Accept(): TransportListener Accept Udp transportConn remote: ", transportConn.RemoteAddr().String())
		return transportConn, nil
	}
	return nil, errors.New("not support connType " + c.connType + " for Accept")
}
