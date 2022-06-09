package tcptlsutil

import "github.com/cpusoft/goutil/jsonutil"

const (
	MSG_TYPE_CLOSE_FORCIBLE             = 1
	MSG_TYPE_CLOSE_GRACEFUL             = 2
	MSG_TYPE_CLOSE_ONE_CONNECT_FORCIBLE = 3
	MSG_TYPE_CLOSE_ONE_CONNECT_GRACEFUL = 4

	MSG_TYPE_ACTIVE_SEND_DATA = 5

//	MSG_TYPE_TO_SERVER_CLOSE_CONNECT_GRACEFUL = 4
)

type TcpTlsMsg struct {
	// common
	MsgType   uint64      `json:"msgType"`
	MsgResult chan string `json:"msgResult"`

	// for close
	ConnKey string `json:"connKey,omitempty"`

	// for send data //
	// NEXT_CONNECT_CLOSE_POLICY_NO  NEXT_CONNECT_CLOSE_POLICY_GRACEFUL  NEXT_CONNECT_CLOSE_POLICY_FORCIBLE
	NextConnectClosePolicy int `json:"nextConnectClosePolicy,omitempty"`
	//NEXT_RW_POLICY_ALL,NEXT_RW_POLICY_WAIT_READ,NEXT_RW_POLICY_WAIT_WRITE
	NextRwPolicy int                     `json:"nextRwPolicy,omitempty"`
	SendData     jsonutil.PrintableBytes `json:"sendData,omitempty"`
}
