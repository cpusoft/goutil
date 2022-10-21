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
func (c *UdpMockListener) Accept() (*net.UDPConn, error) {
	//监听端口
	udpConn, err := net.ListenUDP(c.network, c.udpAddr)
	if err != nil {
		belogs.Error("UdpMockListener.Accept(): ListenUDP fail, network:", c.network, "  udpAddr:", c.udpAddr, err)
		return nil, err
	}
	return udpConn, nil
}

// just set value, not auctually listener
func ListenUDP(network string, laddr *net.UDPAddr) (*UdpMockListener, error) {
	c := &UdpMockListener{}
	c.network = network
	c.udpAddr = laddr
	return c, nil
}
