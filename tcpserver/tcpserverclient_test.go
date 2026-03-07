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
	h.Unlock()
	belogs.Info("[Server] 客户端连接:", conn.RemoteAddr())
	return nil
}

func (h *TestServerHandler) OnReceiveAndSend(conn *net.TCPConn, data []byte) error {
	addr := conn.RemoteAddr().String()
	h.Lock()
	h.recvData[addr] = append(h.recvData[addr], data...)
	h.Unlock()

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
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err := conn.Write([]byte(data))
	return err
}

func (h *TestClientHandler) OnReceive(conn *net.TCPConn, data []byte) error {
	h.Lock()
	h.recvData = append(h.recvData, data...)
	h.Unlock()
	if len(data) > 0 {
		belogs.Debug("[Client] 接收数据:", hex.EncodeToString(data[:min(len(data), 10)])+"...")
	} else {
		belogs.Debug("[Client] 接收空数据")
	}
	return nil
}

// -------------------------- 通用工具函数 --------------------------
func genRandData(size int) []byte {
	data := make([]byte, size)
	rand.Read(data)
	return data
}

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

// -------------------------- 测试用例 --------------------------
func TestTCP_NoTLS(t *testing.T) {
	serverAddr := "127.0.0.1:9999"
	serverHandler := NewTestServerHandler()
	server := NewTcpServer(serverHandler, WithReadWriteTimeout(10*time.Second, 10*time.Second))

	// 异步启动服务端
	go func() {
		if err := server.Start(serverAddr); err != nil && err.Error() != "server already closed" {
			t.Fatal("服务端启动失败:", err)
		}
	}()

	// 等待服务端就绪（增加重试次数）
	if !waitReady(serverAddr, 5*time.Second) {
		t.Fatal("服务端未就绪")
	}

	// 启动客户端（修复连接超时）
	clientHandler := NewTestClientHandler()
	client := NewTcpClient(clientHandler, WithClientReadWriteTimeout(10*time.Second, 10*time.Second))

	// 直接调用Start，无需chan等待（修复后Start不会阻塞）
	clientErr := client.Start(serverAddr)
	if clientErr != nil {
		t.Fatal("客户端启动失败:", clientErr)
	}
	// 短暂等待确保连接完全建立
	time.Sleep(500 * time.Millisecond)

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
			// 清空缓存
			clientHandler.Lock()
			clientHandler.recvData = nil
			clientHandler.Unlock()
			clientAddr := client.conn.RemoteAddr().String()
			serverHandler.Lock()
			serverHandler.recvData[clientAddr] = nil
			serverHandler.Unlock()

			// 连续发送3次
			for i := 0; i < 3; i++ {
				if err := client.CallProcessFunc(tc.data); err != nil {
					t.Fatalf("发送失败(%d): %v", i, err)
				}
				time.Sleep(200 * time.Millisecond)
			}

			// 验证客户端接收
			clientHandler.Lock()
			recv := clientHandler.recvData
			clientHandler.Unlock()
			expected := repeatString(tc.data, 3)

			if string(recv) != expected {
				t.Errorf(
					"接收数据不匹配: 期望长度%d, 实际长度%d | 期望前10字符[%s], 实际[%s]",
					len(expected), len(recv),
					safeSubstr(expected, 10), safeSubstr(string(recv), 10),
				)
			}

			// 验证服务端接收
			serverHandler.Lock()
			serverRecv := serverHandler.recvData[clientAddr]
			serverHandler.Unlock()

			if string(serverRecv) != expected {
				t.Errorf(
					"服务端接收数据不匹配: 期望前10字符[%s], 实际[%s]",
					safeSubstr(expected, 10), safeSubstr(string(serverRecv), 10),
				)
			}
		})
	}

	// 测试主动关闭连接
	clientAddr := client.conn.RemoteAddr().String()
	ok, err := server.CloseConnByAddr(clientAddr)
	if err != nil {
		t.Fatal("主动关闭连接失败:", err)
	}
	if !ok {
		t.Error("未找到指定客户端连接")
	}
	time.Sleep(500 * time.Millisecond)

	if server.GetConnCount() != 0 {
		t.Errorf("连接数未归零: 期望0, 实际%d", server.GetConnCount())
	}

	// 清理资源
	client.CallStop()
	server.Stop()
	time.Sleep(200 * time.Millisecond)
}

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
		if err := server.Start(serverAddr); err != nil && err.Error() != "server already closed" {
			t.Fatal("TLS服务端启动失败:", err)
		}
	}()

	if !waitReady(serverAddr, 3*time.Second) {
		t.Fatal("TLS服务端未就绪")
	}

	clientHandler := NewTestClientHandler()
	tlsClientCfg := &ClientTLSConfig{
		RootCAFile:         filepath.Join(certDir, "ca.crt"),
		ServerName:         "localhost",
		InsecureSkipVerify: false,
	}
	client := NewTcpClient(clientHandler, WithClientTLS(tlsClientCfg))

	var clientErr error
	clientDone := make(chan struct{})
	go func() {
		clientErr = client.Start(serverAddr)
		close(clientDone)
	}()
	select {
	case <-clientDone:
	case <-time.After(2 * time.Second):
		t.Fatal("TLS客户端连接超时")
	}

	if clientErr != nil {
		t.Fatal("TLS客户端连接失败:", clientErr)
	}

	// 发送测试数据
	testData := genRandData(2048)
	err := client.CallProcessFunc(string(testData))
	if err != nil {
		t.Fatal("TLS客户端发送失败:", err)
	}
	time.Sleep(200 * time.Millisecond)

	// 验证数据
	clientHandler.Lock()
	if string(clientHandler.recvData) != string(testData) {
		t.Errorf(
			"TLS客户端接收数据不匹配: 期望长度%d, 实际%d",
			len(testData), len(clientHandler.recvData),
		)
	}
	clientHandler.Unlock()

	// 清理
	client.CallStop()
	server.Stop()
	time.Sleep(200 * time.Millisecond)
}

