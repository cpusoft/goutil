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

type TransportClient struct {
	// both tcp and tls
	connType               string
	transportClientProcess TransportClientProcess

	// for tls
	tlsRootCrtFileName    string
	tlsPublicCrtFileName  string
	tlsPrivateKeyFileName string

	// for close
	transportConn *TransportConn

	// for channel
	TransportMsg chan TransportMsg
}

// server: 0.0.0.0:port
func NewTcpClient(transportClientProcess TransportClientProcess, transportMsg chan TransportMsg) (tc *TransportClient) {

	belogs.Debug("NewTcpClient():transportClientProcess:", transportClientProcess)
	tc = &TransportClient{}
	tc.connType = "tcp"
	tc.transportClientProcess = transportClientProcess
	tc.TransportMsg = transportMsg
	belogs.Info("NewTcpClient():tc:", tc)
	return tc
}

// server: 0.0.0.0:port
func NewTlsClient(tlsRootCrtFileName, tlsPublicCrtFileName, tlsPrivateKeyFileName string,
	transportClientProcess TransportClientProcess, transportMsg chan TransportMsg) (tc *TransportClient, err error) {

	belogs.Debug("NewTlsClient():transportClientProcess:", &transportClientProcess)
	tc = &TransportClient{}
	tc.connType = "tls"
	tc.transportClientProcess = transportClientProcess
	tc.TransportMsg = transportMsg

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
func (tc *TransportClient) StartTcpClient(server string) (err error) {
	belogs.Debug("TransportClient.StartTcpClient(): create client, server is  ", server)

	conn, err := net.DialTimeout("tcp", server, 60*time.Second)
	if err != nil {
		belogs.Error("TransportClient.StartTcpClient(): DialTimeout fail, server:", server, err)
		return err
	}
	belogs.Debug("TransportClient.StartTcpClient(): DialTimeout ok, server is  ", server)

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		belogs.Error("TransportClient.StartTcpClient(): conn cannot conver to tcpConn: ", conn.RemoteAddr().String(), err)
		return err
	}
	belogs.Debug("TransportClient.StartTcpClient(): tcpConn ok, server is  ", server)

	tc.transportConn = NewFromTcpConn(tcpConn)
	//active send to server, and receive from server, loop
	belogs.Debug("TransportClient.StartTcpClient(): NewFromTcpConn ok, server:", server, "   transportConn:", tc.transportConn.RemoteAddr().String())
	go tc.waitTransportMsg()

	// onConnect
	tc.onConnect()

	// onReceive
	go tc.onReceive()

	belogs.Info("TransportClient.StartTcpClient(): onReceive, server is  ", server, "  transportConn:", tc.transportConn.RemoteAddr().String())
	return nil
}

