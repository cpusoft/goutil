package tlsclient

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"time"

	belogs "github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/tcpserverclient/util"
)

type TlsClientMsg struct {
	NextConnectClosePolicy int //NEXT_CONNECT_CLOSE_POLICE_NO  NEXT_CONNECT_CLOSE_POLICE_GRACEFUL  NEXT_CONNECT_CLOSE_POLICE_FORCIBLE
	NextRwPolice           int //NEXT_RW_POLICE_ALL,NEXT_RW_POLICE_WAIT_READ,NEXT_RW_POLICE_WAIT_WRITE
	SendData               []byte
}

type TlsClient struct {
	tlsClientMsgCh chan TlsClientMsg

	rootCrtFileName      string
	publicCrtFileName    string
	privateKeyFileName   string
	tlsClientProcessFunc TlsClientProcessFunc
}

// server: 0.0.0.0:port
func NewTlsClient(rootCrtFileName, publicCrtFileName, privateKeyFileName string,
	tlsClientProcessFunc TlsClientProcessFunc) (tc *TlsClient) {

	belogs.Debug("NewTlsClient():tlsClientProcessFunc:", tlsClientProcessFunc)
	tc = &TlsClient{}
	tc.tlsClientMsgCh = make(chan TlsClientMsg)
	tc.rootCrtFileName = rootCrtFileName
	tc.publicCrtFileName = publicCrtFileName
	tc.privateKeyFileName = privateKeyFileName
	tc.tlsClientProcessFunc = tlsClientProcessFunc
	belogs.Info("NewTlsClient():tc:", tc)
	return tc
}

// server: **.**.**.**:port
func (tc *TlsClient) Start(server string) (err error) {
	belogs.Debug("Start(): tlsclient  create client, server is  ", server)

	cert, err := tls.LoadX509KeyPair(tc.publicCrtFileName, tc.privateKeyFileName)
	if err != nil {
		belogs.Error("Start(): tlsclient  LoadX509KeyPair fail: server:", server,
			"  publicCrtFileName, privateKeyFileName:", tc.publicCrtFileName, tc.privateKeyFileName, err)
		return err
	}
	rootCrtBytes, err := ioutil.ReadFile(tc.rootCrtFileName)
	if err != nil {
		belogs.Error("Start(): tlsclient  ReadFile rootCrtFileName fail, server:", server,
			"  rootCrtFileName:", tc.rootCrtFileName, err)
		return err
	}
	rootCertPool := x509.NewCertPool()
	ok := rootCertPool.AppendCertsFromPEM(rootCrtBytes)
	if !ok {
		belogs.Error("Start(): tlsclient  AppendCertsFromPEM rootCrtFileName fail,server:", server,
			"  rootCrtFileName:", tc.rootCrtFileName, "  len(rootCrtBytes):", len(rootCrtBytes), err)
		return err
	}
	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            rootCertPool,
		InsecureSkipVerify: false,
	}

	tlsConn, err := tls.Dial("tcp", server, config)
	if err != nil {
		belogs.Error("Start(): tlsclient   Dial fail, server:", server, err)
		return err
	}

	tc.OnConnect(tlsConn)
	belogs.Info("Start(): tlsclient  OnConnect, server is  ", server, "  tlsConn:", tlsConn.RemoteAddr().String())

	//active send to server, and receive from server, loop
	go tc.SendAndReceive(tlsConn)
	belogs.Debug("Start(): tlsclient  waite SendAndReceive, server:", server, "   tlsConn:", tlsConn.RemoteAddr().String())
	return nil
}

func (tc *TlsClient) OnConnect(tlsConn *tls.Conn) {
	// call process func OnConnect
	belogs.Debug("OnConnect(): tlsclient  tlsConn: ", tlsConn.RemoteAddr().String(), "   call process func: OnConnect ")
	tc.tlsClientProcessFunc.OnConnectProcess(tlsConn)
	belogs.Info("OnConnect(): tlsclient  after OnConnectProcess, tlsConn: ", tlsConn.RemoteAddr().String())
}

func (tc *TlsClient) OnClose(tlsConn *tls.Conn) {
	// close in the end
	tlsConn.Close()
	close(tc.tlsClientMsgCh)
	belogs.Info("OnClose(): tlsclient , tlsConn: ", tlsConn.RemoteAddr().String())
	tlsConn = nil
}

func (tc *TlsClient) SendMsg(tlsClientMsg *TlsClientMsg) {

	belogs.Debug("SendMsg(): tlsclient, tlsClientMsg:", jsonutil.MarshalJson(*tlsClientMsg))
	tc.tlsClientMsgCh <- *tlsClientMsg
}

