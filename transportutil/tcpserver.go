package transportutil

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

// core struct:
type TcpServer struct {
	// state
	state uint64
	// tcp/tls/udp
	connType string

	// tls/tls/udp
	tcpConnsMutex sync.RWMutex
	tcpConns      map[string]*TcpConn

	// process
	tcpServerProcess TcpServerProcess

	// for tls
	tlsRootCrtFileName    string
	tlsPublicCrtFileName  string
	tlsPrivateKeyFileName string
	tlsVerifyClient       bool

	// https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
	// for close
	tcpListener   *TcpListener
	closeGraceful chan struct{}

	// for channel
	businessToConnMsg chan BusinessToConnMsg
}

func NewTcpServer(tcpServerProcess TcpServerProcess, businessToConnMsg chan BusinessToConnMsg) (ts *TcpServer) {

	belogs.Debug("NewTcpServer():tcpServerProcess:", tcpServerProcess)
	ts = &TcpServer{}
	ts.state = SERVER_STATE_INIT
	ts.connType = "tcp"
	ts.tcpConns = make(map[string]*TcpConn, 16)
	ts.tcpServerProcess = tcpServerProcess
	ts.closeGraceful = make(chan struct{})
	ts.businessToConnMsg = businessToConnMsg
	belogs.Debug("NewTcpServer():ts:", ts)
	return ts
}
func NewTlsServer(tlsRootCrtFileName, tlsPublicCrtFileName, tlsPrivateKeyFileName string, tlsVerifyClient bool,
	tcpServerProcess TcpServerProcess, businessToConnMsg chan BusinessToConnMsg) (ts *TcpServer, err error) {

	belogs.Debug("NewTcpServer():tlsRootCrtFileName:", tlsRootCrtFileName, "  tlsPublicCrtFileName:", tlsPublicCrtFileName,
		"   tlsPrivateKeyFileName:", tlsPrivateKeyFileName, "   tlsVerifyClient:", tlsVerifyClient,
		"   tcpServerProcess:", tcpServerProcess)
	ts = &TcpServer{}
	ts.state = SERVER_STATE_INIT
	ts.connType = "tls"
	ts.tcpConns = make(map[string]*TcpConn, 16)
	ts.closeGraceful = make(chan struct{})
	ts.businessToConnMsg = businessToConnMsg
	ts.tcpServerProcess = tcpServerProcess

	rootExists, _ := osutil.IsExists(tlsRootCrtFileName)
	if !rootExists {
		belogs.Error("NewTcpServer():root cer files not exists:", tlsRootCrtFileName)
		return nil, errors.New("root cer file is not exists")
	}
	publicExists, _ := osutil.IsExists(tlsPublicCrtFileName)
	if !publicExists {
		belogs.Error("NewTcpServer():public cer files not exists:", tlsPublicCrtFileName)
		return nil, errors.New("public cer file is not exists")
	}
	privateExists, _ := osutil.IsExists(tlsPrivateKeyFileName)
	if !privateExists {
		belogs.Error("NewTcpServer():private cer files not exists:", tlsPrivateKeyFileName)
		return nil, errors.New("private cer file is not exists")
	}

	ts.tlsRootCrtFileName = tlsRootCrtFileName
	ts.tlsPublicCrtFileName = tlsPublicCrtFileName
	ts.tlsPrivateKeyFileName = tlsPrivateKeyFileName
	ts.tlsVerifyClient = tlsVerifyClient
	belogs.Debug("NewTlsServer():ts:", &ts)
	return ts, nil
}

// port: `8888` --> `0.0.0.0:8888`
func (ts *TcpServer) StartTcpServer(port string) (err error) {
	tcpServer, err := net.ResolveTCPAddr("tcp", "0.0.0.0:"+port)
	if err != nil {
		belogs.Error("StartTcpServer(): tcpserver  ResolveTCPAddr fail, port:", port, err)
		return err
	}
	belogs.Debug("StartTcpServer(): ResolveTCPAddr ok,  port:", port)

	listener, err := net.ListenTCP("tcp", tcpServer)
	if err != nil {
		belogs.Error("StartTcpServer(): tcpserver  ListenTCP fail, port:", port, err)
		return err
	}
	belogs.Debug("StartTcpServer(): ListenTCP ok,  port:", port)

	// get tcpListener
	ts.tcpListener, err = NewFromTcpListener(listener)
	if err != nil {
		belogs.Error("StartTcpServer(): tcpserver  NewFromTcpListener fail, port:", port, err)
		return err
	}
	belogs.Info("StartTcpServer(): tcpserver  create server ok, port:", port, "  will accept client")

	go ts.waitBusinessToConnMsg()

	// wait new conn
	ts.acceptNewConn()
	return nil
}

