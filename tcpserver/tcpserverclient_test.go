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
	"sync"
	"testing"
	"time"

	"github.com/cpusoft/goutil/belogs"
)

// -------------------------- 核心工具：自动生成测试证书 --------------------------
// generateTestCerts 自动生成CA/服务端/客户端测试证书（临时目录）
func generateTestCerts(t *testing.T) (certDir string) {
	// 创建临时证书目录
	certDir = filepath.Join(t.TempDir(), "test-certs")
	if err := os.MkdirAll(certDir, 0755); err != nil {
		t.Fatalf("创建证书目录失败: %v", err)
	}

	// 1. 生成根CA私钥和证书
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
	// 保存CA证书
	if err := os.WriteFile(filepath.Join(certDir, "ca.crt"), pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}), 0644); err != nil {
		t.Fatalf("保存CA证书失败: %v", err)
	}

	// 2. 生成服务端证书
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
	// 保存服务端证书/私钥
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

	// 3. 生成客户端证书
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
	// 保存客户端证书/私钥
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
// TestServerHandler 服务端通用测试回调
type TestServerHandler struct {
	sync.Mutex
	recvData   map[string][]byte // 按客户端地址存储接收的数据
	connNum    int               // 活跃连接数
	closeCount int               // 关闭连接数
}

func NewTestServerHandler() *TestServerHandler {
	return &TestServerHandler{
		recvData: make(map[string][]byte),
	}
}

func (h *TestServerHandler) OnConnect(conn *net.TCPConn) error {
	h.Lock()
	h.connNum++
	h.Unlock()
	belogs.Info("[Server] 客户端连接:", conn.RemoteAddr())
	return nil
}

func (h *TestServerHandler) OnReceiveAndSend(conn *net.TCPConn, data []byte) error {
	addr := conn.RemoteAddr().String()
	h.Lock()
	h.recvData[addr] = append(h.recvData[addr], data...)
	h.Unlock()

	// 回声响应
	_, err := conn.Write(data)
	if err != nil {
		return fmt.Errorf("回声失败: %w", err)
	}
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
	conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
	_, err := conn.Write(data)
	return err
}

// TestClientHandler 客户端通用测试回调
type TestClientHandler struct {
	sync.Mutex
	recvData  []byte // 接收的所有数据
	sendCount int    // 发送次数
}

func NewTestClientHandler() *TestClientHandler {
	return &TestClientHandler{}
}

func (h *TestClientHandler) ActiveSend(conn *net.TCPConn, data string) error {
	h.Lock()
	h.sendCount++
	h.Unlock()
	_, err := conn.Write([]byte(data))
	return err
}

func (h *TestClientHandler) OnReceive(conn *net.TCPConn, data []byte) error {
	h.Lock()
	h.recvData = append(h.recvData, data...)
	h.Unlock()
	belogs.Debug("[Client] 接收数据:", hex.EncodeToString(data[:min(len(data), 10)])+"...")
	return nil
}

// -------------------------- 通用工具函数 --------------------------
// genRandData 生成指定大小的随机数据
func genRandData(size int) []byte {
	data := make([]byte, size)
	rand.Read(data)
	return data
}

