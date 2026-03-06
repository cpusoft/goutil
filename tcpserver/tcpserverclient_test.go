package tcpserver

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/cpusoft/goutil/belogs"
)

// tcpserver/testdata/ 目录下放测试证书：
// ca.crt (根CA)、server.crt/server.key (服务端证书)、client.crt/client.key (客户端证书)
// # 运行所有测试
//go test -v ./tcpserver
//
//# 运行指定测试
//go test -v ./tcpserver -run TestTCP_CriticalValues

// 测试用的证书生成工具（临时生成测试证书，避免依赖外部文件）
func generateTestCerts(t *testing.T, certDir string) {
	// 创建目录
	if err := os.MkdirAll(certDir, 0755); err != nil {
		t.Fatalf("创建证书目录失败: %v", err)
	}

	// 生成根CA
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
	if err := os.WriteFile(filepath.Join(certDir, "ca.crt"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caBytes}), 0644); err != nil {
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
	// 保存服务端证书/私钥
	if err := os.WriteFile(filepath.Join(certDir, "server.crt"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverBytes}), 0644); err != nil {
		t.Fatalf("保存服务端证书失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(certDir, "server.key"), pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey)}), 0600); err != nil {
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
	// 保存客户端证书/私钥
	if err := os.WriteFile(filepath.Join(certDir, "client.crt"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientBytes}), 0644); err != nil {
		t.Fatalf("保存客户端证书失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(certDir, "client.key"), pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientPrivKey)}), 0600); err != nil {
		t.Fatalf("保存客户端私钥失败: %v", err)
	}
}

// 通用测试配置
type testConfig struct {
	serverAddr string
	certDir    string
}

// 初始化测试配置
func initTestConfig(t *testing.T) testConfig {
	certDir := filepath.Join(t.TempDir(), "certs")
	generateTestCerts(t, certDir)
	return testConfig{
		serverAddr: "127.0.0.1:9999",
		certDir:    certDir,
	}
}

// 服务器处理函数实现（测试用）
type testServerProcess struct{}

func (t *testServerProcess) OnConnect(conn *net.TCPConn) error {
	belogs.Debug("测试服务器: 客户端连接成功 - ", conn.RemoteAddr())
	return nil
}

func (t *testServerProcess) OnReceiveAndSend(conn *net.TCPConn, receiveData []byte) error {
	belogs.Debug("测试服务器: 收到数据 - ", string(receiveData))
	// 回显数据
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err := conn.Write(receiveData)
	return err
}

func (t *testServerProcess) OnClose(conn *net.TCPConn) {
	belogs.Debug("测试服务器: 客户端断开连接 - ", conn.RemoteAddr())
}

func (t *testServerProcess) ActiveSend(conn *net.TCPConn, sendData []byte) error {
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err := conn.Write(sendData)
	return err
}

// 超时测试专用服务器处理函数
type slowTestServerProcess struct{}

func (t *slowTestServerProcess) OnConnect(conn *net.TCPConn) error {
	return nil
}

func (t *slowTestServerProcess) OnReceiveAndSend(conn *net.TCPConn, receiveData []byte) error {
	// 模拟处理超时（超过1秒读写超时）
	time.Sleep(2 * time.Second)
	return nil
}

func (t *slowTestServerProcess) OnClose(conn *net.TCPConn) {}

func (t *slowTestServerProcess) ActiveSend(conn *net.TCPConn, sendData []byte) error {
	return nil
}

// 客户端处理函数实现（测试用）
type testClientProcess struct {
	receivedData chan []byte
}

func newTestClientProcess() *testClientProcess {
	return &testClientProcess{
		receivedData: make(chan []byte, 100),
	}
}

func (t *testClientProcess) ActiveSend(conn *net.TCPConn, processChan string) error {
	belogs.Debug("测试客户端: 主动发送数据 - ", processChan)
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err := conn.Write([]byte(processChan))
	return err
}

func (t *testClientProcess) OnReceive(conn *net.TCPConn, receiveData []byte) error {
	belogs.Debug("测试客户端: 收到数据 - ", string(receiveData))
	t.receivedData <- receiveData
	return nil
}

////////////////////////////////////////////////////////
///////////////////////////////////////////////////////

// 测试1: 无证书TCP通信
func TestTCP_NoTLS(t *testing.T) {
	cfg := initTestConfig(t)

	// 启动服务器
	server := NewTcpServer(&testServerProcess{}, WithReadWriteTimeout(10*time.Second, 5*time.Second))
	go func() {
		if err := server.Start(cfg.serverAddr); err != nil {
			t.Errorf("服务器启动失败: %v", err)
		}
	}()
	defer server.Stop()

	// 等待服务器启动
	time.Sleep(3 * time.Second)

	// 启动客户端
	clientProcess := newTestClientProcess()
	client := NewTcpClient(clientProcess)
	go func() {
		if err := client.Start(cfg.serverAddr); err != nil {
			t.Errorf("客户端启动失败: %v", err)
		}
	}()
	defer client.CallStop()

	// 等待客户端连接
	time.Sleep(3 * time.Second)

	// 测试发送接收
	testData := "test-no-tls-123456"
	client.CallProcessFunc(testData)

	// 验证接收
	select {
	case data := <-clientProcess.receivedData:
		if string(data) != testData {
			t.Errorf("数据回显失败: 期望 %s, 实际 %s", testData, string(data))
		}
	case <-time.After(30 * time.Second):
		t.Fatal("客户端未收到回显数据，超时")
	}
}

// 测试2: 单向TLS认证（服务端验证，客户端不验证）
func TestTLS_OneWayAuth(t *testing.T) {
	cfg := initTestConfig(t)

	// 服务器配置（仅服务端证书，不要求客户端证书）
	serverTLSConfig := &ServerTLSConfig{
		ServerCertFile: filepath.Join(cfg.certDir, "server.crt"),
		ServerKeyFile:  filepath.Join(cfg.certDir, "server.key"),
		RootCAFile:     filepath.Join(cfg.certDir, "ca.crt"),
		ClientAuth:     tls.NoClientCert, // 不要求客户端证书
	}
	server := NewTcpServer(&testServerProcess{},
		WithReadWriteTimeout(10*time.Second, 5*time.Second),
		WithServerTLS(serverTLSConfig),
	)
	go func() {
		if err := server.Start(cfg.serverAddr); err != nil {
			t.Errorf("TLS服务器启动失败: %v", err)
		}
	}()
	defer server.Stop()
	time.Sleep(100 * time.Millisecond)

	// 客户端配置（仅验证服务端证书）
	clientTLSConfig := &ClientTLSConfig{
		RootCAFile:         filepath.Join(cfg.certDir, "ca.crt"),
		ServerName:         "localhost",
		InsecureSkipVerify: false,
	}
	clientProcess := newTestClientProcess()
	client := NewTcpClient(clientProcess, WithClientTLS(clientTLSConfig))
	go func() {
		if err := client.Start(cfg.serverAddr); err != nil {
			t.Errorf("TLS客户端启动失败: %v", err)
		}
	}()
	defer client.CallStop()
	time.Sleep(100 * time.Millisecond)

	// 测试发送接收
	testData := "test-one-way-tls-7890"
	client.CallProcessFunc(testData)

	select {
	case data := <-clientProcess.receivedData:
		if string(data) != testData {
			t.Errorf("单向TLS数据回显失败: 期望 %s, 实际 %s", testData, string(data))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("单向TLS客户端未收到回显数据，超时")
	}
}

// 测试3: 双向TLS认证（服务端验证客户端，客户端验证服务端）
func TestTLS_MutualAuth(t *testing.T) {
	cfg := initTestConfig(t)

	// 服务器配置（双向认证）
	serverTLSConfig := &ServerTLSConfig{
		ServerCertFile: filepath.Join(cfg.certDir, "server.crt"),
		ServerKeyFile:  filepath.Join(cfg.certDir, "server.key"),
		RootCAFile:     filepath.Join(cfg.certDir, "ca.crt"),
		ClientAuth:     tls.RequireAndVerifyClientCert,
	}
	server := NewTcpServer(&testServerProcess{},
		WithReadWriteTimeout(10*time.Second, 5*time.Second),
		WithServerTLS(serverTLSConfig),
	)
	go func() {
		if err := server.Start(cfg.serverAddr); err != nil {
			t.Errorf("双向TLS服务器启动失败: %v", err)
		}
	}()
	defer server.Stop()
	time.Sleep(100 * time.Millisecond)

	// 客户端配置（双向认证）
	clientTLSConfig := &ClientTLSConfig{
		ClientCertFile:     filepath.Join(cfg.certDir, "client.crt"),
		ClientKeyFile:      filepath.Join(cfg.certDir, "client.key"),
		RootCAFile:         filepath.Join(cfg.certDir, "ca.crt"),
		ServerName:         "localhost",
		InsecureSkipVerify: false,
	}
	clientProcess := newTestClientProcess()
	client := NewTcpClient(clientProcess, WithClientTLS(clientTLSConfig))
	go func() {
		if err := client.Start(cfg.serverAddr); err != nil {
			t.Errorf("双向TLS客户端启动失败: %v", err)
		}
	}()
	defer client.CallStop()
	time.Sleep(100 * time.Millisecond)

	// 测试发送接收
	testData := "test-mutual-tls-111222"
	client.CallProcessFunc(testData)

	select {
	case data := <-clientProcess.receivedData:
		if string(data) != testData {
			t.Errorf("双向TLS数据回显失败: 期望 %s, 实际 %s", testData, string(data))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("双向TLS客户端未收到回显数据，超时")
	}
}

// 测试4: 临界值测试（超大/超小数据、超时、断连）
func TestTCP_CriticalValues(t *testing.T) {
	cfg := initTestConfig(t)

	// 启动无证书服务器
	server := NewTcpServer(&testServerProcess{},
		WithReadWriteTimeout(1*time.Second, 1*time.Second), // 短超时
	)
	go func() {
		if err := server.Start(cfg.serverAddr); err != nil {
			t.Errorf("临界值测试服务器启动失败: %v", err)
		}
	}()
	defer server.Stop()
	time.Sleep(100 * time.Millisecond)

	// 启动客户端
	clientProcess := newTestClientProcess()
	client := NewTcpClient(clientProcess)
	go func() {
		if err := client.Start(cfg.serverAddr); err != nil {
			t.Errorf("临界值测试客户端启动失败: %v", err)
		}
	}()
	defer client.CallStop()
	time.Sleep(100 * time.Millisecond)

	// 测试1: 超大数据（超过默认buffer 2048）
	bigData := make([]byte, 4096)
	for i := range bigData {
		bigData[i] = byte(i % 256)
	}
	client.CallProcessFunc(string(bigData))
	select {
	case data := <-clientProcess.receivedData:
		if len(data) != len(bigData) {
			t.Errorf("超大数据接收失败: 期望长度 %d, 实际 %d", len(bigData), len(data))
		}
	case <-time.After(5 * time.Second):
		t.Error("超大数据接收超时")
	}

	// 测试2: 超小数据（空数据）
	client.CallProcessFunc("")
	time.Sleep(100 * time.Millisecond) // 空数据应被忽略，无报错

	// 测试3: 超时测试
	// 启动慢服务器（处理超时）
	slowServerInst := NewTcpServer(&slowTestServerProcess{}, WithReadWriteTimeout(1*time.Second, 1*time.Second))
	go func() {
		if err := slowServerInst.Start("127.0.0.1:8888"); err != nil {
			t.Errorf("慢服务器启动失败: %v", err)
		}
	}()
	defer slowServerInst.Stop()
	time.Sleep(100 * time.Millisecond)

	// 启动慢客户端
	slowClientProcess := newTestClientProcess()
	slowClient := NewTcpClient(slowClientProcess)
	clientDone := make(chan error, 1)
	go func() {
		clientDone <- slowClient.Start("127.0.0.1:8888")
	}()
	time.Sleep(100 * time.Millisecond)

	// 发送测试数据
	slowClient.CallProcessFunc("test-timeout")
	time.Sleep(2 * time.Second)

	// 停止客户端并验证是否因超时退出
	slowClient.CallStop()
	select {
	case err := <-clientDone:
		if err != nil {
			// 超时会导致Read失败，属于预期结果
			t.Logf("超时测试通过: 客户端因超时退出，错误: %v", err)
		} else {
			t.Error("超时测试失败: 客户端未因超时退出")
		}
	case <-time.After(2 * time.Second):
		t.Error("超时测试失败: 客户端无响应")
	}
}

// 测试5: 性能测试（并发连接、高吞吐）
func TestTCP_Performance(t *testing.T) {
	cfg := initTestConfig(t)
	// 性能测试降低日志级别

	// 启动服务器
	server := NewTcpServer(&testServerProcess{})
	go func() {
		if err := server.Start(cfg.serverAddr); err != nil {
			t.Errorf("性能测试服务器启动失败: %v", err)
		}
	}()
	defer server.Stop()
	time.Sleep(100 * time.Millisecond)

	// 配置参数
	const (
		clientCount = 100  // 并发客户端数
		sendCount   = 1000 // 每个客户端发送次数
		dataSize    = 1024 // 每次发送数据大小
	)

	// 生成测试数据
	testData := make([]byte, dataSize)
	_, err := rand.Read(testData)
	if err != nil {
		t.Fatalf("生成测试数据失败: %v", err)
	}

	// 启动并发客户端
	var wg sync.WaitGroup
	startTime := time.Now()
	successCount := 0
	failCount := 0
	mu := sync.Mutex{}

	for i := 0; i < clientCount; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			clientProcess := newTestClientProcess()
			client := NewTcpClient(clientProcess)
			if err := client.Start(cfg.serverAddr); err != nil {
				mu.Lock()
				failCount++
				mu.Unlock()
				t.Errorf("客户端%d启动失败: %v", clientID, err)
				return
			}
			defer client.CallStop()

			// 发送数据
			for j := 0; j < sendCount; j++ {
				client.CallProcessFunc(string(testData))
				select {
				case <-clientProcess.receivedData:
					mu.Lock()
					successCount++
					mu.Unlock()
				case <-time.After(1 * time.Second):
					mu.Lock()
					failCount++
					mu.Unlock()
				}
			}
		}(i)
	}

	// 等待所有客户端完成
	wg.Wait()
	duration := time.Since(startTime)
	totalData := clientCount * sendCount * dataSize
	throughput := float64(totalData) / duration.Seconds() / 1024 / 1024 // MB/s

	// 输出性能指标
	t.Logf("性能测试结果:")
	t.Logf("  并发客户端数: %d", clientCount)
	t.Logf("  每个客户端发送次数: %d", sendCount)
	t.Logf("  总发送次数: %d", clientCount*sendCount)
	t.Logf("  成功次数: %d", successCount)
	t.Logf("  失败次数: %d", failCount)
	t.Logf("  总耗时: %.2f秒", duration.Seconds())
	t.Logf("  总数据量: %.2fMB", float64(totalData)/1024/1024)
	t.Logf("  吞吐量: %.2fMB/s", throughput)

	// 性能阈值校验（可根据实际需求调整）
	if failCount > clientCount*sendCount*0.01 { // 失败率>1%则报错
		t.Errorf("性能测试失败率过高: %d/%d (%.2f%%)", failCount, clientCount*sendCount, float64(failCount)/float64(clientCount*sendCount)*100)
	}
	if throughput < 1 { // 降低阈值适配普通机器，实际可调整为10
		t.Logf("性能警告: 吞吐量偏低 (%.2fMB/s)", throughput)
	}
}