// port: `8888` --> `:8888`
func (ts *TcpServer) StartTlsServer(port string) (err error) {

	belogs.Debug("StartTlsServer(): tlsserver  port:", port)
	cert, err := tls.LoadX509KeyPair(ts.tlsPublicCrtFileName, ts.tlsPrivateKeyFileName)
	if err != nil {
		belogs.Error("StartTlsServer(): tlsserver  LoadX509KeyPair fail: port:", port,
			"  tlsPublicCrtFileName, tlsPrivateKeyFileName:", ts.tlsPublicCrtFileName, ts.tlsPrivateKeyFileName, err)
		return err
	}
	belogs.Debug("StartTlsServer(): tlsserver  cert:", ts.tlsPublicCrtFileName, ts.tlsPrivateKeyFileName)

	rootCrtBytes, err := ioutil.ReadFile(ts.tlsRootCrtFileName)
	if err != nil {
		belogs.Error("StartTlsServer(): tlsserver  ReadFile tlsRootCrtFileName fail, port:", port,
			"  tlsRootCrtFileName:", ts.tlsRootCrtFileName, err)
		return err
	}
	belogs.Debug("StartTlsServer(): tlsserver  len(rootCrtBytes):", len(rootCrtBytes), "  tlsRootCrtFileName:", ts.tlsRootCrtFileName)

	rootCertPool := x509.NewCertPool()
	ok := rootCertPool.AppendCertsFromPEM(rootCrtBytes)
	if !ok {
		belogs.Error("StartTlsServer(): tlsserver  AppendCertsFromPEM tlsRootCrtFileName fail,port:", port,
			"  tlsRootCrtFileName:", ts.tlsRootCrtFileName, "  len(rootCrtBytes):", len(rootCrtBytes), err)
		return err
	}
	belogs.Debug("StartTlsServer(): tlsserver  AppendCertsFromPEM len(rootCrtBytes):", len(rootCrtBytes), "  tlsRootCrtFileName:", ts.tlsRootCrtFileName)

	clientAuthType := tls.NoClientCert
	if ts.tlsVerifyClient {
		clientAuthType = tls.RequireAndVerifyClientCert
	}
	belogs.Debug("StartTlsServer(): tlsserver clientAuthType:", clientAuthType)

	// https://stackoverflow.com/questions/63676241/how-to-set-setkeepaliveperiod-on-a-tls-conn
	setTCPKeepAlive := func(clientHello *tls.ClientHelloInfo) (*tls.Config, error) {
		// Check that the underlying connection really is TCP.
		if tcpConn, ok := clientHello.Conn.(*net.TCPConn); ok {
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(time.Second * 300)
			belogs.Debug("StartTlsServer(): tlsserver SetKeepAlive:")
		}
		// Make sure to return nil, nil to let the caller fall back on the default behavior.
		return nil, nil
	}
	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		ClientAuth:         clientAuthType,
		RootCAs:            rootCertPool,
		InsecureSkipVerify: false,
		GetConfigForClient: setTCPKeepAlive,
	}
	listener, err := tls.Listen("tcp", ":"+port, config)
	if err != nil {
		belogs.Error("StartTlsServer(): tlsserver  Listen fail, port:", port, err)
		return err
	}
	belogs.Debug("StartTlsServer(): Listen ok,  port:", port)

	// get tcpListener
	ts.tcpListener, err = NewFromTlsListener(listener)
	if err != nil {
		belogs.Error("StartTlsServer(): tlsserver  NewFromTlsListener fail, port: ", port, err)
		return err
	}
	belogs.Info("StartTlsServer(): tlsserver  create server ok, port:", port, "  will accept client")

	go ts.waitBusinessToConnMsg()

	// wait new conn
	ts.acceptNewConn()
	return nil
}

