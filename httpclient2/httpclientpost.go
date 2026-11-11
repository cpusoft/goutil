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

// should use: defer CloseResponseBody(resp)
//
// postJson can be empty
//
// resp may be nil, should check
//
// first httpclient try, then curl try, then curl ipv4 try
func PostHttpWithConfig(urlStr string, postJson string,
	httpClientConfig *HttpClientConfig) (resp *resty.Response, body string, err error) {

	if strings.TrimSpace(urlStr) == "" {
		belogs.Error("PostHttpWithConfig(): urlStr is empty")
		return nil, "", errors.New("urlStr is empty")
	}
	//if strings.TrimSpace(postJson) == "" {
	//	return nil, "", errors.New("postJson is empty")
	//}
	if httpClientConfig == nil {
		belogs.Error("PostHttpWithConfig(): httpClientConfig is empty")
		return nil, "", errors.New("httpClientConfig is empty")
	}
	cfg := httpClientConfig

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		belogs.Error("PostHttpWithConfig(): Parse urlStr fail, urlStr:", urlStr,
			"cfg:", jsonutil.MarshalJson(httpClientConfig), err)
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
	// first try httpclient
	start := time.Now()
	resp, err = r.Post(urlStr)
	//defer CloseResponseBody(resp)
	if err == nil {
		belogs.Debug("PostHttpWithConfig(): r.Post ok", urlStr, "statusCode:", resp.StatusCode(),
			"time(s)", time.Since(start))
		return resp, resp.String(), nil
	}

	// when have err, if no try curl, will return err
	if !cfg.TryCurl {
		belogs.Debug("PostHttpWithConfig(): r.Post fail, no try curl, urlStr:", urlStr, "err", err,
			"cfg:", jsonutil.MarshalJson(httpClientConfig), "time(s)", time.Since(start))
		return resp, "", err
	}

	// when have err, try curl
	belogs.Debug("PostHttpWithConfig(): r.Post fail, will try curl, urlStr:", urlStr,
		"cfg:", jsonutil.MarshalJson(httpClientConfig))
	startCurl := time.Now()
	body, err = PostByCurlWithConfig(urlStr, postJson, cfg)
	if err == nil {
		belogs.Debug("PostHttpWithConfig(): trycurl PostByCurlWithConfig ok", urlStr,
			"time(s)", time.Since(startCurl))
		return resp, body, nil
	}

	// if try curl fail, try curl ipv4 again
	cfg.IpType = "ipv4"
	belogs.Debug("PostHttpWithConfig(): trycurl try ipv4 again", urlStr, "cfg:", jsonutil.MarshalJson(httpClientConfig),
		"err", err, "time(s)", time.Since(startCurl))
	body, err = PostByCurlWithConfig(urlStr, postJson, cfg)
	if err == nil {
		belogs.Debug("PostHttpWithConfig(): trycurl PostByCurlWithConfig ok", urlStr,
			"time(s)", time.Since(startCurl))
		return resp, body, nil
	}
	belogs.Error("PostHttpWithConfig(): trycurl try ipv4 fail", urlStr,
		"cfg:", jsonutil.MarshalJson(httpClientConfig), "time(s)", time.Since(startCurl))
	return resp, "", err

}

// should use: defer CloseResponseBody(resp)
//
// postJson can be empty
//
// responseStruct must be a pointer to struct, and not nil
//
// resp may be nil, should check
//
// first httpclient try, then curl try, then curl ipv4 try
func PostHttpUnmarshalStructWithConfig(urlStr, postJson string, responseStruct interface{},
	httpClientConfig *HttpClientConfig) (resp *resty.Response, body string, err error) {

	if responseStruct == nil {
		belogs.Error("PostHttpUnmarshalStructWithConfig(): responseStruct is empty")
		return nil, "", errors.New("response is nil")
	}
	belogs.Debug("PostHttpUnmarshalStructWithConfig(): urlStr:", urlStr, "postJson:", postJson,
		"cfg:", jsonutil.MarshalJson(httpClientConfig))
	start := time.Now()
	resp, body, err = PostHttpWithConfig(urlStr, postJson, httpClientConfig)
	if err != nil {
		belogs.Error("PostHttpUnmarshalStructWithConfig(): r.Post fail", urlStr,
			"postJson:", postJson, "body", body,
			"cfg:", jsonutil.MarshalJson(httpClientConfig), "err", err,
			"time(s)", time.Since(start))
		return resp, "", err
	}
	if strings.TrimSpace(body) == "" {
		belogs.Error("PostHttpUnmarshalStructWithConfig(): body is empty", urlStr,
			"postJson:", postJson,
			"cfg:", jsonutil.MarshalJson(httpClientConfig),
			"err", "body is empty, cannot unmarshal to struct",
			"time(s)", time.Since(start))
		return resp, body, errors.New("body is empty, cannot unmarshal to struct")
	}

	jsonErr := jsonutil.UnmarshalJson(body, responseStruct)
	if jsonErr != nil {
		belogs.Error("PostHttpUnmarshalStructWithConfig(): trycurl UnmarshalJson fail", urlStr,
			"err", jsonErr, "body", body, "time(s)", time.Since(start))
		return resp, body, jsonErr
	}
	belogs.Debug("PostHttpUnmarshalStructWithConfig(): trycurl and unmarshal ok", urlStr, "postJson:", postJson,
		"cfg:", jsonutil.MarshalJson(httpClientConfig), "time(s)", time.Since(start))
	return resp, body, nil

}
