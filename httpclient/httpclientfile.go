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
func PostFile(urlStr string, fileName string, formName string, verifyHttps bool) (gorequest.Response, string, error) {
	if strings.HasPrefix(urlStr, "http://") {
		return PostFileHttp(urlStr, fileName, formName)
	} else if strings.HasPrefix(urlStr, "https://") {
		return PostFileHttps(urlStr, fileName, formName, verifyHttps)
	} else {
		return nil, "", errors.New("unknown protocol")
	}
}

// fileName: file name ; FormName:id in form
func PostFileHttp(urlStr string, fileName string, formName string) (resp gorequest.Response, body string, err error) {

	belogs.Debug("PostFileHttp():url:", urlStr, "   fileName:", fileName, "   formName:", formName)
	b, err := os.ReadFile(fileName)
	if err != nil {
		belogs.Error("PostFileHttp():url:", urlStr, "   fileName:", fileName, "   err:", err)
		return nil, "", err
	}

	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, "", err
	}
	fileNameStr := osutil.Base(fileName)
	belogs.Debug("PostFileHttps():fileNameStr:", fileNameStr)
	return errorsToerror(gorequest.New().Post(urlStr).
		Timeout(time.Duration(globalHttpClientConfig.TimeoutMins)*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(RetryCount, RetryIntervalSeconds*time.Second, RetryHttpStatus...).
		Type("multipart").
		SendFile(b, fileNameStr, formName, true).
		End())

}

// fileName: file name ; FormName:id in form
func PostFileHttps(urlStr string, fileName string, formName string, verify bool) (resp gorequest.Response, body string, err error) {

	belogs.Debug("PostFileHttps():url:", urlStr, "   fileName:", fileName, "   formName:", formName, "  verify:", verify)
	b, err := os.ReadFile(fileName)
	if err != nil {
		return nil, "", err
	}

	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, "", err
	}
	fileNameStr := osutil.Base(fileName)
	belogs.Debug("PostFileHttps():fileNameStr:", fileNameStr)
	config := &tls.Config{InsecureSkipVerify: !verify}
	return errorsToerror(gorequest.New().Post(urlStr).
		TLSClientConfig(config).
		Timeout(time.Duration(globalHttpClientConfig.TimeoutMins)*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(RetryCount, RetryIntervalSeconds*time.Second, RetryHttpStatus...).
		Type("multipart").
		SendFile(b, fileNameStr, formName, true).
		End())

}

// UnTest
// fileName: file name ; FormName:id in form
// v is ResponseModel.Data
func PostFileAndUnmarshalResponseModel(urlStr string, fileName string,
	formName string, verifyHttps bool, v interface{}) (err error) {
	resp, body, err := PostFile(urlStr, fileName, formName, verifyHttps)
	if err != nil {
		belogs.Error("PostFileAndUnmarshalResponseModel():PostFile failed, urlStr:", urlStr,
			"   fileName:", fileName, "   formName:", formName, "   verifyHttps:", verifyHttps, err)
		return err
	}
	if resp != nil {
		resp.Body.Close()
	}

	var responseModel ResponseModel
	err = jsonutil.UnmarshalJson(body, &responseModel)
	if err != nil {
		belogs.Error("PostFileAndUnmarshalResponseModel():UnmarshalJson responseModel failed, urlStr:", urlStr, "  body:", body, err)
		return err
	}
	if responseModel.Result == "fail" {
		belogs.Error("PostFileAndUnmarshalResponseModel():responseModel.Result is fail, err:", jsonutil.MarshalJson(responseModel), body)
		return errors.New(responseModel.Msg)
	}
	if v != nil {
		// UnmarshalJson to get actual ***Response
		data := jsonutil.MarshalJson(responseModel.Data)
		err = jsonutil.UnmarshalJson(data, v)
		if err != nil {
			belogs.Error("PostFileAndUnmarshalResponseModel():UnmarshalJson data failed, urlStr:", urlStr, "  data:", data, err)
			return err
		}
	}
	return nil
}
