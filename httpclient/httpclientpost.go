package httpclient

import (
	"crypto/tls"
	"errors"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/parnurzeal/gorequest"
)

// http or https
func Post(urlStr string, postJson string, verifyHttps bool) (gorequest.Response, string, error) {
	return PostWithConfig(urlStr, postJson,
		NewHttpClientConfigWithParam(5, 3, "all", verifyHttps))
}
func PostWithConfig(urlStr string, postJson string, httpClientConfig *HttpClientConfig) (gorequest.Response, string, error) {
	if strings.HasPrefix(urlStr, "http://") {
		return PostHttpWithConfig(urlStr, postJson, httpClientConfig)
	} else if strings.HasPrefix(urlStr, "https://") {
		return PostHttpsWithConfig(urlStr, postJson, httpClientConfig)
	} else {
		return nil, "", errors.New("unknown protocol")
	}
}

/*
// Http Post Method, complete url
func PostHttp(urlStr string, postJson string) (resp gorequest.Response, body string, err error) {
	belogs.Debug("PostHttp():url:", urlStr, "    len(postJson):", len(postJson))
	return PostHttpWithConfig(urlStr, postJson, nil)
}
*/

func PostHttpWithConfig(urlStr string, postJson string, httpClientConfig *HttpClientConfig) (resp gorequest.Response, body string, err error) {
	belogs.Debug("PostHttpWithConfig():url:", urlStr, "    len(postJson):", len(postJson),
		"  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	url, err := url.Parse(urlStr)
	if err != nil {
		belogs.Error("PostHttpWithConfig(): Parse fail, url:", urlStr, "   postJson:", postJson,
			"   err:", err)
		return nil, "", err
	}
	if httpClientConfig == nil {
		httpClientConfig = NewHttpClientConfig()
	}
	timeOut := time.Duration(httpClientConfig.TimeoutMins) * time.Minute
	if httpClientConfig.TimeoutMillis > 0 {
		timeOut = time.Duration(httpClientConfig.TimeoutMillis) * time.Millisecond
	}
	superAgent := gorequest.New().Post(urlStr).
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
	return errorsToerror(superAgent.Send(postJson).End())
}

/*
// Https Post Method, complete url
// verify: check https or not

	func PostHttps(urlStr string, postJson string, verify bool) (resp gorequest.Response, body string, err error) {
		belogs.Debug("PostHttps():url:", urlStr, "    len(postJson):", len(postJson), "    verify:", verify)
		return PostHttpsWithConfig(urlStr, postJson, verify, nil)
	}
*/
func PostHttpsWithConfig(urlStr string, postJson string,
	httpClientConfig *HttpClientConfig) (resp gorequest.Response, body string, err error) {
	belogs.Debug("PostHttpsWithConfig():url:", urlStr, "    len(postJson):", len(postJson),
		"  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	url, err := url.Parse(urlStr)
	if err != nil {
		belogs.Error("PostHttpsWithConfig(): Parse fail, url:", urlStr, "   postJson:", postJson,
			"   err:", err)
		return nil, "", err
	}
	if httpClientConfig == nil {
		httpClientConfig = NewHttpClientConfig()
	}

	timeOut := time.Duration(httpClientConfig.TimeoutMins) * time.Minute
	if httpClientConfig.TimeoutMillis > 0 {
		timeOut = time.Duration(httpClientConfig.TimeoutMillis) * time.Millisecond
	}

	config := &tls.Config{InsecureSkipVerify: !httpClientConfig.VerifyHttps}
	superAgent := gorequest.New().Post(urlStr).
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
	return errorsToerror(superAgent.Send(postJson).End())
}

// v is ResponseModel.Data
type ResponseModel struct {
	Result string      `json:"result"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data,omitempty"`
}

/*
	func PostAndUnmarshalResponseModel(urlStr, postJson string, verifyHttps bool, v interface{}) (err error) {
		belogs.Debug("PostAndUnmarshalResponseModel(): urlStr:", urlStr, "   postJson:", postJson,
			"   verifyHttps:", verifyHttps)
		return PostAndUnmarshalResponseModelWithConfig(urlStr, postJson, verifyHttps, v, nil)
	}
*/
func PostAndUnmarshalResponseModelWithConfig(urlStr, postJson string, v interface{}, httpClientConfig *HttpClientConfig) (err error) {
	belogs.Debug("PostAndUnmarshalResponseModel(): urlStr:", urlStr, "   postJson:", postJson,
		"   httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))

	resp, body, err := PostWithConfig(urlStr, postJson, httpClientConfig)
	defer CloseResponseBody(resp)
	if err != nil {
		belogs.Error("PostAndUnmarshalResponseModelWithConfig():PostWithConfig failed, urlStr:", urlStr,
			"   postJson:", postJson, err)
		return err
	}
	belogs.Debug("PostAndUnmarshalResponseModelWithConfig(): len(body):", len(body))
	belogs.Debug("PostAndUnmarshalResponseModelWithConfig(): body:", body)

	var responseModel ResponseModel
	err = jsonutil.UnmarshalJson(body, &responseModel)
	if err != nil {
		belogs.Error("PostAndUnmarshalResponseModelWithConfig():UnmarshalJson responseModel failed, urlStr:", urlStr, "  body:", body, err)
		return err
	}
	belogs.Debug("PostAndUnmarshalResponseModelWithConfig():get response, urlStr:", urlStr, "   postJson:", postJson,
		" responseModel:", jsonutil.MarshalJson(responseModel))
	if responseModel.Result == "fail" {
		belogs.Error("PostAndUnmarshalResponseModelWithConfig():responseModel.Result is fail, err:", jsonutil.MarshalJson(responseModel), body)
		return errors.New(responseModel.Msg)
	}

	if v != nil {
		// UnmarshalJson to get actual ***Response
		data := jsonutil.MarshalJson(responseModel.Data)
		belogs.Debug("PostAndUnmarshalResponseModelWithConfig(): v:", reflect.TypeOf(v).Name(), "  len(body):", len(body), "  data:", data)
		err = jsonutil.UnmarshalJson(data, v)
		if err != nil {
			belogs.Error("PostAndUnmarshalResponseModelWithConfig():UnmarshalJson data failed, urlStr:", urlStr, "  data:", data, err)
			return err
		}
	}
	return nil
}

// response is any struct
func PostAndUnmarshalStructWithConfig(urlStr, postJson string, response interface{}, httpClientConfig *HttpClientConfig) (err error) {
	belogs.Debug("PostAndUnmarshalStructWithConfig(): urlStr:", urlStr, "   postJson:", postJson,
		"    response:", reflect.TypeOf(response).Name(), "   httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	resp, body, err := PostWithConfig(urlStr, postJson, httpClientConfig)
	defer CloseResponseBody(resp)
	if err != nil {
		belogs.Error("PostAndUnmarshalStructWithConfig():Post failed, urlStr:", urlStr, "   postJson:", postJson, err)
		return err
	}

	// UnmarshalJson to get actual ***Response
	err = jsonutil.UnmarshalJson(body, response)
	if err != nil {
		belogs.Error("PostAndUnmarshalStructWithConfig():UnmarshalJson failed, urlStr:", urlStr, "  body:", body, err)
		return err
	}
	return nil
}

/*
/////////////////////////////////////////////////////
//Deprecated
func PostAndUnmarshalResponse(urlStr, postJson string, verifyHttps bool, response interface{}) (err error) {
	belogs.Debug("PostAndUnmarshalResponse(): urlStr:", urlStr, "   postJson:", postJson,
		"   verifyHttps:", verifyHttps, "    response:", reflect.TypeOf(response).Name())
	resp, body, err := Post(urlStr, postJson, verifyHttps)
	defer CloseResponseBody(resp)
	if err != nil {
		belogs.Error("PostAndUnmarshalResponse():Post failed, urlStr:", urlStr, "   postJson:", postJson, err)
		return err
	}

	// try UnmarshalJson using HttpResponse to get Result
	var httpResponse httpserver.HttpResponse
	err = jsonutil.UnmarshalJson(body, &httpResponse)
	if err != nil {
		belogs.Error("PostAndUnmarshalResponse():UnmarshalJson failed, urlStr:", urlStr, "  body:", body, err)
		return err
	}
	if httpResponse.Result == "fail" {
		belogs.Error("PostAndUnmarshalResponse():httpResponse.Result is fail, err:", jsonutil.MarshalJson(httpResponse), body)
		return errors.New(httpResponse.Msg)
	}
	// UnmarshalJson to get actual ***Response
	err = jsonutil.UnmarshalJson(body, response)
	if err != nil {
		belogs.Error("PostAndUnmarshalResponse():UnmarshalJson failed, urlStr:", urlStr, "  body:", body, err)
		return err
	}
	return nil
}
*/
