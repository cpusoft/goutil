package transportutil

/*
func TestCreateTcpClient(t *testing.T) {
	clientProcessFunc := new(ClientProcessFunc)
	fmt.Println("CreateTcpClient():", "192.168.83.139:9999")
	//CreateTcpClient("127.0.0.1:9999", ClientProcess1)
	tc := NewTcpClient(clientProcessFunc)
	err := tc.StartTcpClient("192.168.83.139:9999")
	if err != nil {
		fmt.Println("CreateTcpClient(): StartTcpClient tc fail: ", &tc, err)
		return
	}
	fmt.Println("CreateTcpClient(): tcpclient will SendData")
	transportMsg := &TransportMsg{NextConnectClosePolicy: NEXT_CONNECT_POLICY_KEEP,
		NextRwPolicy: NEXT_RW_POLICY_WAIT_READ,
		SendData:     GetTcpClientData(),
	}
	tc.SendTransportMsg(transportMsg)
	time.Sleep(60 * time.Second)

	fmt.Println("CreateTcpClient(): tcpclient will stop")
	tc.SendMsgForCloseConnect()

}

func TestCreateTlsClient(t *testing.T) {
	clientProcessFunc := new(ClientProcessFunc)
	tlsRootCrtFileName := `ca.cer`
	tlsPublicCrtFileName := `client.cer`
	tlsPrivateKeyFileName := `clientkey.pem`
	fmt.Println("CreateTlsClient(): tlsRootCrtFileName:", tlsRootCrtFileName,
		"tlsPublicCrtFileName:", tlsPublicCrtFileName,
		"tlsPrivateKeyFileName:", tlsPrivateKeyFileName)
	//CreateTcpClient("192.168.83.139:9999", ClientProcess1)
	tc, err := NewTlsClient(tlsRootCrtFileName, tlsPublicCrtFileName, tlsPrivateKeyFileName, clientProcessFunc)
	if err != nil {
		fmt.Println("CreateTcpClient(): NewTlsClient tc fail: ", &tc, err)
		return
	}
	err = tc.StartTlsClient("192.168.83.139:9999")
	if err != nil {
		fmt.Println("CreateTcpClient(): StartTlsClient tc fail: ", &tc, err)
		return
	}
	fmt.Println("CreateTcpClient(): tcpclient will SendData")
	transportMsg := &TransportMsg{NextConnectClosePolicy: NEXT_CONNECT_POLICY_KEEP,
		NextRwPolicy: NEXT_RW_POLICY_WAIT_READ,
		SendData:     GetTcpClientData(),
	}
	tc.SendTransportMsg(transportMsg)
	time.Sleep(60 * time.Second)

	fmt.Println("CreateTcpClient(): tcpclient will stop")
	tcpClientSendMsg.NextConnectClosePolicy = NEXT_CONNECT_POLICY_CLOSE_GRACEFUL
	tcpClientSendMsg.SendData = nil
	tc.SendTransportMsg(tcpClientSendMsg)

}

type ClientProcessFunc struct {
}

func (cp *ClientProcessFunc) OnConnectProcess(tcpConn *TcpConn) {

	fmt.Println("OnConnectProcess(): tcpclient tcpConn:", tcpConn.RemoteAddr().String())

}
func (cp *ClientProcessFunc) OnCloseProcess(tcpConn *TcpConn) {
	if tcpConn != nil {
		fmt.Println("OnCloseProcess(): tcpclient tcpConn:", tcpConn.RemoteAddr().String())
	}
}

func (sq *ClientProcessFunc) OnReceiveProcess(tcpConn *TcpConn, receiveData []byte) (nextRwPolicy int, leftData []byte, err error) {

	fmt.Println("OnReceiveProcess() tcpclient  :", tcpConn, convert.Bytes2String(receiveData))
	return NEXT_RW_POLICY_END_READ, make([]byte, 0), nil
}

func GetTcpClientData() (buffer []byte) {

	return []byte{0x00, 0x0b, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x0b, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
}
*/
