package tcpserver

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
)

// TcpClientProcessFunc 客户端业务回调接口
type TcpClientProcessFunc interface {
	ActiveSend(conn *net.TCPConn, processChan string) error
	OnReceive(conn *net.TCPConn, receiveData []byte) error
}

// ClientTLSConfig 客户端TLS配置
type ClientTLSConfig struct {
	ClientCertFile     string // 客户端证书路径
	ClientKeyFile      string // 客户端私钥路径
	RootCAFile         string // 根CA路径（验证服务端）
	ServerName         string // 服务端证书CN
	InsecureSkipVerify bool   // 跳过证书验证（测试用）
}

// ClientOption 客户端配置选项
type ClientOption func(*TcpClient)

// TcpClient TCP/TLS客户端核心结构体
type TcpClient struct {
	stopChan chan struct{}

	// 业务回调
	processFunc TcpClientProcessFunc

	// TLS相关
	isTLS           bool
	clientTLSConfig *ClientTLSConfig
	conn            *net.TCPConn

	// 超时配置
	readTimeout  time.Duration
	writeTimeout time.Duration

	// 并发安全
	mu     sync.Mutex
	closed bool
}

// NewTcpClient 创建客户端实例
func NewTcpClient(processFunc TcpClientProcessFunc, opts ...ClientOption) *TcpClient {
	tc := &TcpClient{
		stopChan:     make(chan struct{}),
		processFunc:  processFunc,
		readTimeout:  30 * time.Second,
		writeTimeout: 30 * time.Second,
		closed:       false,
	}
	for _, opt := range opts {
		opt(tc)
	}
	return tc
}

// WithClientTLS 启用客户端TLS配置
func WithClientTLS(tlsCfg *ClientTLSConfig) ClientOption {
	return func(tc *TcpClient) {
		if tlsCfg == nil {
			return
		}
		tc.clientTLSConfig = tlsCfg
		tc.isTLS = true
	}
}

// WithClientReadWriteTimeout 设置客户端读写超时
func WithClientReadWriteTimeout(readTimeout, writeTimeout time.Duration) ClientOption {
	return func(tc *TcpClient) {
		tc.readTimeout = readTimeout
		tc.writeTimeout = writeTimeout
	}
}

// buildTLSConfig 构建客户端TLS配置
func (tc *TcpClient) buildTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		ServerName:         tc.clientTLSConfig.ServerName,
		InsecureSkipVerify: tc.clientTLSConfig.InsecureSkipVerify,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
	}

	// 加载客户端证书（双向认证）
	if tc.clientTLSConfig.ClientCertFile != "" && tc.clientTLSConfig.ClientKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(tc.clientTLSConfig.ClientCertFile, tc.clientTLSConfig.ClientKeyFile)
		if err != nil {
			return nil, fmt.Errorf("load client cert/key fail: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// 加载根CA验证服务端
	if tc.clientTLSConfig.RootCAFile != "" {
		caPool := x509.NewCertPool()
		caData, err := os.ReadFile(tc.clientTLSConfig.RootCAFile)
		if err != nil {
			return nil, fmt.Errorf("read CA file fail: %w", err)
		}
		if !caPool.AppendCertsFromPEM(caData) {
			return nil, fmt.Errorf("append CA cert fail")
		}
		tlsConfig.RootCAs = caPool
	}

	return tlsConfig, nil
}

// Start 启动客户端连接
func (tc *TcpClient) Start(addr string) error {
	tc.mu.Lock()
	if tc.closed {
		tc.mu.Unlock()
		return fmt.Errorf("client already closed")
	}
	tc.mu.Unlock()

	var conn net.Conn
	var err error

	// 建立连接
	if tc.isTLS {
		tlsCfg, err := tc.buildTLSConfig()
		if err != nil {
			return fmt.Errorf("build TLS config fail: %w", err)
		}
		conn, err = tls.Dial("tcp", addr, tlsCfg)
		if err != nil {
			return fmt.Errorf("TLS dial fail: %w", err)
		}
	} else {
		conn, err = net.Dial("tcp", addr)
		if err != nil {
			return fmt.Errorf("TCP dial fail: %w", err)
		}
	}

	// 转换为TCPConn
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		_ = conn.Close()
		return fmt.Errorf("connection is not TCPConn")
	}
	tc.conn = tcpConn

	belogs.Info("Client connected to:", addr)

	// 启动数据读取协程
	go tc.readLoop()

	// 等待停止信号
	<-tc.stopChan

	// 关闭连接
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.closed = true
	_ = tc.conn.Close()
	belogs.Info("Client disconnected from:", addr)

	return nil
}

// readLoop 客户端读取数据循环
func (tc *TcpClient) readLoop() {
	buf := make([]byte, 4096)
	for {
		tc.conn.SetReadDeadline(time.Now().Add(tc.readTimeout))
		n, err := tc.conn.Read(buf)
		if err != nil {
			if err == net.ErrClosed || err.Error() == "EOF" {
				belogs.Debug("Server closed connection")
			} else {
				belogs.Error("Client read fail:", err)
			}
			return
		}
		if n == 0 {
			continue
		}

		// 业务回调
		if tc.processFunc != nil {
			receiveData := make([]byte, n)
			copy(receiveData, buf[:n])
			if err := tc.processFunc.OnReceive(tc.conn, receiveData); err != nil {
				belogs.Error("OnReceive fail:", err)
				return
			}
		}
	}
}

// CallProcessFunc 调用发送数据方法
func (tc *TcpClient) CallProcessFunc(data string) error {
	tc.mu.Lock()
	if tc.closed || tc.conn == nil {
		tc.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	tc.mu.Unlock()

	if tc.processFunc == nil {
		return fmt.Errorf("processFunc is nil")
	}

	return tc.processFunc.ActiveSend(tc.conn, data)
}

// CallStop 停止客户端
func (tc *TcpClient) CallStop() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.closed {
		belogs.Warn("Client already stopped")
		return
	}
	close(tc.stopChan)
	tc.closed = true
}