func (tc *TlsClient) SendAndReceive(tlsConn *tls.Conn) (err error) {
	belogs.Debug("SendAndReceive(): tlsclient , tlsConn:", tlsConn.RemoteAddr().String())
	for {
		// wait next tlsClientMsgCh: only error or NEXT_CONNECT_POLICE_CLOSE_** will end loop
		select {
		case tlsClientMsg := <-tc.tlsClientMsgCh:
			nextConnectClosePolicy := tlsClientMsg.NextConnectClosePolicy
			nextRwPolice := tlsClientMsg.NextRwPolice
			sendData := tlsClientMsg.SendData
			belogs.Debug("SendAndReceive(): tlsclient , tlsConn:", tlsConn.RemoteAddr().String(),
				"  tlsClientMsg: ", jsonutil.MarshalJson(tlsClientMsg))

			// if close
			if nextConnectClosePolicy == util.NEXT_CONNECT_POLICE_CLOSE_GRACEFUL ||
				nextConnectClosePolicy == util.NEXT_CONNECT_POLICE_CLOSE_FORCIBLE {
				belogs.Info("SendAndReceive(): tlsclient   nextConnectClosePolicy close end client, will end tlsConn: ", tlsConn.RemoteAddr().String(),
					"   nextConnectClosePolicy:", nextConnectClosePolicy)
				tc.OnClose(tlsConn)
				return nil
			}

			// send data
			start := time.Now()
			n, err := tlsConn.Write(sendData)
			if err != nil {
				belogs.Error("SendAndReceive(): tlsclient   Write fail:  tlsConn:", tlsConn.RemoteAddr().String(), err)
				return err
			}
			belogs.Debug("SendAndReceive(): tlsclient   Write to tlsConn:", tlsConn.RemoteAddr().String(),
				"  len(sendData):", len(sendData), "  write n:", n, "   nextRwPolice:", nextRwPolice,
				"  time(s):", time.Now().Sub(start).Seconds())

			// if wait receive, then wait next tlsClientMsgCh
			if nextRwPolice == util.NEXT_RW_POLICE_WAIT_READ {
				// if server tell client: end this loop, or end conn
				err := tc.OnReceive(tlsConn)
				if err != nil {
					belogs.Error("SendAndReceive(): tlsclient   Write fail:  tlsConn:", tlsConn.RemoteAddr().String(), err)
					return err
				}
				belogs.Info("SendAndReceive(): tlsclient  shouldWaitReceive yes, tlsConn:", tlsConn.RemoteAddr().String(),
					"  len(sendData):", len(sendData), "  write n:", n,
					"  time(s):", time.Now().Sub(start).Seconds())
				continue
			} else {
				belogs.Info("SendAndReceive(): tlsclient  OnReceive, shouldWaitReceive no, will return: ", tlsConn.RemoteAddr().String())
				continue
			}
		}
	}

}

func (tc *TlsClient) OnReceive(tlsConn *tls.Conn) (err error) {
	belogs.Debug("OnReceive(): tlsclient  wait for OnReceive, tlsConn:", tlsConn.RemoteAddr().String())
	var leftData []byte
	// one packet
	buffer := make([]byte, 2048)
	// wait for new packet to read

	for {
		n, err := tlsConn.Read(buffer)
		start := time.Now()
		belogs.Debug("OnReceive(): tlsclient  client read: Read n: ", tlsConn.RemoteAddr().String(), n)
		if err != nil {
			if err == io.EOF {
				// is not error, just client close
				belogs.Debug("OnReceive(): tlsclient   io.EOF, client close: ", tlsConn.RemoteAddr().String(), err)
				return nil
			}
			belogs.Error("OnReceive(): tlsclient   Read fail, err ", tlsConn.RemoteAddr().String(), err)
			return err
		}
		if n == 0 {
			continue
		}

		belogs.Debug("OnReceive(): tlsclient  client tlsConn: ", tlsConn.RemoteAddr().String(), "  n:", n,
			" , will call process func: OnReceiveAndSend,  time(s):", time.Now().Sub(start))
		nextRwPolicy, leftData, err := tc.tlsClientProcessFunc.OnReceiveProcess(tlsConn, append(leftData, buffer[:n]...))
		belogs.Info("OnReceive(): tlsclient  tlsClientProcessFunc.OnReceiveProcess, tlsConn: ", tlsConn.RemoteAddr().String(), " receive n: ", n,
			"  len(leftData):", len(leftData), "  nextRwPolicy:", nextRwPolicy, "  time(s):", time.Now().Sub(start))
		if err != nil {
			belogs.Error("OnReceive(): tlsclient  tlsClientProcessFunc.OnReceiveProcess  fail ,will close this tlsConn : ", tlsConn.RemoteAddr().String(), err)
			return err
		}
		if nextRwPolicy == util.NEXT_RW_POLICE_END_READ {
			belogs.Debug("OnReceive(): tlsclient  nextRwPolicy, will end this write/read loop: ", tlsConn.RemoteAddr().String())
			return nil
		}
	}

}
