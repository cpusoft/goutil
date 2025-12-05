package httpclient

import (
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/parnurzeal/gorequest"
)

func DownloadUrlFile(urlFile string, localFile string) (int64, error) {
	belogs.Debug("DownloadUrlFile(): urlFile:", urlFile, "  localFile:", localFile)
	file, err := os.OpenFile(localFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		belogs.Error("DownloadUrlFile(): OpenFile fail, localFile:", localFile, err)
		return 0, err
	}
	defer func() {
		_ = file.Close()
	}()

	rsp, err := http.Get(urlFile)
	defer func() {
		if rsp != nil {
			_ = rsp.Body.Close()
		}
	}()
	if err != nil {
		belogs.Error("DownloadUrlFile(): Get fail, urlFile:", urlFile, err)
		return 0, err
	}
	n, err := io.Copy(file, rsp.Body)
	if err != nil {
		belogs.Error("DownloadUrlFile(): Copy fail, urlFile:", urlFile, "  localFile:", localFile, err)
		return 0, err
	}
	return n, err

}

// fileName: file name ; FormName:id in form
func PostFileWithConfig(urlStr string, fileName string, formName string, httpClientConfig *HttpClientConfig) (gorequest.Response, string, error) {
	if strings.HasPrefix(urlStr, "http://") {
		return PostFileHttpWithConfig(urlStr, fileName, formName, httpClientConfig)
	} else if strings.HasPrefix(urlStr, "https://") {
		return PostFileHttpsWithConfig(urlStr, fileName, formName, httpClientConfig)
	} else {
		return nil, "", errors.New("unknown protocol")
	}
}

// fileName: file name ; FormName:id in form
func PostFileHttpWithConfig(urlStr string, fileName string, formName string, httpClientConfig *HttpClientConfig) (resp gorequest.Response, body string, err error) {

	belogs.Debug("PostFileHttp():url:", urlStr, "   fileName:", fileName,
		"   formName:", formName, " httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	b, err := os.ReadFile(fileName)
	if err != nil {
		belogs.Error("PostFileHttp(): ReadFile fail, url:", urlStr, "   fileName:", fileName,
			"   err:", err)
		return nil, "", err
	}

	url, err := url.Parse(urlStr)
	if err != nil {
		belogs.Error("PostFileHttp(): Parse fail, url:", urlStr, "   fileName:", fileName, "   err:", err)
		return nil, "", err
	}

	if httpClientConfig == nil {
		httpClientConfig = NewHttpClientConfig()
	}
	timeOut := time.Duration(httpClientConfig.TimeoutMins) * time.Minute
	if httpClientConfig.TimeoutMillis > 0 {
		timeOut = time.Duration(httpClientConfig.TimeoutMillis) * time.Millisecond
	}

	fileNameStr := osutil.Base(fileName)
	belogs.Debug("PostFileHttps(): fileName:", fileName, "  fileNameStr:", fileNameStr,
		"  len(b):", len(b))
	superAgent := gorequest.New().Post(urlStr).
		Timeout(timeOut).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(int(httpClientConfig.RetryCount), RetryIntervalSeconds*time.Second, RetryHttpStatus...).
		Type("multipart").
		SendFile(fileNameStr, formName)
	if httpClientConfig.Authorization != "" {
		superAgent = superAgent.Set("Authorization", httpClientConfig.Authorization)
	}
	return errorsToerror(superAgent.End())
}

// fileName: file name ; FormName:id in form
func PostFileHttpsWithConfig(urlStr string, fileName string, formName string,
	httpClientConfig *HttpClientConfig) (resp gorequest.Response, body string, err error) {

	belogs.Debug("PostFileHttps():url:", urlStr, "   fileName:", fileName,
		"   formName:", formName, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	b, err := os.ReadFile(fileName)
	if err != nil {
		belogs.Error("PostFileHttp(): ReadFile fail, url:", urlStr, "   fileName:", fileName,
			"   err:", err)
		return nil, "", err
	}

	url, err := url.Parse(urlStr)
	if err != nil {
		belogs.Error("PostFileHttp(): Parse fail, url:", urlStr, "   fileName:", fileName,
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

	fileNameStr := osutil.Base(fileName)
	belogs.Debug("PostFileHttps(): fileName:", fileName, "  fileNameStr:", fileNameStr,
		" len(b):", len(b))
	config := &tls.Config{InsecureSkipVerify: !httpClientConfig.VerifyHttps}
	superAgent := gorequest.New().Post(urlStr).
		TLSClientConfig(config).
		Timeout(timeOut).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(int(httpClientConfig.RetryCount), RetryIntervalSeconds*time.Second, RetryHttpStatus...).
		Type("multipart").
		SendFile(fileName, formName)
	if httpClientConfig.Authorization != "" {
		superAgent = superAgent.Set("Authorization", httpClientConfig.Authorization)
	}
	return errorsToerror(superAgent.End())

}

// Deprecated: use PostFileAndUnmarshalModelWithConfig
func PostFileAndUnmarshalResponseModel(urlStr string, fileName string,
	formName string, verifyHttps bool, v interface{}) (err error) {
	return PostFileAndUnmarshalModelWithConfig(urlStr, fileName, formName,
		v, NewHttpClientConfigWithParam(5, 3, "all", verifyHttps))
}

// UnTest
// fileName: file name ; FormName:id in form
// v is ResponseModel.Data
func PostFileAndUnmarshalModelWithConfig(urlStr string, fileName string,
	formName string, v interface{}, httpClientConfig *HttpClientConfig) (err error) {
	resp, body, err := PostFileWithConfig(urlStr, fileName, formName, httpClientConfig)
	defer CloseResponseBody(resp)
	if err != nil {
		belogs.Error("PostFileAndUnmarshalModelWithConfig():PostFileWithConfig failed, urlStr:", urlStr,
			"   fileName:", fileName, "   formName:", formName,
			"   httpClientConfig:", jsonutil.MarshalJson(httpClientConfig), err)
		return err
	}

	var responseModel ResponseModel
	err = jsonutil.UnmarshalJson(body, &responseModel)
	if err != nil {
		belogs.Error("PostFileAndUnmarshalModelWithConfig():UnmarshalJson responseModel failed, urlStr:", urlStr, "  body:", body, err)
		return err
	}
	if responseModel.Result == "fail" {
		belogs.Error("PostFileAndUnmarshalModelWithConfig():responseModel.Result is fail, err:", jsonutil.MarshalJson(responseModel), body)
		return errors.New(responseModel.Msg)
	}
	if v != nil {
		// UnmarshalJson to get actual ***Response
		data := jsonutil.MarshalJson(responseModel.Data)
		err = jsonutil.UnmarshalJson(data, v)
		if err != nil {
			belogs.Error("PostFileAndUnmarshalModelWithConfig():UnmarshalJson data failed, urlStr:", urlStr, "  data:", data, err)
			return err
		}
	}
	return nil
}
