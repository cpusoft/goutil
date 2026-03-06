package tcpserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
)

// ServerTLSConfig TLS配置结构体（增强版）
type ServerTLSConfig struct {
	// 服务端证书配置
	ServerCertFile string // 服务端证书文件 (server.crt)
	ServerKeyFile  string // 服务端私钥文件 (server.key)

	// 根CA配置（用于验证对方证书）
	RootCAFile string // 根CA证书文件 (ca.crt)

	// 客户端证书配置（仅客户端使用）
	ClientCertFile string // 客户端证书文件 (client.crt)
	ClientKeyFile  string // 客户端私钥文件 (client.key)

	// TLS安全配置
	ClientAuth   tls.ClientAuthType // 服务端对客户端的认证方式
	MinVersion   uint16             // 最小TLS版本
	MaxVersion   uint16             // 最大TLS版本
	CipherSuites []uint16           // 允许的加密套件
}

// 默认TLS配置（安全合规）
func defaultServerTLSConfig() *ServerTLSConfig {
	return &ServerTLSConfig{
		ClientAuth: tls.RequireAndVerifyClientCert, // 双向认证
		MinVersion: tls.VersionTLS12,               // 禁用TLS1.0/1.1
		MaxVersion: tls.VersionTLS13,               // 启用TLS1.3
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
	}
}

type TcpServer struct {
	tcpConns      map[string]*net.TCPConn
	tcpConnsMutex sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc

	tcpServerProcessFunc TcpServerProcessFunc
	readTimeout          time.Duration
	writeTimeout         time.Duration
	serverTLSConfig      *ServerTLSConfig // TLS配置
	isTLS                bool             // 是否启用TLS
	tlsListener          net.Listener
}

// 可选配置项
type Option func(*TcpServer)

// WithReadWriteTimeout 设置读写超时
func WithReadWriteTimeout(read, write time.Duration) Option {
	return func(ts *TcpServer) {
		ts.readTimeout = read
		ts.writeTimeout = write
	}
}

// WithServerTLS 启用双向TLS认证
func WithServerTLS(tlsCfg *ServerTLSConfig) Option {
	return func(ts *TcpServer) {
		if tlsCfg == nil {
			return
		}

		// 填充默认值
		defaultCfg := defaultServerTLSConfig()
		if tlsCfg.MinVersion == 0 {
			tlsCfg.MinVersion = defaultCfg.MinVersion
		}
		if tlsCfg.MaxVersion == 0 {
			tlsCfg.MaxVersion = defaultCfg.MaxVersion
		}
		if len(tlsCfg.CipherSuites) == 0 {
			tlsCfg.CipherSuites = defaultCfg.CipherSuites
		}
		if tlsCfg.ClientAuth == 0 {
			tlsCfg.ClientAuth = defaultCfg.ClientAuth
		}

		// 校验必要的配置
		if tlsCfg.ServerCertFile == "" || tlsCfg.ServerKeyFile == "" || tlsCfg.RootCAFile == "" {
			belogs.Warn("WithServerTLS(): missing required TLS config: ServerCertFile/ServerKeyFile/RootCAFile must be set")
			return
		}

		ts.serverTLSConfig = tlsCfg
		ts.isTLS = true
	}
}

// NewTcpServer 创建TCP/TLS服务端实例
func NewTcpServer(tcpServerProcessFunc TcpServerProcessFunc, opts ...Option) (ts *TcpServer) {
	belogs.Debug("NewTcpServer():tcpProcessFunc:", tcpServerProcessFunc)
	ctx, cancel := context.WithCancel(context.Background())
	ts = &TcpServer{
		tcpConns:             make(map[string]*net.TCPConn),
		ctx:                  ctx,
		cancel:               cancel,
		tcpServerProcessFunc: tcpServerProcessFunc,
		readTimeout:          30 * time.Second,
		writeTimeout:         10 * time.Second,
		isTLS:                false,
	}

	// 应用配置项
	for _, opt := range opts {
		opt(ts)
	}

	return ts
}

