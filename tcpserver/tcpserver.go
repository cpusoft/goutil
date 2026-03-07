package tcpserver

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
)

// TcpServerProcessFunc 服务器业务回调接口
type TcpServerProcessFunc interface {
	OnConnect(conn *net.TCPConn) (err error)
	OnReceiveAndSend(conn *net.TCPConn, receiveData []byte) (err error)
	OnClose(conn *net.TCPConn)
	ActiveSend(conn *net.TCPConn, sendData []byte) (err error)
}

// ServerTLSConfig 服务端TLS配置
type ServerTLSConfig struct {
	ServerCertFile string             // 服务端证书路径
	ServerKeyFile  string             // 服务端私钥路径
	RootCAFile     string             // 根CA路径（仅双向认证用）
	ClientAuth     tls.ClientAuthType // 客户端认证类型
}

// ServerOption 服务器配置选项
type ServerOption func(*TcpServer)

// TcpServer TCP/TLS服务器核心结构体
type TcpServer struct {
	stopChan chan struct{}

	// 业务回调
	processFunc TcpServerProcessFunc

	// TLS相关
	isTLS           bool
	serverTLSConfig *ServerTLSConfig
	listener        net.Listener // 通用监听器（兼容TLS/非TLS）

	// 超时配置
	readTimeout  time.Duration
	writeTimeout time.Duration

	// 并发安全
	mu     sync.Mutex
	closed bool

	// 新增：客户端连接管理
	tcpConns      map[string]*net.TCPConn // key: 客户端地址(RemoteAddr().String())
	tcpConnsMutex sync.RWMutex            // 读写锁，支持高并发读写
}

// NewTcpServer 创建服务器实例
func NewTcpServer(processFunc TcpServerProcessFunc, opts ...ServerOption) *TcpServer {
	ts := &TcpServer{
		stopChan:     make(chan struct{}),
		processFunc:  processFunc,
		readTimeout:  30 * time.Second,
		writeTimeout: 30 * time.Second,
		closed:       false,
		tcpConns:     make(map[string]*net.TCPConn), // 初始化连接映射表
	}
	for _, opt := range opts {
		opt(ts)
	}
	return ts
}

// WithServerTLS 启用TLS配置
func WithServerTLS(tlsCfg *ServerTLSConfig) ServerOption {
	return func(ts *TcpServer) {
		if tlsCfg == nil {
			return
		}
		if tlsCfg.ClientAuth == 0 {
			tlsCfg.ClientAuth = tls.NoClientCert
		}
		ts.serverTLSConfig = tlsCfg
		ts.isTLS = true
	}
}

// WithReadWriteTimeout 设置读写超时
func WithReadWriteTimeout(readTimeout, writeTimeout time.Duration) ServerOption {
	return func(ts *TcpServer) {
		ts.readTimeout = readTimeout
		ts.writeTimeout = writeTimeout
	}
}

// buildTLSConfig 构建TLS配置
func (ts *TcpServer) buildTLSConfig() (*tls.Config, error) {
	// 加载服务端证书
	cert, err := tls.LoadX509KeyPair(ts.serverTLSConfig.ServerCertFile, ts.serverTLSConfig.ServerKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load server cert/key fail: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS13,
		ClientAuth:   ts.serverTLSConfig.ClientAuth,
	}

	// 仅双向认证时加载ClientCAs
	needClientCA := ts.serverTLSConfig.ClientAuth == tls.RequireAnyClientCert ||
		ts.serverTLSConfig.ClientAuth == tls.RequireAndVerifyClientCert

	if needClientCA {
		if ts.serverTLSConfig.RootCAFile == "" {
			return nil, fmt.Errorf("ClientAuth=%s requires RootCAFile", ts.serverTLSConfig.ClientAuth)
		}
		clientCAPool := x509.NewCertPool()
		caData, err := os.ReadFile(ts.serverTLSConfig.RootCAFile)
		if err != nil {
			return nil, fmt.Errorf("read CA file fail: %w", err)
		}
		if !clientCAPool.AppendCertsFromPEM(caData) {
			return nil, fmt.Errorf("append CA cert fail")
		}
		tlsConfig.ClientCAs = clientCAPool

		// 客户端证书验证逻辑
		tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			if len(rawCerts) == 0 {
				return errors.New("no client certificate provided")
			}
			cert, err := x509.ParseCertificate(rawCerts[0])
			if err != nil {
				return fmt.Errorf("parse client cert fail: %w", err)
			}
			_, err = cert.Verify(x509.VerifyOptions{
				Roots:     clientCAPool,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			})
			if err != nil {
				return fmt.Errorf("client cert verify fail: %w", err)
			}
			belogs.Info("Client cert verified, CN:", cert.Subject.CommonName)
			return nil
		}
	}

	return tlsConfig, nil
}

