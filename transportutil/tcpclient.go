package transportutil

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

type TcpClient struct {
	// both tcp and tls
	connType         string
	tcpClientProcess TcpClientProcess

	// will set tcp length first
	tcptlsLengthDeclaration string
	// if tcptlsLengthDeclaration ==false, then shoud set tcp/tls receive bytes len
	receiveOnePacketLength int

	// for tls
	tlsRootCrtFileName    string
	tlsPublicCrtFileName  string
	tlsPrivateKeyFileName string

	// for close
	tcpConn *TcpConn

	// for channel
	businessToConnMsgCh chan BusinessToConnMsg

	// for onReceive to SendAndReceiveMsg
	connToBusinessMsgCh chan ConnToBusinessMsg
}

// server: 0.0.0.0:port
func NewTcpClient(tcpClientProcess TcpClientProcess,
	businessToConnMsgCh chan BusinessToConnMsg,
	tcptlsLengthDeclaration string, receiveOnePacketLength int) (tc *TcpClient) {

	belogs.Debug("NewTcpClient():tcpClientProcess:", tcpClientProcess, "  receiveOnePacketLength:", receiveOnePacketLength)
	tc = &TcpClient{}
	tc.connType = "tcp"
	tc.tcpClientProcess = tcpClientProcess
	tc.businessToConnMsgCh = businessToConnMsgCh
	tc.connToBusinessMsgCh = make(chan ConnToBusinessMsg)
	tc.tcptlsLengthDeclaration = tcptlsLengthDeclaration
	tc.receiveOnePacketLength = receiveOnePacketLength
	belogs.Info("NewTcpClient():tc:", tc, "  tc.connToBusinessMsgCh:", tc.connToBusinessMsgCh)
	return tc
}

// server: 0.0.0.0:port
func NewTlsClient(tlsRootCrtFileName, tlsPublicCrtFileName, tlsPrivateKeyFileName string,
	tcpClientProcess TcpClientProcess, businessToConnMsgCh chan BusinessToConnMsg,
	tcptlsLengthDeclaration string, receiveOnePacketLength int) (tc *TcpClient, err error) {

	belogs.Debug("NewTlsClient(): tlsRootCrtFileName:", tlsRootCrtFileName, "  tlsPublicCrtFileName:", tlsPublicCrtFileName,
		"   tlsPrivateKeyFileName:", tlsPrivateKeyFileName, "  tcpClientProcess:", &tcpClientProcess,
		"  tcptlsLengthDeclaration:", tcptlsLengthDeclaration, "   receiveOnePacketLength:", receiveOnePacketLength)
	tc = &TcpClient{}
	tc.connType = "tls"
	tc.tcpClientProcess = tcpClientProcess
	tc.businessToConnMsgCh = businessToConnMsgCh
	tc.connToBusinessMsgCh = make(chan ConnToBusinessMsg)
	tc.tcptlsLengthDeclaration = tcptlsLengthDeclaration
	tc.receiveOnePacketLength = receiveOnePacketLength

	rootExists, _ := osutil.IsExists(tlsRootCrtFileName)
	if !rootExists {
		belogs.Error("NewTlsClient():root cer files not exists:", tlsRootCrtFileName)
		return nil, errors.New("root cer file is not exists")
	}
	publicExists, _ := osutil.IsExists(tlsPublicCrtFileName)
	if !publicExists {
		belogs.Error("NewTlsClient():public cer files not exists:", tlsPublicCrtFileName)
		return nil, errors.New("public cer file is not exists")
	}
	privateExists, _ := osutil.IsExists(tlsPrivateKeyFileName)
	if !privateExists {
		belogs.Error("NewTlsClient():private cer files not exists:", tlsPrivateKeyFileName)
		return nil, errors.New("private cer file is not exists")
	}

	tc.tlsRootCrtFileName = tlsRootCrtFileName
	tc.tlsPublicCrtFileName = tlsPublicCrtFileName
	tc.tlsPrivateKeyFileName = tlsPrivateKeyFileName
	belogs.Info("NewTlsClient():tc:", &tc)
	return tc, nil
}

