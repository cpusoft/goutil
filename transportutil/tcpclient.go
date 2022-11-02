package transportutil

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"io/ioutil"
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

	// for tls
	tlsRootCrtFileName    string
	tlsPublicCrtFileName  string
	tlsPrivateKeyFileName string

	// for close
	tcpConn *TcpConn

	// for channel
	businessToConnMsg chan BusinessToConnMsg
}

// server: 0.0.0.0:port
func NewTcpClient(tcpClientProcess TcpClientProcess, businessToConnMsg chan BusinessToConnMsg) (tc *TcpClient) {

	belogs.Debug("NewTcpClient():tcpClientProcess:", tcpClientProcess)
	tc = &TcpClient{}
	tc.connType = "tcp"
	tc.tcpClientProcess = tcpClientProcess
	tc.businessToConnMsg = businessToConnMsg
	belogs.Info("NewTcpClient():tc:", tc)
	return tc
}

// server: 0.0.0.0:port
func NewTlsClient(tlsRootCrtFileName, tlsPublicCrtFileName, tlsPrivateKeyFileName string,
	tcpClientProcess TcpClientProcess, businessToConnMsg chan BusinessToConnMsg) (tc *TcpClient, err error) {

	belogs.Debug("NewTlsClient():tcpClientProcess:", &tcpClientProcess)
	tc = &TcpClient{}
	tc.connType = "tls"
	tc.tcpClientProcess = tcpClientProcess
	tc.businessToConnMsg = businessToConnMsg

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
	go tc.waitBusinessToConnMsg()

	// onConnect
	tc.onConnect()

	// onReceive
	go tc.onReceive()

	belogs.Info("TcpClient.StartTcpClient(): onReceive, server is  ", server, "  tcpConn:", tc.tcpConn.RemoteAddr().String())
	return nil
}

// server: **.**.**.**:port
func (tc *TcpClient) StartTlsClient(server string) (err error) {
	belogs.Debug("TcpClient.StartTlsClient(): create client, server is  ", server,
		"  tlsPublicCrtFileName:", tc.tlsPublicCrtFileName,
		"  tlsPrivateKeyFileName:", tc.tlsPrivateKeyFileName)

	cert, err := tls.LoadX509KeyPair(tc.tlsPublicCrtFileName, tc.tlsPrivateKeyFileName)
	if err != nil {
		belogs.Error("TcpClient.StartTlsClient(): LoadX509KeyPair fail: server:", server,
			"  tlsPublicCrtFileName:", tc.tlsPublicCrtFileName,
			"  tlsPrivateKeyFileName:", tc.tlsPrivateKeyFileName, err)
		return err
	}
	belogs.Debug("TcpClient.StartTlsClient(): LoadX509KeyPair ok, server is  ", server)

	rootCrtBytes, err := ioutil.ReadFile(tc.tlsRootCrtFileName)
	if err != nil {
		belogs.Error("TcpClient.StartTlsClient(): ReadFile tlsRootCrtFileName fail, server:", server,
			"  tlsRootCrtFileName:", tc.tlsRootCrtFileName, err)
		return err
	}
	belogs.Debug("TcpClient.StartTlsClient(): ReadFile ok, server is  ", server)

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
	dialer := &net.Dialer{Timeout: time.Duration(60) * time.Second}
	tlsConn, err := tls.DialWithDialer(dialer, "tcp", server, config)
	if err != nil {
		belogs.Error("TcpClient.StartTlsClient(): DialWithDialer fail, server:", server, err)
		return err
	}
	belogs.Debug("TcpClient.StartTlsClient(): DialWithDialer ok, server is  ", server)

	/*
		tlsConn, err := tls.Dial("tcp", server, config)
		if err != nil {
			belogs.Error("TcpClient.StartTlsClient(): Dial fail, server:", server, err)
			return err
		}
	*/
	tc.tcpConn = NewFromTlsConn(tlsConn)
	belogs.Debug("TcpClient.StartTlsClient(): NewFromTlsConn ok, server:", server, "   tcpConn:", tc.tcpConn.RemoteAddr().String())
	//active send to server, and receive from server, loop
	go tc.waitBusinessToConnMsg()

	// onConnect
	tc.onConnect()

	// onReceive
	go tc.onReceive()

	belogs.Info("TcpClient.StartTlsClient(): onReceive, server is  ", server, "  tcpConn:", tc.tcpConn.RemoteAddr().String())

	return nil
}