// Start 启动服务器
func (ts *TcpServer) Start(addr string) error {
	ts.mu.Lock()
	if ts.closed {
		ts.mu.Unlock()
		return fmt.Errorf("server already closed")
	}
	ts.mu.Unlock()

	var err error

	// 启动监听器
	if ts.isTLS {
		tlsCfg, err := ts.buildTLSConfig()
		if err != nil {
			return fmt.Errorf("build TLS config fail: %w", err)
		}
		ts.listener, err = tls.Listen("tcp", addr, tlsCfg)
		if err != nil {
			return fmt.Errorf("TLS listen fail: %w", err)
		}
	} else {
		ts.listener, err = net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("TCP listen fail: %w", err)
		}
	}

	belogs.Info("Server started, addr:", addr, " isTLS:", ts.isTLS)

	// 接收连接
	go ts.acceptConnections()

	// 等待停止信号
	<-ts.stopChan

	// 关闭资源
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.closed = true

	// 新增：停止时关闭所有客户端连接
	ts.tcpConnsMutex.Lock()
	for addr, conn := range ts.tcpConns {
		_ = conn.Close()
		delete(ts.tcpConns, addr)
	}
	ts.tcpConnsMutex.Unlock()

	if ts.listener != nil {
		_ = ts.listener.Close()
	}
	belogs.Info("Server stopped, addr:", addr)

	return nil
}

// acceptConnections 接收客户端连接
func (ts *TcpServer) acceptConnections() {
	for {
		conn, err := ts.listener.Accept()
		if err != nil {
			select {
			case <-ts.stopChan:
				return
			default:
				belogs.Error("Accept connection fail:", err)
				continue
			}
		}

		// 转换为TCPConn
		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			belogs.Error("Connection is not TCPConn")
			_ = conn.Close()
			continue
		}

		// 新增：将新连接加入连接列表
		clientAddr := tcpConn.RemoteAddr().String()
		ts.tcpConnsMutex.Lock()
		ts.tcpConns[clientAddr] = tcpConn
		ts.tcpConnsMutex.Unlock()
		belogs.Info("Add new connection, client:", clientAddr, " total connections:", ts.GetConnCount())

		// 处理连接
		go ts.handleConn(tcpConn)
	}
}

// handleConnection 处理单个连接
func (ts *TcpServer) handleConn(conn *net.TCPConn) {
	clientAddr := conn.RemoteAddr().String()
	ts.mu.Lock()
	ts.tcpConns[clientAddr] = conn
	ts.mu.Unlock()

	// 触发连接回调
	if ts.processFunc != nil {
		ts.processFunc.OnConnect(conn)
	}

	buf := make([]byte, 4096)
	defer func() {
		ts.mu.Lock()
		delete(ts.tcpConns, clientAddr)
		ts.mu.Unlock()
		conn.Close()
		if ts.processFunc != nil {
			ts.processFunc.OnClose(conn)
		}
	}()

	for {
		conn.SetReadDeadline(time.Now().Add(ts.readTimeout))
		n, err := conn.Read(buf)
		if err != nil {
			// 正常关闭不打印错误
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return
			}
			if !errors.Is(err, io.EOF) {
				belogs.Error("read from client fail:", err)
			}
			return
		}

		if n == 0 {
			continue
		}

		receiveData := make([]byte, n)
		copy(receiveData, buf[:n]) // 深拷贝避免数据覆盖

		// 业务处理回调
		if ts.processFunc != nil {
			if err := ts.processFunc.OnReceiveAndSend(conn, receiveData); err != nil {
				belogs.Error("OnReceiveAndSend fail:", err)
				return
			}
		}
	}
}

