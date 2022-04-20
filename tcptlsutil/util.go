package main

import (
	"container/list"
	"errors"

	belogs "github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
)

const (

	// connect: keep or close graceful/forcible
	NEXT_CONNECT_POLICE_KEEP           = 0
	NEXT_CONNECT_POLICE_CLOSE_GRACEFUL = 1
	NEXT_CONNECT_POLICE_CLOSE_FORCIBLE = 2

	// need wait read/write
	NEXT_RW_POLICE_WAIT_READ  = 3
	NEXT_RW_POLICE_WAIT_WRITE = 4

	// no need more read
	NEXT_RW_POLICE_END_READ = 5

	SERVER_STOP_GRACEFUL = 6 // should --> LISTEN_STOP_GRACEFUL --> all NEXT_CONNECT_POLICE_CLOSE_GRACEFUL
	SERVER_STOP_FORCIBLE = 7 // should --> LISTEN_STOP_FORCIBLE --> all NEXT_CONNECT_POLICE_CLOSE_FORCIBLE
	LISTEN_STOP_GRACEFUL = 8
	LISTEN_STOP_FORCIBLE = 9
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
