package transportutil

import (
	"errors"
	"net"
	"sync"

	"github.com/cpusoft/goutil/belogs"
)

type UdpConn struct {
	// udp
	connType string
	udpConn  *net.UDPConn

	// udp
	serverUdpAddr        *net.UDPAddr
	clientUdpAddrsMutext sync.RWMutex
	clientUdpAddrs       map[string]*net.UDPAddr
}

// when server, should get udpAddr from client
// when client, udpAddr is nil
func NewFromUdpConn(udpConn *net.UDPConn) (c *UdpConn) {
	c = &UdpConn{}
	c.udpConn = udpConn
	c.connType = "udp"
	c.clientUdpAddrs = make(map[string]*net.UDPAddr, 0)
	return c
}
func (c *UdpConn) SetServerUdpAddr(serverUdpAddr *net.UDPAddr) {
	c.serverUdpAddr = serverUdpAddr
}
func (c *UdpConn) AddClientUdpAddr(clientUdpAddr *net.UDPAddr) {
	c.clientUdpAddrsMutext.Lock()
	defer c.clientUdpAddrsMutext.Unlock()
	c.clientUdpAddrs[GetUdpAddrKey(clientUdpAddr)] = clientUdpAddr
}
func (c *UdpConn) DelClientUdpAddr(clientUdpAddr *net.UDPAddr) {
	c.clientUdpAddrsMutext.Lock()
	defer c.clientUdpAddrsMutext.Unlock()
	delete(c.clientUdpAddrs, GetUdpAddrKey(clientUdpAddr))
}

func (c *UdpConn) WriteToClient(b []byte, clientUdpAddrKey string) (n int, err error) {
	c.clientUdpAddrsMutext.Lock()
	defer c.clientUdpAddrsMutext.Unlock()
	belogs.Debug("UdpConn.WriteToClient():  len(b):", len(b), "  clientUdpAddrKey:", clientUdpAddrKey)
	if c.connType == "udp" && c.udpConn != nil && len(c.clientUdpAddrs) > 0 {
		// server write to client
		if len(clientUdpAddrKey) > 0 {
			clientUdpAddr := c.clientUdpAddrs[clientUdpAddrKey]
			if clientUdpAddr != nil {
				n, err = c.udpConn.WriteToUDP(b, clientUdpAddr)
				if err != nil {
					belogs.Error("UdpConn.WriteToClient(): WriteToUDP fail, clientUdpAddr:", clientUdpAddr)
					delete(c.clientUdpAddrs, GetUdpAddrKey(clientUdpAddr))
				}
				return n, err
			}
		} else {
			for key, _ := range c.clientUdpAddrs {
				clientUdpAddr := c.clientUdpAddrs[key]
				if clientUdpAddr != nil {
					n, err = c.udpConn.WriteToUDP(b, clientUdpAddr)
					if err != nil {
						belogs.Error("UdpConn.WriteToClient(): WriteToUDP fail, clientUdpAddr:", clientUdpAddr)
						delete(c.clientUdpAddrs, key)
					}
				}
			}
			return n, err
		}
	}
	return -1, errors.New("fail to write to client")
}

func (c *UdpConn) ReadFromClient(b []byte) (n int, clientUdpAddr *net.UDPAddr, err error) {
	if c.connType == "udp" && c.udpConn != nil {
		// server read from client
		n, clientUdpAddr, err := c.udpConn.ReadFromUDP(b)
		if err != nil {
			return -1, nil, err
		}
		c.AddClientUdpAddr(clientUdpAddr)
		return n, clientUdpAddr, nil
	}
	return -1, nil, errors.New("fail to fread from client")
}

func (c *UdpConn) WriteToServer(b []byte) (n int, err error) {
	if c.connType == "udp" && c.serverUdpAddr != nil {
		// client ,just write
		return c.udpConn.Write(b)
	}
	return -1, errors.New("fail to write to server")
}

func (c *UdpConn) ReadFromServer(b []byte) (n int, err error) {
	if c.connType == "udp" && c.serverUdpAddr != nil {
		// client read from server
		return c.udpConn.Read(b)
	}
	return -1, errors.New("fail to read from server")
}

func (c *UdpConn) Close() (err error) {
	c.clientUdpAddrsMutext.Lock()
	defer c.clientUdpAddrsMutext.Unlock()
	if c.connType == "udp" && c.udpConn != nil {
		c.serverUdpAddr = nil
		c.clientUdpAddrs = make(map[string]*net.UDPAddr, 0)
		err = c.udpConn.Close()
		if err != nil {
			return err
		}
		c.udpConn = nil
		return nil
	}
	return errors.New("fail to close udp conn")
}
