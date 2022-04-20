package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"time"

	belogs "github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

type TcpTlsClientSendMsg struct {
	NextConnectClosePolicy int //NEXT_CONNECT_CLOSE_POLICE_NO  NEXT_CONNECT_CLOSE_POLICE_GRACEFUL  NEXT_CONNECT_CLOSE_POLICE_FORCIBLE
	NextRwPolice           int //NEXT_RW_POLICE_ALL,NEXT_RW_POLICE_WAIT_READ,NEXT_RW_POLICE_WAIT_WRITE
	SendData               []byte
}

type TcpTlsClient struct {
	// both tcp and tls
	isTcpClient             bool
	tcpTlsClientSendMsg     chan TcpTlsClientSendMsg
	tcpTlsClientProcessFunc TcpTlsClientProcessFunc

	// for tls
	tlsRootCrtFileName    string
	tlsPublicCrtFileName  string
	tlsPrivateKeyFileName string

	// for close
	tcpTlsConn *TcpTlsConn
}

// server: 0.0.0.0:port
func NewTcpClient(tcpTlsClientProcessFunc TcpTlsClientProcessFunc) (tc *TcpTlsClient) {

	belogs.Debug("NewTcpClient():tcpTlsClientProcessFunc:", tcpTlsClientProcessFunc)
	tc = &TcpTlsClient{}
	tc.isTcpClient = true
	tc.tcpTlsClientSendMsg = make(chan TcpTlsClientSendMsg)
	tc.tcpTlsClientProcessFunc = tcpTlsClientProcessFunc
	belogs.Info("NewTcpClient():tc:", tc)
	return tc
}

// server: 0.0.0.0:port
func NewTlsClient(tlsRootCrtFileName, tlsPublicCrtFileName, tlsPrivateKeyFileName string,
	tcpTlsClientProcessFunc TcpTlsClientProcessFunc) (tc *TcpTlsClient, err error) {

	belogs.Debug("NewTlsClient():tcpTlsClientProcessFunc:", &tcpTlsClientProcessFunc)
	tc = &TcpTlsClient{}
	tc.isTcpClient = false
	tc.tcpTlsClientSendMsg = make(chan TcpTlsClientSendMsg)
	tc.tcpTlsClientProcessFunc = tcpTlsClientProcessFunc

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
func (tc *TcpTlsClient) StartTcpClient(server string) (err error) {
	belogs.Debug("StartTcpClient(): create client, server is  ", server)

	conn, err := net.DialTimeout("tcp", server, 60*time.Second)
	if err != nil {
		belogs.Error("StartTcpClient(): DialTimeout fail, server:", server, err)
		return err
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		belogs.Error("StartTcpClient(): conn cannot conver to tcpConn: ", conn.RemoteAddr().String(), err)
		return err
	}
	/*
		tcpServer, err := net.ResolveTCPAddr("tcp", server)
		if err != nil {
			belogs.Error("StartTcpClient():  ResolveTCPAddr fail: ", server, err)
			return err
		}
		belogs.Debug("StartTcpClient(): create client, server is  ", server, "  tcpServer:", tcpServer)


		tcpConn, err := net.DialTCP("tcp", nil, tcpServer)
		if err != nil {
			belogs.Error("StartTcpClient(): Dial fail, server:", server, "  tcpServer:", tcpServer, err)
			return err
		}
	*/
	tc.tcpTlsConn = NewFromTcpConn(tcpConn)
	tc.OnConnect()
	belogs.Info("StartTcpClient(): OnConnect, server is  ", server, "  tcpTlsConn:", tc.tcpTlsConn.RemoteAddr().String())

	//active send to server, and receive from server, loop
	go tc.SendAndReceive()
	belogs.Debug("StartTcpClient(): SendAndReceive, server:", server, "   tcpTlsConn:", tc.tcpTlsConn.RemoteAddr().String())
	return nil
}

// server: **.**.**.**:port
func (tc *TcpTlsClient) StartTlsClient(server string) (err error) {
	belogs.Debug("StartTlsClient(): create client, server is  ", server)

	cert, err := tls.LoadX509KeyPair(tc.tlsPublicCrtFileName, tc.tlsPrivateKeyFileName)
	if err != nil {
		belogs.Error("StartTlsClient(): LoadX509KeyPair fail: server:", server,
			"  tlsPublicCrtFileName, tlsPrivateKeyFileName:", tc.tlsPublicCrtFileName, tc.tlsPrivateKeyFileName, err)
		return err
	}
	rootCrtBytes, err := ioutil.ReadFile(tc.tlsRootCrtFileName)
	if err != nil {
		belogs.Error("StartTlsClient(): ReadFile tlsRootCrtFileName fail, server:", server,
			"  tlsRootCrtFileName:", tc.tlsRootCrtFileName, err)
		return err
	}
	rootCertPool := x509.NewCertPool()
	ok := rootCertPool.AppendCertsFromPEM(rootCrtBytes)
	if !ok {
		belogs.Error("StartTlsClient(): AppendCertsFromPEM tlsRootCrtFileName fail,server:", server,
			"  tlsRootCrtFileName:", tc.tlsRootCrtFileName, "  len(rootCrtBytes):", len(rootCrtBytes), err)
		return err
	}
	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            rootCertPool,
		InsecureSkipVerify: false,
	}
	dialer := &net.Dialer{Timeout: time.Duration(60) * time.Second}
	tlsConn, err := tls.DialWithDialer(dialer, "tcp", server, config)
	if err != nil {
		belogs.Error("StartTlsClient(): DialWithDialer fail, server:", server, err)
		return err
	}
	/*
		tlsConn, err := tls.Dial("tcp", server, config)
		if err != nil {
			belogs.Error("StartTlsClient(): Dial fail, server:", server, err)
			return err
		}
	*/
	tc.tcpTlsConn = NewFromTlsConn(tlsConn)
	tc.OnConnect()
	belogs.Info("StartTlsClient(): OnConnect, server is  ", server, "  tcpTlsConn:", tc.tcpTlsConn.RemoteAddr().String())

	//active send to server, and receive from server, loop
	go tc.SendAndReceive()
	belogs.Debug("StartTlsClient(): SendAndReceive, server:", server, "   tcpTlsConn:", tc.tcpTlsConn.RemoteAddr().String())
	return nil
}

