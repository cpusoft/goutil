package tcpserver

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/cpusoft/goutil/belogs"
)

// ClientTLSConfig TLS配置结构体（客户端）
type ClientTLSConfig struct {
	// 客户端证书（双向认证时需要）
	ClientCertFile     string // 客户端证书文件路径
	ClientKeyFile      string // 客户端私钥文件路径
	RootCAFile         string // 根CA证书文件（用于验证服务端证书）
	ServerName         string // 服务端证书的CN名称（用于验证）
	InsecureSkipVerify bool   // 测试环境临时跳过证书验证（生产环境必须为false）
	MinVersion         uint16 // 最小TLS版本
	MaxVersion         uint16 // 最大TLS版本
}

// 默认TLS配置
func defaultClientTLSConfig() *ClientTLSConfig {
	return &ClientTLSConfig{
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
	}
}

type TcpClient struct {
	stopChan chan string

	tcpClientProcessChan chan string
	tcpClientProcessFunc TcpClientProcessFunc

	// 新增TLS相关字段
	clientTLSConfig *ClientTLSConfig // TLS配置（nil表示不启用TLS）
	isTLS           bool             // 是否启用TLS
	tlsConn         *tls.Conn        // TLS连接（启用TLS时使用）
	rawConn         *net.TCPConn     // 底层TCP连接
}

type TcpClientProcessFunc interface {
	ActiveSend(conn *net.TCPConn, tcpClientProcessChan string) (err error)
	OnReceive(conn *net.TCPConn, receiveData []byte) (err error)
}

// 新增Option类型，用于配置客户端
type ClientOption func(*TcpClient)

// WithClientTLS 启用TLS的配置选项
func WithClientTLS(tlsCfg *ClientTLSConfig) ClientOption {
	return func(tc *TcpClient) {
		if tlsCfg == nil {
			return
		}
		// 填充默认值
		defaultCfg := defaultClientTLSConfig()
		if tlsCfg.MinVersion == 0 {
			tlsCfg.MinVersion = defaultCfg.MinVersion
		}
		if tlsCfg.MaxVersion == 0 {
			tlsCfg.MaxVersion = defaultCfg.MaxVersion
		}
		// 测试环境提示
		if tlsCfg.InsecureSkipVerify {
			belogs.Warn("WithClientTLS(): InsecureSkipVerify is true, this is unsafe for production!")
		}

		tc.clientTLSConfig = tlsCfg
		tc.isTLS = true
	}
}

// NewTcpClient 创建TCP/TLS客户端实例（支持可选TLS）
func NewTcpClient(tcpClientProcessFunc TcpClientProcessFunc, opts ...ClientOption) (tc *TcpClient) {
	belogs.Debug("NewTcpClient():tcpClientProcessFuncs:", tcpClientProcessFunc)
	tc = &TcpClient{
		stopChan:             make(chan string),
		tcpClientProcessChan: make(chan string),
		tcpClientProcessFunc: tcpClientProcessFunc,
		isTLS:                false, // 默认不启用TLS
	}

	// 应用配置选项（如TLS配置）
	for _, opt := range opts {
		opt(tc)
	}

	belogs.Debug("NewTcpClient():tc:%p, %p, isTLS:%v:", tc.stopChan, tc.tcpClientProcessChan, tc.isTLS)
	belogs.Debug("NewTcpClient():tc:", tc)
	return tc
}

// 加载CA证书池（用于验证服务端证书）
func (tc *TcpClient) loadCACertPool() (*x509.CertPool, error) {
	if tc.clientTLSConfig.RootCAFile == "" {
		return nil, nil // 未配置根CA，使用系统默认CA池
	}

	// 读取CA证书文件
	caCertData, err := os.ReadFile(tc.clientTLSConfig.RootCAFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA file %s: %w", tc.clientTLSConfig.RootCAFile, err)
	}

	// 创建证书池并添加CA证书
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCertData) {
		return nil, fmt.Errorf("failed to append CA cert from %s", tc.clientTLSConfig.RootCAFile)
	}

	belogs.Debug("loadCACertPool(): successfully loaded CA cert pool from", tc.clientTLSConfig.RootCAFile)
	return caPool, nil
}

