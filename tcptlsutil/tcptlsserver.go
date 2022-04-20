package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"sync"
	"time"

	belogs "github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/osutil"
)

// core struct: Start/OnConnect/ReceiveAndSend....
type TcpTlsServer struct {
	// both tcp and tls
	isTcpServer             bool
	tcpTlsConns             map[string]*TcpTlsConn // map[addr]*net.TCPConn
	tcpTlsConnsMutex        sync.RWMutex
	tcpTlsServerProcessFunc TcpTlsServerProcessFunc

	// for tls
	tlsRootCrtFileName    string
	tlsPublicCrtFileName  string
	tlsPrivateKeyFileName string
	tlsVerifyClient       bool

	// https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
	// for close
	tcpTlsListener *TcpTlsListener
	closeGraceful  chan struct{}
}

//
func NewTcpServer(tcpTlsServerProcessFunc TcpTlsServerProcessFunc) (ts *TcpTlsServer) {

	belogs.Debug("NewTcpServer():tcpTlsServerProcessFunc:", tcpTlsServerProcessFunc)
	ts = &TcpTlsServer{}
	ts.isTcpServer = true
	ts.tcpTlsConns = make(map[string]*TcpTlsConn, 16)
	ts.tcpTlsServerProcessFunc = tcpTlsServerProcessFunc
	ts.closeGraceful = make(chan struct{})
	belogs.Debug("NewTcpServer():ts:", ts)
	return ts
}
func NewTlsServer(tlsRootCrtFileName, tlsPublicCrtFileName, tlsPrivateKeyFileName string, tlsVerifyClient bool,
	tcpTlsServerProcessFunc TcpTlsServerProcessFunc) (ts *TcpTlsServer, err error) {

	belogs.Debug("NewTcpServer():tlsRootCrtFileName:", tlsRootCrtFileName, "  tlsPublicCrtFileName:", tlsPublicCrtFileName,
		"   tlsPrivateKeyFileName:", tlsPrivateKeyFileName, "   tlsVerifyClient:", tlsVerifyClient,
		"   tcpTlsServerProcessFunc:", tcpTlsServerProcessFunc)
	ts = &TcpTlsServer{}
	ts.isTcpServer = false
	ts.tcpTlsConns = make(map[string]*TcpTlsConn, 16)
	ts.closeGraceful = make(chan struct{})
	ts.tcpTlsServerProcessFunc = tcpTlsServerProcessFunc

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

	listener, err := net.ListenTCP("tcp", tcpServer)
	if err != nil {
		belogs.Error("StartTcpServer(): tcpserver  ListenTCP fail, port:", port, err)
		return err
	}

	// get tcpTlsListener
	ts.tcpTlsListener, err = NewFromTcpListener(listener)
	if err != nil {
		belogs.Error("StartTcpServer(): tcpserver  NewFromTcpListener fail, port:", port, err)
		return err
	}
	belogs.Info("StartTcpServer(): tcpserver  create server ok, port:", port, "  will accept client")

	// wait new conn
	ts.AcceptNewConn()
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
	// get tcpTlsListener
	ts.tcpTlsListener, err = NewFromTlsListener(listener)
	if err != nil {
		belogs.Error("StartTlsServer(): tlsserver  NewFromTlsListener fail, port: ", port, err)
		return err
	}
	belogs.Info("StartTlsServer(): tlsserver  create server ok, port:", port, "  will accept client")

	// wait new conn
	ts.AcceptNewConn()
	return nil
}

func (ts *TcpTlsServer) AcceptNewConn() {

	defer ts.tcpTlsListener.Close()
	for {
		tcpTlsConn, err := ts.tcpTlsListener.Accept()
		if err != nil {
			select {
			case <-ts.closeGraceful:
				belogs.Error("AcceptNewConn(): Accept remote fail and closeGraceful, will return: ", err)
				return
			default:
				belogs.Error("AcceptNewConn(): Accept remote fail: ", err)
				continue
			}

		}
		belogs.Info("AcceptNewConn():  Accept remote: ", tcpTlsConn.RemoteAddr().String())

		ts.OnConnect(tcpTlsConn)
		// call func to process tcpTlsConn
		go ts.ReceiveAndSend(tcpTlsConn)

	}
}

