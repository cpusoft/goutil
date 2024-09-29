package transportutil

import (
	"container/list"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"net"
	"os"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

const (

	// connect: keep or close graceful/forcible
	NEXT_CONNECT_POLICY_KEEP           = 0
	NEXT_CONNECT_POLICY_CLOSE_GRACEFUL = 1
	NEXT_CONNECT_POLICY_CLOSE_FORCIBLE = 2

	// need wait read/write
	NEXT_RW_POLICY_WAIT_READ  = 3
	NEXT_RW_POLICY_WAIT_WRITE = 4

	// no need more read
	NEXT_RW_POLICY_END_READ = 5

	SERVER_STATE_INIT    = 0
	SERVER_STATE_RUNNING = 1
	SERVER_STATE_CLOSING = 2
	SERVER_STATE_CLOSED  = 3
)

// packets: if Len==0,means no complete package
func RecombineReceiveData(receiveData []byte, minPacketLen, lengthFieldStart,
	lengthFieldEnd int) (packets *list.List, leftData []byte, err error) {
	belogs.Debug("RecombineReceiveData(): len(receiveData):", len(receiveData),
		"   receiveData:", convert.PrintBytes(receiveData, 8),
		"   minPacketLen:", minPacketLen, "   lengthFieldStart:", lengthFieldStart, "   lengthFieldEnd:", lengthFieldEnd)
	// check parameters
	if len(receiveData) == 0 {
		belogs.Debug("RecombineReceiveData(): len(receiveData) is empty, then return:", len(receiveData))
		return nil, make([]byte, 0), nil
	}
	if minPacketLen <= 0 {
		belogs.Error("RecombineReceiveData(): minPacketLen smaller than 0:", minPacketLen)
		return nil, nil, errors.New("minPacketLen is smaller than 0")
	}
	if lengthFieldStart <= 0 {
		belogs.Error("RecombineReceiveData(): lengthFieldStart smaller than 0:", minPacketLen)
		return nil, nil, errors.New("lengthFieldStart is smaller than 0")
	}
	if lengthFieldEnd <= 0 {
		belogs.Error("RecombineReceiveData(): lengthFieldEnd smaller than 0:", minPacketLen)
		return nil, nil, errors.New("lengthFieldEnd is smaller than 0")
	}
	if lengthFieldStart >= lengthFieldEnd {
		belogs.Error("RecombineReceiveData(): lengthFieldStart lager than lengthFieldEnd:", lengthFieldStart, lengthFieldEnd)
		return nil, nil, errors.New("lengthFieldEnd is smaller than lengthFieldStart")
	}
	packets = list.New()

	for {
		// check
		// unpack: TCP sticky packet

		// if receiveData is smaller than a packet length, packets is empty, receiveData --> leftData
		if len(receiveData) < minPacketLen {
			belogs.Debug("RecombineReceiveData(): len(receiveData) < minPacketLen, then return, len(receiveData), minPacketLen:", len(receiveData), minPacketLen)
			leftData = make([]byte, len(receiveData))
			copy(leftData, receiveData)
			return packets, leftData, nil
		}

		// get length : byte[lengthFieldStart:lengthFieldEnd]
		lengthBuffer := receiveData[lengthFieldStart:lengthFieldEnd]
		length := int(convert.Bytes2Uint64(lengthBuffer))
		belogs.Debug("RecombineReceiveData():lengthBuffer:", lengthBuffer, " length:", length)

		// length is error
		if length < minPacketLen {
			belogs.Error("RecombineReceiveData():length < minPacketLen, then return,length, minPacketLen:", length, minPacketLen)
			return packets, nil, errors.New("length is error")
		}

		// if receiveData is smaller than a packet length, the receiveData --> leftData, return
		if len(receiveData) < length {
			belogs.Debug("RecombineReceiveData(): len(receiveData) smaller then length, then return, len(receiveData), length:", len(receiveData), length)
			leftData = make([]byte, len(receiveData))
			copy(leftData, receiveData)
			return packets, leftData, nil
		} else if len(receiveData) == length {
			// if receiveData is equal to packet length, the receiveData --> packets, return
			belogs.Debug("RecombineReceiveData(): len(receiveData) equal to length, then return, len(receiveData), length:", len(receiveData), length)
			packets.PushBack(receiveData)
			return packets, make([]byte, 0), nil
		} else if len(receiveData) > length {
			// if receiveData is larger than packet length, the receiveData --> packets, leftData --> receiveData, continue
			belogs.Debug("RecombineReceiveData(): len(receiveData) lager than length, then continue, len(receiveData), length:", len(receiveData), length)
			packets.PushBack(receiveData[:length])

			// leftData continue to RecombineReceiveData
			leftData = make([]byte, length)
			copy(leftData, receiveData[length:])
			receiveData = leftData

			belogs.Debug("RecombineReceiveData(): new len(receiveData) lager than length, new(receiveData),length:", len(receiveData), length,
				"   new receiveData:", convert.PrintBytes(receiveData, 8))
		}

	}

}
func TestTcpConnection(address string, port string) (err error) {
	server := net.JoinHostPort(address, port)
	// 3 秒超时
	start := time.Now()
	conn, err := net.DialTimeout("tcp", server, 3*time.Second)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	if err != nil {
		belogs.Error("TestTcpConnection(): DialTimeout fail, server:", server, err, "  time(s):", time.Since(start))
		return err
	}

	return nil
}

func GetTcpConnKey(tcpConn *TcpConn) string {
	if tcpConn == nil {
		return ""
	}
	return tcpConn.LocalAddr().String() + "-" +
		tcpConn.RemoteAddr().String()
}

func GetUdpAddrKey(udpAddr *net.UDPAddr) string {
	if udpAddr == nil {
		return ""
	}
	return udpAddr.String()
}

func getLengthDeclarationSendData(tcptlsLengthDeclaration string, sendData []byte) (sendDataNew []byte) {
	belogs.Debug("getLengthDeclarationSendData(): tcptlsLengthDeclaration:", tcptlsLengthDeclaration,
		"   len(sendData):", len(sendData), convert.PrintBytesOneLine(sendData))
	if tcptlsLengthDeclaration == "true" {
		sendDataNew = make([]byte, 2+len(sendData))
		binary.BigEndian.PutUint16(sendDataNew, uint16(len(sendData)))
		copy(sendDataNew[2:], sendData)
		belogs.Debug("getLengthDeclarationSendData():  len(sendDataNew):", len(sendDataNew))
		return sendDataNew
	}
	return sendData
}

type TlsConfigModel struct {
	TlsPort                string `json:"tlsPort"`
	TlsRootCrtFileName     string `json:"tlsRootCrtFileName"`
	TlsPublicCrtFileName   string `json:"tlsPublicCrtFileName"`
	TlsPrivateKeyFileName  string `json:"tlsPrivateKeyFileName"`
	ClientAuth             string `json:"clientAuth"` //NoClientCert or RequireAndVerifyClientCert
	InsecureSkipVerify     bool   `json:"insecureSkipVerify"`
	KeepAlivePeriodSeconds int    `json:"keepAlivePeriodSeconds"` // if is 0, not set keepalive
}

/*
serverTlsPort := conf.String("dns-server::serverTlsPort")
path := conf.String("dns-server::programDir") + "/conf/cert/"
tlsRootCrtFileName := path + conf.String("dns-server::caTlsRoot")
tlsPublicCrtFileName := path + conf.String("dns-server::serverTlsCrt")
tlsPrivateKeyFileName := path + conf.String("dns-server::serverTlsKey")
dns-server::ClientAuth="NoClientCert" 	"RequestClientCert"
"RequireAnyClientCert"	"VerifyClientCertIfGiven" "RequireAndVerifyClientCert"
ClientAuth := conf.String("dns-server::ClientAuth")
insecureSkipVerify := conf.Bool("dns-server::insecureSkipVerify")
KeepAlivePeriodSeconds:=conf.Int("dns-server::keepAlivePeriodSeconds")
*/
func GetServerTlsConfig(tlsConfigModel TlsConfigModel) (*tls.Config, error) {

	belogs.Debug("GetServerTlsConfig(): tlsConfigModel:", jsonutil.MarshalJson(tlsConfigModel))

	cert, err := tls.LoadX509KeyPair(tlsConfigModel.TlsPublicCrtFileName,
		tlsConfigModel.TlsPrivateKeyFileName)
	if err != nil {
		belogs.Error("GetServerTlsConfig(): tlsserver  LoadX509KeyPair fail:",
			"  tlsPublicCrtFileName, tlsPrivateKeyFileName:",
			tlsConfigModel.TlsPublicCrtFileName, tlsConfigModel.TlsPrivateKeyFileName, err)
		return nil, err
	}
	belogs.Debug("GetServerTlsConfig(): tlsserver  cert:", "  tlsPublicCrtFileName, tlsPrivateKeyFileName:",
		tlsConfigModel.TlsPublicCrtFileName, tlsConfigModel.TlsPrivateKeyFileName)

	rootCrtBytes, err := os.ReadFile(tlsConfigModel.TlsRootCrtFileName)
	if err != nil {
		belogs.Error("GetServerTlsConfig(): tlsserver  ReadFile tlsRootCrtFileName fail:",
			"  tlsRootCrtFileName:", tlsConfigModel.TlsRootCrtFileName, err)
		return nil, err
	}
	belogs.Debug("GetServerTlsConfig(): tlsserver  len(rootCrtBytes):", len(rootCrtBytes),
		"  tlsRootCrtFileName:", tlsConfigModel.TlsRootCrtFileName)

	rootCertPool := x509.NewCertPool()
	ok := rootCertPool.AppendCertsFromPEM(rootCrtBytes)
	if !ok {
		belogs.Error("GetServerTlsConfig(): tlsserver  AppendCertsFromPEM tlsRootCrtFileName fail:",
			"  tlsRootCrtFileName:", tlsConfigModel.TlsRootCrtFileName, "  len(rootCrtBytes):", len(rootCrtBytes), err)
		return nil, err
	}
	belogs.Debug("GetServerTlsConfig(): tlsserver  AppendCertsFromPEM len(rootCrtBytes):", len(rootCrtBytes),
		"  tlsRootCrtFileName:", tlsConfigModel.TlsRootCrtFileName)

	clientAuthType := tls.NoClientCert
	switch tlsConfigModel.ClientAuth {
	case "RequestClientCert":
		clientAuthType = tls.RequestClientCert
	case "RequireAnyClientCert":
		clientAuthType = tls.RequireAnyClientCert
	case "VerifyClientCertIfGiven":
		clientAuthType = tls.VerifyClientCertIfGiven
	case "RequireAndVerifyClientCert":
		clientAuthType = tls.RequireAndVerifyClientCert
	}
	belogs.Debug("GetServerTlsConfig(): tlsserver clientAuthType:", clientAuthType)

	// https://stackoverflow.com/questions/63676241/how-to-set-setkeepaliveperiod-on-a-tls-conn
	var setTCPKeepAlive func(*tls.ClientHelloInfo) (*tls.Config, error)
	if tlsConfigModel.KeepAlivePeriodSeconds > 0 {
		setTCPKeepAlive = func(clientHello *tls.ClientHelloInfo) (*tls.Config, error) {
			// Check that the underlying connection really is TCP.
			if tcpConn, ok := clientHello.Conn.(*net.TCPConn); ok {
				tcpConn.SetKeepAlive(true)
				tcpConn.SetKeepAlivePeriod(time.Second * time.Duration(tlsConfigModel.KeepAlivePeriodSeconds))
				belogs.Debug("GetServerTlsConfig(): SetKeepAlivePeriod KeepAlivePeriodSeconds:", tlsConfigModel.KeepAlivePeriodSeconds)
			}
			// Make sure to return nil, nil to let the caller fall back on the default behavior.
			return nil, nil
		}
	}
	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		ClientAuth:         clientAuthType,
		RootCAs:            rootCertPool,
		GetConfigForClient: setTCPKeepAlive,
		InsecureSkipVerify: tlsConfigModel.InsecureSkipVerify,
	}
	return config, nil
}

