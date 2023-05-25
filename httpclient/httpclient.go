package httpclient

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/httpserver"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/netutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/uuidutil"
	"github.com/parnurzeal/gorequest"
)

var httpClientConfig = NewHttpClientConfig()

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
		Timeout(httpClientConfig.Timeout*time.Minute).
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
		Timeout(httpClientConfig.Timeout*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(RetryCount, RetryIntervalSeconds*time.Second, RetryHttpStatus...).
		End())

}

// http or https
func Post(urlStr string, postJson string, verifyHttps bool) (gorequest.Response, string, error) {
	if strings.HasPrefix(urlStr, "http://") {
		return PostHttp(urlStr, postJson)
	} else if strings.HasPrefix(urlStr, "https://") {
		return PostHttps(urlStr, postJson, verifyHttps)
	} else {
		return nil, "", errors.New("unknown protocol")
	}
}

// response need HttpResponse{}
func PostAndUnmarshalResponse(urlStr, postJson string, verifyHttps bool, response interface{}) (err error) {
	belogs.Debug("PostAndUnmarshalResponse(): urlStr:", urlStr, "   postJson:", postJson,
		"   verifyHttps:", verifyHttps, "    response:", reflect.TypeOf(response).Name())
	resp, body, err := Post(urlStr, postJson, verifyHttps)
	if err != nil {
		belogs.Error("PostAndUnmarshalResponse():Post failed, urlStr:", urlStr, "   postJson:", postJson, err)
		return err
	}
	resp.Body.Close()

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

// response is any struct
func PostAndUnmarshalStruct(urlStr, postJson string, verifyHttps bool, response interface{}) (err error) {
	belogs.Debug("PostAndUnmarshalStruct(): urlStr:", urlStr, "   postJson:", postJson,
		"   verifyHttps:", verifyHttps, "    response:", reflect.TypeOf(response).Name())
	resp, body, err := Post(urlStr, postJson, verifyHttps)
	if err != nil {
		belogs.Error("PostAndUnmarshalStruct():Post failed, urlStr:", urlStr, "   postJson:", postJson, err)
		return err
	}
	resp.Body.Close()

	// UnmarshalJson to get actual ***Response
	err = jsonutil.UnmarshalJson(body, response)
	if err != nil {
		belogs.Error("PostAndUnmarshalStruct():UnmarshalJson failed, urlStr:", urlStr, "  body:", body, err)
		return err
	}
	return nil
}

type ResponseModel struct {
	Result string      `json:"result"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data,omitempty"`
}

// v is ResponseModel.Data
func PostAndUnmarshalResponseModel(urlStr, postJson string, verifyHttps bool, v interface{}) (err error) {
	belogs.Debug("PostAndUnmarshalResponseModel(): urlStr:", urlStr, "   postJson:", postJson,
		"   verifyHttps:", verifyHttps)
	resp, body, err := Post(urlStr, postJson, verifyHttps)
	if err != nil {
		belogs.Error("PostAndUnmarshalResponseModel():Post failed, urlStr:", urlStr, "   postJson:", postJson, err)
		return err
	}
	if resp != nil {
		resp.Body.Close()
	}
	belogs.Debug("PostAndUnmarshalResponseModel(): len(body):", len(body))
	belogs.Debug("PostAndUnmarshalResponseModel(): body:", body)

	var responseModel ResponseModel
	err = jsonutil.UnmarshalJson(body, &responseModel)
	if err != nil {
		belogs.Error("PostAndUnmarshalResponseModel():UnmarshalJson responseModel failed, urlStr:", urlStr, "  body:", body, err)
		return err
	}
	belogs.Debug("PostAndUnmarshalResponseModel():get response, urlStr:", urlStr, "   postJson:", postJson,
		" responseModel:", jsonutil.MarshalJson(responseModel))
	if responseModel.Result == "fail" {
		belogs.Error("PostAndUnmarshalResponseModel():responseModel.Result is fail, err:", jsonutil.MarshalJson(responseModel), body)
		return errors.New(responseModel.Msg)
	}

	if v != nil {
		belogs.Debug("PostAndUnmarshalResponseModel(): v:", reflect.TypeOf(v).Name(), "  len(body):", len(body))
		// UnmarshalJson to get actual ***Response
		data := jsonutil.MarshalJson(responseModel.Data)
		err = jsonutil.UnmarshalJson(data, v)
		if err != nil {
			belogs.Error("PostAndUnmarshalResponseModel():UnmarshalJson data failed, urlStr:", urlStr, "  data:", data, err)
			return err
		}
	}
	return nil
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
		Timeout(httpClientConfig.Timeout*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(RetryCount, RetryIntervalSeconds*time.Second, RetryHttpStatus...).
		Send(postJson).
		End())

}

/*
// Https Post Method, complete url
func PostHttps(urlStr string, postJson string) (resp gorequest.Response, body string, err error) {
	return PostHttpsVerify(urlStr, postJson, false)
}
*/

// Https Post Method, complete url
// verify: check https or not
func PostHttps(urlStr string, postJson string, verify bool) (resp gorequest.Response, body string, err error) {
	belogs.Debug("PostHttps():url:", urlStr, "    len(postJson):", len(postJson), "    verify:", verify)
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, "", err
	}
	config := &tls.Config{InsecureSkipVerify: !verify}
	return errorsToerror(gorequest.New().Post(urlStr).
		TLSClientConfig(config).
		Timeout(httpClientConfig.Timeout*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(RetryCount, RetryIntervalSeconds*time.Second, RetryHttpStatus...).
		Send(postJson).
		End())

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

/*
// Http/Https Post Method,
// protocol: "http" or "https"
func PostFile(protocol string, address string, port int, path string, fileName string, formName string, verify bool) (gorequest.Response, string, error) {
	if protocol == "http" {
		return PostFileHttp(protocol+"://"+address+":"+strconv.Itoa(port)+path, fileName, formName)
	} else if protocol == "https" {
		return PostFileHttps(protocol+"://"+address+":"+strconv.Itoa(port)+path, fileName, formName)
	} else {
		return nil, "", errors.New("unknown protocol")
	}
}
*/

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
	fileNameStr := osutil.Base(fileName)
	belogs.Debug("PostFileHttps():fileNameStr:", fileNameStr)
	return errorsToerror(gorequest.New().Post(urlStr).
		Timeout(httpClientConfig.Timeout*time.Minute).
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
	b, err := ioutil.ReadFile(fileName)
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
		Timeout(httpClientConfig.Timeout*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Set("Connection", "keep-alive").
		Retry(RetryCount, RetryIntervalSeconds*time.Second, RetryHttpStatus...).
		Type("multipart").
		SendFile(b, fileNameStr, formName, true).
		End())

}

func GetByCurl(url string) (result string, err error) {
	if len(url) == 0 {
		return "", errors.New("url is emtpy")
	}

	url = strings.TrimSpace(url)
	belogs.Debug("GetByCurl(): cmd:  curl:'" + url + "'")
	tmpFile := os.TempDir() + string(os.PathSeparator) + uuidutil.GetUuid()
	defer os.Remove(tmpFile)
	belogs.Debug("GetByCurl():will curl url:", url, "   tmpFile:", tmpFile)

	// -s: slient mode  --no use
	// -4: ipv4  --no use
	// --connect-timeout: SECONDS  Maximum time allowed for connection
	// --ignore-content-length: Ignore the Content-Length header  --no use
	// --retry:
	// -o : output file
	// --limit-rate:  100k  --no use
	// -m: --max-time SECONDS  Maximum time allowed for the transfer
	/*
		cmd := exec.Command("curl", "-4", "-v", "-o", tmpFile, url)
	*/
	start := time.Now()
	cmd := exec.Command("curl", "--connect-timeout", "600",
		"-m", "600", "--retry", "3", "-4", "--compressed", "-v", "-o", tmpFile, url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		belogs.Error("GetByCurl(): exec.Command fail, curl:", url, "  ipAddrs:", netutil.LookupIpByUrl(url),
			"  tmpFile:", tmpFile, "  time(s):", time.Since(start), "   err: ", err,
			"  Output  is:", string(output))
		return "", errors.New("Fail to get by curl. Error is `" + err.Error() + "`. Output  is `" + string(output) + "`")
	}
	belogs.Debug("GetByCurl(): curl ok, url:", url, "   tmpFile:", tmpFile, "  time(s):", time.Since(start),
		" Output  is:", string(output))

	b, err := fileutil.ReadFileToBytes(tmpFile)
	if err != nil {
		belogs.Error("GetByCurl(): ReadFileToBytes fail, url", url, "   tmpFile:", tmpFile, "   err: ", err, "   output: "+string(output))
		return "", errors.New("Fail to get by curl. Error is `" + err.Error() + "`. Output  is `" + string(output) + "`")
	}
	belogs.Debug("GetByCurl(): ReadFileToBytes ok, url:", url, "   tmpFile:", tmpFile, "  len(b):", len(b), "  time(s):", time.Since(start))
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

func NewHttpClientConfig() *HttpClientConfig {
	httpClientConfig := new(HttpClientConfig)
	httpClientConfig.Timeout = time.Duration(DefaultTimeout)
	httpClientConfig.RetryCount = RetryCount
	return httpClientConfig
}

// Minutes
func SetTimeout(minute uint64) {
	if minute > 0 {
		httpClientConfig.Timeout = time.Duration(minute)
	}
}
func ResetTimeout() {
	httpClientConfig.Timeout = time.Duration(DefaultTimeout)
}

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