// 加载CA证书池（核心函数）
func loadCACertPool(caFile string) (*x509.CertPool, error) {
	// 读取CA证书文件
	caCertData, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA file %s: %w", caFile, err)
	}

	// 创建证书池并添加CA证书
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCertData) {
		return nil, fmt.Errorf("failed to append CA cert from %s", caFile)
	}

	belogs.Debug("loadCACertPool(): successfully loaded CA cert pool from", caFile)
	return caPool, nil
}

// 构建服务端TLS配置（双向认证）
func (ts *TcpServer) buildServerTLSConfig() (*tls.Config, error) {
	// 1. 加载服务端证书和私钥
	serverCert, err := tls.LoadX509KeyPair(ts.serverTLSConfig.ServerCertFile, ts.serverTLSConfig.ServerKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server cert/key: %w", err)
	}

	// 2. 加载根CA证书池（用于验证客户端证书）
	caPool, err := loadCACertPool(ts.serverTLSConfig.RootCAFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA pool: %w", err)
	}

	// 3. 构建TLS配置
	serverTLSConfig := &tls.Config{
		// 服务端证书
		Certificates: []tls.Certificate{serverCert},

		// 验证客户端证书的配置
		ClientCAs:  caPool,                        // 用于验证客户端证书的CA池
		ClientAuth: ts.serverTLSConfig.ClientAuth, // 要求并验证客户端证书

		// 安全配置
		MinVersion:         ts.serverTLSConfig.MinVersion,
		MaxVersion:         ts.serverTLSConfig.MaxVersion,
		CipherSuites:       ts.serverTLSConfig.CipherSuites,
		InsecureSkipVerify: false, // 生产环境必须为false（验证服务端证书）
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			// 自定义验证逻辑：确保客户端证书由根CA签发
			if len(rawCerts) == 0 {
				return errors.New("no client certificate provided")
			}

			// 解析客户端证书
			cert, err := x509.ParseCertificate(rawCerts[0])
			if err != nil {
				return fmt.Errorf("failed to parse client cert: %w", err)
			}

			/*
				// 验证客户端证书的CN值（可选，增强安全性）
				allowedCNs := []string{"test-client", "prod-client", "dev-client"} // 允许的CN列表
				clientCN := cert.Subject.CommonName
				isValidCN := false
				for _, cn := range allowedCNs {
					if clientCN == cn {
						isValidCN = true
						break
					}
				}
				if !isValidCN {
					return fmt.Errorf("client cert CN %s is not allowed (allowed: %v)", clientCN, allowedCNs)
				}
			*/

			// 验证证书签名
			_, err = cert.Verify(x509.VerifyOptions{
				Roots:     caPool,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			})
			if err != nil {
				return fmt.Errorf("client cert verification failed: %w", err)
			}

			belogs.Info("VerifyPeerCertificate(): client cert verified successfully, subject:", cert.Subject.CommonName)
			return nil
		},
	}

	return serverTLSConfig, nil
}