func (ts *TcpTlsServer) ReceiveAndSend(tcpTlsConn *TcpTlsConn) {

	defer ts.OnClose(tcpTlsConn)

	var leftData []byte
	// one packet
	buffer := make([]byte, 2048)
	// wait for new packet to read
	//https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
	//https://stackoverflow.com/questions/66755407/cancelling-a-net-listener-via-context-in-golang
ReadLoop:
	for {
		select {
		case <-ts.closeGraceful:
			belogs.Info("ReceiveAndSend(): tcptlsserver closeGraceful, will return: ", tcpTlsConn.RemoteAddr().String())
			return
		default:
			tcpTlsConn.SetDeadline(time.Now().Add(500 * time.Millisecond))
			start := time.Now()
			n, err := tcpTlsConn.Read(buffer)
			//	if n == 0 {
			//		continue
			//	}
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					belogs.Debug("ReceiveAndSend(): tcptlsserver Timeout,err:", opErr) //should //
					continue ReadLoop
				} else if err == io.EOF {
					// is not error, just client close
					belogs.Info("ReceiveAndSend(): tcptlsserver Read io.EOF, client close: ", tcpTlsConn.RemoteAddr().String(), err)
					return
				}
				belogs.Error("ReceiveAndSend(): tcptlsserver Read fail, err ", tcpTlsConn.RemoteAddr().String(), err)
				return
			}

			// call process func OnReceiveAndSend
			// copy to leftData
			belogs.Debug("ReceiveAndSend(): tcptlsserver tcpTlsConn: ", tcpTlsConn.RemoteAddr().String(),
				" , Read n:", n, "  time(s):", time.Since(start))
			nextConnectPolicy, leftData, err := ts.tcpTlsServerProcessFunc.ReceiveAndSendProcess(tcpTlsConn, append(leftData, buffer[:n]...))
			belogs.Debug("ReceiveAndSend(): tcptlsserver  after ReceiveAndSendProcess,server tcpTlsConn: ", tcpTlsConn.RemoteAddr().String(), " receive n: ", n,
				"  len(leftData):", len(leftData), "  time(s):", time.Since(start))
			if err != nil {
				belogs.Error("ReceiveAndSend(): tcptlsserver ReceiveAndSendProcess fail ,will remove this tcpTlsConn : ", tcpTlsConn.RemoteAddr().String(), err)
				return
			}

			if nextConnectPolicy == NEXT_CONNECT_POLICE_CLOSE_GRACEFUL ||
				nextConnectPolicy == NEXT_CONNECT_POLICE_CLOSE_FORCIBLE {
				belogs.Info("ReceiveAndSend(): tcptlsserver  nextConnectPolicy return : ", tcpTlsConn.RemoteAddr().String(), nextConnectPolicy)
				return
			}
		}
	}
}

// connKey is "": send to all clients
// connKey is net.Conn.Address.String(): send this client
func (ts *TcpTlsServer) ActiveSend(sendData []byte, connKey string) (err error) {
	ts.tcpTlsConnsMutex.RLock()
	defer ts.tcpTlsConnsMutex.RUnlock()
	start := time.Now()

	belogs.Debug("ActiveSend(): tcptlsserver ,len(sendData):", len(sendData), "   len(tcpTlsConns): ", len(ts.tcpTlsConns), "  connKey:", connKey)
	if len(connKey) == 0 {
		belogs.Debug("ActiveSend(): tcptlsserver to all, len(sendData):", len(sendData), "   len(tcpConns): ", len(ts.tcpTlsConns))
		for i := range ts.tcpTlsConns {
			belogs.Debug("ActiveSend(): tcptlsserver   to all, client: ", i, "    ts.tcpConns[i]:", ts.tcpTlsConns[i], "   call process func: ActiveSend ")
			err = ts.tcpTlsServerProcessFunc.ActiveSendProcess(ts.tcpTlsConns[i], sendData)
			if err != nil {
				// just logs, not return or break
				belogs.Error("ActiveSend(): tcptlsserver  ActiveSendProcess fail, to all, client: ", i, "    ts.tcpTlsConns[i]:", ts.tcpTlsConns[i], err)
			}
		}
		belogs.Info("ActiveSend(): tcptlsserver  send to all clients ok,  len(sendData):", len(sendData), "   len(tcpTlsConns): ", len(ts.tcpTlsConns),
			"  time(s):", time.Now().Sub(start).Seconds())
		return
	} else {
		belogs.Debug("ActiveSend(): tcptlsserver  to connKey:", connKey)
		if tcpTlsConn, ok := ts.tcpTlsConns[connKey]; ok {
			err = ts.tcpTlsServerProcessFunc.ActiveSendProcess(tcpTlsConn, sendData)
			if err != nil {
				// just logs, not return or break
				belogs.Error("ActiveSend(): tcptlsserver  fail, to connKey: ", connKey, "   tcpTlsConn:", tcpTlsConn.RemoteAddr().String(), err)
			}
		}
		belogs.Info("ActiveSend(): tcptlsserver  send to connKey ok,  len(sendData):", len(sendData), "   connKey: ", connKey,
			"  time(s):", time.Now().Sub(start).Seconds())
		return
	}

}

