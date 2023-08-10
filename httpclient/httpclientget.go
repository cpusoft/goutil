package httpclient

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/netutil"
	"github.com/parnurzeal/gorequest"
)

func Get(urlStr string, verifyHttps bool) (resp gorequest.Response, body string, err error) {
	return GetWithConfig(urlStr, verifyHttps, nil)
}
func GetWithConfig(urlStr string, verifyHttps bool, httpClientConfig *HttpClientConfig) (resp gorequest.Response, body string, err error) {
	if strings.HasPrefix(urlStr, "http://") {
		return GetHttpWithConfig(urlStr, httpClientConfig)
	} else if strings.HasPrefix(urlStr, "https://") {
		return GetHttpsVerifyWithConfig(urlStr, verifyHttps, httpClientConfig)
	} else {
		return nil, "", errors.New("unknown protocol")
	}
}

// Http Get Method, complete url
func GetHttp(urlStr string) (resp gorequest.Response, body string, err error) {
	belogs.Debug("GetHttp():url:", urlStr)
	return GetHttpWithConfig(urlStr, nil)
}
func GetHttpWithConfig(urlStr string, httpClientConfig *HttpClientConfig) (resp gorequest.Response, body string, err error) {
	belogs.Debug("GetHttpWithConfig():url:", urlStr, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, "", err
	}
	if httpClientConfig == nil {
		httpClientConfig = globalHttpClientConfig
	}
	return errorsToerror(gorequest.New().Get(urlStr).
		Timeout(time.Duration(httpClientConfig.TimeoutMins)*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(int(httpClientConfig.RetryCount), RetryIntervalSeconds*time.Second, RetryHttpStatus...).
		End())

}

// Https Get Method, complete url
func GetHttps(urlStr string) (resp gorequest.Response, body string, err error) {
	return GetHttpsVerify(urlStr, false)
}

// Https Get Method, complete url
// verify: check https or not
func GetHttpsVerify(urlStr string, verify bool) (resp gorequest.Response, body string, err error) {
	belogs.Debug("GetHttpsVerify():url:", urlStr, "    verify:", verify)
	return GetHttpsVerifyWithConfig(urlStr, verify, nil)
}
func GetHttpsVerifyWithConfig(urlStr string, verify bool, httpClientConfig *HttpClientConfig) (resp gorequest.Response, body string, err error) {
	belogs.Debug("GetHttpsVerifyWithConfig():url:", urlStr, "    verify:", verify, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, "", err
	}
	if httpClientConfig == nil {
		httpClientConfig = globalHttpClientConfig
	}

	config := &tls.Config{InsecureSkipVerify: !verify}
	return errorsToerror(gorequest.New().Get(urlStr).
		TLSClientConfig(config).
		Timeout(time.Duration(httpClientConfig.TimeoutMins)*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(int(httpClientConfig.RetryCount), RetryIntervalSeconds*time.Second, RetryHttpStatus...).
		End())

}

// get by Curl
func GetByCurl(url string) (result string, err error) {
	return GetByCurlWithConfig(url, nil)
}
func GetByCurlWithConfig(url string, httpClientConfig *HttpClientConfig) (result string, err error) {
	url = strings.TrimSpace(url)
	if len(url) == 0 {
		return "", errors.New("url is emtpy")
	}
	if httpClientConfig == nil {
		httpClientConfig = globalHttpClientConfig
	}
	// mins --> seconds
	timeout := convert.ToString(httpClientConfig.TimeoutMins * 60)
	retryCount := convert.ToString(httpClientConfig.RetryCount)
	//	tmpFile := os.TempDir() + string(os.PathSeparator) + uuidutil.GetUuid()
	tmpFile, err := ioutil.TempFile("", "_tmp_") // temp file
	if err != nil {
		belogs.Error("GetByCurlWithConfig(): TempFile fail", err)
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	ipType := ""
	if httpClientConfig.IpType == "ipv4" {
		ipType = "-4"
	} else if httpClientConfig.IpType == "ipv6" {
		ipType = "-6"
	}
	belogs.Info("GetByCurlWithConfig():will curl, url:", url, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig),
		"  httpClientConfig.TimeoutMins(m):", int64(httpClientConfig.TimeoutMins), "  timeout as seconds:", timeout,
		"  retryCount:", retryCount, "  ipType:", ipType, "   tmpFile:", tmpFile.Name())

	// -s: slient mode  --no use
	// -4: ipv4  --no use
	// --connect-timeout: SECONDS  Maximum time allowed for connection
	// --ignore-content-length: Ignore the Content-Length header  --no use
	// --retry:
	// -o : output file
	// --limit-rate:  100k  --no use
	// --keepalive-time: <seconds> Interval time for keepalive probes
	// -m: --max-time SECONDS  Maximum time allowed for the transfer
	/*
		cmd := exec.Command("curl", "-4", "-v", "-o", tmpFile, url)
	*/
	// minute-->second

	start := time.Now()
	cmd := exec.Command("curl", "--connect-timeout", timeout, "--keepalive-time", timeout,
		"-m", timeout, ipType, "--retry", retryCount, "--compressed", "-v", "-o", tmpFile.Name(), url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		belogs.Error("GetByCurlWithConfig(): exec.Command fail, curl:", url, "  ipAddrs:", netutil.LookupIpByUrl(url),
			"  tmpFile:", tmpFile.Name(), "  timeout:", timeout, "  retryCount:", retryCount, "  ipType:", ipType, "  time(s):", time.Since(start), "   err: ", err,
			"  Output  is:", string(output))
		return "", errors.New("Fail to get by curl. Error is `" + err.Error() + "`. Output  is `" + string(output) + "`")
	}
	belogs.Debug("GetByCurlWithConfig(): curl ok, url:", url, "   tmpFile:", tmpFile.Name(), "  timeout:", timeout, "  time(s):", time.Since(start),
		" Output  is:", string(output))

	b, err := fileutil.ReadFileToBytes(tmpFile.Name())
	if err != nil {
		belogs.Error("GetByCurlWithConfig(): ReadFileToBytes fail, url", url, "   tmpFile:", tmpFile.Name(), "   err: ", err, "   output: "+string(output))
		return "", errors.New("Fail to get by curl. Error is `" + err.Error() + "`. Output  is `" + string(output) + "`")
	}
	belogs.Debug("GetByCurlWithConfig(): ReadFileToBytes ok, url:", url, "   tmpFile:", tmpFile.Name(), "  len(b):", len(b), "  time(s):", time.Since(start))
	return string(b), nil
}
