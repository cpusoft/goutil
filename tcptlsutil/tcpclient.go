package tcptlsutil

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
	ActiveSend(conn net.Conn, processChan string) error
	OnReceive(conn net.Conn, receiveData []byte) error
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
	stopChan        chan struct{}
	processFunc     TcpClientProcessFunc
	isTLS           bool
	clientTLSConfig *ClientTLSConfig
	conn            net.Conn // 改为 net.Conn
	readTimeout     time.Duration
	writeTimeout    time.Duration
	mu              sync.Mutex
	closed          bool
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

// Start 启动客户端连接（修复阻塞问题：连接成功后不阻塞，通过独立goroutine监听stopChan）
func (tc *TcpClient) Start(addr string) error {
	tc.mu.Lock()
	if tc.closed {
		tc.mu.Unlock()
		return fmt.Errorf("client already closed")
	}
	tc.mu.Unlock()

	var conn net.Conn
	var err error

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

	tc.conn = conn // 直接保存，不再解包

	belogs.Info("TcpClient.Start(): connected to:", addr)

	go func() {
		defer func() {
			tc.mu.Lock()
			tc.closed = true
			tc.conn.Close()
			tc.conn = nil
			tc.mu.Unlock()
			belogs.Info("TcpClient.Start(): Client disconnected from:", addr)
		}()

		go tc.readLoop()
		<-tc.stopChan
	}()

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
				belogs.Debug("TcpClient.readLoop(): connection closed")
			} else {
				belogs.Error("TcpClient.readLoop(): read fail:", err)
			}
			return
		}
		if n == 0 {
			continue
		}

		if tc.processFunc != nil {
			receiveData := make([]byte, n)
			copy(receiveData, buf[:n])
			if err := tc.processFunc.OnReceive(tc.conn, receiveData); err != nil {
				belogs.Error("TcpClient.readLoop(): OnReceive fail:", err)
				return
			}
		}
	}
}

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

func (tc *TcpClient) CallStop() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.closed {
		belogs.Warn("TcpClient.CallStop(): Client already stopped")
		return
	}
	close(tc.stopChan)
	tc.closed = true
}
