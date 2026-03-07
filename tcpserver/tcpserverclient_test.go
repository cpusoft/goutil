package tcpserver

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cpusoft/goutil/belogs"
)

// -------------------------- 核心工具：自动生成测试证书 --------------------------
func generateTestCerts(t *testing.T) (certDir string) {
	certDir = filepath.Join(t.TempDir(), "test-certs")
	if err := os.MkdirAll(certDir, 0755); err != nil {
		t.Fatalf("创建证书目录失败: %v", err)
	}

	// 生成CA私钥和证书
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成CA私钥失败: %v", err)
	}
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test CA",
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		t.Fatalf("生成CA证书失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(certDir, "ca.crt"), pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}), 0644); err != nil {
		t.Fatalf("保存CA证书失败: %v", err)
	}

	// 生成服务端证书
	serverPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成服务端私钥失败: %v", err)
	}
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName:   "localhost",
			Organization: []string{"Test Org"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:    []string{"localhost"},
	}
	serverBytes, err := x509.CreateCertificate(rand.Reader, serverTemplate, caTemplate, &serverPrivKey.PublicKey, caPrivKey)
	if err != nil {
		t.Fatalf("生成服务端证书失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(certDir, "server.crt"), pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverBytes,
	}), 0644); err != nil {
		t.Fatalf("保存服务端证书失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(certDir, "server.key"), pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	}), 0600); err != nil {
		t.Fatalf("保存服务端私钥失败: %v", err)
	}

	// 生成客户端证书
	clientPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成客户端私钥失败: %v", err)
	}
	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			CommonName:   "test-client",
			Organization: []string{"Test Org"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientBytes, err := x509.CreateCertificate(rand.Reader, clientTemplate, caTemplate, &clientPrivKey.PublicKey, caPrivKey)
	if err != nil {
		t.Fatalf("生成客户端证书失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(certDir, "client.crt"), pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: clientBytes,
	}), 0644); err != nil {
		t.Fatalf("保存客户端证书失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(certDir, "client.key"), pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(clientPrivKey),
	}), 0600); err != nil {
		t.Fatalf("保存客户端私钥失败: %v", err)
	}

	return certDir
}

// -------------------------- 通用回调实现 --------------------------
// TestServerHandler 实现TcpServerProcessFunc接口
type TestServerHandler struct {
	sync.Mutex
	recvData   map[string][]byte
	connNum    int
	closeCount int
}

func NewTestServerHandler() *TestServerHandler {
	return &TestServerHandler{
		recvData: make(map[string][]byte),
	}
}

func (h *TestServerHandler) OnConnect(conn *net.TCPConn) error {
	h.Lock()
	h.connNum++
	addr := conn.RemoteAddr().String()
	h.recvData[addr] = make([]byte, 0) // 初始化空切片
	h.Unlock()
	belogs.Info("[Server] 客户端连接:", addr)
	return nil
}

func (h *TestServerHandler) OnReceiveAndSend(conn *net.TCPConn, data []byte) error {
	addr := conn.RemoteAddr().String()

	// 深拷贝数据，避免缓冲区覆盖
	recvData := make([]byte, len(data))
	copy(recvData, data)

	// 记录接收数据
	h.Lock()
	h.recvData[addr] = append(h.recvData[addr], recvData...)
	h.Unlock()

	// 回声给客户端（禁用Nagle算法，立即发送）
	conn.SetNoDelay(true)
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Write(recvData)
	if err != nil {
		return fmt.Errorf("回声失败: %w", err)
	}
	if n != len(recvData) {
		return fmt.Errorf("回声数据不完整: 发送%d字节，总长度%d", n, len(recvData))
	}

	belogs.Debug("[Server] 接收并回声数据，客户端:", addr, " 长度:", len(data))
	return nil
}

func (h *TestServerHandler) OnClose(conn *net.TCPConn) {
	h.Lock()
	h.connNum--
	h.closeCount++
	delete(h.recvData, conn.RemoteAddr().String())
	h.Unlock()
	belogs.Info("[Server] 客户端断开:", conn.RemoteAddr())
}

func (h *TestServerHandler) ActiveSend(conn *net.TCPConn, data []byte) error {
	conn.SetNoDelay(true)
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write(data)
	return err
}

// TestClientHandler 实现TcpClientProcessFunc接口
type TestClientHandler struct {
	sync.Mutex
	recvData  []byte
	sendCount int
}

func NewTestClientHandler() *TestClientHandler {
	return &TestClientHandler{}
}

func (h *TestClientHandler) ActiveSend(conn *net.TCPConn, data string) error {
	h.Lock()
	h.sendCount++
	h.Unlock()

	// 禁用Nagle算法，确保数据立即发送
	conn.SetNoDelay(true)
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	// 分块发送大数据，避免单次发送失败
	buf := []byte(data)
	total := len(buf)
	sent := 0
	for sent < total {
		n, err := conn.Write(buf[sent:])
		if err != nil {
			return fmt.Errorf("发送失败: %w", err)
		}
		sent += n
	}

	belogs.Debug("[Client] 发送数据成功，长度:", total)
	return nil
}

