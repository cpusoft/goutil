package httpclient2

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/imroc/req/v3"
)

// ---------------------------------------------------------------
// 1) 纯提交 POST JSON
// should use: defer CloseResponseBody(resp)
// ---------------------------------------------------------------
func PostHttpWithConfig(urlStr string, postJson string,
	httpClientConfig *HttpClientConfig) (resp *req.Response, body string, err error) {

	if strings.TrimSpace(urlStr) == "" {
		return nil, "", errors.New("urlStr is empty")
	}
	if strings.TrimSpace(postJson) == "" {
		return nil, "", errors.New("postJson is empty")
	}
	if httpClientConfig == nil {
		return nil, "", errors.New("httpClientConfig is empty")
	}
	cfg := httpClientConfig

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		belogs.Error("PostHttpWithConfig(): Parse urlStr fail, urlStr:", urlStr, err)
		return nil, "", err
	}

	client, err := buildClientForConfig(parsedURL, cfg, "application/json")
	if err != nil {
		belogs.Error("PostHttpWithConfig(): buildClientForConfig fail", err)
		return nil, "", err
	}

	r := client.R()
	// SetBodyJsonString：直接发送原始 JSON 字符串，不会二次序列化 / 转义
	r.SetBodyJsonString(postJson)
	if cfg.ContentType != "" {
		// Request 级覆盖，优先级高于 SetBodyJsonString 的默认值
		r.SetContentType(cfg.ContentType)
	}
	start := time.Now()
	resp, err = r.Post(urlStr)
	if err != nil {
		belogs.Error("PostHttpWithConfig(): r.Post fail", urlStr, "err", err, "time(s)", time.Since(start))
		return nil, "", err
	}
	belogs.Debug("PostHttpWithConfig(): r.Post ok", urlStr, "statusCode:", resp.GetStatusCode(),
		"time(s)", time.Since(start))
	return resp, resp.String(), nil
}

// ---------------------------------------------------------------
// 2) POST JSON + 自动反序列化响应到 struct
// should use: defer CloseResponseBody(resp)
// ---------------------------------------------------------------
func PostHttpUnmarshalStructWithConfig(urlStr, postJson string, responseStruct interface{},
	httpClientConfig *HttpClientConfig) (resp *req.Response, body string, err error) {

	if strings.TrimSpace(urlStr) == "" {
		return nil, "", errors.New("urlStr is empty")
	}
	if strings.TrimSpace(postJson) == "" {
		return nil, "", errors.New("postJson is empty")
	}
	if responseStruct == nil {
		return nil, "", errors.New("response is nil")
	}
	if httpClientConfig == nil {
		return nil, "", errors.New("httpClientConfig is empty")
	}
	cfg := httpClientConfig

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		belogs.Error("PostHttpUnmarshalStructWithConfig(): Parse urlStr fail, urlStr:", urlStr, err)
		return nil, "", err
	}

	client, err := buildClientForConfig(parsedURL, cfg, "application/json")
	if err != nil {
		belogs.Error("PostHttpUnmarshalStructWithConfig(): buildClientForConfig fail", err)
		return nil, "", err
	}

	r := client.R()
	r.SetBodyJsonString(postJson)
	if cfg.ContentType != "" {
		// Request 级覆盖，优先级高于 SetBodyJsonString 的默认值
		r.SetContentType(cfg.ContentType)
	}
	r.SetSuccessResult(responseStruct) // 2xx 时自动反序列化 JSON body 到 response

	start := time.Now()
	resp, err = r.Post(urlStr)
	if err != nil {
		belogs.Error("PostHttpUnmarshalStructWithConfig(): r.Post fail", urlStr, "err", err, "time(s)", time.Since(start))
		return nil, "", err
	}
	belogs.Debug("PostHttpUnmarshalStructWithConfig(): r.Post ok", urlStr, "statusCode:", resp.GetStatusCode(),
		"time(s)", time.Since(start))
	return resp, resp.String(), nil
}
