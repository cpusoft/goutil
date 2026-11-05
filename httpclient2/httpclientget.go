package httpclient2

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/go-resty/resty/v2"
)

// should use: defer CloseResponseBody(resp)
func GetHttpWithConfig(urlStr string, httpClientConfig *HttpClientConfig) (resp *resty.Response,
	body string, err error) {
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
		if cfg.TryCurl {
			belogs.Debug("GetHttpWithConfig(): r.Get fail, will try curl, urlStr:", urlStr)
			startCurl := time.Now()
			body, err = GetByCurlWithConfig(urlStr, cfg)
			if err != nil {
				belogs.Error("GetHttpWithConfig(): trycurl GetByCurlWithConfig fail", urlStr, "err", err,
					"time(s)", time.Since(startCurl))
				return nil, "", err
			}
			belogs.Debug("GetHttpWithConfig(): trycurl GetByCurlWithConfig ok", urlStr, "time(s)", time.Since(startCurl))
			return nil, body, nil
		}
		return nil, "", err
	}
	belogs.Debug("GetHttpWithConfig(): r.Get ok", urlStr, "statusCode:", GetStatusCode(resp),
		"time(s)", time.Since(start))
	return resp, resp.String(), nil
}