// Start 启动双向TLS认证的TCP服务
// Start 启动双向TLS认证的TCP服务
func (ts *TcpServer) Start(server string) (err error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", server)
	if err != nil {
		belogs.Error("Start(): ResolveTCPAddr fail: ", server, err)
		return err
	}

	// 根据是否启用TLS创建监听器
	var listener net.Listener
	var underlyingTCPListener *net.TCPListener // 底层TCP监听器（统一用于SetDeadline）
	if ts.isTLS {
		// 构建双向TLS配置
		serverTLSConfig, err := ts.buildServerTLSConfig()
		if err != nil {
			belogs.Error("Start(): buildServerTLSConfig fail: ", err)
			return err
		}

		// 1. 先创建TCP监听器（作为TLS监听器的底层）
		tcpListener, err := net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			belogs.Error("Start(): ListenTCP fail (TLS base): ", server, err)
			return err
		}
		underlyingTCPListener = tcpListener

		// 2. 基于TCP监听器创建TLS监听器
		listener = tls.NewListener(tcpListener, serverTLSConfig)
		ts.tlsListener = listener
		belogs.Info("Start(): create mutual TLS server ok, server is ", server)
	} else {
		// 创建普通TCP监听器
		tcpListener, err := net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			belogs.Error("Start(): ListenTCP fail: ", server, err)
			return err
		}
		listener = tcpListener
		underlyingTCPListener = tcpListener
		belogs.Info("Start(): create TCP server ok, server is ", server)
	}

	// 优雅退出
	go func() {
		<-ts.ctx.Done()
		belogs.Info("Start(): closing listener...")
		listener.Close()
	}()

	// 接受客户端连接
	for {
		// 统一设置超时（核心修复：使用底层TCP监听器）
		if underlyingTCPListener != nil {
			_ = underlyingTCPListener.SetDeadline(time.Now().Add(1 * time.Second))
		}

		conn, err := listener.Accept()
		if err != nil {
			if ts.ctx.Err() != nil {
				belogs.Info("Start(): server is shutting down, exit accept loop")
				return nil
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			belogs.Error("Start(): Accept remote fail: ", server, err)
			continue
		}

		// 处理TLS连接握手和验证（原有逻辑不变）
		var tcpConn *net.TCPConn
		if tlsConn, ok := conn.(*tls.Conn); ok {
			// 强制完成TLS握手
			if err := tlsConn.Handshake(); err != nil {
				belogs.Error("Start(): TLS Handshake fail: ", conn.RemoteAddr(), err)
				conn.Close()
				continue
			}

			// 打印TLS连接信息
			state := tlsConn.ConnectionState()
			belogs.Info("Start(): TLS handshake success: ", conn.RemoteAddr(),
				"  version:", tlsVersionToString(state.Version),
				"  cipher suite:", tls.CipherSuiteName(state.CipherSuite),
				"  client cert subject:", state.PeerCertificates[0].Subject.CommonName)

			// 获取底层TCP连接
			netConn := tlsConn.NetConn()
			if tc, ok := netConn.(*net.TCPConn); ok {
				tcpConn = tc
			} else {
				belogs.Error("Start(): invalid underlying connection type")
				conn.Close()
				continue
			}
		} else if tc, ok := conn.(*net.TCPConn); ok {
			tcpConn = tc
		} else {
			belogs.Error("Start(): invalid connection type")
			conn.Close()
			continue
		}

		belogs.Info("Start(): Accept remote: ", tcpConn.RemoteAddr().String())

		// 处理新连接
		err = ts.OnConnect(tcpConn)
		if err != nil {
			belogs.Error("Start(): OnConnect fail: ", server, tcpConn.RemoteAddr().String(), err)
			ts.OnClose(tcpConn)
			continue
		}

		// 启动goroutine处理连接
		go ts.ReceiveAndSend(tcpConn)
	}
}

// ------------------- 以下是保留的原有功能代码 -------------------
func (ts *TcpServer) Stop() {
	belogs.Info("Stop(): stopping tcp/tls server...")
	ts.cancel()

	// 关闭所有连接
	ts.tcpConnsMutex.Lock()
	defer ts.tcpConnsMutex.Unlock()
	for addr, conn := range ts.tcpConns {
		belogs.Info("Stop(): closing conn: ", addr)
		conn.Close()
	}
	ts.tcpConns = make(map[string]*net.TCPConn)

	if ts.tlsListener != nil {
		ts.tlsListener.Close()
	}
}

type TcpServerProcessFunc interface {
	OnConnect(conn *net.TCPConn) (err error)
	OnReceiveAndSend(conn *net.TCPConn, receiveData []byte) (err error)
	OnClose(conn *net.TCPConn)
	ActiveSend(conn *net.TCPConn, sendData []byte) (err error)
}

func (ts *TcpServer) OnConnect(conn *net.TCPConn) (err error) {
	start := time.Now()

	err = ts.tcpServerProcessFunc.OnConnect(conn)
	if err != nil {
		belogs.Error("OnConnect(): tcpServerProcessFunc.OnConnect fail:", conn, err,
			"   time(s):", time.Since(start))
		return err
	}

	ts.tcpConnsMutex.Lock()
	defer ts.tcpConnsMutex.Unlock()
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(300 * time.Second)
	conn.SetNoDelay(true)

	addr := conn.RemoteAddr().String()
	ts.tcpConns[addr] = conn
	belogs.Info("OnConnect():add conn: ", addr, "   len(tcpConns): ", len(ts.tcpConns),
		"   time(s):", time.Since(start))
	return nil
}