// Stop 停止服务器
func (ts *TcpServer) Stop() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.closed {
		belogs.Warn("Server already stopped")
		return
	}
	close(ts.stopChan)
	ts.closed = true
}

// ActiveSend 主动发送数据
func (ts *TcpServer) ActiveSend(conn *net.TCPConn, sendData []byte) error {
	ts.mu.Lock()
	if ts.closed {
		ts.mu.Unlock()
		return fmt.Errorf("server closed")
	}
	ts.mu.Unlock()

	if ts.processFunc == nil {
		return fmt.Errorf("processFunc is nil")
	}

	conn.SetWriteDeadline(time.Now().Add(ts.writeTimeout))
	return ts.processFunc.ActiveSend(conn, sendData)
}

// 新增方法：获取当前连接数
func (ts *TcpServer) GetConnCount() int {
	ts.tcpConnsMutex.RLock()
	defer ts.tcpConnsMutex.RUnlock()
	return len(ts.tcpConns)
}

// 新增方法：获取所有客户端连接
func (ts *TcpServer) GetAllConns() []*net.TCPConn {
	ts.tcpConnsMutex.RLock()
	defer ts.tcpConnsMutex.RUnlock()

	conns := make([]*net.TCPConn, 0, len(ts.tcpConns))
	for _, conn := range ts.tcpConns {
		conns = append(conns, conn)
	}
	return conns
}

// 新增方法：根据地址获取指定客户端连接
func (ts *TcpServer) GetConnByAddr(clientAddr string) (*net.TCPConn, bool) {
	ts.tcpConnsMutex.RLock()
	defer ts.tcpConnsMutex.RUnlock()

	conn, exists := ts.tcpConns[clientAddr]
	return conn, exists
}

// 新增方法：向所有客户端广播数据
func (ts *TcpServer) Broadcast(sendData []byte) error {
	ts.mu.Lock()
	if ts.closed {
		ts.mu.Unlock()
		return fmt.Errorf("server closed")
	}
	ts.mu.Unlock()

	if ts.processFunc == nil {
		return fmt.Errorf("processFunc is nil")
	}

	// 遍历所有连接发送数据
	ts.tcpConnsMutex.RLock()
	conns := make([]*net.TCPConn, 0, len(ts.tcpConns))
	for _, conn := range ts.tcpConns {
		conns = append(conns, conn)
	}
	ts.tcpConnsMutex.RUnlock()

	var errMsg string
	for _, conn := range conns {
		conn.SetWriteDeadline(time.Now().Add(ts.writeTimeout))
		if err := ts.processFunc.ActiveSend(conn, sendData); err != nil {
			errMsg += fmt.Sprintf("send to %s fail: %v; ", conn.RemoteAddr().String(), err)
		}
	}

	if errMsg != "" {
		return fmt.Errorf("broadcast fail: %s", errMsg)
	}
	return nil
}

// CloseConnByIP 根据客户端IP关闭连接（匹配所有该IP的连接）
// ip: 客户端IP地址（如 "192.168.1.100"）
// 返回：关闭的连接数、错误信息
func (ts *TcpServer) CloseConnByIP(ip string) (int, error) {
	ts.mu.Lock()
	if ts.closed {
		ts.mu.Unlock()
		return 0, fmt.Errorf("server already closed")
	}
	ts.mu.Unlock()

	ts.tcpConnsMutex.RLock()
	// 先收集需要关闭的连接（避免遍历过程中map修改）
	var connsToClose []*net.TCPConn
	var addrsToDelete []string
	for addr, conn := range ts.tcpConns {
		// 解析地址中的IP部分（addr格式："IP:Port"）
		remoteAddr := conn.RemoteAddr().String()
		host, _, err := net.SplitHostPort(remoteAddr)
		if err != nil {
			belogs.Warn("Split host port fail:", err, " addr:", remoteAddr)
			continue
		}
		if host == ip {
			connsToClose = append(connsToClose, conn)
			addrsToDelete = append(addrsToDelete, addr)
		}
	}
	ts.tcpConnsMutex.RUnlock()

	if len(connsToClose) == 0 {
		return 0, fmt.Errorf("no connection found for IP: %s", ip)
	}

	// 关闭连接并从map中删除
	closedCount := 0
	var errMsg string
	for i, conn := range connsToClose {
		belogs.Info("Closing connection for IP:", ip, " addr:", addrsToDelete[i])

		// 触发业务层的OnClose回调
		if ts.processFunc != nil {
			ts.processFunc.OnClose(conn)
		}

		// 关闭连接
		if err := conn.Close(); err != nil {
			errMsg += fmt.Sprintf("close conn %s fail: %v; ", addrsToDelete[i], err)
		} else {
			closedCount++
		}

		// 从map中删除
		ts.tcpConnsMutex.Lock()
		delete(ts.tcpConns, addrsToDelete[i])
		ts.tcpConnsMutex.Unlock()
	}

	if errMsg != "" {
		return closedCount, fmt.Errorf("partial close fail: %s", errMsg)
	}
	return closedCount, nil
}

