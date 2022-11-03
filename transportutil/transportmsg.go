package transportutil

import "github.com/cpusoft/goutil/jsonutil"

const (
	// common
	BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_DATA = 1

	// server
	BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_FORCIBLE             = 100
	BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_GRACEFUL             = 101
	BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_FORCIBLE = 102
	BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_GRACEFUL = 103

	// client
	BUSINESS_TO_CONN_MSG_TYPE_CLIENT_CLOSE_CONNECT = 200

	// client conn to business
	CONN_TO_BUSINESS_MSG_TYPE_DNS = 300
)

// from upper business send to lower conn, such as 'send data to conn'
// used in client and server
type BusinessToConnMsg struct {
	// common
	BusinessToConnMsgType uint64 `json:"businessToConnMsgType"`

	// for send data //
	SendData jsonutil.HexBytes `json:"sendData,omitempty"`

	// for server to choose which conn
	// if is "", will send all conns
	ServerConnKey string `json:"serverConnKey,omitempty"`
}