func (tc *TcpClient) onReceive() (err error) {
	belogs.Debug("TcpClient.onReceive(): wait for onReceive, tcpConn:", tc.tcpConn.RemoteAddr().String())
	var leftData []byte
	// one packet
	buffer := make([]byte, 2048)
	// wait for new packet to read

	// when end onReceive, will onClose
	defer tc.onClose()
	for {
		start := time.Now()
		n, err := tc.tcpConn.Read(buffer)
		//	if n == 0 {
		//		continue
		//	}
		if err != nil {
			if err == io.EOF {
				// is not error, just client close
				belogs.Debug("TcpClient.onReceive(): io.EOF, client close: ", tc.tcpConn.RemoteAddr().String(), err)
				return nil
			}
			belogs.Error("TcpClient.onReceive(): Read fail or connect is closing, err ", tc.tcpConn.RemoteAddr().String(), err)
			return err
		}

		belogs.Debug("TcpClient.onReceive(): Read n :", n, " from tcpConn: ", tc.tcpConn.RemoteAddr().String(),
			"  time(s):", time.Now().Sub(start))
		nextRwPolicy, leftData, err := tc.tcpClientProcess.OnReceiveProcess(tc.tcpConn, append(leftData, buffer[:n]...))
		belogs.Info("TcpClient.onReceive(): tcpClientProcess.OnReceiveProcess, tcpConn: ", tc.tcpConn.RemoteAddr().String(), " receive n: ", n,
			"  len(leftData):", len(leftData), "  nextRwPolicy:", nextRwPolicy, "  time(s):", time.Now().Sub(start))
		if err != nil {
			belogs.Error("TcpClient.onReceive(): tcpClientProcess.OnReceiveProcess  fail ,will close this tcpConn : ", tc.tcpConn.RemoteAddr().String(), err)
			return err
		}
		if nextRwPolicy == NEXT_RW_POLICY_END_READ {
			belogs.Info("TcpClient.onReceive():  nextRwPolicy is NEXT_RW_POLICY_END_READ, will close connect: ", tc.tcpConn.RemoteAddr().String())
			return nil
		}

		// reset buffer
		buffer = make([]byte, 2048)
		belogs.Debug("TcpClient.onReceive(): will reset buffer and wait for Read from tcpConn: ", tc.tcpConn.RemoteAddr().String(),
			"  time(s):", time.Now().Sub(start))

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

func (tc *TcpClient) SendBusinessToConnMsg(businessToConnMsg *BusinessToConnMsg) {

	belogs.Debug("TcpClient.SendBusinessToConnMsg(): businessToConnMsg:", jsonutil.MarshalJson(*businessToConnMsg))
	tc.businessToConnMsg <- *businessToConnMsg
}

func (tc *TcpClient) IsConnected() bool {

	belogs.Debug("TcpClient.IsConnected():")
	if tc.tcpConn == nil {
		return false
	}
	b := tc.tcpConn.IsConnected()
	belogs.Debug("TcpClient.IsConnected(): connected:", b)
	return b
}

func (tc *TcpClient) SendMsgForCloseConnect() {
	// send channel, and wait listener and conns end itself process and close loop
	belogs.Info("TcpClient.SendMsgForCloseConnect(): will close graceful")
	if tc.IsConnected() {
		businessToConnMsg := &BusinessToConnMsg{
			MsgType: MSG_TYPE_CLIENT_CLOSE_CONNECT,
		}
		tc.SendBusinessToConnMsg(businessToConnMsg)
	}
}

func (tc *TcpClient) waitBusinessToConnMsg() (err error) {
	belogs.Debug("TcpClient.waitBusinessToConnMsg(): tcpConn:", tc.tcpConn.RemoteAddr().String())
	for {
		// wait next businessToConnMsg: only error or NEXT_CONNECT_POLICY_CLOSE_** will end loop
		select {
		case businessToConnMsg := <-tc.businessToConnMsg:
			belogs.Info("TcpClient.waitBusinessToConnMsg(): businessToConnMsg:", jsonutil.MarshalJson(businessToConnMsg),
				"  tcpConn: ", tc.tcpConn.RemoteAddr().String())

			switch businessToConnMsg.MsgType {
			case MSG_TYPE_CLIENT_CLOSE_CONNECT:
				belogs.Info("TcpClient.waitBusinessToConnMsg(): msgType is MSG_TYPE_CLIENT_CLOSE_CONNECT,",
					" will close for tcpConn: ", tc.tcpConn.RemoteAddr().String(), " will return, close waitBusinessToConnMsg")
				tc.onClose()
				// end for/select
				// will return, close waitBusinessToConnMsg
				return nil
			case MSG_TYPE_COMMON_SEND_DATA:
				belogs.Info("TcpClient.waitBusinessToConnMsg(): msgType is MSG_TYPE_COMMON_SEND_DATA,",
					" will send to tcpConn: ", tc.tcpConn.RemoteAddr().String())
				sendData := businessToConnMsg.SendData
				belogs.Debug("TcpClient.waitBusinessToConnMsg(): send to server:", tc.tcpConn.RemoteAddr().String(),
					"   sendData:", convert.PrintBytesOneLine(sendData))

				// send data
				start := time.Now()
				n, err := tc.tcpConn.Write(sendData)
				if err != nil {
					belogs.Error("TcpClient.waitBusinessToConnMsg(): Write fail, will close  tcpConn:", tc.tcpConn.RemoteAddr().String(), err)
					tc.onClose()
					return err
				}
				belogs.Info("TcpClient.waitBusinessToConnMsg(): Write to tcpConn:", tc.tcpConn.RemoteAddr().String(),
					"  len(sendData):", len(sendData), "  write n:", n,
					"  time(s):", time.Since(start))

			}
		}
	}

}
