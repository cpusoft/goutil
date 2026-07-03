package tcpserver

import (
	"crypto/tls"
	"fmt"
	"net"
)

func getUnderlyingTCPConn(conn net.Conn) (*net.TCPConn, bool) {
	for {
		switch c := conn.(type) {
		case *net.TCPConn:
			return c, true
		case *tls.Conn:
			inner := c.NetConn()
			if tcpConn, ok := inner.(*net.TCPConn); ok {
				return tcpConn, true
			}
			// 如果底层不是 TCPConn，尝试继续 unwrap（理论上不会发生）
			conn = inner
		default:
			return nil, false
		}
	}
}

// 辅助函数：TLS版本转字符串
func tlsVersionToString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown(0x%x)", version)
	}
}
