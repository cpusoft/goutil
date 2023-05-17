package tcptlsutil

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

// core struct: Start/onConnect/receiveAndSend....
type TcpTlsServer struct {
	// state
	state uint64
	// both tcp and tls
	isTcpServer         bool
	tcpTlsConns         map[string]*TcpTlsConn // map[addr]*net.TCPConn
	tcpTlsConnsMutex    sync.RWMutex
	tcpTlsServerProcess TcpTlsServerProcess

	// for tls
	tlsRootCrtFileName    string
	tlsPublicCrtFileName  string
	tlsPrivateKeyFileName string
	tlsVerifyClient       bool

	// https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
	// for close
	tcpTlsListener *TcpTlsListener
	closeGraceful  chan struct{}

	// for channel
	TcpTlsMsg chan TcpTlsMsg
}

//
func NewTcpServer(tcpTlsServerProcess TcpTlsServerProcess, tcpTlsMsg chan TcpTlsMsg) (ts *TcpTlsServer) {

	belogs.Debug("NewTcpServer():tcpTlsServerProcess:", tcpTlsServerProcess)
	ts = &TcpTlsServer{}
	ts.state = SERVER_STATE_INIT
	ts.isTcpServer = true
	ts.tcpTlsConns = make(map[string]*TcpTlsConn, 16)
	ts.tcpTlsServerProcess = tcpTlsServerProcess
	ts.closeGraceful = make(chan struct{})
	ts.TcpTlsMsg = tcpTlsMsg
	belogs.Debug("NewTcpServer():ts:", ts)
	return ts
}
func NewTlsServer(tlsRootCrtFileName, tlsPublicCrtFileName, tlsPrivateKeyFileName string, tlsVerifyClient bool,
	tcpTlsServerProcess TcpTlsServerProcess, tcpTlsMsg chan TcpTlsMsg) (ts *TcpTlsServer, err error) {

	belogs.Debug("NewTcpServer():tlsRootCrtFileName:", tlsRootCrtFileName, "  tlsPublicCrtFileName:", tlsPublicCrtFileName,
		"   tlsPrivateKeyFileName:", tlsPrivateKeyFileName, "   tlsVerifyClient:", tlsVerifyClient,
		"   tcpTlsServerProcess:", tcpTlsServerProcess)
	ts = &TcpTlsServer{}
	ts.state = SERVER_STATE_INIT
	ts.isTcpServer = false
	ts.tcpTlsConns = make(map[string]*TcpTlsConn, 16)
	ts.closeGraceful = make(chan struct{})
	ts.TcpTlsMsg = tcpTlsMsg
	ts.tcpTlsServerProcess = tcpTlsServerProcess

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
func (ts *TcpTlsServer) StartTcpServer(port string) (err error) {
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

	// get tcpTlsListener
	ts.tcpTlsListener, err = NewFromTcpListener(listener)
	if err != nil {
		belogs.Error("StartTcpServer(): tcpserver  NewFromTcpListener fail, port:", port, err)
		return err
	}
	belogs.Info("StartTcpServer(): tcpserver  create server ok, port:", port, "  will accept client")

	go ts.waitTcpTlsMsg()

	// wait new conn
	ts.acceptNewConn()
	return nil
}

// port: `8888` --> `:8888`
func (ts *TcpTlsServer) StartTlsServer(port string) (err error) {

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

	// get tcpTlsListener
	ts.tcpTlsListener, err = NewFromTlsListener(listener)
	if err != nil {
		belogs.Error("StartTlsServer(): tlsserver  NewFromTlsListener fail, port: ", port, err)
		return err
	}
	belogs.Info("StartTlsServer(): tlsserver  create server ok, port:", port, "  will accept client")

	go ts.waitTcpTlsMsg()

	// wait new conn
	ts.acceptNewConn()
	return nil
}

func (ts *TcpTlsServer) acceptNewConn() {

	defer ts.tcpTlsListener.Close()
	belogs.Debug("acceptNewConn(): will accept client")
	ts.state = SERVER_STATE_RUNNING
	for {
		tcpTlsConn, err := ts.tcpTlsListener.Accept()
		if err != nil {
			select {
			case <-ts.closeGraceful:
				belogs.Info("acceptNewConn(): Accept remote fail and closeGraceful, will return: ", err)
				return
			default:
				belogs.Error("acceptNewConn(): Accept remote fail: ", err)
				continue
			}

		}
		belogs.Info("acceptNewConn():  Accept remote: ", tcpTlsConn.RemoteAddr().String())

		ts.onConnect(tcpTlsConn)
		// call func to process tcpTlsConn
		go ts.receiveAndSend(tcpTlsConn)

	}
}

// close directly
func (ts *TcpTlsServer) receiveAndSend(tcpTlsConn *TcpTlsConn) {

	defer ts.onClose(tcpTlsConn)

	var leftData []byte
	// one packet
	buffer := make([]byte, 2048)
	belogs.Debug("receiveAndSend(): recive from tcpTlsConn: ", tcpTlsConn.RemoteAddr().String())

	// wait for new packet to read
	for {
		start := time.Now()
		n, err := tcpTlsConn.Read(buffer)
		belogs.Debug("receiveAndSend():server read: Read from : ", tcpTlsConn.RemoteAddr().String(), "  read n:", n)
		//	if n == 0 {
		//		continue
		//	}
		if err != nil {
			if err == io.EOF {
				// is not error, just client close
				belogs.Info("receiveAndSend(): tcptlsserver Read io.EOF, client close: ", tcpTlsConn.RemoteAddr().String(), err)
				return
			}
			belogs.Error("receiveAndSend(): tcptlsserver Read fail, err ", tcpTlsConn.RemoteAddr().String(), err)
			return
		}

		// call process func OnReceiveAndSend
		// copy to leftData
		belogs.Debug("receiveAndSend(): tcptlsserver tcpTlsConn: ", tcpTlsConn.RemoteAddr().String(),
			" , Read n:", n, "  time(s):", time.Since(start))
		nextConnectPolicy, leftData, err := ts.tcpTlsServerProcess.OnReceiveAndSendProcess(tcpTlsConn, append(leftData, buffer[:n]...))
		belogs.Debug("receiveAndSend(): tcptlsserver  after OnReceiveAndSendProcess,server tcpTlsConn: ", tcpTlsConn.RemoteAddr().String(), " receive n: ", n,
			"  len(leftData):", len(leftData), "  time(s):", time.Since(start))
		if err != nil {
			belogs.Error("receiveAndSend(): tcptlsserver OnReceiveAndSendProcess fail ,will remove this tcpTlsConn : ", tcpTlsConn.RemoteAddr().String(), err)
			return
		}

		if nextConnectPolicy == NEXT_CONNECT_POLICY_CLOSE_GRACEFUL ||
			nextConnectPolicy == NEXT_CONNECT_POLICY_CLOSE_FORCIBLE {
			belogs.Info("receiveAndSend(): tcptlsserver  nextConnectPolicy close,  return : ", tcpTlsConn.RemoteAddr().String(), nextConnectPolicy)
			return
		}
		// check state
		if ts.state != SERVER_STATE_RUNNING {
			belogs.Debug("receiveAndSend(): state is not running, will close from tcpTlsConn: ", tcpTlsConn.RemoteAddr().String(),
				"  state:", ts.state, "  time(s):", time.Since(start))
			return
		}
		belogs.Debug("receiveAndSend(): tcptlsserver, will wait for Read from tcpTlsConn: ", tcpTlsConn.RemoteAddr().String(),
			"  time(s):", time.Since(start))

	}
}

/*
// close gracefully
func (ts *TcpTlsServer) receiveAndSend(tcpTlsConn *TcpTlsConn) {

	defer ts.onClose(tcpTlsConn)

	var leftData []byte
	// one packet
	buffer := make([]byte, 2048)
	belogs.Debug("receiveAndSend(): recive from tcpTlsConn: ", tcpTlsConn.RemoteAddr().String())

	// wait for new packet to read
	//https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
	//https://stackoverflow.com/questions/66755407/cancelling-a-net-listener-via-context-in-golang
ReadLoop:
	for {
		select {
		case <-ts.closeGraceful:
			belogs.Info("receiveAndSend(): tcptlsserver closeGraceful, will return: ", tcpTlsConn.RemoteAddr().String())
			return
		default:
			tcpTlsConn.SetDeadline(time.Now().Add(60 * time.Second))
			start := time.Now()
			n, err := tcpTlsConn.Read(buffer)
			//	if n == 0 {
			//		continue
			//	}
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					belogs.Debug("receiveAndSend(): tcptlsserver Timeout :", tcpTlsConn.RemoteAddr().String(), opErr) //should //
					continue ReadLoop
				} else if err == io.EOF {
					// is not error, just client close
					belogs.Info("receiveAndSend(): tcptlsserver Read io.EOF, client close: ", tcpTlsConn.RemoteAddr().String(), err)
					return
				}
				belogs.Error("receiveAndSend(): tcptlsserver Read fail, err ", tcpTlsConn.RemoteAddr().String(), err)
				return
			}

			// call process func OnReceiveAndSend
			// copy to leftData
			belogs.Debug("receiveAndSend(): tcptlsserver tcpTlsConn: ", tcpTlsConn.RemoteAddr().String(),
				" , Read n:", n, "  time(s):", time.Since(start))
			nextConnectPolicy, leftData, err := ts.tcpTlsServerProcess.OnReceiveAndSendProcess(tcpTlsConn, append(leftData, buffer[:n]...))
			belogs.Debug("receiveAndSend(): tcptlsserver  after OnReceiveAndSendProcess,server tcpTlsConn: ", tcpTlsConn.RemoteAddr().String(), " receive n: ", n,
				"  len(leftData):", len(leftData), "  time(s):", time.Since(start))
			if err != nil {
				belogs.Error("receiveAndSend(): tcptlsserver OnReceiveAndSendProcess fail ,will remove this tcpTlsConn : ", tcpTlsConn.RemoteAddr().String(), err)
				return
			}

			if nextConnectPolicy == NEXT_CONNECT_POLICY_CLOSE_GRACEFUL ||
				nextConnectPolicy == NEXT_CONNECT_POLICY_CLOSE_FORCIBLE {
				belogs.Info("receiveAndSend(): tcptlsserver  nextConnectPolicy close,  return : ", tcpTlsConn.RemoteAddr().String(), nextConnectPolicy)
				return
			}
			// reset buffer
			buffer = make([]byte, 2048)
			belogs.Debug("receiveAndSend(): tcptlsserver, will reset buffer, tcpTlsConn: ", tcpTlsConn.RemoteAddr().String())

			belogs.Debug("onReceive(): tcptlsserver, will wait for Read from tcpTlsConn: ", tcpTlsConn.RemoteAddr().String(),
				"  time(s):", time.Since(start))
		}
	}
}
*/
func (ts *TcpTlsServer) onConnect(tcpTlsConn *TcpTlsConn) {
	start := time.Now()
	belogs.Debug("onConnect(): new tcpTlsConn: ", tcpTlsConn.RemoteAddr().String())

	// add new tcpTlsConn to tcpTlsConns
	ts.tcpTlsConnsMutex.Lock()
	defer ts.tcpTlsConnsMutex.Unlock()

	connKey := GetConnKey(tcpTlsConn)
	ts.tcpTlsConns[connKey] = tcpTlsConn
	belogs.Debug("onConnect(): tcptlsserver tcpTlsConn: ", tcpTlsConn.RemoteAddr().String(), ", connKey:", connKey, "  ts.tcpTlsConns: ", ts.tcpTlsConns)
	ts.tcpTlsServerProcess.OnConnectProcess(tcpTlsConn)
	belogs.Info("onConnect(): tcptlsserver add tcpTlsConn: ", tcpTlsConn.RemoteAddr().String(), "   len(tcpTlsConns): ", len(ts.tcpTlsConns), "   time(s):", time.Since(start).Seconds())

}

func (ts *TcpTlsServer) onClose(tcpTlsConn *TcpTlsConn) {
	// close in the end
	if tcpTlsConn == nil {
		return
	}

	defer func() {
		tcpTlsConn.Close()
		//tcpTlsConn.SetNil()
	}()

	start := time.Now()
	// call process func onClose
	belogs.Debug("onClose(): tcptlsserver tcpTlsConn: ", tcpTlsConn.RemoteAddr().String(), "   call process func: onClose ")

	// remove tcpTlsConn from tcpConns
	ts.tcpTlsConnsMutex.Lock()
	defer ts.tcpTlsConnsMutex.Unlock()
	belogs.Debug("onClose(): tcptlsserver will close old tcpTlsConns, tcpTlsConn: ", tcpTlsConn.RemoteAddr().String(), "   old len(tcpTlsConns): ", len(ts.tcpTlsConns))
	delete(ts.tcpTlsConns, GetConnKey(tcpTlsConn))
	ts.tcpTlsServerProcess.OnCloseProcess(tcpTlsConn)
	belogs.Info("onClose(): tcptlsserver new len(tcpTlsConns): ", len(ts.tcpTlsConns), "  time(s):", time.Since(start).Seconds())
}

func (ts *TcpTlsServer) SendMsg(tcpTlsMsg *TcpTlsMsg) {

	belogs.Debug("SendMsg(): tcptlsserver, tcpTlsMsg:", jsonutil.MarshalJson(*tcpTlsMsg))
	ts.TcpTlsMsg <- *tcpTlsMsg
}

// msgType:MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_GRACEFUL, //
// MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_FORCIBLE
func (ts *TcpTlsServer) SendMsgForCloseConnect(msgType uint64, serverConnKey string) {
	// send channel, and wait listener and conns end itself process and close loop
	belogs.Info("SendMsgForCloseConnect(): tcptlsserver will close, msgType:", msgType, "  serverConnKey:", serverConnKey)
	tcpTlsMsg := &TcpTlsMsg{
		MsgType:       msgType,
		ServerConnKey: serverConnKey,
	}
	ts.SendMsg(tcpTlsMsg)
}

func (ts *TcpTlsServer) waitTcpTlsMsg() {
	belogs.Debug("waitTcpTlsMsg(): tcptlsserver will waitTcpTlsMsg")
	for {
		select {
		case tcpTlsMsg := <-ts.TcpTlsMsg:
			belogs.Info("waitTcpTlsMsg(): tcptlsserver tcpTlsMsg:", jsonutil.MarshalJson(tcpTlsMsg))

			switch tcpTlsMsg.MsgType {
			case MSG_TYPE_SERVER_CLOSE_FORCIBLE:
				// ignore conns's writing/reading, just close
				belogs.Info("waitTcpTlsMsg(): tcptlsserver msgType is MSG_TYPE_SERVER_CLOSE_FORCIBLE")
				// just close
				ts.state = SERVER_STATE_CLOSING
				ts.tcpTlsListener.Close()
				for connKey := range ts.tcpTlsConns {
					ts.onClose(ts.tcpTlsConns[connKey])
				}
				close(ts.closeGraceful)
				belogs.Info("waitTcpTlsMsg(): tcptlsserver will close server forcible, will return waitTcpTlsMsg:")
				// end for/select
				ts.state = SERVER_STATE_CLOSED
				// will return, close waitTcpTlsMsg
				return
			case MSG_TYPE_SERVER_CLOSE_GRACEFUL:
				// close and wait connect.Read and Accept
				belogs.Info("waitTcpTlsMsg(): tcptlsserver msgType is MSG_TYPE_SERVER_CLOSE_GRACEFUL")
				ts.state = SERVER_STATE_CLOSING
				close(ts.closeGraceful)
				time.Sleep(5 * time.Second)
				ts.tcpTlsListener.Close()
				for connKey := range ts.tcpTlsConns {
					ts.onClose(ts.tcpTlsConns[connKey])
				}
				belogs.Info("waitTcpTlsMsg(): tcptlsserver will close server graceful, will return waitTcpTlsMsg:")
				// end for/select
				ts.state = SERVER_STATE_CLOSED
				// will return, close waitTcpTlsMsg
				return
			case MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_GRACEFUL:
				belogs.Info("waitTcpTlsMsg(): tcptlsserver msgType is MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_GRACEFUL")
				fallthrough
			case MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_FORCIBLE:
				// close and wait connect.Read and Accept
				belogs.Info("waitTcpTlsMsg(): tcptlsserver msgType is MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_FORCIBLE")
				if len(tcpTlsMsg.ServerConnKey) > 0 {
					ts.onClose(ts.tcpTlsConns[tcpTlsMsg.ServerConnKey])
				}
				belogs.Info("waitTcpTlsMsg(): tcptlsserver close connect, serverConnKey:", tcpTlsMsg.ServerConnKey)
				// close one connect, no return
				// return
			case MSG_TYPE_COMMON_SEND_DATA:

				serverConnKey := tcpTlsMsg.ServerConnKey
				sendData := tcpTlsMsg.SendData
				belogs.Info("waitTcpTlsMsg(): tcptlsserver msgType is MSG_TYPE_COMMON_SEND_DATA, serverConnKey:", serverConnKey,
					"  sendData:", convert.PrintBytesOneLine(sendData))
				err := ts.activeSend(serverConnKey, sendData)
				if err != nil {
					belogs.Error("waitTcpTlsMsg(): tcptlsserver activeSend fail, serverConnKey:", serverConnKey,
						"  sendData:", convert.PrintBytesOneLine(sendData), err)
					// err, no return
					// return
				} else {
					belogs.Info("waitTcpTlsMsg(): tcptlsserver activeSend ok, serverConnKey:", serverConnKey,
						"  sendData:", convert.PrintBytesOneLine(sendData))
				}
			}
		}
	}
}

// connKey is "": send to all clients
// connKey is net.Conn.Address.String(): send this client
func (ts *TcpTlsServer) activeSend(connKey string, sendData []byte) (err error) {
	ts.tcpTlsConnsMutex.RLock()
	defer ts.tcpTlsConnsMutex.RUnlock()
	start := time.Now()

	belogs.Debug("activeSend(): tcptlsserver ,len(sendData):", len(sendData),
		"   tcpTlsConns: ", ts.tcpTlsConns, "  connKey:", connKey)
	if len(connKey) == 0 {
		belogs.Debug("activeSend(): tcptlsserver to all, len(sendData):", len(sendData), "   len(tcpConns): ", len(ts.tcpTlsConns))
		for i := range ts.tcpTlsConns {
			belogs.Debug("activeSend(): tcptlsserver to all, client: ", i, "    ts.tcpConns[i]:", ts.tcpTlsConns[i].RemoteAddr().String())
			startOne := time.Now()
			n, err := ts.tcpTlsConns[i].Write(sendData)
			if err != nil {
				belogs.Error("activeSend(): server to all, tcpTlsConn.Write fail, will ignore, tcpTlsConn:", ts.tcpTlsConns[i].RemoteAddr().String(),
					"   n:", n, "   sendData:", convert.PrintBytesOneLine(sendData), "   time(s):", time.Since(startOne), err)
				continue
			} else {
				belogs.Info("activeSend(): server to all, tcpTlsConn.Write ok, tcpTlsConn:", ts.tcpTlsConns[i].RemoteAddr().String(),
					"   n:", n, "   sendData:", convert.PrintBytesOneLine(sendData), "   time(s):", time.Since(startOne))
			}
		}
		belogs.Info("activeSend(): tcptlsserver  send to all clients ok,  len(sendData):", len(sendData), "   len(tcpTlsConns): ", len(ts.tcpTlsConns),
			"  time(s):", time.Since(start).Seconds())
		return
	} else {
		belogs.Debug("activeSend(): tcptlsserver  to connKey:", connKey, "   ts.tcpTlsConns:", ts.tcpTlsConns)
		if tcpTlsConn, ok := ts.tcpTlsConns[connKey]; ok {
			startOne := time.Now()
			belogs.Debug("activeSend():  tcptlsserver  found connKey: ", connKey, "   sendData:", convert.PrintBytesOneLine(sendData))
			n, err := tcpTlsConn.Write(sendData)
			if err != nil {
				belogs.Error("activeSend(): tcptlsserver to ", connKey, " tcpTlsConn.Write fail: tcpTlsConn:", tcpTlsConn.RemoteAddr().String(),
					"   n:", n, "   sendData:", convert.PrintBytesOneLine(sendData), "   time(s):", time.Since(startOne), err)
			} else {
				belogs.Info("activeSend():  tcptlsserver to ", connKey, " tcpTlsConn.Write ok, tcpTlsConn:", tcpTlsConn.RemoteAddr().String(),
					"   n:", n, "   sendData:", convert.PrintBytesOneLine(sendData), "   time(s):", time.Since(startOne))
			}
		} else {
			belogs.Error("activeSend(): tcptlsserver not found connKey: ", connKey, " fail: tcpTlsConn:", tcpTlsConn.RemoteAddr().String(),
				"   sendData:", convert.PrintBytesOneLine(sendData))
		}
		belogs.Info("activeSend(): tcptlsserver  send to connKey ok,  len(sendData):", len(sendData), "   connKey: ", connKey,
			"  time(s):", time.Since(start).Seconds())
		return
	}

}