// CloseConnByAddr 根据完整地址（IP:Port）关闭指定连接
// addr: 客户端完整地址（如 "192.168.1.100:8080"）
// 返回：是否找到并关闭连接
// 修复CloseConnByAddr方法（使用正确的读写锁）
func (ts *TcpServer) CloseConnByAddr(addr string) (bool, error) {
	ts.mu.Lock()
	if ts.closed {
		ts.mu.Unlock()
		return false, fmt.Errorf("server already closed")
	}
	ts.mu.Unlock()

	ts.tcpConnsMutex.RLock()
	defer ts.tcpConnsMutex.RUnlock()

	// 精确匹配或后缀匹配
	for clientAddr, conn := range ts.tcpConns {
		if clientAddr == addr || strings.HasSuffix(clientAddr, ":"+strings.Split(addr, ":")[1]) {
			// 先触发OnClose回调
			if ts.processFunc != nil {
				ts.processFunc.OnClose(conn)
			}
			// 关闭连接
			if err := conn.Close(); err != nil {
				return true, fmt.Errorf("关闭连接失败: %w", err)
			}
			// 删除映射（需要写锁）
			delete(ts.tcpConns, clientAddr)
			return true, nil
		}
	}
	return false, fmt.Errorf("未找到连接: %s", addr)
}

// CloseAllConns 关闭所有客户端连接（保留服务器监听）
// 返回：关闭的连接总数
func (ts *TcpServer) CloseAllConns() (int, error) {
	ts.mu.Lock()
	if ts.closed {
		ts.mu.Unlock()
		return 0, fmt.Errorf("server already closed")
	}
	ts.mu.Unlock()

	ts.tcpConnsMutex.RLock()
	// 复制所有连接信息，避免遍历中修改map
	conns := make(map[string]*net.TCPConn)
	for addr, conn := range ts.tcpConns {
		conns[addr] = conn
	}
	ts.tcpConnsMutex.RUnlock()

	if len(conns) == 0 {
		return 0, nil
	}

	closedCount := 0
	var errMsg string
	for addr, conn := range conns {
		belogs.Info("Closing connection:", addr)

		// 触发业务回调
		if ts.processFunc != nil {
			ts.processFunc.OnClose(conn)
		}

		// 关闭连接
		if err := conn.Close(); err != nil {
			errMsg += fmt.Sprintf("close conn %s fail: %v; ", addr, err)
		} else {
			closedCount++
		}

		// 从map中删除
		ts.tcpConnsMutex.Lock()
		delete(ts.tcpConns, addr)
		ts.tcpConnsMutex.Unlock()
	}

	if errMsg != "" {
		return closedCount, fmt.Errorf("partial close fail: %s", errMsg)
	}
	return closedCount, nil
}

// 辅助方法：获取所有客户端IP（去重）
func (ts *TcpServer) GetAllClientIPs() []string {
	ts.tcpConnsMutex.RLock()
	defer ts.tcpConnsMutex.RUnlock()

	ipSet := make(map[string]bool)
	for _, conn := range ts.tcpConns {
		host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err == nil {
			ipSet[host] = true
		}
	}

	ips := make([]string, 0, len(ipSet))
	for ip := range ipSet {
		ips = append(ips, ip)
	}
	return ips
}
