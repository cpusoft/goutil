package main

import (
	"crypto/tls"
	"net"
	"time"

	belogs "github.com/cpusoft/goutil/belogs"
)

// https://stackoverflow.com/questions/66755407/cancelling-a-net-listener-via-context-in-golang

type TcpTlsListener struct {
	// tcp: true
	// tls: false
	isTcpConn   bool
	tcpLisenter *net.TCPListener
	tlsListener net.Listener
}

func NewFromTcpListener(tcpListener *net.TCPListener) (c *TcpTlsListener, err error) {
	c = &TcpTlsListener{}
	c.isTcpConn = true
	c.tcpLisenter = tcpListener
	return c, nil
}
func NewFromTlsListener(tlsListener net.Listener) (c *TcpTlsListener, err error) {
	c = &TcpTlsListener{}
	c.isTcpConn = false
	c.tlsListener = tlsListener
	return c, nil
}

func (c *TcpTlsListener) Close() error {
	if c.isTcpConn && c.tcpLisenter != nil {
		return c.tcpLisenter.Close()
	} else {
		return c.tlsListener.Close()
	}
}

func (c *TcpTlsListener) Accept() (tcpTlsConn *TcpTlsConn, err error) {
	if c.isTcpConn && c.tcpLisenter != nil {
		tcpConn, err := c.tcpLisenter.AcceptTCP()
		if err != nil {
			belogs.Error("Accept(): TcpTlsListener Accept TCP tcpConn remote fail: ", err)
			return nil, err
		}
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(time.Second * 300)
		tcpTlsConn = NewFromTcpConn(tcpConn)
		belogs.Info("Accept(): TcpTlsListener Accept TCP tcpTlsConn remote: ", tcpTlsConn.RemoteAddr().String())
		return tcpTlsConn, nil
	} else {
		conn, err := c.tlsListener.Accept()
		if err != nil {
			belogs.Error("Accept(): TcpTlsListener  Accept Tls remote fail: ", err)
			return nil, err
		}

		tlsConn, ok := conn.(*tls.Conn)
		if !ok {
			belogs.Error("Accept(): TcpTlsListener  Accept Tls remote , conn cannot conver to tlsConn: ", conn.RemoteAddr().String(), err)
			return nil, err
		}
		tcpTlsConn = NewFromTlsConn(tlsConn)
		belogs.Info("Accept(): TcpTlsListener Accept Tls tcpTlsConn remote: ", tcpTlsConn.RemoteAddr().String())
		return tcpTlsConn, nil
	}
}