func (ts *TcpServer) acceptNewConn() {

	defer ts.tcpListener.Close()
	belogs.Debug("TcpServer.acceptNewConn(): will accept client")
	ts.state = SERVER_STATE_RUNNING
	for {
		tcpConn, err := ts.tcpListener.Accept()
		if err != nil {
			select {
			case <-ts.closeGraceful:
				belogs.Info("TcpServer.acceptNewConn(): Accept remote fail and closeGraceful, will return: ", err)
				return
			default:
				belogs.Error("TcpServer.acceptNewConn(): Accept remote fail: ", err)
				continue
			}

		}
		belogs.Info("TcpServer.acceptNewConn():  Accept remote: ", tcpConn.RemoteAddr().String())

		ts.onConnect(tcpConn)
		// call func to process tcpConn
		go ts.receiveAndSend(tcpConn)

	}
}

// close directly
func (ts *TcpServer) receiveAndSend(tcpConn *TcpConn) {

	defer ts.onClose(tcpConn)

	var leftData []byte
	// one packet
	buffer := make([]byte, 2048)
	belogs.Debug("TcpServer.receiveAndSend(): recive from tcpConn: ", tcpConn.RemoteAddr().String())

	// wait for new packet to read
	for {
		start := time.Now()
		n, err := tcpConn.Read(buffer)
		belogs.Debug("TcpServer.receiveAndSend():server read: Read from : ", tcpConn.RemoteAddr().String(), "  read n:", n)
		//	if n == 0 {
		//		continue
		//	}
		if err != nil {
			if err == io.EOF {
				// is not error, just client close
				belogs.Info("TcpServer.receiveAndSend(): Read io.EOF, client close: ", tcpConn.RemoteAddr().String(), err)
				return
			}
			belogs.Error("TcpServer.receiveAndSend(): Read fail, err ", tcpConn.RemoteAddr().String(), err)
			return
		}

		// call process func OnReceiveAndSend
		// copy to leftData
		belogs.Debug("TcpServer.receiveAndSend(): tcpConn: ", tcpConn.RemoteAddr().String(),
			" , Read n:", n, "  time(s):", time.Since(start))
		nextConnectPolicy, leftData, err := ts.tcpServerProcess.OnReceiveAndSendProcess(tcpConn, append(leftData, buffer[:n]...))
		belogs.Debug("TcpServer.receiveAndSend(): after OnReceiveAndSendProcess,server tcpConn: ", tcpConn.RemoteAddr().String(), " receive n: ", n,
			"  len(leftData):", len(leftData), "  time(s):", time.Since(start))
		if err != nil {
			belogs.Error("TcpServer.receiveAndSend(): OnReceiveAndSendProcess fail ,will remove this tcpConn : ", tcpConn.RemoteAddr().String(), err)
			return
		}

		if nextConnectPolicy == NEXT_CONNECT_POLICY_CLOSE_GRACEFUL ||
			nextConnectPolicy == NEXT_CONNECT_POLICY_CLOSE_FORCIBLE {
			belogs.Info("TcpServer.receiveAndSend(): nextConnectPolicy close,  return : ", tcpConn.RemoteAddr().String(), nextConnectPolicy)
			return
		}
		// check state
		if ts.state != SERVER_STATE_RUNNING {
			belogs.Debug("TcpServer.receiveAndSend(): state is not running, will close from tcpConn: ", tcpConn.RemoteAddr().String(),
				"  state:", ts.state, "  time(s):", time.Now().Sub(start))
			return
		}
		belogs.Debug("TcpServer.receiveAndSend(): will wait for Read from tcpConn: ", tcpConn.RemoteAddr().String(),
			"  time(s):", time.Now().Sub(start))

	}
}

func (ts *TcpServer) onConnect(tcpConn *TcpConn) {
	start := time.Now()
	belogs.Debug("TcpServer.onConnect(): new tcpConn: ", tcpConn.RemoteAddr().String())

	// add new tcpConn to tcpConns
	ts.tcpConnsMutex.Lock()
	defer ts.tcpConnsMutex.Unlock()

	connKey := GetTcpConnKey(tcpConn)
	ts.tcpConns[connKey] = tcpConn
	belogs.Debug("TcpServer.onConnect(): tcpConn: ", tcpConn.RemoteAddr().String(), ", connKey:", connKey, "  ts.tcpConns: ", ts.tcpConns)
	ts.tcpServerProcess.OnConnectProcess(tcpConn)
	belogs.Info("TcpServer.onConnect(): add tcpConn: ", tcpConn.RemoteAddr().String(), "   len(tcpConns): ", len(ts.tcpConns), "   time(s):", time.Now().Sub(start).Seconds())

}

