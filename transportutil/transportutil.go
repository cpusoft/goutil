package transportutil

import (
	"container/list"
	"encoding/binary"
	"errors"
	"net"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
)

const (

	// connect: keep or close graceful/forcible
	NEXT_CONNECT_POLICY_KEEP           = 0
	NEXT_CONNECT_POLICY_CLOSE_GRACEFUL = 1
	NEXT_CONNECT_POLICY_CLOSE_FORCIBLE = 2

	// need wait read/write
	NEXT_RW_POLICY_WAIT_READ  = 3
	NEXT_RW_POLICY_WAIT_WRITE = 4

	// no need more read
	NEXT_RW_POLICY_END_READ = 5

	SERVER_STATE_INIT    = 0
	SERVER_STATE_RUNNING = 1
	SERVER_STATE_CLOSING = 2
	SERVER_STATE_CLOSED  = 3
)

// packets: if Len==0,means no complete package
func RecombineReceiveData(receiveData []byte, minPacketLen, lengthFieldStart,
	lengthFieldEnd int) (packets *list.List, leftData []byte, err error) {
	belogs.Debug("RecombineReceiveData(): len(receiveData):", len(receiveData),
		"   receiveData:", convert.PrintBytes(receiveData, 8),
		"   minPacketLen:", minPacketLen, "   lengthFieldStart:", lengthFieldStart, "   lengthFieldEnd:", lengthFieldEnd)
	// check parameters
	if len(receiveData) == 0 {
		belogs.Debug("RecombineReceiveData(): len(receiveData) is empty, then return:", len(receiveData))
		return nil, make([]byte, 0), nil
	}
	if minPacketLen <= 0 {
		belogs.Error("RecombineReceiveData(): minPacketLen smaller than 0:", minPacketLen)
		return nil, nil, errors.New("minPacketLen is smaller than 0")
	}
	if lengthFieldStart <= 0 {
		belogs.Error("RecombineReceiveData(): lengthFieldStart smaller than 0:", minPacketLen)
		return nil, nil, errors.New("lengthFieldStart is smaller than 0")
	}
	if lengthFieldEnd <= 0 {
		belogs.Error("RecombineReceiveData(): lengthFieldEnd smaller than 0:", minPacketLen)
		return nil, nil, errors.New("lengthFieldEnd is smaller than 0")
	}
	if lengthFieldStart >= lengthFieldEnd {
		belogs.Error("RecombineReceiveData(): lengthFieldStart lager than lengthFieldEnd:", lengthFieldStart, lengthFieldEnd)
		return nil, nil, errors.New("lengthFieldEnd is smaller than lengthFieldStart")
	}
	packets = list.New()

	for {
		// check
		// unpack: TCP sticky packet

		// if receiveData is smaller than a packet length, packets is empty, receiveData --> leftData
		if len(receiveData) < minPacketLen {
			belogs.Debug("RecombineReceiveData(): len(receiveData) < minPacketLen, then return, len(receiveData), minPacketLen:", len(receiveData), minPacketLen)
			leftData = make([]byte, len(receiveData))
			copy(leftData, receiveData)
			return packets, leftData, nil
		}

		// get length : byte[lengthFieldStart:lengthFieldEnd]
		lengthBuffer := receiveData[lengthFieldStart:lengthFieldEnd]
		length := int(convert.Bytes2Uint64(lengthBuffer))
		belogs.Debug("RecombineReceiveData():lengthBuffer:", lengthBuffer, " length:", length)

		// length is error
		if length < minPacketLen {
			belogs.Error("RecombineReceiveData():length < minPacketLen, then return,length, minPacketLen:", length, minPacketLen)
			return packets, nil, errors.New("length is error")
		}

		// if receiveData is smaller than a packet length, the receiveData --> leftData, return
		if len(receiveData) < length {
			belogs.Debug("RecombineReceiveData(): len(receiveData) smaller then length, then return, len(receiveData), length:", len(receiveData), length)
			leftData = make([]byte, len(receiveData))
			copy(leftData, receiveData)
			return packets, leftData, nil
		} else if len(receiveData) == length {
			// if receiveData is equal to packet length, the receiveData --> packets, return
			belogs.Debug("RecombineReceiveData(): len(receiveData) equal to length, then return, len(receiveData), length:", len(receiveData), length)
			packets.PushBack(receiveData)
			return packets, make([]byte, 0), nil
		} else if len(receiveData) > length {
			// if receiveData is larger than packet length, the receiveData --> packets, leftData --> receiveData, continue
			belogs.Debug("RecombineReceiveData(): len(receiveData) lager than length, then continue, len(receiveData), length:", len(receiveData), length)
			packets.PushBack(receiveData[:length])

			// leftData continue to RecombineReceiveData
			leftData = make([]byte, length)
			copy(leftData, receiveData[length:])
			receiveData = leftData

			belogs.Debug("RecombineReceiveData(): new len(receiveData) lager than length, new(receiveData),length:", len(receiveData), length,
				"   new receiveData:", convert.PrintBytes(receiveData, 8))
		}

	}

}
func TestTcpConnection(address string, port string) (err error) {
	server := net.JoinHostPort(address, port)
	// 3 秒超时
	start := time.Now()
	conn, err := net.DialTimeout("tcp", server, 3*time.Second)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	if err != nil {
		belogs.Error("TestTcpConnection(): DialTimeout fail, server:", server, err, "  time(s):", time.Since(start))
		return err
	}

	return nil
}

func GetTcpConnKey(tcpConn *TcpConn) string {
	if tcpConn == nil {
		return ""
	}
	return tcpConn.LocalAddr().String() + "-" +
		tcpConn.RemoteAddr().String()
}

func GetUdpAddrKey(udpAddr *net.UDPAddr) string {
	if udpAddr == nil {
		return ""
	}
	return udpAddr.String()
}

func getLengthDeclarationSendData(tcptlsLengthDeclaration string, sendData []byte) (sendDataNew []byte) {
	belogs.Debug("getLengthDeclarationSendData(): tcptlsLengthDeclaration:", tcptlsLengthDeclaration,
		"   len(sendData):", len(sendData), convert.PrintBytesOneLine(sendData))
	if tcptlsLengthDeclaration == "true" {
		sendDataNew = make([]byte, 2+len(sendData))
		binary.BigEndian.PutUint16(sendDataNew, uint16(len(sendData)))
		copy(sendDataNew[2:], sendData)
		belogs.Debug("getLengthDeclarationSendData():  len(sendDataNew):", len(sendDataNew))
		return sendDataNew
	}
	return sendData
}
