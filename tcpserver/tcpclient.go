package tcpserver

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
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
	stopChan chan struct{} // 改为struct{}，仅用于信号，不传递数据

	tcpClientProcessChan chan string
	tcpClientProcessFunc TcpClientProcessFunc

	// 新增TLS相关字段
	clientTLSConfig *ClientTLSConfig // TLS配置（nil表示不启用TLS）
	isTLS           bool             // 是否启用TLS
	tlsConn         *tls.Conn        // TLS连接（启用TLS时使用）
	rawConn         *net.TCPConn     // 底层TCP连接
	mu              sync.Mutex       // 保护通道/连接关闭
	closed          bool             // 标记是否已关闭
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
		stopChan:             make(chan struct{}),    // 无缓冲通道，仅用于退出信号
		tcpClientProcessChan: make(chan string, 100), // 增加缓冲区，避免阻塞
		tcpClientProcessFunc: tcpClientProcessFunc,
		isTLS:                false, // 默认不启用TLS
		closed:               false,
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

	// 加锁防止并发关闭
	tc.mu.Lock()
	if tc.closed {
		tc.mu.Unlock()
		return fmt.Errorf("client already closed")
	}
	tc.mu.Unlock()

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

	// 启动主动发送协程
	go tc.waitActiveSend()

	// 启动接收协程
	go tc.waitReceive()

	// 等待退出信号
	<-tc.stopChan // 阻塞直到收到退出信号
	belogs.Info("Start(): client exiting, server:", server)

	// 关闭连接（加锁保护）
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.isTLS && tc.tlsConn != nil {
		_ = tc.tlsConn.Close()
	}
	if tc.rawConn != nil {
		_ = tc.rawConn.Close()
	}
	tc.closed = true

	return nil
}

// exit: to quit
func (tc *TcpClient) CallProcessFunc(clientProcessFunc string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	// 检查是否已关闭，避免向关闭的通道发送数据
	if tc.closed {
		belogs.Warn("CallProcessFunc(): client is closed, skip send")
		return
	}
	// 使用非阻塞发送，避免协程阻塞
	select {
	case tc.tcpClientProcessChan <- clientProcessFunc:
	default:
		belogs.Warn("CallProcessFunc(): process chan is full, skip send")
	}
}

func (tc *TcpClient) CallStop() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.closed {
		belogs.Warn("CallStop(): client already closed")
		return
	}
	belogs.Debug("CallStop():")
	// 关闭退出信号通道（仅关闭一次）
	close(tc.stopChan)
	tc.closed = true
}

// 改造waitActiveSend：适配TLS/非TLS连接，增加关闭保护
func (tc *TcpClient) waitActiveSend() {
	belogs.Debug("waitActiveSend():wait for ActiveSend, conn:", tc.rawConn, " isTLS:", tc.isTLS)
	for {
		select {
		case <-tc.stopChan: // 优先处理退出信号
			belogs.Info("waitActiveSend(): client stopping, exit send loop")
			return
		case tcpClientProcessChan, ok := <-tc.tcpClientProcessChan:
			if !ok { // 通道已关闭
				belogs.Info("waitActiveSend(): process chan closed, exit")
				return
			}
			belogs.Debug("waitActiveSend():  tcpClientProcess:", tcpClientProcessChan)
			start := time.Now()

			// 加锁检查连接是否有效
			tc.mu.Lock()
			if tc.closed || tc.rawConn == nil {
				tc.mu.Unlock()
				belogs.Warn("waitActiveSend(): client closed, skip send")
				continue
			}
			conn := tc.rawConn
			tc.mu.Unlock()

			// 调用回调函数（始终传入底层TCP连接，保持接口兼容）
			err := tc.tcpClientProcessFunc.ActiveSend(conn, tcpClientProcessChan)
			if err != nil {
				belogs.Error("waitActiveSend(): tcpClientProcessFunc.ActiveSend fail:  conn:", conn, err)
				// 仅在未关闭时触发停止
				tc.mu.Lock()
				notClosed := !tc.closed
				tc.mu.Unlock()
				if notClosed {
					tc.CallStop()
				}
			} else {
				belogs.Info("waitActiveSend(): tcpClientProcessChan:", tcpClientProcessChan, "  time(s):", time.Since(start))
			}
		}
	}
}

// 改造waitReceive：适配TLS/非TLS连接的读取逻辑，增加关闭保护
func (tc *TcpClient) waitReceive() {
	belogs.Debug("waitReceive():wait for OnReceive, conn:", tc.rawConn, " isTLS:", tc.isTLS)
	// one packet
	buffer := make([]byte, 2048)

	// 选择读取的连接（TLS或原始TCP）
	var readConn io.Reader
	tc.mu.Lock()
	if tc.isTLS && tc.tlsConn != nil {
		readConn = tc.tlsConn
	} else {
		readConn = tc.rawConn
	}
	connValid := readConn != nil
	tc.mu.Unlock()

	if !connValid {
		belogs.Error("waitReceive(): no valid read conn")
		tc.CallStop()
		return
	}

	// wait for new packet to read
	for {
		select {
		case <-tc.stopChan: // 优先处理退出信号
			belogs.Info("waitReceive(): client stopping, exit receive loop")
			return
		default:
		}

		n, err := readConn.Read(buffer)
		start := time.Now()
		belogs.Debug("waitReceive(): tcpClient Read n Bytes:", tc.rawConn, "   n:", n)
		if err != nil {
			if err == io.EOF {
				// is not error, just server close
				belogs.Debug("waitReceive(): tcpClient Read io.EOF, close: ", tc.rawConn, err)
			} else if tlsErr, ok := err.(tls.RecordHeaderError); ok {
				belogs.Error("waitReceive(): TLS record error: ", tc.rawConn, tlsErr)
			} else {
				belogs.Error("waitReceive(): tcpClient Read fail, fail:", tc.rawConn, err)
			}
			// 仅在未关闭时触发停止
			tc.mu.Lock()
			notClosed := !tc.closed
			tc.mu.Unlock()
			if notClosed {
				tc.CallStop()
			}
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

		// 加锁检查连接有效性
		tc.mu.Lock()
		conn := tc.rawConn
		tc.mu.Unlock()
		err = tc.tcpClientProcessFunc.OnReceive(conn, receiveData)
		belogs.Debug("waitReceive():tcpClient OnReceive, conn: ", tc.rawConn, "  len(receiveData): ", len(receiveData), "  time(s):", time.Since(start))
		if err != nil {
			belogs.Error("waitReceive(): tcpClient OnReceive fail ,will stop client : ", tc.rawConn, err)
			// 仅在未关闭时触发停止
			tc.mu.Lock()
			notClosed := !tc.closed
			tc.mu.Unlock()
			if notClosed {
				tc.CallStop()
			}
			break
		}
	}
}