// server: **.**.**.**:port
func (tc *TcpClient) StartTcpClient(server string) (err error) {
	belogs.Debug("TcpClient.StartTcpClient(): create client, server is  ", server)

	conn, err := net.DialTimeout("tcp", server, 60*time.Second)
	if err != nil {
		belogs.Error("TcpClient.StartTcpClient(): DialTimeout fail, server:", server, err)
		return err
	}
	belogs.Debug("TcpClient.StartTcpClient(): DialTimeout ok, server is  ", server)

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		belogs.Error("TcpClient.StartTcpClient(): conn cannot conver to tcpConn: ", conn.RemoteAddr().String(), err)
		return err
	}
	belogs.Debug("TcpClient.StartTcpClient(): tcpConn ok, server is  ", server)

	tc.tcpConn = NewFromTcpConn(tcpConn)
	//active send to server, and receive from server, loop
	belogs.Debug("TcpClient.StartTcpClient(): NewFromTcpConn ok, server:", server, "   tcpConn:", tc.tcpConn.RemoteAddr().String())

	// onConnect
	tc.onConnect()

	// onReceive
	go tc.onReceive()

	belogs.Info("TcpClient.StartTcpClient(): ok, server is  ", server, "  tcpConn:", tc.tcpConn.RemoteAddr().String())
	return nil
}

// server: **.**.**.**:port
func (tc *TcpClient) StartTlsClient(server string) (err error) {
	belogs.Debug("TcpClient.StartTlsClient(): create client, server is  ", server,
		"  tlsPublicCrtFileName:", tc.tlsPublicCrtFileName,
		"  tlsPrivateKeyFileName:", tc.tlsPrivateKeyFileName)

	tlsConfigModel := TlsConfigModel{
		TlsRootCrtFileName:    tc.tlsRootCrtFileName,
		TlsPublicCrtFileName:  tc.tlsPublicCrtFileName,
		TlsPrivateKeyFileName: tc.tlsPublicCrtFileName,
		InsecureSkipVerify:    false,
	}
	config, err := GetClientTlsConfig(tlsConfigModel)
	if err != nil {
		belogs.Error("TcpClient.StartTlsClient(): GetClientTlsConfig fail, tlsConfigModel:", jsonutil.MarshalJson(tlsConfigModel), err)
		return err
	}
	/*
		cert, err := tls.LoadX509KeyPair(tc.tlsPublicCrtFileName, tc.tlsPrivateKeyFileName)
		if err != nil {
			belogs.Error("TcpClient.StartTlsClient(): LoadX509KeyPair fail: server:", server,
				"  tlsPublicCrtFileName:", tc.tlsPublicCrtFileName,
				"  tlsPrivateKeyFileName:", tc.tlsPrivateKeyFileName, err)
			return err
		}
		belogs.Debug("TcpClient.StartTlsClient(): LoadX509KeyPair ok, server is  ", server)

		rootCrtBytes, err := os.ReadFile(tc.tlsRootCrtFileName)
		if err != nil {
			belogs.Error("TcpClient.StartTlsClient(): ReadFile tlsRootCrtFileName fail, server:", server,
				"  tlsRootCrtFileName:", tc.tlsRootCrtFileName, err)
			return err
		}
		belogs.Debug("TcpClient.StartTlsClient(): ReadFile ok, tlsRootCrtFileName:", tc.tlsRootCrtFileName)

		rootCertPool := x509.NewCertPool()
		ok := rootCertPool.AppendCertsFromPEM(rootCrtBytes)
		if !ok {
			belogs.Error("TcpClient.StartTlsClient(): AppendCertsFromPEM tlsRootCrtFileName fail,server:", server,
				"  tlsRootCrtFileName:", tc.tlsRootCrtFileName, "  len(rootCrtBytes):", len(rootCrtBytes), err)
			return err
		}
		belogs.Debug("TcpClient.StartTlsClient(): AppendCertsFromPEM ok, server is  ", server)

		config := &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            rootCertPool,
			InsecureSkipVerify: false,
		}
	*/

	dialer := &net.Dialer{Timeout: time.Duration(60) * time.Second}
	tlsConn, err := tls.DialWithDialer(dialer, "tcp", server, config)
	if err != nil {
		belogs.Error("TcpClient.StartTlsClient(): DialWithDialer fail, server:", server, err)
		return err
	}
	belogs.Debug("TcpClient.StartTlsClient(): DialWithDialer ok, server is  ", server)

	tc.tcpConn = NewFromTlsConn(tlsConn)
	belogs.Debug("TcpClient.StartTlsClient(): NewFromTlsConn ok, server:", server, "   tcpConn:", tc.tcpConn.RemoteAddr().String())

	// onConnect
	tc.onConnect()

	// onReceive
	go tc.onReceive()

	belogs.Info("TcpClient.StartTlsClient(): onReceive, server is  ", server, "  tcpConn:", tc.tcpConn.RemoteAddr().String())

	return nil
}

