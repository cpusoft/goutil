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
	"github.com/cpusoft/goutil/transportutil/udpmock"
)

// core struct: Start/onConnect/receiveAndSend....
type TransportServer struct {
	// state
	state uint64
	// tcp/tls/udp
	connType               string
	transportConns         map[string]*TransportConn // map[addr]*net.TCPConn
	transportConnsMutex    sync.RWMutex
	transportServerProcess TransportServerProcess

	// for tls
	tlsRootCrtFileName    string
	tlsPublicCrtFileName  string
	tlsPrivateKeyFileName string
	tlsVerifyClient       bool

	// https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
	// for close
	transportListener *TransportListener
	closeGraceful     chan struct{}

	// for channel
	TransportMsg chan TransportMsg
}

//
func NewTcpServer(transportServerProcess TransportServerProcess, transportMsg chan TransportMsg) (ts *TransportServer) {

	belogs.Debug("NewTcpServer():transportServerProcess:", transportServerProcess)
	ts = &TransportServer{}
	ts.state = SERVER_STATE_INIT
	ts.connType = "tcp"
	ts.transportConns = make(map[string]*TransportConn, 16)
	ts.transportServerProcess = transportServerProcess
	ts.closeGraceful = make(chan struct{})
	ts.TransportMsg = transportMsg
	belogs.Debug("NewTcpServer():ts:", ts)
	return ts
}
func NewTlsServer(tlsRootCrtFileName, tlsPublicCrtFileName, tlsPrivateKeyFileName string, tlsVerifyClient bool,
	transportServerProcess TransportServerProcess, transportMsg chan TransportMsg) (ts *TransportServer, err error) {

	belogs.Debug("NewTcpServer():tlsRootCrtFileName:", tlsRootCrtFileName, "  tlsPublicCrtFileName:", tlsPublicCrtFileName,
		"   tlsPrivateKeyFileName:", tlsPrivateKeyFileName, "   tlsVerifyClient:", tlsVerifyClient,
		"   transportServerProcess:", transportServerProcess)
	ts = &TransportServer{}
	ts.state = SERVER_STATE_INIT
	ts.connType = "tls"
	ts.transportConns = make(map[string]*TransportConn, 16)
	ts.closeGraceful = make(chan struct{})
	ts.TransportMsg = transportMsg
	ts.transportServerProcess = transportServerProcess

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

//
func NewUdpServer(transportServerProcess TransportServerProcess, transportMsg chan TransportMsg) (ts *TransportServer) {

	belogs.Debug("NewUdpServer():transportServerProcess:", transportServerProcess)
	ts = &TransportServer{}
	ts.state = SERVER_STATE_INIT
	ts.connType = "udp"
	ts.transportConns = make(map[string]*TransportConn, 16)
	ts.transportServerProcess = transportServerProcess
	ts.closeGraceful = make(chan struct{})
	ts.TransportMsg = transportMsg
	belogs.Debug("NewUdpServer():ts:", ts)
	return ts
}

// port: `8888` --> `0.0.0.0:8888`
func (ts *TransportServer) StartTcpServer(port string) (err error) {
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

	// get transportListener
	ts.transportListener, err = NewFromTcpListener(listener)
	if err != nil {
		belogs.Error("StartTcpServer(): tcpserver  NewFromTcpListener fail, port:", port, err)
		return err
	}
	belogs.Info("StartTcpServer(): tcpserver  create server ok, port:", port, "  will accept client")

	go ts.waitTransportMsg()

	// wait new conn
	ts.acceptNewConn()
	return nil
}

// port: `8888` --> `:8888`
func (ts *TransportServer) StartTlsServer(port string) (err error) {

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

	// get transportListener
	ts.transportListener, err = NewFromTlsListener(listener)
	if err != nil {
		belogs.Error("StartTlsServer(): tlsserver  NewFromTlsListener fail, port: ", port, err)
		return err
	}
	belogs.Info("StartTlsServer(): tlsserver  create server ok, port:", port, "  will accept client")

	go ts.waitTransportMsg()

	// wait new conn
	ts.acceptNewConn()
	return nil
}

// port: `8888` --> `0.0.0.0:8888`
func (ts *TransportServer) StartUdpServer(port string) (err error) {
	udpServer, err := net.ResolveUDPAddr("udp", "0.0.0.0:"+port)
	if err != nil {
		belogs.Error("StartUdpServer(): ResolveUDPAddr fail, port:", port, err)
		return err
	}
	belogs.Debug("StartUdpServer(): ResolveUDPAddr ok,  port:", port)

	listener, err := udpmock.ListenUDP("udp", udpServer)
	if err != nil {
		belogs.Error("StartUdpServer(): ListenUDP fail, port:", port, err)
		return err
	}
	belogs.Debug("StartUdpServer(): ListenUDP ok,  port:", port)

	// get transportListener
	ts.transportListener, err = NewFromUdpListener(listener)
	if err != nil {
		belogs.Error("StartTlsServer(): tlsserver  NewFromTlsListener fail, port: ", port, err)
		return err
	}
	belogs.Info("StartTlsServer(): tlsserver  create server ok, port:", port, "  will accept client")

	go ts.waitTransportMsg()

	// wait new conn
	ts.acceptNewConn()
	return nil
}

func (ts *TransportServer) acceptNewConn() {

	defer ts.transportListener.Close()
	belogs.Debug("TransportServer.acceptNewConn(): will accept client")
	ts.state = SERVER_STATE_RUNNING
	for {
		transportConn, err := ts.transportListener.Accept()
		if err != nil {
			select {
			case <-ts.closeGraceful:
				belogs.Info("TransportServer.acceptNewConn(): Accept remote fail and closeGraceful, will return: ", err)
				return
			default:
				belogs.Error("TransportServer.acceptNewConn(): Accept remote fail: ", err)
				continue
			}

		}
		belogs.Info("TransportServer.acceptNewConn():  Accept remote: ", transportConn.RemoteAddr().String())

		ts.onConnect(transportConn)
		// call func to process transportConn
		go ts.receiveAndSend(transportConn)

	}
}

// close directly
func (ts *TransportServer) receiveAndSend(transportConn *TransportConn) {

	defer ts.onClose(transportConn)

	var leftData []byte
	// one packet
	buffer := make([]byte, 2048)
	belogs.Debug("TransportServer.receiveAndSend(): recive from transportConn: ", transportConn.RemoteAddr().String())

	// wait for new packet to read
	for {
		start := time.Now()
		n, err := transportConn.Read(buffer)
		belogs.Debug("TransportServer.receiveAndSend():server read: Read from : ", transportConn.RemoteAddr().String(), "  read n:", n)
		//	if n == 0 {
		//		continue
		//	}
		if err != nil {
			if err == io.EOF {
				// is not error, just client close
				belogs.Info("TransportServer.receiveAndSend(): Read io.EOF, client close: ", transportConn.RemoteAddr().String(), err)
				return
			}
			belogs.Error("TransportServer.receiveAndSend(): Read fail, err ", transportConn.RemoteAddr().String(), err)
			return
		}

		// call process func OnReceiveAndSend
		// copy to leftData
		belogs.Debug("TransportServer.receiveAndSend(): transportConn: ", transportConn.RemoteAddr().String(),
			" , Read n:", n, "  time(s):", time.Since(start))
		nextConnectPolicy, leftData, err := ts.transportServerProcess.OnReceiveAndSendProcess(transportConn, append(leftData, buffer[:n]...))
		belogs.Debug("TransportServer.receiveAndSend(): after OnReceiveAndSendProcess,server transportConn: ", transportConn.RemoteAddr().String(), " receive n: ", n,
			"  len(leftData):", len(leftData), "  time(s):", time.Since(start))
		if err != nil {
			belogs.Error("TransportServer.receiveAndSend(): OnReceiveAndSendProcess fail ,will remove this transportConn : ", transportConn.RemoteAddr().String(), err)
			return
		}

		if nextConnectPolicy == NEXT_CONNECT_POLICY_CLOSE_GRACEFUL ||
			nextConnectPolicy == NEXT_CONNECT_POLICY_CLOSE_FORCIBLE {
			belogs.Info("TransportServer.receiveAndSend(): nextConnectPolicy close,  return : ", transportConn.RemoteAddr().String(), nextConnectPolicy)
			return
		}
		// check state
		if ts.state != SERVER_STATE_RUNNING {
			belogs.Debug("TransportServer.receiveAndSend(): state is not running, will close from transportConn: ", transportConn.RemoteAddr().String(),
				"  state:", ts.state, "  time(s):", time.Now().Sub(start))
			return
		}
		belogs.Debug("TransportServer.receiveAndSend(): will wait for Read from transportConn: ", transportConn.RemoteAddr().String(),
			"  time(s):", time.Now().Sub(start))

	}
}

func (ts *TransportServer) onConnect(transportConn *TransportConn) {
	start := time.Now()
	belogs.Debug("TransportServer.onConnect(): new transportConn: ", transportConn.RemoteAddr().String())

	// add new transportConn to transportConns
	ts.transportConnsMutex.Lock()
	defer ts.transportConnsMutex.Unlock()

	connKey := GetConnKey(transportConn)
	ts.transportConns[connKey] = transportConn
	belogs.Debug("TransportServer.onConnect(): transportConn: ", transportConn.RemoteAddr().String(), ", connKey:", connKey, "  ts.transportConns: ", ts.transportConns)
	ts.transportServerProcess.OnConnectProcess(transportConn)
	belogs.Info("TransportServer.onConnect(): add transportConn: ", transportConn.RemoteAddr().String(), "   len(transportConns): ", len(ts.transportConns), "   time(s):", time.Now().Sub(start).Seconds())

}

func (ts *TransportServer) onClose(transportConn *TransportConn) {
	// close in the end
	if transportConn == nil {
		return
	}

	defer func() {
		transportConn.Close()
		//transportConn.SetNil()
	}()

	start := time.Now()
	// call process func onClose
	belogs.Debug("TransportServer.onClose(): transportConn: ", transportConn.RemoteAddr().String(), "   call process func: onClose ")

	// remove transportConn from tcpConns
	ts.transportConnsMutex.Lock()
	defer ts.transportConnsMutex.Unlock()
	belogs.Debug("TransportServer.onClose(): will close old transportConns, transportConn: ", transportConn.RemoteAddr().String(), "   old len(transportConns): ", len(ts.transportConns))
	delete(ts.transportConns, GetConnKey(transportConn))
	ts.transportServerProcess.OnCloseProcess(transportConn)
	belogs.Info("TransportServer.onClose(): new len(transportConns): ", len(ts.transportConns), "  time(s):", time.Now().Sub(start).Seconds())
}

func (ts *TransportServer) SendMsg(transportMsg *TransportMsg) {

	belogs.Debug("TransportServer.SendMsg():, transportMsg:", jsonutil.MarshalJson(*transportMsg))
	ts.TransportMsg <- *transportMsg
}

// msgType:MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_GRACEFUL, //
// MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_FORCIBLE
func (ts *TransportServer) SendMsgForCloseConnect(msgType uint64, serverConnKey string) {
	// send channel, and wait listener and conns end itself process and close loop
	belogs.Info("TransportServer.SendMsgForCloseConnect(): will close, msgType:", msgType, "  serverConnKey:", serverConnKey)
	transportMsg := &TransportMsg{
		MsgType:       msgType,
		ServerConnKey: serverConnKey,
	}
	ts.SendMsg(transportMsg)
}

func (ts *TransportServer) waitTransportMsg() {
	belogs.Debug("TransportServer.waitTransportMsg(): will waitTransportMsg")
	for {
		select {
		case transportMsg := <-ts.TransportMsg:
			belogs.Info("TransportServer.waitTransportMsg(): transportMsg:", jsonutil.MarshalJson(transportMsg))

			switch transportMsg.MsgType {
			case MSG_TYPE_SERVER_CLOSE_FORCIBLE:
				// ignore conns's writing/reading, just close
				belogs.Info("TransportServer.waitTransportMsg(): msgType is MSG_TYPE_SERVER_CLOSE_FORCIBLE")
				// just close
				ts.state = SERVER_STATE_CLOSING
				ts.transportListener.Close()
				for connKey := range ts.transportConns {
					ts.onClose(ts.transportConns[connKey])
				}
				close(ts.closeGraceful)
				belogs.Info("TransportServer.waitTransportMsg(): will close server forcible, will return waitTransportMsg:")
				// end for/select
				ts.state = SERVER_STATE_CLOSED
				// will return, close waitTransportMsg
				return
			case MSG_TYPE_SERVER_CLOSE_GRACEFUL:
				// close and wait connect.Read and Accept
				belogs.Info("TransportServer.waitTransportMsg(): msgType is MSG_TYPE_SERVER_CLOSE_GRACEFUL")
				ts.state = SERVER_STATE_CLOSING
				close(ts.closeGraceful)
				time.Sleep(5 * time.Second)
				ts.transportListener.Close()
				for connKey := range ts.transportConns {
					ts.onClose(ts.transportConns[connKey])
				}
				belogs.Info("TransportServer.waitTransportMsg(): will close server graceful, will return waitTransportMsg:")
				// end for/select
				ts.state = SERVER_STATE_CLOSED
				// will return, close waitTransportMsg
				return
			case MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_GRACEFUL:
				belogs.Info("TransportServer.waitTransportMsg(): msgType is MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_GRACEFUL")
				fallthrough
			case MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_FORCIBLE:
				// close and wait connect.Read and Accept
				belogs.Info("TransportServer.waitTransportMsg(): msgType is MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_FORCIBLE")
				if len(transportMsg.ServerConnKey) > 0 {
					ts.onClose(ts.transportConns[transportMsg.ServerConnKey])
				}
				belogs.Info("TransportServer.waitTransportMsg(): close connect, serverConnKey:", transportMsg.ServerConnKey)
				// close one connect, no return
				// return
			case MSG_TYPE_COMMON_SEND_DATA:

				serverConnKey := transportMsg.ServerConnKey
				sendData := transportMsg.SendData
				belogs.Info("TransportServer.waitTransportMsg(): msgType is MSG_TYPE_COMMON_SEND_DATA, serverConnKey:", serverConnKey,
					"  sendData:", convert.PrintBytesOneLine(sendData))
				err := ts.activeSend(serverConnKey, sendData)
				if err != nil {
					belogs.Error("TransportServer.waitTransportMsg(): activeSend fail, serverConnKey:", serverConnKey,
						"  sendData:", convert.PrintBytesOneLine(sendData), err)
					// err, no return
					// return
				} else {
					belogs.Info("TransportServer.waitTransportMsg(): activeSend ok, serverConnKey:", serverConnKey,
						"  sendData:", convert.PrintBytesOneLine(sendData))
				}
			}
		}
	}
}

// connKey is "": send to all clients
// connKey is net.Conn.Address.String(): send this client
func (ts *TransportServer) activeSend(connKey string, sendData []byte) (err error) {
	ts.transportConnsMutex.RLock()
	defer ts.transportConnsMutex.RUnlock()
	start := time.Now()

	belogs.Debug("TransportServer.activeSend(): ,len(sendData):", len(sendData),
		"   transportConns: ", ts.transportConns, "  connKey:", connKey)
	if len(connKey) == 0 {
		belogs.Debug("TransportServer.activeSend(): to all, len(sendData):", len(sendData), "   len(tcpConns): ", len(ts.transportConns))
		for i := range ts.transportConns {
			belogs.Debug("TransportServer.activeSend(): to all, client: ", i, "    ts.tcpConns[i]:", ts.transportConns[i].RemoteAddr().String())
			startOne := time.Now()
			n, err := ts.transportConns[i].Write(sendData)
			if err != nil {
				belogs.Error("TransportServer.activeSend(): server to all, transportConn.Write fail, will ignore, transportConn:", ts.transportConns[i].RemoteAddr().String(),
					"   n:", n, "   sendData:", convert.PrintBytesOneLine(sendData), "   time(s):", time.Since(startOne), err)
				continue
			} else {
				belogs.Info("TransportServer.activeSend(): server to all, transportConn.Write ok, transportConn:", ts.transportConns[i].RemoteAddr().String(),
					"   n:", n, "   sendData:", convert.PrintBytesOneLine(sendData), "   time(s):", time.Since(startOne))
			}
		}
		belogs.Info("TransportServer.activeSend(): send to all clients ok,  len(sendData):", len(sendData), "   len(transportConns): ", len(ts.transportConns),
			"  time(s):", time.Now().Sub(start).Seconds())
		return
	} else {
		belogs.Debug("TransportServer.activeSend(): to connKey:", connKey, "   ts.transportConns:", ts.transportConns)
		if transportConn, ok := ts.transportConns[connKey]; ok {
			startOne := time.Now()
			belogs.Debug("TransportServer.activeSend(): found connKey: ", connKey, "   sendData:", convert.PrintBytesOneLine(sendData))
			n, err := transportConn.Write(sendData)
			if err != nil {
				belogs.Error("TransportServer.activeSend(): to ", connKey, " transportConn.Write fail: transportConn:", transportConn.RemoteAddr().String(),
					"   n:", n, "   sendData:", convert.PrintBytesOneLine(sendData), "   time(s):", time.Since(startOne), err)
			} else {
				belogs.Info("TransportServer.activeSend(): to ", connKey, " transportConn.Write ok, transportConn:", transportConn.RemoteAddr().String(),
					"   n:", n, "   sendData:", convert.PrintBytesOneLine(sendData), "   time(s):", time.Since(startOne))
			}
		} else {
			belogs.Error("TransportServer.activeSend(): not found connKey: ", connKey, " fail: transportConn:", transportConn.RemoteAddr().String(),
				"   sendData:", convert.PrintBytesOneLine(sendData))
		}
		belogs.Info("TransportServer.activeSend(): send to connKey ok,  len(sendData):", len(sendData), "   connKey: ", connKey,
			"  time(s):", time.Now().Sub(start).Seconds())
		return
	}

}
