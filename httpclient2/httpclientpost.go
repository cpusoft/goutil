package httpclient2

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/go-resty/resty/v2"
)

// ---------------------------------------------------------------
// 1) 纯提交 POST JSON
// should use: defer CloseResponseBody(resp)
// ---------------------------------------------------------------
func PostHttpWithConfig(urlStr string, postJson string,
	httpClientConfig *HttpClientConfig) (resp *resty.Response, body string, err error) {

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
	r.SetBody(postJson)
	if cfg.ContentType != "" {
		r.SetHeader("Content-Type", cfg.ContentType)
	}
	start := time.Now()
	resp, err = r.Post(urlStr)
	if err != nil {
		belogs.Error("PostHttpWithConfig(): r.Post fail", urlStr, "err", err, "time(s)", time.Since(start))
		if cfg.TryCurl {
			belogs.Debug("PostHttpWithConfig(): r.Post fail, will try curl", urlStr)
			startCurl := time.Now()

			// 修改：使用 = 替代 :=，防止变量遮蔽
			body, err = PostByCurlWithConfig(urlStr, postJson, cfg)
			if err != nil {
				belogs.Error("PostHttpWithConfig(): trycurl PostByCurlWithConfig fail", urlStr, "err", err,
					"time(s)", time.Since(startCurl))
				return nil, "", err
			}
			belogs.Debug("PostHttpWithConfig(): trycurl PostByCurlWithConfig ok", urlStr, "time(s)", time.Since(startCurl))
			return nil, body, nil
		}
		return nil, "", err
	}
	belogs.Debug("PostHttpWithConfig(): r.Post ok", urlStr, "statusCode:", resp.StatusCode(),
		"time(s)", time.Since(start))
	return resp, resp.String(), nil
}

// ---------------------------------------------------------------
// 2) POST JSON + 自动反序列化响应到 struct
// should use: defer CloseResponseBody(resp)
// ---------------------------------------------------------------
func PostHttpUnmarshalStructWithConfig(urlStr, postJson string, responseStruct interface{},
	httpClientConfig *HttpClientConfig) (resp *resty.Response, body string, err error) {

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
	r.SetBody(postJson)
	if cfg.ContentType != "" {
		r.SetHeader("Content-Type", cfg.ContentType)
	}
	r.SetResult(responseStruct)

	start := time.Now()
	resp, err = r.Post(urlStr)
	if err != nil {
		belogs.Error("PostHttpUnmarshalStructWithConfig(): r.Post fail", urlStr, "err", err, "time(s)", time.Since(start))
		if cfg.TryCurl {
			belogs.Debug("PostHttpUnmarshalStructWithConfig(): r.Post fail, will try curl", urlStr)
			startCurl := time.Now()

			// 修改：使用 = 替代 :=
			body, err = PostByCurlWithConfig(urlStr, postJson, cfg)
			if err != nil {
				belogs.Error("PostHttpUnmarshalStructWithConfig(): trycurl PostByCurlWithConfig fail", urlStr, "err", err,
					"time(s)", time.Since(startCurl))
				return nil, "", err
			}

			jsonErr := jsonutil.UnmarshalJson(body, responseStruct)
			if jsonErr != nil {
				belogs.Error("PostHttpUnmarshalStructWithConfig(): trycurl UnmarshalJson fail", urlStr,
					"err", jsonErr, "body", body, "time(s)", time.Since(startCurl))
				return nil, body, jsonErr
			}
			belogs.Debug("PostHttpUnmarshalStructWithConfig(): trycurl and unmarshal ok", urlStr,
				"time(s)", time.Since(startCurl))
			return nil, body, nil
		}
		return nil, "", err
	}
	belogs.Debug("PostHttpUnmarshalStructWithConfig(): r.Post ok", urlStr, "statusCode:", resp.StatusCode(),
		"time(s)", time.Since(start))
	return resp, resp.String(), nil
}
