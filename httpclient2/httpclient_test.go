package httpclient2

import (
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// ---------------- GET ----------------

func TestGetHttpWithConfig_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if ua := r.Header.Get("User-Agent"); ua != DefaultUserAgent {
			t.Errorf("unexpected User-Agent: %s", ua)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"hello":"world"}`))
	}))
	defer server.Close()

	cfg := &HttpClientConfig{AllTimeoutSeconds: 5, VerifyHttps: true}
	resp, body, err := GetHttpWithConfig(server.URL, cfg)
	if err != nil {
		t.Fatalf("GetHttpWithConfig error: %v", err)
	}
	defer CloseResponseBody(resp)

	if resp.GetStatusCode() != http.StatusOK {
		t.Errorf("unexpected status: %d", resp.GetStatusCode())
	}
	if body != `{"hello":"world"}` {
		t.Errorf("unexpected body: %s", body)
	}
	if !IsSuccessState(resp) {
		t.Error("expected success state")
	}
}

func TestGetHttpWithConfig_NoContentTypeByDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "" {
			t.Errorf("GET should not set Content-Type by default, got: %s", ct)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	cfg := &HttpClientConfig{AllTimeoutSeconds: 5, VerifyHttps: true}
	resp, _, err := GetHttpWithConfig(server.URL, cfg)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer CloseResponseBody(resp)
}

func TestGetHttpWithConfig_EmptyURL(t *testing.T) {
	if _, _, err := GetHttpWithConfig("   ", &HttpClientConfig{}); err == nil {
		t.Error("expected error for empty urlStr")
	}
}

func TestGetHttpWithConfig_NilConfig(t *testing.T) {
	if _, _, err := GetHttpWithConfig("http://example.com", nil); err == nil {
		t.Error("expected error for nil config")
	}
}

func TestGetHttpWithConfig_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	cfg := &HttpClientConfig{AllTimeoutSeconds: 1, VerifyHttps: true}
	start := time.Now()
	_, _, err := GetHttpWithConfig(server.URL, cfg)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected timeout error")
	}
	if elapsed > 2500*time.Millisecond {
		t.Errorf("timeout took too long: %v", elapsed)
	}
}

// ---------------- POST（纯 JSON） ----------------

func TestPostHttpWithConfig_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected default Content-Type application/json, got: %s", ct)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"user":"alice"}` {
			t.Errorf("unexpected request body: %s", body)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	defer server.Close()

	cfg := &HttpClientConfig{AllTimeoutSeconds: 5, VerifyHttps: true}
	resp, body, err := PostHttpWithConfig(server.URL, `{"user":"alice"}`, cfg)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer CloseResponseBody(resp)

	if resp.GetStatusCode() != http.StatusOK {
		t.Errorf("unexpected status: %d", resp.GetStatusCode())
	}
	if body != `{"code":0,"msg":"ok"}` {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestPostHttpWithConfig_AuthorizationHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "Bearer xyz" {
			t.Errorf("unexpected Authorization: %s", auth)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	cfg := &HttpClientConfig{
		AllTimeoutSeconds: 5,
		VerifyHttps:       true,
		Authorization:     "Bearer xyz",
	}
	resp, _, err := PostHttpWithConfig(server.URL, `{}`, cfg)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer CloseResponseBody(resp)
}

func TestPostHttpWithConfig_EmptyJson(t *testing.T) {
	if _, _, err := PostHttpWithConfig("http://example.com", "   ", &HttpClientConfig{}); err == nil {
		t.Error("expected error for empty postJson")
	}
}

func TestPostHttpWithConfig_NilConfig(t *testing.T) {
	if _, _, err := PostHttpWithConfig("http://example.com", `{}`, nil); err == nil {
		t.Error("expected error for nil config")
	}
}

// ---------------- POST + 反序列化 ----------------

type loginResp struct {
	Code  int    `json:"code"`
	Token string `json:"token"`
	Msg   string `json:"msg"`
}

func TestPostHttpUnmarshalStruct_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":0,"token":"abc123","msg":"ok"}`))
	}))
	defer server.Close()

	cfg := &HttpClientConfig{AllTimeoutSeconds: 5, VerifyHttps: true}
	var lr loginResp
	resp, _, err := PostHttpUnmarshalStructWithConfig(server.URL, `{"user":"alice"}`, &lr, cfg)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer CloseResponseBody(resp)

	if lr.Code != 0 || lr.Token != "abc123" || lr.Msg != "ok" {
		t.Errorf("unexpected deserialized: %+v", lr)
	}
}

