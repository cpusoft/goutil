package httpclient

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/fileutil"
	osutil "github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/uuidutil"
	"github.com/parnurzeal/gorequest"
)

const (
	DefaultUserAgent     = "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36"
	DefaultTimeout       = 30
	RetryCount           = 3
	RetryIntervalSeconds = 5
)

var RetryHttpStatus = []int{http.StatusBadRequest, http.StatusInternalServerError,
	http.StatusRequestTimeout, http.StatusBadGateway, http.StatusGatewayTimeout}

func Get(urlStr string, verifyHttps bool) (resp gorequest.Response, body string, err error) {
	if strings.HasPrefix(urlStr, "http://") {
		return GetHttp(urlStr)
	} else if strings.HasPrefix(urlStr, "https://") {
		return GetHttpsVerify(urlStr, verifyHttps)
	} else {
		return nil, "", errors.New("unknown protocol")
	}
}

/*
// Http/Https Get Method,
// protocol: "http" or "https"
func Get(protocol string, address string, port int, path string) (gorequest.Response, string, error) {
	if protocol == "http" {
		return GetHttp(protocol + "://" + address + ":" + strconv.Itoa(port) + path)
	} else if protocol == "https" {
		return GetHttps(protocol + "://" + address + ":" + strconv.Itoa(port) + path)
	} else {
		return nil, "", errors.New("unknown protocol")
	}
}
*/

// Http Get Method, complete url
func GetHttp(urlStr string) (resp gorequest.Response, body string, err error) {
	belogs.Debug("GetHttp():url:", urlStr)
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, "", err
	}
	return errorsToerror(gorequest.New().Get(urlStr).
		Timeout(DefaultTimeout*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(RetryCount, RetryIntervalSeconds*time.Second, RetryHttpStatus...).
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
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, "", err
	}
	config := &tls.Config{InsecureSkipVerify: !verify}

	return errorsToerror(gorequest.New().Get(urlStr).
		TLSClientConfig(config).
		Timeout(DefaultTimeout*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(RetryCount, RetryIntervalSeconds*time.Second, RetryHttpStatus...).
		End())

}

//  http or https
func Post(urlStr string, postJson string, verifyHttps bool) (gorequest.Response, string, error) {
	if strings.HasPrefix(urlStr, "http://") {
		return PostHttp(urlStr, postJson)
	} else if strings.HasPrefix(urlStr, "https://") {
		return PostHttpsVerify(urlStr, postJson, verifyHttps)
	} else {
		return nil, "", errors.New("unknown protocol")
	}
}

/*
// Http/Https Post Method,
// protocol: "http" or "https"
func Post(protocol string, address string, port int, path string, postJson string) (gorequest.Response, string, error) {
	if protocol == "http" {
		return PostHttp(protocol+"://"+address+":"+strconv.Itoa(port)+path, postJson)
	} else if protocol == "https" {
		return PostHttps(protocol+"://"+address+":"+strconv.Itoa(port)+path, postJson)
	} else {
		return nil, "", errors.New("unknown protocol")
	}
}
*/
// Http Post Method, complete url
func PostHttp(urlStr string, postJson string) (resp gorequest.Response, body string, err error) {
	belogs.Debug("PostHttp():url:", urlStr, "    len(postJson):", len(postJson))
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, "", err
	}
	return errorsToerror(gorequest.New().Post(urlStr).
		Timeout(DefaultTimeout*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(RetryCount, RetryIntervalSeconds*time.Second, RetryHttpStatus...).
		Send(postJson).
		End())

}

// Https Post Method, complete url
func PostHttps(urlStr string, postJson string) (resp gorequest.Response, body string, err error) {
	return PostHttpsVerify(urlStr, postJson, false)

}

// Https Post Method, complete url
// verify: check https or not
func PostHttpsVerify(urlStr string, postJson string, verify bool) (resp gorequest.Response, body string, err error) {
	belogs.Debug("PostHttpsVerify():url:", urlStr, "    len(postJson):", len(postJson), "    verify:", verify)
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, "", err
	}
	config := &tls.Config{InsecureSkipVerify: !verify}
	return errorsToerror(gorequest.New().Post(urlStr).
		TLSClientConfig(config).
		Timeout(DefaultTimeout*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(RetryCount, RetryIntervalSeconds*time.Second, RetryHttpStatus...).
		Send(postJson).
		End())

}

// Http/Https Post Method,
// protocol: "http" or "https"
func PostFile(protocol string, address string, port int, path string, fileName string, formName string) (gorequest.Response, string, error) {
	if protocol == "http" {
		return PostFileHttp(protocol+"://"+address+":"+strconv.Itoa(port)+path, fileName, formName)
	} else if protocol == "https" {
		return PostFileHttps(protocol+"://"+address+":"+strconv.Itoa(port)+path, fileName, formName)
	} else {
		return nil, "", errors.New("unknown protocol")
	}
}

// fileName: file name ; FormName:id in form
func PostFileHttp(urlStr string, fileName string, formName string) (resp gorequest.Response, body string, err error) {

	belogs.Debug("PostFileHttp():url:", urlStr, "   fileName:", fileName, "   formName:", formName)
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		belogs.Error("PostFileHttp():url:", urlStr, "   fileName:", fileName, "   err:", err)
		return nil, "", err
	}

	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, "", err
	}
	file := osutil.Base(fileName)
	belogs.Debug("PostFileHttps():file:", file)
	return errorsToerror(gorequest.New().Post(urlStr).
		Timeout(DefaultTimeout*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(RetryCount, RetryIntervalSeconds*time.Second, RetryHttpStatus...).
		Type("multipart").
		SendFile(b, file).
		End())

}