func (h *TestClientHandler) OnReceive(conn *net.TCPConn, data []byte) error {
	h.Lock()
	// 深拷贝避免数据覆盖
	recvData := make([]byte, len(data))
	copy(recvData, data)
	h.recvData = append(h.recvData, recvData...)
	h.Unlock()

	if len(data) > 0 {
		belogs.Debug("[Client] 接收数据，长度:", len(data), " 前10字节:", hex.EncodeToString(data[:min(len(data), 10)]))
	} else {
		belogs.Debug("[Client] 接收空数据")
	}
	return nil
}

// -------------------------- 通用工具函数 --------------------------
func genRandData(size int) []byte {
	data := make([]byte, size)
	_, err := rand.Read(data)
	if err != nil {
		panic(fmt.Sprintf("生成随机数据失败: %v", err))
	}
	return data
}

// 增强版等待就绪函数
func waitReady(addr string, timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}

func repeatString(s string, n int) string {
	var res string
	for i := 0; i < n; i++ {
		res += s
	}
	return res
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 安全截取字符串前n个字符（处理空字符串）
func safeSubstr(s string, n int) string {
	if len(s) == 0 {
		return "<空数据>"
	}
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// -------------------------- 核心测试用例 --------------------------
func TestTCP_NoTLS(t *testing.T) {
	// 禁用日志干扰测试

	serverAddr := "127.0.0.1:9999"
	serverHandler := NewTestServerHandler()
	server := NewTcpServer(serverHandler, WithReadWriteTimeout(10*time.Second, 10*time.Second))

	// 异步启动服务端
	go func() {
		if err := server.Start(serverAddr); err != nil && !strings.Contains(err.Error(), "server already closed") {
			t.Fatal("服务端启动失败:", err)
		}
	}()

	// 等待服务端就绪
	if !waitReady(serverAddr, 10*time.Second) {
		t.Fatal("服务端启动超时")
	}

	// 启动客户端
	clientHandler := NewTestClientHandler()
	client := NewTcpClient(clientHandler, WithClientReadWriteTimeout(10*time.Second, 10*time.Second))

	// 连接服务端
	clientErr := client.Start(serverAddr)
	if clientErr != nil {
		t.Fatal("客户端连接失败:", clientErr)
	}
	// 确保连接完全建立
	time.Sleep(1 * time.Second)

	// 获取真实客户端地址
	if client.conn == nil {
		t.Fatal("客户端连接未建立")
	}
	clientRealAddr := client.conn.RemoteAddr().String()
	t.Log("客户端真实地址:", clientRealAddr)

	// 验证服务端已注册该连接
	conn, exists := server.GetConnByAddr(clientRealAddr)
	if !exists || conn == nil {
		t.Fatalf("服务端未注册客户端连接: %s", clientRealAddr)
	}

	// 测试连续数据发送
	testCases := []struct {
		name string
		data string
	}{
		{"小数据(16B)", "hello tcp server"},
		{"中数据(1KB)", string(genRandData(1024))},
		{"大数据(8KB)", string(genRandData(8 * 1024))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 清空接收缓存
			clientHandler.Lock()
			clientHandler.recvData = nil
			clientHandler.Unlock()
			serverHandler.Lock()
			serverHandler.recvData[clientRealAddr] = nil
			serverHandler.Unlock()

			// 连续发送3次
			sendCount := 3
			for i := 0; i < sendCount; i++ {
				err := client.CallProcessFunc(tc.data)
				if err != nil {
					t.Fatalf("第%d次发送失败: %v", i+1, err)
				}
				// 等待数据接收和回声
				time.Sleep(1 * time.Second)
			}

			// 验证客户端接收（回声数据）
			clientHandler.Lock()
			clientRecv := clientHandler.recvData
			clientHandler.Unlock()
			expected := repeatString(tc.data, sendCount)

			if string(clientRecv) != expected {
				t.Errorf(
					"客户端接收数据不匹配:\n期望长度: %d, 实际长度: %d\n期望前10字符: [%s]\n实际前10字符: [%s]",
					len(expected), len(clientRecv),
					safeSubstr(expected, 10), safeSubstr(string(clientRecv), 10),
				)
			}

			// 验证服务端接收
			serverHandler.Lock()
			serverRecv := serverHandler.recvData[clientRealAddr]
			serverHandler.Unlock()

			if string(serverRecv) != expected {
				t.Errorf(
					"服务端接收数据不匹配:\n期望长度: %d, 实际长度: %d\n期望前10字符: [%s]\n实际前10字符: [%s]",
					len(expected), len(serverRecv),
					safeSubstr(expected, 10), safeSubstr(string(serverRecv), 10),
				)
			}
		})
	}

	// 测试主动关闭连接
	t.Log("尝试关闭客户端连接:", clientRealAddr)
	ok, err := server.CloseConnByAddr(clientRealAddr)
	if err != nil {
		t.Fatal("关闭连接失败:", err)
	}
	if !ok {
		t.Error("未找到要关闭的客户端连接")
	}
	time.Sleep(1 * time.Second)

	// 验证连接数归零
	if count := server.GetConnCount(); count != 0 {
		t.Errorf("连接数未归零: 期望0, 实际%d", count)
	}

	// 清理资源
	client.CallStop()
	server.Stop()
	time.Sleep(500 * time.Millisecond)
}

// -------------------------- 其他测试用例（简化版） --------------------------
func TestTLS_OneWay(t *testing.T) {

	certDir := generateTestCerts(t)
	serverAddr := "127.0.0.1:9998"

	serverHandler := NewTestServerHandler()
	tlsServerCfg := &ServerTLSConfig{
		ServerCertFile: filepath.Join(certDir, "server.crt"),
		ServerKeyFile:  filepath.Join(certDir, "server.key"),
		ClientAuth:     tls.NoClientCert,
	}
	server := NewTcpServer(serverHandler, WithServerTLS(tlsServerCfg))

	go func() {
		if err := server.Start(serverAddr); err != nil && !strings.Contains(err.Error(), "server already closed") {
			t.Fatal("TLS服务端启动失败:", err)
		}
	}()

	if !waitReady(serverAddr, 10*time.Second) {
		t.Fatal("TLS服务端启动超时")
	}

	clientHandler := NewTestClientHandler()
	tlsClientCfg := &ClientTLSConfig{
		RootCAFile:         filepath.Join(certDir, "ca.crt"),
		ServerName:         "localhost",
		InsecureSkipVerify: false,
	}
	client := NewTcpClient(clientHandler, WithClientTLS(tlsClientCfg), WithClientReadWriteTimeout(10*time.Second, 10*time.Second))

	clientErr := client.Start(serverAddr)
	if clientErr != nil {
		t.Fatal("TLS客户端连接失败:", clientErr)
	}
	time.Sleep(1 * time.Second)

	// 发送测试数据
	testData := string(genRandData(2048))
	err := client.CallProcessFunc(testData)
	if err != nil {
		t.Fatal("TLS客户端发送失败:", err)
	}
	time.Sleep(1 * time.Second)

	// 验证数据
	clientHandler.Lock()
	if string(clientHandler.recvData) != testData {
		t.Errorf(
			"TLS客户端接收数据不匹配: 期望长度%d, 实际%d",
			len(testData), len(clientHandler.recvData),
		)
	}
	clientHandler.Unlock()

	// 清理
	client.CallStop()
	server.Stop()
	time.Sleep(500 * time.Millisecond)
}

func TestEdgeCases(t *testing.T) {

	serverAddr := "127.0.0.1:9996"
	serverHandler := NewTestServerHandler()
	server := NewTcpServer(serverHandler, WithReadWriteTimeout(1*time.Second, 1*time.Second))

	go func() {
		_ = server.Start(serverAddr)
	}()
	time.Sleep(1 * time.Second)

	// 空数据发送
	t.Run("空数据发送", func(t *testing.T) {
		client := NewTcpClient(NewTestClientHandler())
		err := client.Start(serverAddr)
		if err != nil {
			t.Fatal("客户端连接失败:", err)
		}
		time.Sleep(500 * time.Millisecond)

		err = client.CallProcessFunc("")
		if err != nil {
			t.Error("空数据发送失败:", err)
		}
		time.Sleep(500 * time.Millisecond)
		client.CallStop()
	})

	// 超大数据
	t.Run("超大数据(16KB)", func(t *testing.T) {
		client := NewTcpClient(NewTestClientHandler())
		err := client.Start(serverAddr)
		if err != nil {
			t.Fatal("客户端连接失败:", err)
		}
		time.Sleep(500 * time.Millisecond)

		largeData := genRandData(16 * 1024)
		err = client.CallProcessFunc(string(largeData))
		if err != nil {
			t.Fatal("超大数据发送失败:", err)
		}
		time.Sleep(2 * time.Second)

		clientHandler := client.processFunc.(*TestClientHandler)
		clientHandler.Lock()
		if len(clientHandler.recvData) != len(largeData) {
			t.Errorf(
				"超大数据接收长度不匹配: 期望%d, 实际%d",
				len(largeData), len(clientHandler.recvData),
			)
		}
		clientHandler.Unlock()
		client.CallStop()
	})

	server.Stop()
	time.Sleep(500 * time.Millisecond)
}

// -------------------------- 性能测试 --------------------------
func BenchmarkTCP_Throughput(b *testing.B) {

	serverAddr := "127.0.0.1:9995"
	serverHandler := NewTestServerHandler()
	server := NewTcpServer(serverHandler)

	go func() {
		_ = server.Start(serverAddr)
	}()
	time.Sleep(1 * time.Second)

	clientHandler := NewTestClientHandler()
	client := NewTcpClient(clientHandler)
	err := client.Start(serverAddr)
	if err != nil {
		b.Fatal("客户端连接失败:", err)
	}
	time.Sleep(1 * time.Second)

	testData := string(genRandData(1024))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.CallProcessFunc(testData)
	}
	b.StopTimer()

	client.CallStop()
	server.Stop()
}
