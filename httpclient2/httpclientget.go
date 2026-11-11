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
// must check resp is nil
//
// first httpclient try, then curl try, then curl ipv4 try
func GetHttpWithConfig(urlStr string, httpClientConfig *HttpClientConfig) (resp *resty.Response,
	body string, err error) {
	if strings.TrimSpace(urlStr) == "" {
		belogs.Error("GetHttpWithConfig(): urlStr is empty")
		return nil, "", errors.New("urlStr is empty")
	}
	if httpClientConfig == nil {
		belogs.Error("GetHttpWithConfig(): httpClientConfig is empty")
		return nil, "", errors.New("httpClientConfig is empty")
	}
	cfg := httpClientConfig

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		belogs.Error("GetHttpWithConfig(): Parse urlStr fail, urlStr:", urlStr, err)
		return nil, "", err
	}

	belogs.Debug("GetHttpWithConfig():  parsedURL:", parsedURL, "urlStr:", urlStr,
		"cfg:", jsonutil.MarshalJson(httpClientConfig))
	client, err := buildClientForConfig(parsedURL, cfg, "")
	if err != nil {
		belogs.Error("GetHttpWithConfig(): buildClientForConfig fail, parsedURL:", parsedURL,
			"cfg:", jsonutil.MarshalJson(httpClientConfig), err)
		return nil, "", err
	}

	// first try httpclient
	start := time.Now()
	resp, err = client.R().Get(urlStr)
	//defer CloseResponseBody(resp)
	if err == nil {
		belogs.Debug("GetHttpWithConfig(): r.Get ok", urlStr, "statusCode:", GetStatusCode(resp),
			"cfg:", jsonutil.MarshalJson(httpClientConfig), "time(s)", time.Since(start))
		return resp, resp.String(), nil
	}
	// when have err, if no try curl, will return err
	if !cfg.TryCurl {
		belogs.Debug("GetHttpWithConfig(): r.Get fail, no try curl, urlStr:", urlStr, "err", err,
			"cfg:", jsonutil.MarshalJson(httpClientConfig), "time(s)", time.Since(start))
		return resp, "", err
	}

	// when have err, try curl
	belogs.Debug("GetHttpWithConfig(): r.Get fail, will try curl, urlStr:", urlStr,
		"cfg:", jsonutil.MarshalJson(httpClientConfig))
	startCurl := time.Now()
	body, err = GetByCurlWithConfig(urlStr, cfg)
	if err == nil {
		belogs.Debug("GetHttpWithConfig(): trycurl GetByCurlWithConfig ok", urlStr, "time(s)", time.Since(startCurl))
		return resp, body, nil
	}
	// if try curl fail, try curl ipv4 again
	cfg.IpType = "ipv4"
	belogs.Debug("GetHttpWithConfig(): trycurl try ipv4 again", urlStr, "cfg:", jsonutil.MarshalJson(httpClientConfig),
		"err", err, "time(s)", time.Since(startCurl))
	body, err = GetByCurlWithConfig(urlStr, cfg)
	if err == nil {
		belogs.Debug("GetHttpWithConfig(): trycurl try ipv4 ok", urlStr, "time(s)", time.Since(startCurl))
		return resp, body, nil
	}
	belogs.Error("GetHttpWithConfig(): trycurl try ipv4 fail", urlStr,
		"cfg:", jsonutil.MarshalJson(httpClientConfig), "time(s)", time.Since(startCurl))
	return resp, "", err

}
