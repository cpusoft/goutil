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

// GetByCurlWithConfig get 请求，通过 curl 实现
func GetByCurlWithConfig(url string, httpClientConfig *HttpClientConfig) (result string, err error) {
	return doCurlWithConfig("GET", url, "", httpClientConfig)
}

// PostByCurlWithConfig post 请求，通过 curl 实现
func PostByCurlWithConfig(url, postJson string, httpClientConfig *HttpClientConfig) (result string, err error) {
	return doCurlWithConfig("POST", url, postJson, httpClientConfig)
}

// doCurlWithConfig 提取出的公共 curl 执行逻辑
func doCurlWithConfig(method, url, postBody string, httpClientConfig *HttpClientConfig) (result string, err error) {
	belogs.Debug("doCurlWithConfig(): method:", method, "  url:", url, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))

	url = strings.TrimSpace(url)
	if len(url) == 0 {
		return "", errors.New("url is empty")
	}
	if httpClientConfig == nil {
		return "", errors.New("httpClientConfig is empty")
	}
	if method == "POST" && strings.TrimSpace(postBody) == "" {
		return "", errors.New("postBody is empty")
	}

	cfg := httpClientConfig

	var ctx context.Context
	var cancel context.CancelFunc
	if cfg.AllTimeoutSeconds > 0 {
		goTimeout := time.Duration(cfg.AllTimeoutSeconds)*time.Second + 30*time.Second
		ctx, cancel = context.WithTimeout(context.Background(), goTimeout)
		belogs.Debug("doCurlWithConfig(): goTimeout(s):", goTimeout.Seconds())
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	// 创建输出临时文件
	tmpFile, err := os.CreateTemp("", "_tmp_resp_")
	if err != nil {
		belogs.Error("doCurlWithConfig(): TempFile fail", err)
		return "", err
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	args := make([]string, 0)
	args = append(args, "-X", method)

	if cfg.AllTimeoutSeconds > 0 {
		args = append(args, "-m", convert.ToString(cfg.AllTimeoutSeconds))
	}
	if cfg.ConnectionTimeoutSeconds > 0 {
		args = append(args, "--connect-timeout", convert.ToString(cfg.ConnectionTimeoutSeconds))
	}
	if cfg.IpType == "ipv4" {
		args = append(args, "-4")
	} else if cfg.IpType == "ipv6" {
		args = append(args, "-6")
	}
	if !cfg.VerifyHttps {
		args = append(args, "--insecure")
	} else if cfg.RootFile != "" {
		args = append(args, "--cacert", cfg.RootFile)
	}
	if cfg.CrtFile != "" && cfg.KeyFile != "" {
		args = append(args, "--cert", cfg.CrtFile, "--key", cfg.KeyFile)
	}
	if cfg.RetryCount > 0 {
		args = append(args, "--retry", convert.ToString(cfg.RetryCount))
	}
	if cfg.ContentType != "" {
		args = append(args, "-H", "Content-Type: "+cfg.ContentType)
	} else if method == "POST" {
		args = append(args, "-H", "Content-Type: application/json")
	}
	if cfg.Authorization != "" {
		args = append(args, "-H", "Authorization: "+cfg.Authorization)
	}

	// ---- 请求体处理 (POST) ----
	if method == "POST" {
		bodyFile, err := os.CreateTemp("", "_tmp_body_")
		if err != nil {
			belogs.Error("doCurlWithConfig(): TempBodyFile fail", err)
			return "", err
		}
		// 注意：这里不要 defer close，必须在执行 cmd 前主动 close
		bodyFileName := bodyFile.Name()
		defer os.Remove(bodyFileName) // 确保退出时删除

		_, err = bodyFile.WriteString(postBody)
		if err != nil {
			bodyFile.Close()
			belogs.Error("doCurlWithConfig(): WriteString to bodyFile fail", err)
			return "", err
		}
		// 【关键修复】必须在 curl 读取前关闭文件，否则 Windows 下 curl 无法读取！
		err = bodyFile.Close()
		if err != nil {
			belogs.Error("doCurlWithConfig(): bodyFile.Close fail", err)
			return "", err
		}

		args = append(args, "--data-binary", "@"+bodyFileName)
	}

	args = append(args, "--compressed", "-o", tmpFile.Name(), url)

	belogs.Debug("doCurlWithConfig(): will curl, args:", args)

	start := time.Now()
	cmd := exec.CommandContext(ctx, "curl", args...)
	output, err := cmd.CombinedOutput()
	outputStr := GetOutputStr(output)

	if ctx.Err() != nil {
		belogs.Error("doCurlWithConfig(): CURL fail, ctx err, method:", method, "  url:", url, "  ctxErr:", ctx.Err(),
			"  tmpFile:", tmpFile.Name(), "  time(s):", time.Since(start))
		return "", fmt.Errorf("curl %s timeout/canceled: %w", url, ctx.Err())
	}

	if err != nil {
		belogs.Error("doCurlWithConfig(): exec.Command fail, method:", method, "  curl:", url, "  ipAddrs:", netutil.LookupIpByUrl(url),
			"  tmpFile:", tmpFile.Name(), "  cfg:", jsonutil.MarshalJson(cfg),
			"  len(output):", len(output), "  outputStr:", outputStr, "  time(s):", time.Since(start), "   err:", err)
		return "", errors.New("Fail to " + method + " by curl. url is " + url + ". " + err.Error())
	}

	b, err := fileutil.ReadFileToBytes(tmpFile.Name())
	if err != nil {
		belogs.Error("doCurlWithConfig(): ReadFileToBytes fail, url", url, "   tmpFile:", tmpFile.Name(),
			"   err: ", err, "   output: "+string(output))
		return "", errors.New("Fail to " + method + " by curl. Error is `" + err.Error() + "`. Output is `" + string(output) + "`")
	}
	belogs.Debug("doCurlWithConfig(): ReadFileToBytes ok, url:", url, "   tmpFile:", tmpFile.Name(),
		"  len(b):", len(b), "  time(s):", time.Since(start))
	return string(b), nil
}
