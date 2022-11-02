package transportutil

import "github.com/cpusoft/goutil/jsonutil"

const (
	// common
	MSG_TYPE_COMMON_SEND_DATA = 1

	// server
	MSG_TYPE_SERVER_CLOSE_FORCIBLE             = 10
	MSG_TYPE_SERVER_CLOSE_GRACEFUL             = 11
	MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_FORCIBLE = 12
	MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_GRACEFUL = 13

	// client
	MSG_TYPE_CLIENT_CLOSE_CONNECT = 20
)

// from upper business send to lower conn, such as 'send data to conn'
// used in client and server
type BusinessToConnMsg struct {
	// common
	MsgType uint64 `json:"msgType"`

	// for send data //
	SendData jsonutil.HexBytes `json:"sendData,omitempty"`

	// for server to choose which conn
	// if is "", will send all conns
	ServerConnKey string `json:"serverConnKey,omitempty"`
}

// from lower conn send to upper business, such as 'receive data from conn, will send to business'
// used in client
type ConnToBusinessMsg struct {
	// common
	MsgType     uint64      `json:"msgType"`
	ReceiveData interface{} `json:"receiveData"`
}