// server: **.**.**.**:port
func (tc *TransportClient) StartTlsClient(server string) (err error) {
	belogs.Debug("TransportClient.StartTlsClient(): create client, server is  ", server,
		"  tlsPublicCrtFileName:", tc.tlsPublicCrtFileName,
		"  tlsPrivateKeyFileName:", tc.tlsPrivateKeyFileName)

	cert, err := tls.LoadX509KeyPair(tc.tlsPublicCrtFileName, tc.tlsPrivateKeyFileName)
	if err != nil {
		belogs.Error("TransportClient.StartTlsClient(): LoadX509KeyPair fail: server:", server,
			"  tlsPublicCrtFileName:", tc.tlsPublicCrtFileName,
			"  tlsPrivateKeyFileName:", tc.tlsPrivateKeyFileName, err)
		return err
	}
	belogs.Debug("TransportClient.StartTlsClient(): LoadX509KeyPair ok, server is  ", server)

	rootCrtBytes, err := ioutil.ReadFile(tc.tlsRootCrtFileName)
	if err != nil {
		belogs.Error("TransportClient.StartTlsClient(): ReadFile tlsRootCrtFileName fail, server:", server,
			"  tlsRootCrtFileName:", tc.tlsRootCrtFileName, err)
		return err
	}
	belogs.Debug("TransportClient.StartTlsClient(): ReadFile ok, server is  ", server)

	rootCertPool := x509.NewCertPool()
	ok := rootCertPool.AppendCertsFromPEM(rootCrtBytes)
	if !ok {
		belogs.Error("TransportClient.StartTlsClient(): AppendCertsFromPEM tlsRootCrtFileName fail,server:", server,
			"  tlsRootCrtFileName:", tc.tlsRootCrtFileName, "  len(rootCrtBytes):", len(rootCrtBytes), err)
		return err
	}
	belogs.Debug("TransportClient.StartTlsClient(): AppendCertsFromPEM ok, server is  ", server)

	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            rootCertPool,
		InsecureSkipVerify: false,
	}
	dialer := &net.Dialer{Timeout: time.Duration(60) * time.Second}
	tlsConn, err := tls.DialWithDialer(dialer, "tcp", server, config)
	if err != nil {
		belogs.Error("TransportClient.StartTlsClient(): DialWithDialer fail, server:", server, err)
		return err
	}
	belogs.Debug("TransportClient.StartTlsClient(): DialWithDialer ok, server is  ", server)

	/*
		tlsConn, err := tls.Dial("tcp", server, config)
		if err != nil {
			belogs.Error("TransportClient.StartTlsClient(): Dial fail, server:", server, err)
			return err
		}
	*/
	tc.transportConn = NewFromTlsConn(tlsConn)
	belogs.Debug("TransportClient.StartTlsClient(): NewFromTlsConn ok, server:", server, "   transportConn:", tc.transportConn.RemoteAddr().String())
	//active send to server, and receive from server, loop
	go tc.waitTransportMsg()

	// onConnect
	tc.onConnect()

	// onReceive
	go tc.onReceive()

	belogs.Info("TransportClient.StartTlsClient(): onReceive, server is  ", server, "  transportConn:", tc.transportConn.RemoteAddr().String())

	return nil
}

func (tc *TransportClient) onReceive() (err error) {
	belogs.Debug("TransportClient.onReceive(): wait for onReceive, transportConn:", tc.transportConn.RemoteAddr().String())
	var leftData []byte
	// one packet
	buffer := make([]byte, 2048)
	// wait for new packet to read

	// when end onReceive, will onClose
	defer tc.onClose()
	for {
		start := time.Now()
		n, err := tc.transportConn.Read(buffer)
		//	if n == 0 {
		//		continue
		//	}
		if err != nil {
			if err == io.EOF {
				// is not error, just client close
				belogs.Debug("TransportClient.onReceive(): io.EOF, client close: ", tc.transportConn.RemoteAddr().String(), err)
				return nil
			}
			belogs.Error("TransportClient.onReceive(): Read fail or connect is closing, err ", tc.transportConn.RemoteAddr().String(), err)
			return err
		}

		belogs.Debug("TransportClient.onReceive(): Read n :", n, " from transportConn: ", tc.transportConn.RemoteAddr().String(),
			"  time(s):", time.Now().Sub(start))
		nextRwPolicy, leftData, err := tc.transportClientProcess.OnReceiveProcess(tc.transportConn, append(leftData, buffer[:n]...))
		belogs.Info("TransportClient.onReceive(): transportClientProcess.OnReceiveProcess, transportConn: ", tc.transportConn.RemoteAddr().String(), " receive n: ", n,
			"  len(leftData):", len(leftData), "  nextRwPolicy:", nextRwPolicy, "  time(s):", time.Now().Sub(start))
		if err != nil {
			belogs.Error("TransportClient.onReceive(): transportClientProcess.OnReceiveProcess  fail ,will close this transportConn : ", tc.transportConn.RemoteAddr().String(), err)
			return err
		}
		if nextRwPolicy == NEXT_RW_POLICY_END_READ {
			belogs.Info("TransportClient.onReceive():  nextRwPolicy is NEXT_RW_POLICY_END_READ, will close connect: ", tc.transportConn.RemoteAddr().String())
			return nil
		}

		// reset buffer
		buffer = make([]byte, 2048)
		belogs.Debug("TransportClient.onReceive(): will reset buffer and wait for Read from transportConn: ", tc.transportConn.RemoteAddr().String(),
			"  time(s):", time.Now().Sub(start))

	}

}

