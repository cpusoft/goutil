package httpclient

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
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
		belogs.Error("GetHttpWithConfig(): Parse fail, urlStr:", urlStr, err)
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
		belogs.Error("GetHttpsVerifyWithConfig(): Parse fail, urlStr:", urlStr, err)
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

func GetHttpsVerifyResponseWithConfig(urlStr string, verify bool, httpClientConfig *HttpClientConfig) (resp gorequest.Response, err error) {
	belogs.Debug("GetHttpsVerifyResponseWithConfig():url:", urlStr, "    verify:", verify, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	url, err := url.Parse(urlStr)
	if err != nil {
		belogs.Error("GetHttpsVerifyResponseWithConfig(): Parse fail, urlStr:", urlStr, err)
		return nil, err
	}
	if httpClientConfig == nil {
		httpClientConfig = globalHttpClientConfig
	}

	config := &tls.Config{InsecureSkipVerify: !verify}
	resp, _, err = errorsToerror(gorequest.New().Head(urlStr).
		TLSClientConfig(config).
		Timeout(time.Duration(httpClientConfig.TimeoutMins)*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(int(httpClientConfig.RetryCount), RetryIntervalSeconds*time.Second, RetryHttpStatus...).
		End())
	return resp, err
}
func GetHttpsVerifySupportRangeWithConfig(urlStr string, verify bool, httpClientConfig *HttpClientConfig) (resp gorequest.Response,
	supportRange bool, contentLength uint64, err error) {
	belogs.Debug("GetHttpsVerifySupportRangeWithConfig():url:", urlStr, "    verify:", verify, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	resp, err = GetHttpsVerifyResponseWithConfig(urlStr, verify, httpClientConfig)
	if err != nil {
		belogs.Error("GetHttpsVerifySupportRangeWithConfig(): GetHttpsVerifyResponseWithConfig fail, urlStr:", urlStr, err)
		return nil, false, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		belogs.Error("GetHttpsVerifySupportRangeWithConfig(): StatusCode is not 200, urlStr:", urlStr, "  resp.StatusCode:", resp.StatusCode)
		return nil, false, 0, errors.New("StatusCode is not 200")
	}
	acceptRanges := resp.Header.Get("Accept-Ranges")
	if acceptRanges != "bytes" {
		belogs.Debug("GetHttpsVerifySupportRangeWithConfig(): not support, urlStr:", urlStr, "  resp.Header:", jsonutil.MarshalJson(resp.Header))
		return nil, false, 0, errors.New("Accept-Ranges is not supported")
	}
	contentLengthStr := resp.Header.Get("Content-Length")
	len, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		belogs.Error("GetHttpsVerifySupportRangeWithConfig(): contentLengthStr is not number, urlStr:", urlStr,
			"  contentLengthStr:", contentLengthStr, err)
		return nil, false, 0, err
	}
	belogs.Debug("GetHttpsVerifySupportRangeWithConfig(): support range, urlStr:", urlStr, "  contentLength:", len)
	return resp, true, uint64(len), nil
}

type rangeBody struct {
	Index uint64
	Body  string
}
type rangeBodySort []rangeBody

func (v rangeBodySort) Len() int {
	return len(v)
}

func (v rangeBodySort) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// default comparison.
func (v rangeBodySort) Less(i, j int) bool {
	return v[i].Index < v[j].Index
}

// contentLength: all bytes len
// oneRangeLength: one range download bytes len
func GetHttpsVerifyRangeWithConfig(urlStr string, contentLength uint64,
	oneRangeLength uint64, verify bool,
	httpClientConfig *HttpClientConfig) (resp gorequest.Response, body string, err error) {
	start := time.Now()
	belogs.Debug("GetHttpsVerifyResponseWithConfig():url:", urlStr,
		"    contentLength:", contentLength, "  oneRangeLength:", oneRangeLength,
		"    verify:", verify, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	url, err := url.Parse(urlStr)
	if err != nil {
		belogs.Error("GetHttpsVerifyRangeWithConfig(): Parse fail, urlStr:", urlStr, err)
		return nil, "", err
	}
	if httpClientConfig == nil {
		httpClientConfig = globalHttpClientConfig
	}

	config := &tls.Config{InsecureSkipVerify: !verify}
	count := contentLength / oneRangeLength
	if contentLength%oneRangeLength != 0 {
		count++
	}
	belogs.Debug("GetHttpsVerifyResponseWithConfig(): get count, url:", urlStr,
		"    contentLength:", contentLength, "  oneRangeLength:", oneRangeLength,
		"    count:", count)
	var wg sync.WaitGroup
	rangeBodyCh := make(chan rangeBody, count)
	for i := uint64(0); i < count; i++ {
		startLen := i * oneRangeLength
		endLen := (i+1)*oneRangeLength - 1
		if endLen > contentLength {
			endLen = contentLength
		}
		rangeStr := "bytes=" + convert.ToString(startLen) + "-" + convert.ToString(endLen)
		wg.Add(1)
		go func(rangeStrTmp string, iTmp uint64, rangeBodyCh chan rangeBody, wg *sync.WaitGroup) {
			defer wg.Done()
			resp, body, err = errorsToerror(gorequest.New().Get(urlStr).
				TLSClientConfig(config).
				Timeout(time.Duration(httpClientConfig.TimeoutMins)*time.Minute).
				Set("User-Agent", DefaultUserAgent).
				Set("Referrer", url.Host).
				Set("Connection", "keep-alive").
				Set("Range", rangeStrTmp).
				Retry(int(httpClientConfig.RetryCount), RetryIntervalSeconds*time.Second, RetryHttpStatus...).
				End())
			if err != nil {
				belogs.Error("GetHttpsVerifyRangeWithConfig(): go Get fail, iTmp:", iTmp, "  urlStr:", urlStr, err)
				// no return
				rangeBodyCh <- rangeBody{}
			} else {
				belogs.Debug("GetHttpsVerifyResponseWithConfig(): go Get, iTmp:", iTmp,
					"  url:", urlStr,
					"  contentLength:", contentLength,
					"  startLen:", startLen, "  endLen:", endLen, " rangeStrTmp:", rangeStrTmp,
					"  len(body):", len(body), "  time(s):", time.Since(start))
				rangeBodyCh <- rangeBody{
					Index: iTmp,
					Body:  body,
				}
			}
		}(rangeStr, i, rangeBodyCh, &wg)
	}
	wg.Wait()
	close(rangeBodyCh)
	belogs.Debug("GetHttpsVerifyResponseWithConfig(): after get all url:", urlStr,
		"  len(rangeBodyCh):", len(rangeBodyCh), "  time(s):", time.Since(start))

	rangeBodys := make([]rangeBody, 0, count)
	for b := range rangeBodyCh {
		rangeBodys = append(rangeBodys, b)
	}
	// sort, from newest to oldest
	sort.Sort(rangeBodySort(rangeBodys))
	var sbuilder strings.Builder
	for r := range rangeBodys {
		sbuilder.WriteString(rangeBodys[r].Body)
	}
	body = sbuilder.String()
	belogs.Debug("GetHttpsVerifyResponseWithConfig(): done get url:", urlStr,
		"  len(body):", len(body), "  time(s):", time.Since(start))
	return resp, body, err
}