func (tc *TcpTlsClient) OnConnect() {
	// call process func OnConnect
	tc.tcpTlsClientProcessFunc.OnConnectProcess(tc.tcpTlsConn)
	belogs.Info("OnConnect(): tcptlsclient  after OnConnectProcess, tcpTlsConn: ", tc.tcpTlsConn.RemoteAddr().String())
}

func (tc *TcpTlsClient) OnClose() {
	// close in the end
	belogs.Info("OnClose(): tcptlsclient , tcpTlsConn: ", tc.tcpTlsConn.RemoteAddr().String())
	tc.tcpTlsConn.Close()
}

func (tc *TcpTlsClient) SendMsg(tcpTlsClientSendMsg *TcpTlsClientSendMsg) {

	belogs.Debug("SendMsg(): tcptlsclient, tcpTlsClientSendMsg:", jsonutil.MarshalJson(*tcpTlsClientSendMsg))
	tc.tcpTlsClientSendMsg <- *tcpTlsClientSendMsg
}

func (tc *TcpTlsClient) SendAndReceive() (err error) {
	belogs.Debug("SendAndReceive(): tcptlsclient , tcpTlsConn:", tc.tcpTlsConn.RemoteAddr().String())
	for {
		// wait next tcpTlsClientSendMsg: only error or NEXT_CONNECT_POLICE_CLOSE_** will end loop
		select {
		case tcpTlsClientSendMsg := <-tc.tcpTlsClientSendMsg:
			nextConnectClosePolicy := tcpTlsClientSendMsg.NextConnectClosePolicy
			nextRwPolice := tcpTlsClientSendMsg.NextRwPolice
			sendData := tcpTlsClientSendMsg.SendData
			belogs.Debug("SendAndReceive(): tcptlsclient , tcpTlsConn:", tc.tcpTlsConn.RemoteAddr().String(),
				"  tcpTlsClientSendMsg: ", jsonutil.MarshalJson(tcpTlsClientSendMsg))

			// if close
			if nextConnectClosePolicy == NEXT_CONNECT_POLICE_CLOSE_GRACEFUL ||
				nextConnectClosePolicy == NEXT_CONNECT_POLICE_CLOSE_FORCIBLE {
				belogs.Info("SendAndReceive(): tcptlsclient   nextConnectClosePolicy close end client, will end tcpTlsConn: ", tc.tcpTlsConn.RemoteAddr().String(),
					"   nextConnectClosePolicy:", nextConnectClosePolicy)
				tc.OnClose()
				return nil
			}

			// send data
			start := time.Now()
			n, err := tc.tcpTlsConn.Write(sendData)
			if err != nil {
				belogs.Error("SendAndReceive(): tcptlsclient   Write fail:  tcpTlsConn:", tc.tcpTlsConn.RemoteAddr().String(), err)
				return err
			}
			belogs.Debug("SendAndReceive(): tcptlsclient   Write to tcpTlsConn:", tc.tcpTlsConn.RemoteAddr().String(),
				"  len(sendData):", len(sendData), "  write n:", n, "   nextRwPolice:", nextRwPolice,
				"  time(s):", time.Since(start))

			// if wait receive, then wait next tcpTlsClientSendMsg
			if nextRwPolice == NEXT_RW_POLICE_WAIT_READ {
				// if server tell client: end this loop, or end conn
				err := tc.OnReceive()
				if err != nil {
					belogs.Error("SendAndReceive(): tcptlsclient   Write fail:  tcpTlsConn:", tc.tcpTlsConn.RemoteAddr().String(), err)
					return err
				}
				belogs.Info("SendAndReceive(): tcptlsclient NEXT_RW_POLICE_WAIT_READ, OnReceive, tcpTlsConn:", tc.tcpTlsConn.RemoteAddr().String(),
					"  len(sendData):", len(sendData), "  write n:", n,
					"  time(s):", time.Since(start))
				continue
			} else {
				belogs.Info("SendAndReceive(): tcptlsclient no NEXT_RW_POLICE_WAIT_READ, OnReceive, will return: ", tc.tcpTlsConn.RemoteAddr().String())
				continue
			}
		}
	}

}

