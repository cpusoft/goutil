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

/*
	func Get(urlStr string, verifyHttps bool) (resp gorequest.Response, body string, err error) {
		return GetWithConfig(urlStr, nil)
	}
*/
func GetWithConfig(urlStr string, httpClientConfig *HttpClientConfig) (resp gorequest.Response, body string, err error) {
	if strings.HasPrefix(urlStr, "http://") {
		return GetHttpWithConfig(urlStr, httpClientConfig)
	} else if strings.HasPrefix(urlStr, "https://") {
		return GetHttpsWithConfig(urlStr, httpClientConfig)
	} else {
		return nil, "", errors.New("unknown protocol")
	}
}

/*
// Http Get Method, complete url

	func GetHttp(urlStr string) (resp gorequest.Response, body string, err error) {
		belogs.Debug("GetHttp():url:", urlStr)
		return GetHttpWithConfig(urlStr, nil)
	}
*/
func GetHttpWithConfig(urlStr string, httpClientConfig *HttpClientConfig) (resp gorequest.Response, body string, err error) {
	belogs.Debug("GetHttpWithConfig():url:", urlStr, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	url, err := url.Parse(urlStr)
	if err != nil {
		belogs.Error("GetHttpWithConfig(): Parse fail, urlStr:", urlStr, err)
		return nil, "", err
	}
	if httpClientConfig == nil {
		httpClientConfig = NewHttpClientConfig()
	}
	timeOut := time.Duration(httpClientConfig.TimeoutMins) * time.Minute
	if httpClientConfig.TimeoutMillis > 0 {
		timeOut = time.Duration(httpClientConfig.TimeoutMillis) * time.Millisecond
	}
	superAgent := gorequest.New().Get(urlStr).
		Timeout(timeOut).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(int(httpClientConfig.RetryCount), RetryIntervalSeconds*time.Second, RetryHttpStatus...)
	if httpClientConfig.ContentType != "" {
		superAgent = superAgent.Set("Content-Type", httpClientConfig.ContentType)
	}
	if httpClientConfig.Authorization != "" {
		superAgent = superAgent.Set("Authorization", httpClientConfig.Authorization)
	}
	return errorsToerror(superAgent.End())

}

/*
// Https Get Method, complete url
func GetHttps(urlStr string) (resp gorequest.Response, body string, err error) {
	return GetHttpsVerify(urlStr, false)
}
*/

/*
// Https Get Method, complete url
// verify: check https or not

	func GetHttpsVerify(urlStr string, verifyHttps bool) (resp gorequest.Response, body string, err error) {
		belogs.Debug("GetHttpsVerify():url:", urlStr, "    verifyHttps:", verifyHttps)
		httpClientConfig := NewHttpClientConfig()
		httpClientConfig.VerifyHttps = verifyHttps
		return GetHttpsWithConfig(urlStr, httpClientConfig)
	}
*/
func GetHttpsWithConfig(urlStr string, httpClientConfig *HttpClientConfig) (resp gorequest.Response, body string, err error) {
	belogs.Debug("GetHttpsWithConfig():url:", urlStr, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	url, err := url.Parse(urlStr)
	if err != nil {
		belogs.Error("GetHttpsWithConfig(): Parse fail, urlStr:", urlStr, err)
		return nil, "", err
	}
	if httpClientConfig == nil {
		httpClientConfig = NewHttpClientConfig()
	}

	config := &tls.Config{InsecureSkipVerify: !httpClientConfig.VerifyHttps}
	timeOut := time.Duration(httpClientConfig.TimeoutMins) * time.Minute
	if httpClientConfig.TimeoutMillis > 0 {
		timeOut = time.Duration(httpClientConfig.TimeoutMillis) * time.Millisecond
	}
	superAgent := gorequest.New().Get(urlStr).
		TLSClientConfig(config).
		Timeout(timeOut).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(int(httpClientConfig.RetryCount), RetryIntervalSeconds*time.Second, RetryHttpStatus...)
	if httpClientConfig.ContentType != "" {
		superAgent = superAgent.Set("Content-Type", httpClientConfig.ContentType)
	}
	if httpClientConfig.Authorization != "" {
		superAgent = superAgent.Set("Authorization", httpClientConfig.Authorization)
	}
	return errorsToerror(superAgent.End())

}

func GetHttpsResponseWithConfig(urlStr string, httpClientConfig *HttpClientConfig) (resp gorequest.Response, err error) {
	belogs.Debug("GetHttpsResponseWithConfig():url:", urlStr, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	url, err := url.Parse(urlStr)
	if err != nil {
		belogs.Error("GetHttpsResponseWithConfig(): Parse fail, urlStr:", urlStr, err)
		return nil, err
	}
	if httpClientConfig == nil {
		httpClientConfig = NewHttpClientConfig()
	}

	config := &tls.Config{InsecureSkipVerify: !httpClientConfig.VerifyHttps}
	superAgent := gorequest.New().Head(urlStr).
		TLSClientConfig(config).
		Timeout(time.Duration(httpClientConfig.TimeoutMins)*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(int(httpClientConfig.RetryCount), RetryIntervalSeconds*time.Second, RetryHttpStatus...)
	if httpClientConfig.ContentType != "" {
		superAgent = superAgent.Set("Content-Type", httpClientConfig.ContentType)
	}
	if httpClientConfig.Authorization != "" {
		superAgent = superAgent.Set("Authorization", httpClientConfig.Authorization)
	}
	resp, _, err = errorsToerror(superAgent.End())
	return resp, err
}
func GetHttpsSupportRangeWithConfig(urlStr string, httpClientConfig *HttpClientConfig) (resp gorequest.Response,
	supportRange bool, contentLength uint64, err error) {
	belogs.Debug("GetHttpsSupportRangeWithConfig():url:", urlStr, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	resp, err = GetHttpsResponseWithConfig(urlStr, httpClientConfig)
	if err != nil {
		belogs.Error("GetHttpsSupportRangeWithConfig(): GetHttpsResponseWithConfig fail, urlStr:", urlStr, err)
		return nil, false, 0, err
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		belogs.Error("GetHttpsSupportRangeWithConfig(): StatusCode is not 200, urlStr:", urlStr,
			"  statusCode:", GetStatusCode(resp))
		return nil, false, 0, errors.New("StatusCode is not 200")
	}
	acceptRanges := resp.Header.Get("Accept-Ranges")
	if acceptRanges != "bytes" {
		belogs.Debug("GetHttpsSupportRangeWithConfig(): not support, urlStr:", urlStr, "  resp.Header:", jsonutil.MarshalJson(resp.Header))
		return nil, false, 0, errors.New("Accept-Ranges is not supported")
	}
	contentLengthStr := resp.Header.Get("Content-Length")
	len, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		belogs.Error("GetHttpsSupportRangeWithConfig(): contentLengthStr is not number, urlStr:", urlStr,
			"  contentLengthStr:", contentLengthStr, err)
		return nil, false, 0, err
	}
	belogs.Debug("GetHttpsSupportRangeWithConfig(): support range, urlStr:", urlStr, "  contentLength:", len)
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

// resp: one range response, should ignore
// contentLength: all bytes len
// oneRangeLength: one range download bytes len
func GetHttpsRangeWithConfig(urlStr string, contentLength uint64,
	oneRangeLength uint64, httpClientConfig *HttpClientConfig) (resp gorequest.Response, body string, err error) {
	start := time.Now()
	belogs.Debug("GetHttpsResponseWithConfig():url:", urlStr,
		"  contentLength:", contentLength, "  oneRangeLength:", oneRangeLength,
		"  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	url, err := url.Parse(urlStr)
	if err != nil {
		belogs.Error("GetHttpsRangeWithConfig(): Parse fail, urlStr:", urlStr, err)
		return nil, "", err
	}
	if httpClientConfig == nil {
		httpClientConfig = NewHttpClientConfig()
	}

	config := &tls.Config{InsecureSkipVerify: !httpClientConfig.VerifyHttps}
	count := contentLength / oneRangeLength
	if contentLength%oneRangeLength != 0 {
		count++
	}
	belogs.Debug("GetHttpsResponseWithConfig(): get count, url:", urlStr,
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
			superAgent := gorequest.New().Get(urlStr).
				TLSClientConfig(config).
				Timeout(time.Duration(httpClientConfig.TimeoutMins)*time.Minute).
				Set("User-Agent", DefaultUserAgent).
				Set("Referrer", url.Host).
				Set("Connection", "keep-alive").
				Set("Range", rangeStrTmp).
				Retry(int(httpClientConfig.RetryCount), RetryIntervalSeconds*time.Second, RetryHttpStatus...)
			if httpClientConfig.ContentType != "" {
				superAgent = superAgent.Set("Content-Type", httpClientConfig.ContentType)
			}
			if httpClientConfig.Authorization != "" {
				superAgent = superAgent.Set("Authorization", httpClientConfig.Authorization)
			}
			resp, body, err := errorsToerror(superAgent.End())
			if err != nil {
				belogs.Error("GetHttpsRangeWithConfig(): go Get range fail, iTmp:", iTmp,
					"  urlStr:", urlStr, "  rangeStrTmp:", rangeStrTmp,
					"  statusCode:", GetStatusCode(resp), err)
				// no return
				rangeBodyCh <- rangeBody{Index: iTmp, Body: ""}
			} else {
				belogs.Debug("GetHttpsResponseWithConfig(): go Get range ok, iTmp:", iTmp,
					"  urlStr:", urlStr, " rangeStrTmp:", rangeStrTmp,
					"  contentLength:", contentLength, "  startLen:", startLen, "  endLen:", endLen,
					"  statusCode:", GetStatusCode(resp),
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
	belogs.Debug("GetHttpsResponseWithConfig(): after get all url:", urlStr,
		"  len(rangeBodyCh):", len(rangeBodyCh), "  time(s):", time.Since(start))

	rangeBodys := make([]rangeBody, 0, count)
	for b := range rangeBodyCh {
		if len(b.Body) == 0 {
			belogs.Error("GetHttpsRangeWithConfig(): get empty body in range",
				"  urlStr:", urlStr,
				"  statusCode:", GetStatusCode(resp))
			return resp, "", errors.New("get empty body in range")
		}
		rangeBodys = append(rangeBodys, b)
	}
	// sort, from newest to oldest
	sort.Sort(rangeBodySort(rangeBodys))
	var sbuilder strings.Builder
	for r := range rangeBodys {
		sbuilder.WriteString(rangeBodys[r].Body)
	}
	body = sbuilder.String()
	belogs.Debug("GetHttpsResponseWithConfig(): done get url:", urlStr,
		"  len(body):", len(body), "  time(s):", time.Since(start))
	return resp, body, err
}

func CloseResponseBody(resp gorequest.Response) {
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
}
