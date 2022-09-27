package tcptlsutil

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

type TcpTlsMsg struct {
	// common
	MsgType   uint64      `json:"msgType"`
	MsgResult chan string `json:"-"` // must ignore

	// NEXT_CONNECT_CLOSE_POLICY_NO  NEXT_CONNECT_CLOSE_POLICY_GRACEFUL  NEXT_CONNECT_CLOSE_POLICY_FORCIBLE
	//NextConnectClosePolicy int `json:"nextConnectClosePolicy,omitempty"`
	//NEXT_RW_POLICY_ALL,NEXT_RW_POLICY_WAIT_READ,NEXT_RW_POLICY_WAIT_WRITE
	//NextRwPolicy int `json:"nextRwPolicy,omitempty"`

	// for send data //
	SendData jsonutil.HexBytes `json:"sendData,omitempty"`

	// for server to choose which conn
	// if is "", will send all conns
	ServerConnKey string `json:"serverConnKey,omitempty"`
}