// 构建TLS配置
func (tc *TcpClient) buildTLSConfig() (*tls.Config, error) {
	clientTLSConfig := &tls.Config{
		InsecureSkipVerify: tc.clientTLSConfig.InsecureSkipVerify,
		MinVersion:         tc.clientTLSConfig.MinVersion,
		MaxVersion:         tc.clientTLSConfig.MaxVersion,
		ServerName:         tc.clientTLSConfig.ServerName,
	}

	// 加载客户端证书（双向认证）
	if tc.clientTLSConfig.ClientCertFile != "" && tc.clientTLSConfig.ClientKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(tc.clientTLSConfig.ClientCertFile, tc.clientTLSConfig.ClientKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert/key: %w", err)
		}
		clientTLSConfig.Certificates = []tls.Certificate{cert}
		belogs.Debug("buildTLSConfig(): loaded client certificate")
	}

	// 加载根CA证书池（验证服务端证书）
	caPool, err := tc.loadCACertPool()
	if err != nil {
		return nil, err
	}
	if caPool != nil {
		clientTLSConfig.RootCAs = caPool
	}

	return clientTLSConfig, nil
}

// server: **.**.**.**:port
func (tc *TcpClient) Start(server string) (err error) {
	belogs.Debug("Start():create client, server is:", server, " isTLS:", tc.isTLS)

	tcpServer, err := net.ResolveTCPAddr("tcp", server)
	if err != nil {
		belogs.Error("Start(): ResolveTCPAddr fail: ", server, err)
		return err
	}

	// 根据是否启用TLS创建不同的连接
	if tc.isTLS {
		// 构建TLS配置
		clientTLSConfig, err := tc.buildTLSConfig()
		if err != nil {
			belogs.Error("Start(): buildTLSConfig fail: ", err)
			return err
		}

		// 先建立底层TCP连接
		rawConn, err := net.DialTCP("tcp4", nil, tcpServer)
		if err != nil {
			belogs.Error("Start(): Dial TCP fail (TLS): ", server, err)
			return err
		}
		tc.rawConn = rawConn

		// 创建TLS连接并完成握手
		tlsConn := tls.Client(rawConn, clientTLSConfig)
		if err := tlsConn.Handshake(); err != nil {
			rawConn.Close()
			belogs.Error("Start(): TLS Handshake fail: ", server, err)
			return err
		}

		// 验证TLS连接状态
		state := tlsConn.ConnectionState()
		belogs.Info("Start(): TLS handshake success: ", server,
			"  version:", tlsVersionToString(state.Version),
			"  cipher suite:", tls.CipherSuiteName(state.CipherSuite))
		/*
			// 验证server证书的CN值
			if len(state.PeerCertificates) == 0 {
				return errors.New("no server certificate provided")
			}
			serverCert := state.PeerCertificates[0]
			expectedServerCN := "localhost" // 预期的服务端CN
			if serverCert.Subject.CommonName != expectedServerCN {
				return fmt.Errorf("server cert CN %s does not match expected %s",
					serverCert.Subject.CommonName, expectedServerCN)
			}
		*/
		tc.tlsConn = tlsConn
		belogs.Debug("Start():create TLS client ok, server:", server, "   tlsConn:", tlsConn)
	} else {
		// 创建普通TCP连接
		conn, err := net.DialTCP("tcp4", nil, tcpServer)
		if err != nil {
			belogs.Error("Start(): Dial fail, server:", server, "  tcpServer:", tcpServer, err)
			return err
		}
		tc.rawConn = conn
		belogs.Debug("Start():create TCP client ok, server:", server, "   conn:", conn)
	}

	// 确保连接最终关闭
	defer func() {
		if tc.isTLS && tc.tlsConn != nil {
			_ = tc.tlsConn.Close()
		}
		if tc.rawConn != nil && !tc.isTLS { // TLS连接关闭会自动关闭底层TCP
			_ = tc.rawConn.Close()
		}
	}()

	// 启动主动发送协程
	go tc.waitActiveSend()

	// 启动接收协程
	go tc.waitReceive()

	// 等待退出信号
	for {
		belogs.Debug("Start():wait for stop  ", server, "   conn:", tc.rawConn)
		select {
		case stop := <-tc.stopChan:
			if stop == "stop" {
				close(tc.stopChan)
				close(tc.tcpClientProcessChan)
				belogs.Info("Start(): end client: ", server)
				return nil
			}
		}
	}
}

