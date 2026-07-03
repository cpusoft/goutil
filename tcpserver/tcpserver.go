package tcpserver

import (
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

// TcpServerProcessFunc 服务器业务回调接口
type TcpServerProcessFunc interface {
	OnConnect(conn net.Conn) (err error)
	PreCheckConn(conn net.Conn) (err error)
	OnReceiveAndSend(conn net.Conn, receiveData []byte) (err error)
	OnClose(conn net.Conn)
	ActiveSend(conn net.Conn, sendData []byte) (err error)
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
	setReadTimeout bool
	readTimeout    time.Duration
	writeTimeout   time.Duration

	// 并发安全
	mu     sync.Mutex
	closed bool

	// 新增：客户端连接管理
	conns      map[string]net.Conn // 改为 net.Conn, key: 客户端地址(RemoteAddr().String())
	connsMutex sync.RWMutex        // 读写锁，支持高并发读写
}

// NewTcpServer 创建服务器实例
func NewTcpServer(processFunc TcpServerProcessFunc, opts ...ServerOption) *TcpServer {
	ts := &TcpServer{
		stopChan:     make(chan struct{}),
		processFunc:  processFunc,
		readTimeout:  30 * time.Second,
		writeTimeout: 30 * time.Second,
		closed:       false,
		conns:        make(map[string]net.Conn), // 初始化连接映射表
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
func WithReadWriteTimeout(setReadTimeout bool, readTimeout, writeTimeout time.Duration) ServerOption {
	return func(ts *TcpServer) {
		ts.setReadTimeout = setReadTimeout
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
			belogs.Info("TcpServer.buildTLSConfig(): Client cert verified, CN:", cert.Subject.CommonName)
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

	belogs.Info("TcpServer.Start(): Server started, addr:", addr,
		" isTLS:", ts.isTLS)

	// 接收连接
	go ts.acceptConnections()

	// 等待停止信号
	<-ts.stopChan

	// 先关闭监听器，阻止新连接进入
	if ts.listener != nil {
		_ = ts.listener.Close()
	}

	// 新增：停止时关闭所有客户端连接
	ts.CloseAllConns()

	// 关闭资源，注意顺序不能在CloseAllConns之前
	ts.mu.Lock()
	ts.closed = true
	ts.mu.Unlock()

	belogs.Info("TcpServer.Start(): Server stopped, addr:", addr)

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

		// 不再解包，直接传递 net.Conn
		if err := ts.preCheckConn(conn); err != nil {
			belogs.Error("Connection preCheckConn failed:", err)
			_ = conn.Close()
			continue
		}

		go ts.handleConn(conn)
	}
}

// handleConnection 处理单个连接
func (ts *TcpServer) preCheckConn(conn net.Conn) error {
	if ts.processFunc != nil {
		return ts.processFunc.PreCheckConn(conn)
	}
	return nil
}

// handleConnection 处理单个连接
func (ts *TcpServer) handleConn(conn net.Conn) {
	clientAddr := conn.RemoteAddr().String()
	ts.connsMutex.Lock()
	ts.conns[clientAddr] = conn
	ts.connsMutex.Unlock()
	belogs.Info("TcpServer.handleConn(): Add new connection, client:",
		clientAddr, " total connections:", ts.GetConnCount())

	if ts.processFunc != nil {
		ts.processFunc.OnConnect(conn)
	}

	buf := make([]byte, 4096)
	defer func() {
		conn.Close()
		ts.connsMutex.Lock()
		delete(ts.conns, clientAddr)
		ts.connsMutex.Unlock()
		if ts.processFunc != nil {
			ts.processFunc.OnClose(conn)
		}
	}()

	for {
		if ts.setReadTimeout {
			conn.SetReadDeadline(time.Now().Add(ts.readTimeout))
		}
		n, err := conn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				belogs.Info("TcpServer.handleConn(): read timeout so close", err)
				return
			}
			if !errors.Is(err, io.EOF) {
				belogs.Error("TcpServer.handleConn(): read from client no io.EOF fail", err)
				return
			}
			belogs.Error("TcpServer.handleConn(): read from client io.EOF fail", err)
			return
		}

		if n <= 0 {
			belogs.Debug("TcpServer.handleConn(): read 0 bytes, n<=0, will return", n)
			return
		}

		receiveData := make([]byte, n)
		copy(receiveData, buf[:n])

		if ts.processFunc != nil {
			if err := ts.processFunc.OnReceiveAndSend(conn, receiveData); err != nil {
				belogs.Error("TcpServer.handleConn(): OnReceiveAndSend fail:", err)
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
func (ts *TcpServer) ActiveSend(conn net.Conn, sendData []byte) error {
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
func (ts *TcpServer) GetConnCount() int {
	ts.connsMutex.RLock()
	defer ts.connsMutex.RUnlock()
	return len(ts.conns)
}

func (ts *TcpServer) GetAllConns() []net.Conn {
	ts.connsMutex.RLock()
	defer ts.connsMutex.RUnlock()
	conns := make([]net.Conn, 0, len(ts.conns))
	for _, conn := range ts.conns {
		conns = append(conns, conn)
	}
	return conns
}

// 获取所有连接的客户端IP地址（去重）
func (ts *TcpServer) GetDistinctConnIps() []string {
	ipMap := make(map[string]string)
	ts.connsMutex.RLock()
	for _, conn := range ts.conns {
		remoteAddr := conn.RemoteAddr().String()
		host, _, err := net.SplitHostPort(remoteAddr)
		if err != nil {
			belogs.Warn("Split host port fail:", err, " addr:", remoteAddr)
			continue
		}
		if _, ok := ipMap[host]; !ok {
			ipMap[host] = host
		}
	}
	ts.connsMutex.RUnlock()
	ips := make([]string, 0, len(ipMap))
	for k := range ipMap {
		ips = append(ips, k)
	}
	return ips
}

/* use GetDistinctConnIps
func (ts *TcpServer) GetAllClientIPs() []string {
	ts.connsMutex.RLock()
	defer ts.connsMutex.RUnlock()

	ipSet := make(map[string]bool)
	for _, conn := range ts.conns {
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
*/

// 新增方法：根据地址获取指定客户端连接
func (ts *TcpServer) GetConnByAddr(clientAddr string) (conn net.Conn, exists bool) {
	ts.connsMutex.RLock()
	defer ts.connsMutex.RUnlock()

	conn, exists = ts.conns[clientAddr]
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
	ts.connsMutex.RLock()
	conns := make([]net.Conn, 0, len(ts.conns))
	for _, conn := range ts.conns {
		conns = append(conns, conn)
	}
	ts.connsMutex.RUnlock()

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

/*
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

	ts.connsMutex.RLock()
	// 先收集需要关闭的连接（避免遍历过程中map修改）
	var connsToClose []net.Conn
	var addrsToDelete []string
	for addr, conn := range ts.conns {
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
	ts.connsMutex.RUnlock()

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
		ts.connsMutex.Lock()
		delete(ts.conns, addrsToDelete[i])
		ts.connsMutex.Unlock()
	}

	if errMsg != "" {
		return closedCount, fmt.Errorf("partial close fail: %s", errMsg)
	}
	return closedCount, nil
}
*/

// addr: 客户端完整地址（如 "192.168.1.100"），关闭192.168.1.100的所有连接
// 返回： 关闭了几个，是否有错误
func (ts *TcpServer) CloseConnByIP(ip string) (int, error) {
	ts.mu.Lock()
	if ts.closed {
		ts.mu.Unlock()
		return 0, fmt.Errorf("server already closed")
	}
	ts.mu.Unlock()

	shouldCloseAddres := make([]string, 0)
	ts.connsMutex.Lock()
	defer ts.connsMutex.Unlock()
	for clientAddr := range ts.conns {
		// clientAddr: 客户端完整地址（如 "192.168.1.100:8080"），关闭192.168.1.100的所有连接（后缀匹配）
		clientHost, _, _ := net.SplitHostPort(clientAddr)
		if clientHost == ip {
			shouldCloseAddres = append(shouldCloseAddres, clientAddr)
		}
	}
	if len(shouldCloseAddres) == 0 {
		return 0, nil
	}
	closedCount := 0
	var errMsg string
	for _, addr := range shouldCloseAddres {
		conn := ts.conns[addr]
		// 先触发OnClose回调
		if ts.processFunc != nil {
			ts.processFunc.OnClose(conn)
		}
		// 关闭连接
		if err := conn.Close(); err != nil {
			errMsg += fmt.Sprintf("close %s fail: %v; ", addr, err)
		} else {
			closedCount++
		}
		// 删除映射（需要写锁）
		delete(ts.conns, addr)
		//	return true, nil
	}
	if errMsg != "" {
		return closedCount, fmt.Errorf("partial close fail: %s", errMsg)
	}
	return closedCount, nil
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

	ts.connsMutex.RLock()
	// 复制所有连接信息，避免遍历中修改map
	conns := make(map[string]net.Conn)
	for addr, conn := range ts.conns {
		conns[addr] = conn
	}
	ts.connsMutex.RUnlock()

	if len(conns) == 0 {
		return 0, nil
	}

	closedCount := 0
	var errMsg string
	for addr, conn := range conns {
		belogs.Info("TcpServer.CloseAllConns(): Closing connection:", addr)

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
		ts.connsMutex.Lock()
		delete(ts.conns, addr)
		ts.connsMutex.Unlock()
	}

	if errMsg != "" {
		return closedCount, fmt.Errorf("partial close fail: %s", errMsg)
	}
	return closedCount, nil
}