func TestPostHttpUnmarshalStruct_Non2xx_ShouldNotUnmarshal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":401,"msg":"unauthorized"}`))
	}))
	defer server.Close()

	cfg := &HttpClientConfig{AllTimeoutSeconds: 5, VerifyHttps: true}
	var lr loginResp
	resp, body, err := PostHttpUnmarshalStructWithConfig(server.URL, `{}`, &lr, cfg)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer CloseResponseBody(resp)

	if resp.GetStatusCode() != http.StatusUnauthorized {
		t.Errorf("unexpected status: %d", resp.GetStatusCode())
	}
	// SetSuccessResult 仅在 2xx 时反序列化，非 2xx 应保持零值
	if lr.Code != 0 || lr.Token != "" || lr.Msg != "" {
		t.Errorf("expected zero value, got: %+v", lr)
	}
	if !strings.Contains(body, "unauthorized") {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestPostHttpUnmarshalStruct_NilResult(t *testing.T) {
	if _, _, err := PostHttpUnmarshalStructWithConfig("http://example.com", `{}`, nil, &HttpClientConfig{}); err == nil {
		t.Error("expected error for nil responseStruct")
	}
}

// ---------------- HTTPS ----------------

func TestGetHttpWithConfig_HTTPS_SkipVerify(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("secure"))
	}))
	defer server.Close()

	cfg := &HttpClientConfig{AllTimeoutSeconds: 5, VerifyHttps: false}
	resp, body, err := GetHttpWithConfig(server.URL, cfg)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer CloseResponseBody(resp)
	if body != "secure" {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestGetHttpWithConfig_HTTPS_VerifyFail(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("secure"))
	}))
	defer server.Close()

	// 自签证书不被系统信任，VerifyHttps=true 时应失败
	cfg := &HttpClientConfig{AllTimeoutSeconds: 5, VerifyHttps: true}
	if _, _, err := GetHttpWithConfig(server.URL, cfg); err == nil {
		t.Error("expected TLS verify error")
	}
}

func TestGetHttpWithConfig_HTTPS_CustomCA(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("secure"))
	}))
	defer server.Close()

	// 把 server 自签证书写成 PEM 文件，用作 RootFile
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: server.Certificate().Raw,
	})
	tmpFile := filepath.Join(t.TempDir(), "ca.pem")
	if err := os.WriteFile(tmpFile, pemBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &HttpClientConfig{
		AllTimeoutSeconds: 5,
		VerifyHttps:       true,
		RootFile:          tmpFile,
	}
	resp, body, err := GetHttpWithConfig(server.URL, cfg)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer CloseResponseBody(resp)
	if body != "secure" {
		t.Errorf("unexpected body: %s", body)
	}
}

// ---------------- 重试 ----------------

func TestGetHttpWithConfig_RetryOnConnClose(t *testing.T) {
	var count int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&count, 1)
		if n == 1 {
			// 第一次直接关闭底层连接，模拟网络错误
			if hj, ok := w.(http.Hijacker); ok {
				conn, _, err := hj.Hijack()
				if err == nil {
					_ = conn.Close()
				}
			}
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	cfg := &HttpClientConfig{
		AllTimeoutSeconds: 5,
		VerifyHttps:       true,
		RetryCount:        2,
	}
	resp, body, err := GetHttpWithConfig(server.URL, cfg)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer CloseResponseBody(resp)
	if body != "ok" {
		t.Errorf("unexpected body: %s", body)
	}
	if atomic.LoadInt32(&count) < 2 {
		t.Errorf("expected retry, actual request count=%d", count)
	}
}