func (tc *TcpClient) onReceive() (err error) {
	belogs.Debug("TcpClient.onReceive(): wait for onReceive, tcptlsLengthDeclaration:", tc.tcptlsLengthDeclaration,
		" receiveOnePacketLength:", tc.receiveOnePacketLength, " tcpConn:", tc.tcpConn.RemoteAddr().String())
	var leftData []byte
	// when end onReceive, will onClose
	defer tc.onClose()
	var buffer []byte
	var length uint16
	for {
		start := time.Now()
		if tc.tcptlsLengthDeclaration == "true" {
			binary.Read(tc.tcpConn, binary.BigEndian, &length)
		} else {
			length = uint16(tc.receiveOnePacketLength)
		}
		buffer = make([]byte, length)
		n, err := tc.tcpConn.Read(buffer)
		belogs.Debug("TcpClient.onReceive(): Read n :", n, "  length:", length, " from tcpConn: ", tc.tcpConn.RemoteAddr().String(),
			" buffer:", convert.PrintBytesOneLine(buffer), "  time(s):", time.Since(start))
		if err != nil {
			if err == io.EOF {
				// is not error, just client close
				belogs.Debug("TcpClient.onReceive(): io.EOF, client close: ", tc.tcpConn.RemoteAddr().String(),
					" length:", length, "  time(s):", time.Since(start), err)
				return nil
			}
			belogs.Error("TcpClient.onReceive(): Read fail or connect is closing, err ", tc.tcpConn.RemoteAddr().String(),
				" length:", length, "  time(s):", time.Since(start), err)
			return err
		}

		belogs.Debug("TcpClient.onReceive(): Read n :", n, "  length:", length, " from tcpConn: ", tc.tcpConn.RemoteAddr().String(),
			"  time(s):", time.Since(start))
		nextRwPolicy, leftData, connToBusinessMsg, err := tc.tcpClientProcess.OnReceiveProcess(tc.tcpConn, append(leftData, buffer[:n]...))
		belogs.Info("TcpClient.onReceive(): tcpClientProcess.OnReceiveProcess, tcpConn: ", tc.tcpConn.RemoteAddr().String(), " receive n: ", n,
			"  len(leftData):", len(leftData), "  nextRwPolicy:", nextRwPolicy, "  connToBusinessMsg:", jsonutil.MarshalJson(connToBusinessMsg), "  time(s):", time.Since(start))
		if err != nil {
			belogs.Error("TcpClient.onReceive(): tcpClientProcess.OnReceiveProcess  fail ,will close this tcpConn : ", tc.tcpConn.RemoteAddr().String(), err)
			return err
		}
		if nextRwPolicy == NEXT_RW_POLICY_END_READ {
			belogs.Info("TcpClient.onReceive():  nextRwPolicy is NEXT_RW_POLICY_END_READ, will close connect: ", tc.tcpConn.RemoteAddr().String())
			return nil
		}

		// reset buffer
		buffer = make([]byte, length)
		belogs.Debug("TcpClient.onReceive(): will reset buffer and wait for Read from tcpConn: ", tc.tcpConn.RemoteAddr().String(),
			"  time(s):", time.Since(start))
		go func() {
			if !connToBusinessMsg.IsActiveSendFromServer {
				belogs.Debug("TcpClient.onReceive(): tcpClientProcess.OnReceiveProcess, will send to tc.businessToConnMsgCh:", tc.businessToConnMsgCh,
					"   connToBusinessMsg:", jsonutil.MarshalJson(connToBusinessMsg))
				tc.connToBusinessMsgCh <- *connToBusinessMsg
				belogs.Debug("TcpClient.onReceive(): tcpClientProcess.OnReceiveProcess, have send to connToBusinessMsg:", jsonutil.MarshalJson(connToBusinessMsg))
			}
		}()
	}

}

func (tc *TcpClient) onConnect() {
	// call process func onConnect
	tc.tcpClientProcess.OnConnectProcess(tc.tcpConn)
	belogs.Info("TcpClient.onConnect(): after OnConnectProcess, tcpConn: ", tc.tcpConn.RemoteAddr().String())
}

func (tc *TcpClient) onClose() {
	// close in the end
	belogs.Info("TcpClient.onClose(): tcpConn: ", tc.tcpConn.RemoteAddr().String())
	tc.tcpClientProcess.OnCloseProcess(tc.tcpConn)
	tc.tcpConn.Close()

}