func (tc *TransportClient) onConnect() {
	// call process func onConnect
	tc.transportClientProcess.OnConnectProcess(tc.transportConn)
	belogs.Info("TransportClient.onConnect(): after OnConnectProcess, transportConn: ", tc.transportConn.RemoteAddr().String())
}

func (tc *TransportClient) onClose() {
	// close in the end
	belogs.Info("TransportClient.onClose(): transportConn: ", tc.transportConn.RemoteAddr().String())
	tc.transportClientProcess.OnCloseProcess(tc.transportConn)
	tc.transportConn.Close()

}

func (tc *TransportClient) SendMsg(transportMsg *TransportMsg) {

	belogs.Debug("TransportClient.SendMsg(): transportMsg:", jsonutil.MarshalJson(*transportMsg))
	tc.TransportMsg <- *transportMsg
}

func (tc *TransportClient) IsConnected() bool {

	belogs.Debug("TransportClient.IsConnected():")
	if tc.transportConn == nil {
		return false
	}
	b := tc.transportConn.IsConnected()
	belogs.Debug("TransportClient.IsConnected(): connected:", b)
	return b
}

func (tc *TransportClient) SendMsgForCloseConnect() {
	// send channel, and wait listener and conns end itself process and close loop
	belogs.Info("TransportClient.SendMsgForCloseConnect(): will close graceful")
	if tc.IsConnected() {
		transportMsg := &TransportMsg{
			MsgType: MSG_TYPE_CLIENT_CLOSE_CONNECT,
		}
		tc.SendMsg(transportMsg)
	}
}

func (tc *TransportClient) waitTransportMsg() (err error) {
	belogs.Debug("TransportClient.waitTransportMsg(): transportConn:", tc.transportConn.RemoteAddr().String())
	for {
		// wait next transportMsg: only error or NEXT_CONNECT_POLICY_CLOSE_** will end loop
		select {
		case transportMsg := <-tc.TransportMsg:
			belogs.Info("TransportClient.waitTransportMsg(): transportMsg:", jsonutil.MarshalJson(transportMsg),
				"  transportConn: ", tc.transportConn.RemoteAddr().String())

			switch transportMsg.MsgType {
			case MSG_TYPE_CLIENT_CLOSE_CONNECT:
				belogs.Info("TransportClient.waitTransportMsg(): msgType is MSG_TYPE_CLIENT_CLOSE_CONNECT,",
					" will close for transportConn: ", tc.transportConn.RemoteAddr().String(), " will return, close waitTransportMsg")
				tc.onClose()
				// end for/select
				// will return, close waitTransportMsg
				return nil
			case MSG_TYPE_COMMON_SEND_DATA:
				belogs.Info("TransportClient.waitTransportMsg(): msgType is MSG_TYPE_COMMON_SEND_DATA,",
					" will send to transportConn: ", tc.transportConn.RemoteAddr().String())
				sendData := transportMsg.SendData
				belogs.Debug("TransportClient.waitTransportMsg(): send to server:", tc.transportConn.RemoteAddr().String(),
					"   sendData:", convert.PrintBytesOneLine(sendData))

				// send data
				start := time.Now()
				n, err := tc.transportConn.Write(sendData)
				if err != nil {
					belogs.Error("TransportClient.waitTransportMsg(): Write fail, will close  transportConn:", tc.transportConn.RemoteAddr().String(), err)
					tc.onClose()
					return err
				}
				belogs.Info("TransportClient.waitTransportMsg(): Write to transportConn:", tc.transportConn.RemoteAddr().String(),
					"  len(sendData):", len(sendData), "  write n:", n,
					"  time(s):", time.Since(start))

			}
		}
	}

}