func (ts *TcpTlsServer) OnConnect(tcpTlsConn *TcpTlsConn) {
	start := time.Now()
	belogs.Debug("OnConnect(): new tcpTlsConn: ", tcpTlsConn)

	// add new tcpTlsConn to tcpTlsConns
	ts.tcpTlsConnsMutex.Lock()
	defer ts.tcpTlsConnsMutex.Unlock()

	connKey := tcpTlsConn.RemoteAddr().String()
	ts.tcpTlsConns[connKey] = tcpTlsConn
	belogs.Debug("OnConnect(): tcptlsserver tcpTlsConn: ", tcpTlsConn.RemoteAddr().String(), ", connKey:", connKey, "  new len(tcpTlsConns): ", len(ts.tcpTlsConns))
	ts.tcpTlsServerProcessFunc.OnConnectProcess(tcpTlsConn)
	belogs.Info("OnConnect(): tcptlsserver add tcpTlsConn: ", tcpTlsConn.RemoteAddr().String(), "   len(tcpTlsConns): ", len(ts.tcpTlsConns), "   time(s):", time.Now().Sub(start).Seconds())

}

func (ts *TcpTlsServer) OnClose(tcpTlsConn *TcpTlsConn) {
	// close in the end
	defer func() {
		tcpTlsConn.Close()
		//tcpTlsConn.SetNil()
	}()
	start := time.Now()

	// call process func OnClose
	belogs.Debug("OnClose(): tcptlsserver tcpTlsConn: ", tcpTlsConn.RemoteAddr().String(), "   call process func: OnClose ")
	ts.tcpTlsServerProcessFunc.OnCloseProcess(tcpTlsConn)

	// remove tcpTlsConn from tcpConns
	ts.tcpTlsConnsMutex.Lock()
	defer ts.tcpTlsConnsMutex.Unlock()
	belogs.Debug("OnClose(): tcptlsserver will new tcpTlsConns, tcpTlsConn: ", tcpTlsConn.RemoteAddr().String(), "   old len(tcpTlsConns): ", len(ts.tcpTlsConns))
	newTlsTcpConns := make(map[string]*TcpTlsConn, len(ts.tcpTlsConns))
	for i := range ts.tcpTlsConns {
		if ts.tcpTlsConns[i] != tcpTlsConn {
			connKey := tcpTlsConn.RemoteAddr().String()
			newTlsTcpConns[connKey] = tcpTlsConn
		}
	}
	ts.tcpTlsConns = newTlsTcpConns

	belogs.Info("OnClose(): tcptlsserver new len(tcpTlsConns): ", len(ts.tcpTlsConns), "  time(s):", time.Now().Sub(start).Seconds())
}

func (ts *TcpTlsServer) CloseGraceful() {
	// send channel, and wait listener and conns end itself process and close loop
	belogs.Info("CloseGraceful(): tcptlsserver will close graceful")
	close(ts.closeGraceful)
}

func (ts *TcpTlsServer) CloseForceful() {
	belogs.Info("CloseForceful(): tcptlsserver will close forceful")
	// close listener/conns loop
	go ts.CloseGraceful()
	// ignore conns's writing/reading, just close
	ts.tcpTlsListener.Close()
	for i := range ts.tcpTlsConns {
		ts.tcpTlsConns[i].Close()
	}

}