func (ts *TcpServer) ReceiveAndSend(conn *net.TCPConn) {
	addr := conn.RemoteAddr().String()
	belogs.Debug("ReceiveAndSend(): start processing conn: ", addr)
	defer func() {
		belogs.Debug("ReceiveAndSend(): stop processing conn: ", addr)
		ts.OnClose(conn)
	}()

	buffer := make([]byte, 2048)
	for {
		select {
		case <-ts.ctx.Done():
			belogs.Info("ReceiveAndSend(): server is shutting down, exit conn: ", addr)
			return
		default:
		}

		conn.SetReadDeadline(time.Now().Add(ts.readTimeout))
		n, err := conn.Read(buffer)
		start := time.Now()
		belogs.Debug("ReceiveAndSend():server read: Read n: ", addr, n)

		if err != nil {
			if err == io.EOF {
				belogs.Debug("ReceiveAndSend():server Read io.EOF, client close: ", addr)
				return
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				belogs.Debug("ReceiveAndSend():server Read timeout: ", addr)
				continue
			}
			belogs.Error("ReceiveAndSend():server Read fail: ", addr, err)
			return
		}

		if n == 0 {
			continue
		}

		receiveData := make([]byte, n)
		copy(receiveData, buffer[:n])
		belogs.Debug("ReceiveAndSend():server conn: ", addr, "  len(receiveData):", len(receiveData),
			" , will call process func: OnReceiveAndSend,  time(s):", time.Since(start))

		err = ts.tcpServerProcessFunc.OnReceiveAndSend(conn, receiveData)
		if err != nil {
			belogs.Error("ReceiveAndSend(): OnReceiveAndSend fail: ", addr, err)
			break
		}
	}
}

func (ts *TcpServer) OnClose(conn *net.TCPConn) {
	if conn == nil {
		return
	}
	addr := conn.RemoteAddr().String()
	start := time.Now()

	ts.tcpServerProcessFunc.OnClose(conn)

	ts.tcpConnsMutex.Lock()
	defer ts.tcpConnsMutex.Unlock()
	delete(ts.tcpConns, addr)
	belogs.Info("OnClose():server,conn: ", addr, "   new len(tcpConns): ", len(ts.tcpConns),
		"  time(s):", time.Since(start))

	if err := conn.Close(); err != nil {
		belogs.Error("OnClose(): close conn fail: ", addr, err)
	}
}

func (ts *TcpServer) ActiveSend(sendData []byte) (err error) {
	ts.tcpConnsMutex.RLock()
	conns := make([]*net.TCPConn, 0, len(ts.tcpConns))
	for _, conn := range ts.tcpConns {
		conns = append(conns, conn)
	}
	ts.tcpConnsMutex.RUnlock()

	start := time.Now()
	belogs.Debug("ActiveSend():server,len(sendData):", len(sendData), "   len(tcpConns): ", len(conns))

	var wg sync.WaitGroup
	for _, conn := range conns {
		wg.Add(1)
		go func(c *net.TCPConn) {
			defer wg.Done()
			addr := c.RemoteAddr().String()
			c.SetWriteDeadline(time.Now().Add(ts.writeTimeout))
			err := ts.tcpServerProcessFunc.ActiveSend(c, sendData)
			if err != nil {
				belogs.Error("ActiveSend(): fail, client: ", addr, err)
				ts.OnClose(c)
			} else {
				belogs.Debug("ActiveSend(): success, client: ", addr)
			}
		}(conn)
	}
	wg.Wait()

	belogs.Info("ActiveSend(): send to all clients ok,  len(sendData):", len(sendData), "   len(tcpConns): ", len(conns),
		"  time(s):", time.Since(start))
	return
}
