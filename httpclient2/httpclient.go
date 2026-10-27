package httpclient2

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cpusoft/goutil/stringutil"
	"github.com/imroc/req/v3"
)

func GetStatusCode(resp *req.Response) int {
	if resp == nil {
		return 0
	}
	return resp.GetStatusCode()
}
func IsSuccessState(resp *req.Response) bool {
	if resp == nil {
		return false
	}
	return resp.IsSuccessState()
}
func GetOutputStr(output []byte) string {
	var outputStr string
	if len(output) <= 100 {
		outputStr = string(output)
	} else {
		outputStr = string(output[:100])
	}
	return stringutil.TrimNewLine(outputStr)
}
func CloseResponseBody(resp *req.Response) {
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
}

// loadRootCAsWithSystem 获取系统根证书池，并将额外的 PEM 文件追加进去
func loadRootCAsWithSystem(pemFiles ...string) (*x509.CertPool, error) {
	// 1. 获取系统默认证书池（基于当前操作系统的信任库）
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("获取系统证书池失败: %w", err)
	}
	if pool == nil {
		// 在部分平台上 SystemCertPool 可能返回 nil
		pool = x509.NewCertPool()
	}

	// 2. 逐个读取并追加 PEM 证书
	for _, file := range pemFiles {
		pemData, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("读取证书文件 %q 失败: %w", file, err)
		}

		// AppendCertsFromPEM 返回 true 表示至少有一个证书被成功解析并追加
		if !pool.AppendCertsFromPEM(pemData) {
			return nil, fmt.Errorf("证书文件 %q 中未找到有效的 PEM 证书", file)
		}
	}

	return pool, nil
}
func buildClientForConfig(parsedURL *url.URL, cfg *HttpClientConfig, defaultJsonContentType string) (*req.Client, error) {
	client := req.C()

	// 1) 整体超时
	if cfg.AllTimeoutSeconds > 0 {
		client.SetTimeout(time.Duration(cfg.AllTimeoutSeconds) * time.Second)
	}

	// 2) TLS 握手超时
	if cfg.TlsHandshakeTimeoutSeconds > 0 {
		client.SetTLSHandshakeTimeout(time.Duration(cfg.TlsHandshakeTimeoutSeconds) * time.Second)
	}

	// 3) 连接超时 & IP 协议类型
	if cfg.ConnectionTimeoutSeconds > 0 || cfg.IpType == "ipv4" || cfg.IpType == "ipv6" {
		connTimeout := time.Duration(cfg.ConnectionTimeoutSeconds) * time.Second
		network := "tcp" // 双栈；IpType 为 "all"/"" 时保持默认 tcp
		switch cfg.IpType {
		case "ipv4":
			network = "tcp4"
		case "ipv6":
			network = "tcp6"
		case "all":
			network = "tcp"
		}
		client.SetDial(func(ctx context.Context, _, addr string) (net.Conn, error) {
			d := &net.Dialer{}
			if connTimeout > 0 {
				d.Timeout = connTimeout
			}
			if cfg.AllTimeoutSeconds > 0 {
				d.KeepAlive = time.Duration(cfg.AllTimeoutSeconds) * time.Second
			}
			return d.DialContext(ctx, network, addr)
		})
	}

	// 4) HTTPS 相关：校验开关 / 根证书 / 客户端证书，统一放进同一个 tls.Config
	if strings.EqualFold(parsedURL.Scheme, "https") {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		if !cfg.VerifyHttps {
			tlsConfig.InsecureSkipVerify = true
		} else if cfg.RootFile != "" {
			pool, err := loadRootCAsWithSystem(cfg.RootFile)
			if err != nil {
				return nil, fmt.Errorf("loadRootCAsWithSystem fail: %w", err)
			}
			tlsConfig.RootCAs = pool
		}
		if cfg.CrtFile != "" && cfg.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(cfg.CrtFile, cfg.KeyFile)
			if err != nil {
				return nil, fmt.Errorf("LoadX509KeyPair fail: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		client.SetTLSClientConfig(tlsConfig)
	}

	// 5) 公共 Header 与重试（原 applyCommonHeaders 的等价 Client 级写法）
	if cfg.ContentType != "" {
		client.SetCommonHeader("Content-Type", cfg.ContentType)
	} else {
		if defaultJsonContentType != "" {
			client.SetCommonHeader("Content-Type", defaultJsonContentType)
		}
	}
	if cfg.Authorization != "" {
		client.SetCommonHeader("Authorization", cfg.Authorization)
	}
	if cfg.RetryCount > 0 {
		client.SetCommonRetryCount(int(cfg.RetryCount))
	}
	client.SetCommonHeader("User-Agent", DefaultUserAgent)
	// no referer
	//client.SetCommonHeader("Referer", parsedURL.Scheme+"://"+parsedURL.Host)

	return client, nil
}