func (ts *TcpServer) onClose(tcpConn *TcpConn) {
	// close in the end
	if tcpConn == nil {
		return
	}

	defer func() {
		tcpConn.Close()
		//tcpConn.SetNil()
	}()

	start := time.Now()
	// call process func onClose
	belogs.Debug("TcpServer.onClose(): tcpConn: ", tcpConn.RemoteAddr().String(), "   call process func: onClose ")

	// remove tcpConn from tcpConns
	ts.tcpConnsMutex.Lock()
	defer ts.tcpConnsMutex.Unlock()
	belogs.Debug("TcpServer.onClose(): will close old tcpConns, tcpConn: ", tcpConn.RemoteAddr().String(), "   old len(tcpConns): ", len(ts.tcpConns))
	delete(ts.tcpConns, GetTcpConnKey(tcpConn))
	ts.tcpServerProcess.OnCloseProcess(tcpConn)
	belogs.Info("TcpServer.onClose(): new len(tcpConns): ", len(ts.tcpConns), "  time(s):", time.Now().Sub(start).Seconds())
}

func (ts *TcpServer) SendBusinessToConnMsg(businessToConnMsg *BusinessToConnMsg) {

	belogs.Debug("TcpServer.SendBusinessToConnMsg():, businessToConnMsg:", jsonutil.MarshalJson(*businessToConnMsg))
	ts.businessToConnMsg <- *businessToConnMsg
}

// businessToConnMsgType:BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_GRACEFUL, //
// BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_FORCIBLE
func (ts *TcpServer) SendMsgForCloseConnect(businessToConnMsgType string, serverConnKey string) {
	// send channel, and wait listener and conns end itself process and close loop
	belogs.Info("TcpServer.SendMsgForCloseConnect(): will close, businessToConnMsgType:", businessToConnMsgType, "  serverConnKey:", serverConnKey)
	businessToConnMsg := &BusinessToConnMsg{
		BusinessToConnMsgType: businessToConnMsgType,
		ServerConnKey:         serverConnKey,
	}
	ts.SendBusinessToConnMsg(businessToConnMsg)
}

func (ts *TcpServer) waitBusinessToConnMsg() {
	belogs.Debug("TcpServer.waitBusinessToConnMsg(): will waitBusinessToConnMsg")
	for {
		select {
		case businessToConnMsg := <-ts.businessToConnMsg:
			belogs.Info("TcpServer.waitBusinessToConnMsg(): businessToConnMsg:", jsonutil.MarshalJson(businessToConnMsg))

			switch businessToConnMsg.BusinessToConnMsgType {
			case BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_FORCIBLE:
				// ignore conns's writing/reading, just close
				belogs.Info("TcpServer.waitBusinessToConnMsg(): businessToConnMsgType is BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_FORCIBLE")
				// just close
				ts.state = SERVER_STATE_CLOSING
				ts.tcpListener.Close()
				for connKey := range ts.tcpConns {
					ts.onClose(ts.tcpConns[connKey])
				}
				close(ts.closeGraceful)
				belogs.Info("TcpServer.waitBusinessToConnMsg(): will close server forcible, will return waitBusinessToConnMsg:")
				// end for/select
				ts.state = SERVER_STATE_CLOSED
				// will return, close waitBusinessToConnMsg
				return
			case BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_GRACEFUL:
				// close and wait connect.Read and Accept
				belogs.Info("TcpServer.waitBusinessToConnMsg(): businessToConnMsgType is BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_GRACEFUL")
				ts.state = SERVER_STATE_CLOSING
				close(ts.closeGraceful)
				time.Sleep(5 * time.Second)
				ts.tcpListener.Close()
				for connKey := range ts.tcpConns {
					ts.onClose(ts.tcpConns[connKey])
				}
				belogs.Info("TcpServer.waitBusinessToConnMsg(): will close server graceful, will return waitBusinessToConnMsg:")
				// end for/select
				ts.state = SERVER_STATE_CLOSED
				// will return, close waitBusinessToConnMsg
				return
			case BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_GRACEFUL:
				belogs.Info("TcpServer.waitBusinessToConnMsg(): businessToConnMsgType is BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_GRACEFUL")
				fallthrough
			case BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_FORCIBLE:
				// close and wait connect.Read and Accept
				belogs.Info("TcpServer.waitBusinessToConnMsg(): businessToConnMsgType is BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_FORCIBLE")
				if len(businessToConnMsg.ServerConnKey) > 0 {
					ts.onClose(ts.tcpConns[businessToConnMsg.ServerConnKey])
				}
				belogs.Info("TcpServer.waitBusinessToConnMsg(): close connect, serverConnKey:", businessToConnMsg.ServerConnKey)
				// close one connect, no return
				// return
			case BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_DATA:

				serverConnKey := businessToConnMsg.ServerConnKey
				sendData := businessToConnMsg.SendData
				belogs.Info("TcpServer.waitBusinessToConnMsg(): businessToConnMsgType is BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_DATA, serverConnKey:", serverConnKey,
					"  sendData:", convert.PrintBytesOneLine(sendData))
				err := ts.activeSend(serverConnKey, sendData)
				if err != nil {
					belogs.Error("TcpServer.waitBusinessToConnMsg(): activeSend fail, serverConnKey:", serverConnKey,
						"  sendData:", convert.PrintBytesOneLine(sendData), err)
					// err, no return
					// return
				} else {
					belogs.Info("TcpServer.waitBusinessToConnMsg(): activeSend ok, serverConnKey:", serverConnKey,
						"  sendData:", convert.PrintBytesOneLine(sendData))
				}
			}
		}
	}
}