// exit: to quit
func (tc *TcpClient) CallProcessFunc(clientProcessFunc string) {
	belogs.Debug("CallProcessFunc():  clientProcessFunc:", clientProcessFunc)
	tc.tcpClientProcessChan <- clientProcessFunc
}

func (tc *TcpClient) CallStop() {
	belogs.Debug("CallStop():")
	tc.stopChan <- "stop"
}

// 改造waitActiveSend：适配TLS/非TLS连接
func (tc *TcpClient) waitActiveSend() {
	belogs.Debug("waitActiveSend():wait for ActiveSend, conn:", tc.rawConn, " isTLS:", tc.isTLS)
	for {
		select {
		case tcpClientProcessChan := <-tc.tcpClientProcessChan:
			belogs.Debug("waitActiveSend():  tcpClientProcess:", tcpClientProcessChan)
			start := time.Now()

			// 调用回调函数（始终传入底层TCP连接，保持接口兼容）
			err := tc.tcpClientProcessFunc.ActiveSend(tc.rawConn, tcpClientProcessChan)
			if err != nil {
				belogs.Error("waitActiveSend(): tcpClientProcessFunc.ActiveSend fail:  conn:", tc.rawConn, err)
				// 发送失败时触发停止
				tc.CallStop()
				return
			}
			belogs.Info("waitActiveSend(): tcpClientProcessChan:", tcpClientProcessChan, "  time(s):", time.Since(start))
		}
	}
}

// 改造waitReceive：适配TLS/非TLS连接的读取逻辑
func (tc *TcpClient) waitReceive() {
	belogs.Debug("waitReceive():wait for OnReceive, conn:", tc.rawConn, " isTLS:", tc.isTLS)
	// one packet
	buffer := make([]byte, 2048)

	// 选择读取的连接（TLS或原始TCP）
	var readConn io.Reader
	if tc.isTLS && tc.tlsConn != nil {
		readConn = tc.tlsConn
	} else {
		readConn = tc.rawConn
	}

	// wait for new packet to read
	for {
		n, err := readConn.Read(buffer)
		start := time.Now()
		belogs.Debug("waitReceive(): tcpClient Read n Bytes:", tc.rawConn, "   n:", n)
		if err != nil {
			if err == io.EOF {
				// is not error, just server close
				belogs.Debug("waitReceive(): tcpClient Read io.EOF, close: ", tc.rawConn, err)
				tc.CallStop()
				return
			}
			// 处理TLS特定错误
			if tlsErr, ok := err.(tls.RecordHeaderError); ok {
				belogs.Error("waitReceive(): TLS record error: ", tc.rawConn, tlsErr)
				tc.CallStop()
				return
			}
			belogs.Error("waitReceive(): tcpClient Read fail, fail:", tc.rawConn, err)
			tc.CallStop()
			return
		}
		if n == 0 {
			continue
		}

		// copy to new []byte
		receiveData := make([]byte, n)
		copy(receiveData, buffer[0:n])
		belogs.Info("waitReceive():tcpClient conn: ", tc.rawConn, "  len(receiveData): ", len(receiveData),
			" , will call client tcpClientProcessFunc,  time(s):", time.Since(start))
		err = tc.tcpClientProcessFunc.OnReceive(tc.rawConn, receiveData)
		belogs.Debug("waitReceive():tcpClient OnReceive, conn: ", tc.rawConn, "  len(receiveData): ", len(receiveData), "  time(s):", time.Since(start))
		if err != nil {
			belogs.Error("waitReceive(): tcpClient OnReceive fail ,will stop client : ", tc.rawConn, err)
			tc.CallStop()
			break
		}
	}
}
