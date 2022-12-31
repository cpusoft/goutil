package transportutil

import "github.com/cpusoft/goutil/jsonutil"

const (
	// common
	BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_DATA             = "sendData"
	BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_AND_RECEIVE_DATA = "sendAndReceiveData"
	// server
	BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_FORCIBLE             = "serverCloseForcible"
	BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_GRACEFUL             = "serverCloseGraceful"
	BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_FORCIBLE = "serverCloseOneConnectForcible"
	BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_GRACEFUL = "serverCloseOneConnectGraceful"

	// client
	BUSINESS_TO_CONN_MSG_TYPE_CLIENT_CLOSE_CONNECT = "clientCloseConnect"
)

// from upper business send to lower conn, such as 'send data to conn'
// used in client and server
type BusinessToConnMsg struct {
	// common
	BusinessToConnMsgType string `json:"businessToConnMsgType"`

	// for send data //
	SendData jsonutil.HexBytes `json:"sendData,omitempty"`

	// for server to choose which conn
	// if is "", will send all conns
	ServerConnKey string `json:"serverConnKey,omitempty"`

	// for client,need wait for server's response
	NeedClientWaitForServerResponse bool `json:"needClientWaitForServerResponse,omitempty"`
}

type ConnToBusinessMsg struct {
	// true: active from server;  false: is response from server
	IsActiveSendFromServer bool `json:"isActiveSendFromServer"`
	// common
	ConnToBusinessMsgType string      `json:"connToBusinessMsgType"`
	ReceiveData           interface{} `json:"receiveData"`
}