// connKey is "": send to all clients
// connKey is net.Conn.Address.String(): send this client
func (ts *TcpServer) activeSend(connKey string, sendData []byte) (err error) {
	ts.tcpConnsMutex.RLock()
	defer ts.tcpConnsMutex.RUnlock()
	start := time.Now()

	belogs.Debug("TcpServer.activeSend(): ,len(sendData):", len(sendData),
		"   tcpConns: ", ts.tcpConns, "  connKey:", connKey)
	if len(connKey) == 0 {
		belogs.Debug("TcpServer.activeSend(): to all, len(sendData):", len(sendData), "   len(tcpConns): ", len(ts.tcpConns))
		for i := range ts.tcpConns {
			belogs.Debug("TcpServer.activeSend(): to all, client: ", i, "    ts.tcpConns[i]:", ts.tcpConns[i].RemoteAddr().String())
			startOne := time.Now()
			n, err := ts.tcpConns[i].Write(sendData)
			if err != nil {
				belogs.Error("TcpServer.activeSend(): server to all, tcpConn.Write fail, will ignore, tcpConn:", ts.tcpConns[i].RemoteAddr().String(),
					"   n:", n, "   sendData:", convert.PrintBytesOneLine(sendData), "   time(s):", time.Since(startOne), err)
				continue
			} else {
				belogs.Info("TcpServer.activeSend(): server to all, tcpConn.Write ok, tcpConn:", ts.tcpConns[i].RemoteAddr().String(),
					"   n:", n, "   sendData:", convert.PrintBytesOneLine(sendData), "   time(s):", time.Since(startOne))
			}
		}
		belogs.Info("TcpServer.activeSend(): send to all clients ok,  len(sendData):", len(sendData), "   len(tcpConns): ", len(ts.tcpConns),
			"  time(s):", time.Now().Sub(start).Seconds())
		return
	} else {
		belogs.Debug("TcpServer.activeSend(): to connKey:", connKey, "   ts.tcpConns:", ts.tcpConns)
		if tcpConn, ok := ts.tcpConns[connKey]; ok {
			startOne := time.Now()
			belogs.Debug("TcpServer.activeSend(): found connKey: ", connKey, "   sendData:", convert.PrintBytesOneLine(sendData))
			n, err := tcpConn.Write(sendData)
			if err != nil {
				belogs.Error("TcpServer.activeSend(): to ", connKey, " tcpConn.Write fail: tcpConn:", tcpConn.RemoteAddr().String(),
					"   n:", n, "   sendData:", convert.PrintBytesOneLine(sendData), "   time(s):", time.Since(startOne), err)
			} else {
				belogs.Info("TcpServer.activeSend(): to ", connKey, " tcpConn.Write ok, tcpConn:", tcpConn.RemoteAddr().String(),
					"   n:", n, "   sendData:", convert.PrintBytesOneLine(sendData), "   time(s):", time.Since(startOne))
			}
		} else {
			belogs.Error("TcpServer.activeSend(): not found connKey: ", connKey, " fail: tcpConn:", tcpConn.RemoteAddr().String(),
				"   sendData:", convert.PrintBytesOneLine(sendData))
		}
		belogs.Info("TcpServer.activeSend(): send to connKey ok,  len(sendData):", len(sendData), "   connKey: ", connKey,
			"  time(s):", time.Now().Sub(start).Seconds())
		return
	}

}