func (tc *TcpTlsClient) OnReceive() (err error) {
	belogs.Debug("OnReceive(): tcptlsclient  wait for OnReceive, tcpTlsConn:", tc.tcpTlsConn.RemoteAddr().String())
	var leftData []byte
	// one packet
	buffer := make([]byte, 2048)
	// wait for new packet to read

	for {
		start := time.Now()
		n, err := tc.tcpTlsConn.Read(buffer)
		//	if n == 0 {
		//		continue
		//	}
		if err != nil {
			if err == io.EOF {
				// is not error, just client close
				belogs.Debug("OnReceive(): tcptlsclient   io.EOF, client close: ", tc.tcpTlsConn.RemoteAddr().String(), err)
				return nil
			}
			belogs.Error("OnReceive(): tcptlsclient   Read fail, err ", tc.tcpTlsConn.RemoteAddr().String(), err)
			return err
		}

		belogs.Debug("OnReceive(): tcptlsclient, tcpTlsConn: ", tc.tcpTlsConn.RemoteAddr().String(),
			"  Read n", n, "  time(s):", time.Now().Sub(start))
		nextRwPolicy, leftData, err := tc.tcpTlsClientProcessFunc.OnReceiveProcess(tc.tcpTlsConn, append(leftData, buffer[:n]...))
		belogs.Info("OnReceive(): tcptlsclient  tcpTlsClientProcessFunc.OnReceiveProcess, tcpTlsConn: ", tc.tcpTlsConn.RemoteAddr().String(), " receive n: ", n,
			"  len(leftData):", len(leftData), "  nextRwPolicy:", nextRwPolicy, "  time(s):", time.Now().Sub(start))
		if err != nil {
			belogs.Error("OnReceive(): tcptlsclient  tcpTlsClientProcessFunc.OnReceiveProcess  fail ,will close this tcpTlsConn : ", tc.tcpTlsConn.RemoteAddr().String(), err)
			return err
		}
		if nextRwPolicy == NEXT_RW_POLICE_END_READ {
			belogs.Debug("OnReceive(): tcptlsclient  nextRwPolicy, will end this write/read loop: ", tc.tcpTlsConn.RemoteAddr().String())
			return nil
		}

	}

}

func (tc *TcpTlsClient) CloseGraceful() {
	// send channel, and wait listener and conns end itself process and close loop
	belogs.Info("CloseGraceful(): tcptlsclient will close graceful")
	tcpClientSendMsg := &TcpTlsClientSendMsg{
		NextConnectClosePolicy: NEXT_CONNECT_POLICE_CLOSE_GRACEFUL,
		SendData:               nil,
	}
	tc.SendMsg(tcpClientSendMsg)

}

func (tc *TcpTlsClient) CloseForceful() {
	belogs.Info("CloseForceful(): tcptlsclient will close forceful")
	go tc.CloseGraceful()
	tc.OnClose()
}
