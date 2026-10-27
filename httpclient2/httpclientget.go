package httpclient2

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/imroc/req/v3"
)

// should use: defer CloseResponseBody(resp)
func GetHttpWithConfig(urlStr string, httpClientConfig *HttpClientConfig) (resp *req.Response, body string, err error) {
	if strings.TrimSpace(urlStr) == "" {
		return nil, "", errors.New("urlStr is empty")
	}
	if httpClientConfig == nil {
		return nil, "", errors.New("httpClientConfig is empty")
	}
	cfg := httpClientConfig

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		belogs.Error("GetHttpWithConfig(): Parse urlStr fail, urlStr:", urlStr, err)
		return nil, "", err
	}

	client, err := buildClientForConfig(parsedURL, cfg, "")
	if err != nil {
		belogs.Error("GetHttpWithConfig(): buildClientForConfig fail", err)
		return nil, "", err
	}

	start := time.Now()
	resp, err = client.R().Get(urlStr)
	if err != nil {
		belogs.Error("GetHttpWithConfig(): r.Get fail", urlStr, "err", err, "time(s)", time.Since(start))
		return nil, "", err
	}
	belogs.Debug("GetHttpWithConfig(): r.Get ok", urlStr, "statusCode:", resp.GetStatusCode(),
		"time(s)", time.Since(start))
	return resp, resp.String(), nil
}
