package tcptlsutil

const (
	MSG_TYPE_TO_SERVER_CLOSE_SERVER_FORCIBLE  = 1
	MSG_TYPE_TO_SERVER_CLOSE_SERVER_GRACEFUL  = 2
	MSG_TYPE_TO_SERVER_CLOSE_CONNECT_FORCIBLE = 3

//	MSG_TYPE_TO_SERVER_CLOSE_CONNECT_GRACEFUL = 4
)

type TcpTlsMsg struct {
	MsgType   uint64      `json:"msgType"`
	MsgDetail interface{} `json:"msgDetail"`
}