// waitReady 等待服务端启动就绪
func waitReady(addr string, timeout time.Duration) bool {
	timeoutChan := time.After(timeout)
	for {
		select {
		case <-timeoutChan:
			return false
		default:
			conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
			if err == nil {
				conn.Close()
				return true
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// repeatString 重复字符串n次
func repeatString(s string, n int) string {
	var res string
	for i := 0; i < n; i++ {
		res += s
	}
	return res
}

// min 取最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// -------------------------- 测试用例 --------------------------
// TestTCP_NoTLS 无证书TCP连接测试（基础功能+连续数据发送）
func TestTCP_NoTLS(t *testing.T) {
	// 1. 启动服务端
	serverAddr := "127.0.0.1:9999"
	serverHandler := NewTestServerHandler()
	server := NewTcpServer(serverHandler, WithReadWriteTimeout(10*time.Second, 10*time.Second))

	// 异步启动服务端
	go func() {
		if err := server.Start(serverAddr); err != nil && err.Error() != "server already closed" {
			t.Fatal("服务端启动失败:", err)
		}
	}()

	// 等待服务端就绪
	if !waitReady(serverAddr, 3*time.Second) {
		t.Fatal("服务端未就绪")
	}

	// 2. 启动客户端
	clientHandler := NewTestClientHandler()
	client := NewTcpClient(clientHandler, WithClientReadWriteTimeout(10*time.Second, 10*time.Second))

	var clientErr error
	go func() {
		clientErr = client.Start(serverAddr)
	}()
	time.Sleep(500 * time.Millisecond)

	// 3. 测试连续数据发送（小/中/大）
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
			// 连续发送3次
			for i := 0; i < 3; i++ {
				if err := client.CallProcessFunc(tc.data); err != nil {
					t.Fatalf("发送失败(%d): %v", i, err)
				}
				time.Sleep(100 * time.Millisecond)
			}

			// 验证客户端接收（回声）
			clientHandler.Lock()
			recv := clientHandler.recvData
			clientHandler.Unlock()
			expected := repeatString(tc.data, 3)
			if string(recv) != expected {
				t.Errorf("接收数据不匹配: 期望长度%d, 实际长度%d", len(expected), len(recv))
			}

			// 验证服务端接收
			clientAddr := client.conn.RemoteAddr().String()
			serverHandler.Lock()
			serverRecv := serverHandler.recvData[clientAddr]
			serverHandler.Unlock()
			if string(serverRecv) != expected {
				t.Errorf("服务端接收数据不匹配: 期望前10字符[%s], 实际[%s]", expected[:10], string(serverRecv)[:10])
			}

			// 清空缓存
			clientHandler.Lock()
			clientHandler.recvData = nil
			clientHandler.Unlock()
			serverHandler.Lock()
			serverHandler.recvData[clientAddr] = nil
			serverHandler.Unlock()
		})
	}

	// 4. 测试主动关闭连接
	clientAddr := client.conn.RemoteAddr().String()
	ok, err := server.CloseConnByAddr(clientAddr)
	if err != nil {
		t.Fatal("主动关闭连接失败:", err)
	}
	if !ok {
		t.Error("未找到指定客户端连接")
	}
	time.Sleep(500 * time.Millisecond)

	// 验证连接数归零
	if server.GetConnCount() != 0 {
		t.Errorf("连接数未归零: 期望0, 实际%d", server.GetConnCount())
	}

	// 5. 清理资源
	client.CallStop()
	server.Stop()
	time.Sleep(200 * time.Millisecond)

	if clientErr != nil && clientErr.Error() != "client already closed" {
		t.Error("客户端异常:", clientErr)
	}
}

// TestTLS_OneWay TLS单向认证测试（服务端证书）
func TestTLS_OneWay(t *testing.T) {
	// 1. 自动生成测试证书
	certDir := generateTestCerts(t)
	serverAddr := "127.0.0.1:9998"

	// 2. 启动TLS服务端（单向认证）
	serverHandler := NewTestServerHandler()
	tlsServerCfg := &ServerTLSConfig{
		ServerCertFile: filepath.Join(certDir, "server.crt"),
		ServerKeyFile:  filepath.Join(certDir, "server.key"),
		ClientAuth:     tls.NoClientCert, // 不验证客户端证书
	}
	server := NewTcpServer(serverHandler, WithServerTLS(tlsServerCfg))

	go func() {
		if err := server.Start(serverAddr); err != nil && err.Error() != "server already closed" {
			t.Fatal("TLS服务端启动失败:", err)
		}
	}()

	if !waitReady(serverAddr, 3*time.Second) {
		t.Fatal("TLS服务端未就绪")
	}

	// 3. 启动TLS客户端（验证服务端证书）
	clientHandler := NewTestClientHandler()
	tlsClientCfg := &ClientTLSConfig{
		RootCAFile:         filepath.Join(certDir, "ca.crt"),
		ServerName:         "localhost", // 匹配服务端证书CN
		InsecureSkipVerify: false,       // 验证服务端证书
	}
	client := NewTcpClient(clientHandler, WithClientTLS(tlsClientCfg))

	var clientErr error
	go func() {
		clientErr = client.Start(serverAddr)
	}()
	time.Sleep(500 * time.Millisecond)

	if clientErr != nil {
		t.Fatal("TLS客户端连接失败:", clientErr)
	}

	// 4. 发送测试数据（2KB）
	testData := genRandData(2048)
	err := client.CallProcessFunc(string(testData))
	if err != nil {
		t.Fatal("TLS客户端发送失败:", err)
	}
	time.Sleep(200 * time.Millisecond)

	// 5. 验证数据完整性
	clientHandler.Lock()
	if string(clientHandler.recvData) != string(testData) {
		t.Errorf("TLS客户端接收数据不匹配: 期望长度%d, 实际%d", len(testData), len(clientHandler.recvData))
	}
	clientHandler.Unlock()

	// 6. 清理资源
	client.CallStop()
	server.Stop()
	time.Sleep(200 * time.Millisecond)
}

// TestTLS_TwoWay TLS双向认证测试
func TestTLS_TwoWay(t *testing.T) {
	// 1. 自动生成测试证书
	certDir := generateTestCerts(t)
	serverAddr := "127.0.0.1:9997"

	// 2. 启动双向认证服务端
	serverHandler := NewTestServerHandler()
	tlsServerCfg := &ServerTLSConfig{
		ServerCertFile: filepath.Join(certDir, "server.crt"),
		ServerKeyFile:  filepath.Join(certDir, "server.key"),
		RootCAFile:     filepath.Join(certDir, "ca.crt"),
		ClientAuth:     tls.RequireAndVerifyClientCert, // 强制验证客户端证书
	}
	server := NewTcpServer(serverHandler, WithServerTLS(tlsServerCfg))

	go func() {
		if err := server.Start(serverAddr); err != nil && err.Error() != "server already closed" {
			t.Fatal("双向认证服务端启动失败:", err)
		}
	}()

	if !waitReady(serverAddr, 3*time.Second) {
		t.Fatal("双向认证服务端未就绪")
	}

	// 3. 启动双向认证客户端
	clientHandler := NewTestClientHandler()
	tlsClientCfg := &ClientTLSConfig{
		ClientCertFile:     filepath.Join(certDir, "client.crt"),
		ClientKeyFile:      filepath.Join(certDir, "client.key"),
		RootCAFile:         filepath.Join(certDir, "ca.crt"),
		ServerName:         "localhost",
		InsecureSkipVerify: false,
	}
	client := NewTcpClient(clientHandler, WithClientTLS(tlsClientCfg))

	var clientErr error
	go func() {
		clientErr = client.Start(serverAddr)
	}()
	time.Sleep(500 * time.Millisecond)

	if clientErr != nil {
		t.Fatal("双向认证客户端连接失败:", clientErr)
	}

	// 4. 连续发送5次测试数据
	for i := 0; i < 5; i++ {
		testData := fmt.Sprintf("双向认证测试-%d", i)
		err := client.CallProcessFunc(testData)
		if err != nil {
			t.Errorf("双向认证发送失败(%d): %v", i, err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 5. 验证接收数据
	clientHandler.Lock()
	expected := "双向认证测试-0双向认证测试-1双向认证测试-2双向认证测试-3双向认证测试-4"
	if string(clientHandler.recvData) != expected {
		t.Errorf("双向认证接收数据不匹配: 期望[%s], 实际[%s]", expected, string(clientHandler.recvData))
	}
	clientHandler.Unlock()

	// 6. 清理资源
	client.CallStop()
	server.Stop()
	time.Sleep(200 * time.Millisecond)
}

// TestEdgeCases 临界值测试（空数据/超大数据/超时/并发）
func TestEdgeCases(t *testing.T) {
	serverAddr := "127.0.0.1:9996"
	serverHandler := NewTestServerHandler()
	// 超短超时配置（1秒）
	server := NewTcpServer(serverHandler, WithReadWriteTimeout(1*time.Second, 1*time.Second))

	go func() {
		_ = server.Start(serverAddr)
	}()
	time.Sleep(300 * time.Millisecond)

	// 测试1：空数据发送
	t.Run("空数据发送", func(t *testing.T) {
		client := NewTcpClient(NewTestClientHandler())
		go func() {
			_ = client.Start(serverAddr)
		}()
		time.Sleep(200 * time.Millisecond)

		err := client.CallProcessFunc("")
		if err != nil {
			t.Error("空数据发送失败:", err)
		}
		client.CallStop()
	})

	// 测试2：超大数据（16KB）
	t.Run("超大数据(16KB)", func(t *testing.T) {
		client := NewTcpClient(NewTestClientHandler())
		go func() {
			_ = client.Start(serverAddr)
		}()
		time.Sleep(200 * time.Millisecond)

		largeData := genRandData(16 * 1024)
		err := client.CallProcessFunc(string(largeData))
		if err != nil {
			t.Fatal("超大数据发送失败:", err)
		}
		time.Sleep(200 * time.Millisecond)

		// 验证接收长度
		clientHandler := client.processFunc.(*TestClientHandler)
		clientHandler.Lock()
		if len(clientHandler.recvData) != len(largeData) {
			t.Errorf("超大数据接收长度不匹配: 期望%d, 实际%d", len(largeData), len(clientHandler.recvData))
		}
		clientHandler.Unlock()
		client.CallStop()
	})

	// 测试3：超时断开
	t.Run("读超时断开", func(t *testing.T) {
		client := NewTcpClient(NewTestClientHandler())
		go func() {
			_ = client.Start(serverAddr)
		}()
		time.Sleep(200 * time.Millisecond)

		// 等待超时（1.5秒）
		time.Sleep(1500 * time.Millisecond)
		if server.GetConnCount() != 0 {
			t.Errorf("超时后未自动断开连接: 连接数%d", server.GetConnCount())
		}
		client.CallStop()
	})

	// 测试4：并发连接（10个客户端）
	t.Run("并发连接(10个)", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				client := NewTcpClient(NewTestClientHandler())
				err := client.Start(serverAddr)
				if err == nil {
					_ = client.CallProcessFunc("并发测试")
					time.Sleep(100 * time.Millisecond)
					client.CallStop()
				}
			}()
		}
		wg.Wait()

		// 验证关闭计数
		if serverHandler.closeCount < 10 {
			t.Errorf("并发连接关闭计数不匹配: 期望≥10, 实际%d", serverHandler.closeCount)
		}
	})

	// 清理资源
	server.Stop()
	time.Sleep(200 * time.Millisecond)
}

// -------------------------- 性能测试 --------------------------
// BenchmarkTCP_Throughput TCP吞吐性能测试（1KB数据并发发送）
func BenchmarkTCP_Throughput(b *testing.B) {
	serverAddr := "127.0.0.1:9995"
	serverHandler := NewTestServerHandler()
	server := NewTcpServer(serverHandler)

	go func() {
		_ = server.Start(serverAddr)
	}()
	time.Sleep(300 * time.Millisecond)

	// 启动客户端
	clientHandler := NewTestClientHandler()
	client := NewTcpClient(clientHandler)
	go func() {
		_ = client.Start(serverAddr)
	}()
	time.Sleep(300 * time.Millisecond)

	// 测试数据（1KB）
	testData := genRandData(1024)
	testStr := string(testData)

	// 重置基准计时器
	b.ResetTimer()

	// 并发压测
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = client.CallProcessFunc(testStr)
		}
	})

	// 停止计时器
	b.StopTimer()

	// 输出内存分配
	b.ReportAllocs()

	// 清理
	client.CallStop()
	server.Stop()
}

// BenchmarkTCP_ConcurrentConn 并发连接性能测试
func BenchmarkTCP_ConcurrentConn(b *testing.B) {
	serverAddr := "127.0.0.1:9994"
	serverHandler := NewTestServerHandler()
	server := NewTcpServer(serverHandler)

	go func() {
		_ = server.Start(serverAddr)
	}()
	time.Sleep(300 * time.Millisecond)

	b.ResetTimer()

	// 并发创建/关闭连接
	var wg sync.WaitGroup
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				client := NewTcpClient(NewTestClientHandler())
				err := client.Start(serverAddr)
				if err == nil {
					_ = client.CallProcessFunc("ping")
					client.CallStop()
				}
			}()
		}
	})
	wg.Wait()

	b.StopTimer()
	b.ReportAllocs()

	// 验证连接数归零
	if server.GetConnCount() != 0 {
		b.Error("并发压测后连接数未归零")
	}

	server.Stop()
}
