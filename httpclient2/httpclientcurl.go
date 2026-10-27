package httpclient2

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/netutil"
)

// get by Curl
func GetByCurlWithConfig(url string, httpClientConfig *HttpClientConfig) (result string, err error) {
	belogs.Debug("GetByCurlWithConfig(): url:", url, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	url = strings.TrimSpace(url)
	if len(url) == 0 {
		return "", errors.New("url is empty")
	}
	if httpClientConfig == nil {
		return "", errors.New("httpClientConfig is empty")
	}
	cfg := httpClientConfig

	// -------------------------- Go 层面的超时兜底（双保险） --------------------------
	// AllTimeoutSeconds > 0 时：goTimeout = curl 的 -m + 30s 缓冲
	// AllTimeoutSeconds == 0 时：不施加 Go 超时，仅可被 cancel 打断
	var ctx context.Context
	var cancel context.CancelFunc
	if cfg.AllTimeoutSeconds > 0 {
		goTimeout := time.Duration(cfg.AllTimeoutSeconds)*time.Second + 30*time.Second
		ctx, cancel = context.WithTimeout(context.Background(), goTimeout)
		belogs.Debug("GetByCurlWithConfig(): goTimeout(s):", goTimeout.Seconds())
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	tmpFile, err := os.CreateTemp("", "_tmp_") // temp file
	if err != nil {
		belogs.Error("GetByCurlWithConfig(): TempFile fail", err)
		return "", err
	}
	defer tmpFile.Close() // 关闭 Go 打开的句柄，避免 fd 泄漏
	defer os.Remove(tmpFile.Name())

	args := make([]string, 0)

	// ---- 1) 整体传输最大耗时（-m / --max-time） ----
	if cfg.AllTimeoutSeconds > 0 {
		args = append(args, "-m", convert.ToString(cfg.AllTimeoutSeconds))
	}

	// ---- 2) 连接阶段超时（TCP 连接，不含 TLS 握手） ----
	// 注：TlsHandshakeTimeoutSeconds 在 curl 中没有独立选项，
	//     --connect-timeout 只覆盖 TCP 连接阶段；如需限制 TLS 握手，
	//     需在 Go 侧用 context 控制整体耗时，或依赖 -m 总超时。
	if cfg.ConnectionTimeoutSeconds > 0 {
		args = append(args, "--connect-timeout", convert.ToString(cfg.ConnectionTimeoutSeconds))
	}

	// ---- 3) IP 协议类型 ----
	if cfg.IpType == "ipv4" {
		args = append(args, "-4")
	} else if cfg.IpType == "ipv6" {
		args = append(args, "-6")
	}

	// ---- 4) HTTPS 相关 ----
	if !cfg.VerifyHttps {
		// 跳过服务端证书校验
		args = append(args, "--insecure")
	} else if cfg.RootFile != "" {
		// 信任指定的 CA 文件
		// 注意：curl 的 --cacert 会替换默认信任库，而非像 Go 那样追加；
		//       若要“系统 CA + 自定义 CA”，需另外拼接或设置 CURL_CA_BUNDLE。
		args = append(args, "--cacert", cfg.RootFile)
	}
	if cfg.CrtFile != "" && cfg.KeyFile != "" {
		// mTLS 客户端证书
		args = append(args, "--cert", cfg.CrtFile, "--key", cfg.KeyFile)
	}

	// ---- 5) 自动重试 ----
	if cfg.RetryCount > 0 {
		args = append(args, "--retry", convert.ToString(cfg.RetryCount))
	}

	// ---- 6) 业务请求头 ----
	if cfg.ContentType != "" {
		args = append(args, "-H", "Content-Type: "+cfg.ContentType)
	}
	if cfg.Authorization != "" {
		args = append(args, "-H", "Authorization: "+cfg.Authorization)
	}

	// ---- 7) 响应体 gzip 自动解压 + 输出文件 ----
	args = append(args, "--compressed", "-o", tmpFile.Name(), url)

	belogs.Debug("GetByCurlWithConfig(): will curl, args:", args)

	start := time.Now()
	// 用 CommandContext 保证 Go 兜底超时能杀掉 curl 进程
	cmd := exec.CommandContext(ctx, "curl", args...)
	output, err := cmd.CombinedOutput()
	outputStr := GetOutputStr(output)

	// 判断是否为 Go 层面的超时/取消
	if ctx.Err() != nil {
		belogs.Error("GetByCurlWithConfig(): CURL fail, ctx err, url:", url, "  ctxErr:", ctx.Err(),
			"  tmpFile:", tmpFile.Name(), "  time(s):", time.Since(start))
		return "", fmt.Errorf("curl %s timeout/canceled: %w", url, ctx.Err())
	}

	if err != nil {
		belogs.Error("GetByCurlWithConfig(): exec.Command fail, curl:", url, "  ipAddrs:", netutil.LookupIpByUrl(url),
			"  tmpFile:", tmpFile.Name(), "  cfg:", jsonutil.MarshalJson(cfg),
			"  len(output):", len(output), "  outputStr:", outputStr, "  time(s):", time.Since(start), "   err:", err)
		return "", errors.New("Fail to get by curl. url is " + url + ". " + err.Error())
	}
	belogs.Debug("GetByCurlWithConfig(): curl ok, url:", url, "   tmpFile:", tmpFile.Name(),
		"  len(output):", len(output), "  outputStr:", outputStr, "  time(s):", time.Since(start))

	b, err := fileutil.ReadFileToBytes(tmpFile.Name())
	if err != nil {
		belogs.Error("GetByCurlWithConfig(): ReadFileToBytes fail, url", url, "   tmpFile:", tmpFile.Name(),
			"   err: ", err, "   output: "+string(output))
		return "", errors.New("Fail to get by curl. Error is `" + err.Error() + "`. Output is `" + string(output) + "`")
	}
	belogs.Debug("GetByCurlWithConfig(): ReadFileToBytes ok, url:", url, "   tmpFile:", tmpFile.Name(),
		"  len(b):", len(b), "  time(s):", time.Since(start))
	return string(b), nil
}