func (tc *TcpClient) SendAndReceiveMsg(businessToConnMsg *BusinessToConnMsg) (connToBusinessMsg *ConnToBusinessMsg, err error) {
	belogs.Info("TcpClient.SendAndReceiveMsg(): businessToConnMsg:", jsonutil.MarshalJson(businessToConnMsg),
		"  tcpConn: ", tc.tcpConn.RemoteAddr().String())

	switch businessToConnMsg.BusinessToConnMsgType {
	case BUSINESS_TO_CONN_MSG_TYPE_CLIENT_CLOSE_CONNECT:
		belogs.Info("TcpClient.SendAndReceiveMsg(): businessToConnMsgType is BUSINESS_TO_CONN_MSG_TYPE_CLIENT_CLOSE_CONNECT,",
			" will close for tcpConn: ", tc.tcpConn.RemoteAddr().String(), " will return, close SendAndReceiveMsg")
		tc.onClose()
		// end for/select
		// will return, close SendAndReceiveMsg
		return nil, nil
	case BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_DATA:
		belogs.Info("TcpClient.SendAndReceiveMsg(): businessToConnMsgType is BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_DATA,",
			" will send to tcpConn: ", tc.tcpConn.RemoteAddr().String(), "  tcptlsLengthDeclaration:", tc.tcptlsLengthDeclaration)

		var n int
		sendDataNew := getLengthDeclarationSendData(tc.tcptlsLengthDeclaration, businessToConnMsg.SendData)
		length := len(sendDataNew)
		belogs.Debug("TcpClient.SendAndReceiveMsg(): send to server with LengthDeclaration:", tc.tcpConn.RemoteAddr().String(),
			"   tcptlsLengthDeclaration:", tc.tcptlsLengthDeclaration,
			"   len(sendDataNew):", length, "   sendDataNew:", convert.PrintBytesOneLine(sendDataNew))
		belogs.Info("TcpClient.SendAndReceiveMsg(): send to server with LengthDeclaration:", tc.tcpConn.RemoteAddr().String(),
			"   tcptlsLengthDeclaration:", tc.tcptlsLengthDeclaration,
			"   len(sendDataNew):", length)
		start := time.Now()
		n, err = tc.tcpConn.Write(sendDataNew)
		if err != nil {
			belogs.Error("TcpClient.SendAndReceiveMsg(): Write fail, will close tcpConn:", tc.tcpConn.RemoteAddr().String(), "  time(s):", time.Since(start), err)
			tc.onClose()
			return nil, err
		}
		belogs.Info("TcpClient.SendAndReceiveMsg(): Write to tcpConn:", tc.tcpConn.RemoteAddr().String(),
			"  tcptlsLengthDeclaration:", tc.tcptlsLengthDeclaration, "  len(sendDataNew):", length,
			"  write n:", n, ",  and wait for receive connToBusinessMsg",
			"  time(s):", time.Since(start))
		if !businessToConnMsg.NeedClientWaitForServerResponse {
			belogs.Debug("TcpClient.SendAndReceiveMsg(): isnot NeedClientWaitForServerResponse, just return, businessToConnMsg:", jsonutil.MarshalJson(businessToConnMsg))
			return nil, nil
		}
		// wait receive msg from "onReceive"
		belogs.Debug("TcpClient.SendAndReceiveMsg(): will receive from uc.connToBusinessMsg: ")

		//connToBusinessMsg := <-tc.connToBusinessMsgCh
		for {
			belogs.Debug("TcpClient.SendAndReceiveMsg(): for select,  tc.connToBusinessMsgCh:", tc.connToBusinessMsgCh)
			select {
			case connToBusinessMsg := <-tc.connToBusinessMsgCh:
				belogs.Debug("TcpClient.SendAndReceiveMsg(): receive from tc.connToBusinessMsg, connToBusinessMsg:", jsonutil.MarshalJson(connToBusinessMsg),
					"  time(s):", time.Since(start))
				return &connToBusinessMsg, nil

			case <-time.After(5 * time.Second):
				belogs.Debug("TcpClient.SendAndReceiveMsg(): receive fail, timeout")
				return nil, errors.New("server response is timeout")
			}
		}
	}
	return nil, errors.New("BusinessToConnMsgType is not supported")
}
func (tc *TcpClient) GetTcpConnKey() string {
	return GetTcpConnKey(tc.tcpConn)
}