func GetClientTlsConfig(tlsConfigModel TlsConfigModel) (*tls.Config, error) {

	belogs.Debug("GetClientTlsConfig(): tlsConfigModel:", jsonutil.MarshalJson(tlsConfigModel))

	cert, err := tls.LoadX509KeyPair(tlsConfigModel.TlsPublicCrtFileName, tlsConfigModel.TlsPrivateKeyFileName)
	if err != nil {
		belogs.Error("GetClientTlsConfig(): LoadX509KeyPair fail:",
			"  tlsPublicCrtFileName:", tlsConfigModel.TlsPublicCrtFileName,
			"  tlsPrivateKeyFileName:", tlsConfigModel.TlsPrivateKeyFileName, err)
		return nil, err
	}
	belogs.Debug("GetClientTlsConfig(): LoadX509KeyPair ok,  tlsPublicCrtFileName:", tlsConfigModel.TlsPublicCrtFileName,
		"  tlsPrivateKeyFileName:", tlsConfigModel.TlsPrivateKeyFileName)

	rootCrtBytes, err := os.ReadFile(tlsConfigModel.TlsRootCrtFileName)
	if err != nil {
		belogs.Error("GetClientTlsConfig(): ReadFile tlsRootCrtFileName fail:",
			"  tlsRootCrtFileName:", tlsConfigModel.TlsRootCrtFileName, err)
		return nil, err
	}
	belogs.Debug("GetClientTlsConfig(): ReadFile ok, tlsRootCrtFileName:", tlsConfigModel.TlsRootCrtFileName)

	rootCertPool := x509.NewCertPool()
	ok := rootCertPool.AppendCertsFromPEM(rootCrtBytes)
	if !ok {
		belogs.Error("GetClientTlsConfig(): AppendCertsFromPEM tlsRootCrtFileName fail,",
			"  tlsRootCrtFileName:", tlsConfigModel.TlsRootCrtFileName, "  len(rootCrtBytes):", len(rootCrtBytes), err)
		return nil, err
	}
	belogs.Debug("GetClientTlsConfig(): AppendCertsFromPEM ok, tlsRootCrtFileName:", tlsConfigModel.TlsRootCrtFileName)

	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            rootCertPool,
		InsecureSkipVerify: tlsConfigModel.InsecureSkipVerify,
	}
	belogs.Debug("GetClientTlsConfig(): config:", jsonutil.MarshalJson(config))
	return config, nil
}
