package udpmock

import (
	"net"

	"github.com/cpusoft/goutil/belogs"
)

type UdpMockListener struct {
	network string
	udpAddr *net.UDPAddr
}

func (c *UdpMockListener) Close() error {
	return nil
}

// get udpConn
func (c *UdpMockListener) Accept() (*net.UDPConn, *net.UDPAddr, error) {
	//监听端口
	udpConn, err := net.ListenUDP(c.network, c.udpAddr)
	if err != nil {
		belogs.Error("UdpMockListener.Accept(): ListenUDP fail, network:", c.network, "  udpAddr:", c.udpAddr, err)
		return nil, nil, err
	}
	belogs.Debug("UdpMockListener.Accept(): ListenUDP, network:", c.network, "  udpAddr:", c.udpAddr, "   udpConn:", udpConn.RemoteAddr().String())
	return udpConn, c.udpAddr, nil
}

// just set value, not auctually listener
func ListenUDP(network string, laddr *net.UDPAddr) (*UdpMockListener, error) {
	c := &UdpMockListener{}
	c.network = network
	c.udpAddr = laddr
	belogs.Debug("UdpMockListener.ListenUDP(): UdpMockListener:", c)
	return c, nil
}