// fileName: file name ; FormName:id in form
func PostFileHttps(urlStr string, fileName string, formName string) (resp gorequest.Response, body string, err error) {

	belogs.Debug("PostFileHttps():url:", urlStr, "   fileName:", fileName, "   formName:", formName)
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, "", err
	}

	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, "", err
	}
	file := osutil.Base(fileName)
	belogs.Debug("PostFileHttps():file:", file)
	config := &tls.Config{InsecureSkipVerify: true}
	return errorsToerror(gorequest.New().Post(urlStr).
		TLSClientConfig(config).
		Timeout(DefaultTimeout*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(RetryCount, RetryIntervalSeconds*time.Second, RetryHttpStatus...).
		Type("multipart").
		SendFile(b, file).
		End())

}

func GetByCurl(url string) (result string, err error) {
	belogs.Debug("GetByCurl(): cmd:  curl ", url)
	tmpFile := os.TempDir() + string(os.PathSeparator) + uuidutil.GetUuid()
	defer os.Remove(tmpFile)
	belogs.Debug("GetByCurl(): url:", url, "   tmpFile:", tmpFile)

	// -s: slient mode
	// -4: ipv4
	// --connect-timeout: connect seconds
	// --ignore-content-length: Ignore the Content-Length header
	// --retry:
	// -o : output file
	// --limit-rate:  100k
	cmd := exec.Command("curl", "-s", "-4", "--connect-timeout", "120", "--ignore-content-length",
		"--retry", "3", "--limit-rate", "1000k", "-o", tmpFile, url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		belogs.Error("GetByCurl(): exec.Command fail, curl:", url, "   tmpFile:", tmpFile, "   err: ", err, "   output: "+string(output))
		return "", errors.New("Fail to get by curl. Error is `" + err.Error() + "`. Output  is `" + string(output) + "`")
	}
	b, err := fileutil.ReadFileToBytes(tmpFile)
	if err != nil {
		belogs.Error("GetByCurl(): ReadFileToBytes fail, url", url, "   tmpFile:", tmpFile, "   err: ", err, "   output: "+string(output))
		return "", errors.New("Fail to get by curl. Error is `" + err.Error() + "`. Output  is `" + string(output) + "`")
	}
	return string(b), nil
}

// convert many erros to on error
func errorsToerror(resps gorequest.Response, bodys string, errs []error) (resp gorequest.Response, body string, err error) {
	if errs != nil && len(errs) > 0 {
		buffer := bytes.NewBufferString("")
		for _, er := range errs {
			buffer.WriteString(er.Error())
			buffer.WriteString("; ")
		}
		return resps, bodys, errors.New(buffer.String())
	}
	return resps, bodys, nil
}