// ---------------- 辅助函数 ----------------

func TestGetStatusCode_Nil(t *testing.T) {
	if got := GetStatusCode(nil); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestIsSuccessState_Nil(t *testing.T) {
	if IsSuccessState(nil) {
		t.Error("expected false for nil resp")
	}
}

func TestCloseResponseBody_Nil(t *testing.T) {
	CloseResponseBody(nil) // 不应 panic
}

func TestGetOutputStr_Truncation(t *testing.T) {
	short := []byte("hello\nworld")
	if got := GetOutputStr(short); strings.ContainsAny(got, "\r\n") {
		t.Errorf("newlines should be trimmed, got: %q", got)
	}

	long := make([]byte, 300)
	for i := range long {
		long[i] = 'a'
	}
	if got := GetOutputStr(long); len(got) != 100 {
		t.Errorf("expected length 100, got %d", len(got))
	}
}

// ---------------- 配置构造 ----------------

func TestHttpClientConfig_Setters(t *testing.T) {
	cfg := (&HttpClientConfig{}).
		SetAllTimeoutSeconds(30).
		SetTlsHandshakeTimeoutSeconds(5).
		SetConnectionTimeoutSeconds(3).
		SetRetryCount(2).
		SetIpType("ipv4").
		SetVerifyHttps(true).
		SetRootFile("/tmp/ca.pem").
		SetKeyFile("/tmp/client.key").
		SetCrtFile("/tmp/client.crt").
		SetContentType("application/json").
		SetAuthorization("Bearer xxx")

	checks := []struct {
		name string
		got  any
		want any
	}{
		{"AllTimeoutSeconds", cfg.AllTimeoutSeconds, uint64(30)},
		{"TlsHandshakeTimeoutSeconds", cfg.TlsHandshakeTimeoutSeconds, uint64(5)},
		{"ConnectionTimeoutSeconds", cfg.ConnectionTimeoutSeconds, uint64(3)},
		{"RetryCount", cfg.RetryCount, uint64(2)},
		{"IpType", cfg.IpType, "ipv4"},
		{"VerifyHttps", cfg.VerifyHttps, true},
		{"RootFile", cfg.RootFile, "/tmp/ca.pem"},
		{"KeyFile", cfg.KeyFile, "/tmp/client.key"},
		{"CrtFile", cfg.CrtFile, "/tmp/client.crt"},
		{"ContentType", cfg.ContentType, "application/json"},
		{"Authorization", cfg.Authorization, "Bearer xxx"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestHttpClientConfig_Options(t *testing.T) {
	cfg := NewHttpClientConfig(
		WithAllTimeoutSeconds(30),
		WithTlsHandshakeTimeoutSeconds(5),
		WithConnectionTimeoutSeconds(3),
		WithRetryCount(2),
		WithIpType("ipv4"),
		WithVerifyHttps(true),
		WithRootFile("/tmp/ca.pem"),
		WithKeyFile("/tmp/client.key"),
		WithCrtFile("/tmp/client.crt"),
		WithContentType("application/json"),
		WithAuthorization("Bearer xxx"),
	)
	if cfg.AllTimeoutSeconds != 30 ||
		cfg.TlsHandshakeTimeoutSeconds != 5 ||
		cfg.ConnectionTimeoutSeconds != 3 ||
		cfg.RetryCount != 2 ||
		cfg.IpType != "ipv4" ||
		!cfg.VerifyHttps ||
		cfg.RootFile != "/tmp/ca.pem" ||
		cfg.KeyFile != "/tmp/client.key" ||
		cfg.CrtFile != "/tmp/client.crt" ||
		cfg.ContentType != "application/json" ||
		cfg.Authorization != "Bearer xxx" {
		t.Errorf("unexpected config: %+v", cfg)
	}
}