func TestTLS_TwoWay(t *testing.T) {
	certDir := generateTestCerts(t)
	serverAddr := "127.0.0.1:9997"

	serverHandler := NewTestServerHandler()
	tlsServerCfg := &ServerTLSConfig{
		ServerCertFile: filepath.Join(certDir, "server.crt"),
		ServerKeyFile:  filepath.Join(certDir, "server.key"),
		RootCAFile:     filepath.Join(certDir, "ca.crt"),
		ClientAuth:     tls.RequireAndVerifyClientCert,
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
	clientDone := make(chan struct{})
	go func() {
		clientErr = client.Start(serverAddr)
		close(clientDone)
	}()
	select {
	case <-clientDone:
	case <-time.After(2 * time.Second):
		t.Fatal("双向认证客户端连接超时")
	}

	if clientErr != nil {
		t.Fatal("双向认证客户端连接失败:", clientErr)
	}

	// 连续发送数据
	for i := 0; i < 5; i++ {
		testData := fmt.Sprintf("双向认证测试-%d", i)
		err := client.CallProcessFunc(testData)
		if err != nil {
			t.Errorf("双向认证发送失败(%d): %v", i, err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 验证数据
	clientHandler.Lock()
	expected := "双向认证测试-0双向认证测试-1双向认证测试-2双向认证测试-3双向认证测试-4"
	if string(clientHandler.recvData) != expected {
		t.Errorf(
			"双向认证接收数据不匹配: 期望[%s], 实际[%s]",
			expected, safeSubstr(string(clientHandler.recvData), 50),
		)
	}
	clientHandler.Unlock()

	// 清理
	client.CallStop()
	server.Stop()
	time.Sleep(200 * time.Millisecond)
}

func TestEdgeCases(t *testing.T) {
	serverAddr := "127.0.0.1:9996"
	serverHandler := NewTestServerHandler()
	server := NewTcpServer(serverHandler, WithReadWriteTimeout(1*time.Second, 1*time.Second))

	go func() {
		_ = server.Start(serverAddr)
	}()
	time.Sleep(300 * time.Millisecond)

	// 空数据发送
	t.Run("空数据发送", func(t *testing.T) {
		client := NewTcpClient(NewTestClientHandler())
		clientDone := make(chan struct{})
		go func() {
			_ = client.Start(serverAddr)
			close(clientDone)
		}()
		select {
		case <-clientDone:
		case <-time.After(2 * time.Second):
			t.Fatal("客户端连接超时")
		}

		err := client.CallProcessFunc("")
		if err != nil {
			t.Error("空数据发送失败:", err)
		}
		time.Sleep(200 * time.Millisecond)
		client.CallStop()
	})

	// 超大数据
	t.Run("超大数据(16KB)", func(t *testing.T) {
		client := NewTcpClient(NewTestClientHandler())
		clientDone := make(chan struct{})
		go func() {
			_ = client.Start(serverAddr)
			close(clientDone)
		}()
		select {
		case <-clientDone:
		case <-time.After(2 * time.Second):
			t.Fatal("客户端连接超时")
		}

		largeData := genRandData(16 * 1024)
		err := client.CallProcessFunc(string(largeData))
		if err != nil {
			t.Fatal("超大数据发送失败:", err)
		}
		time.Sleep(300 * time.Millisecond)

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

	// 超时断开
	t.Run("读超时断开", func(t *testing.T) {
		client := NewTcpClient(NewTestClientHandler())
		clientDone := make(chan struct{})
		go func() {
			_ = client.Start(serverAddr)
			close(clientDone)
		}()
		select {
		case <-clientDone:
		case <-time.After(2 * time.Second):
			t.Fatal("客户端连接超时")
		}

		time.Sleep(1500 * time.Millisecond)
		if server.GetConnCount() != 0 {
			t.Errorf("超时后未自动断开连接: 连接数%d", server.GetConnCount())
		}
		client.CallStop()
	})

	// 并发连接
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

		if serverHandler.closeCount < 10 {
			t.Errorf("并发连接关闭计数不匹配: 期望≥10, 实际%d", serverHandler.closeCount)
		}
	})

	server.Stop()
	time.Sleep(200 * time.Millisecond)
}

// -------------------------- 性能测试 --------------------------
func BenchmarkTCP_Throughput(b *testing.B) {
	serverAddr := "127.0.0.1:9995"
	serverHandler := NewTestServerHandler()
	server := NewTcpServer(serverHandler)

	go func() {
		_ = server.Start(serverAddr)
	}()
	time.Sleep(300 * time.Millisecond)

	clientHandler := NewTestClientHandler()
	client := NewTcpClient(clientHandler)
	go func() {
		_ = client.Start(serverAddr)
	}()
	time.Sleep(300 * time.Millisecond)

	testData := genRandData(1024)
	testStr := string(testData)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = client.CallProcessFunc(testStr)
		}
	})

	b.StopTimer()
	b.ReportAllocs()

	client.CallStop()
	server.Stop()
}

func BenchmarkTCP_ConcurrentConn(b *testing.B) {
	serverAddr := "127.0.0.1:9994"
	serverHandler := NewTestServerHandler()
	server := NewTcpServer(serverHandler)

	go func() {
		_ = server.Start(serverAddr)
	}()
	time.Sleep(300 * time.Millisecond)

	b.ResetTimer()

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

	if server.GetConnCount() != 0 {
		b.Error("并发压测后连接数未归零")
	}

	server.Stop()
}
