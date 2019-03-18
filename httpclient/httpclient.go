package httpclient

import (
	"bytes"
	"crypto/tls"
	"errors"
	"net/url"
	"strconv"
	"time"

	belogs "github.com/astaxie/beego/logs"
	"github.com/parnurzeal/gorequest"
)

const (
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36"
	DefaultTimeout   = 30
)

func Get(protocol string, address string, port int, path string) (gorequest.Response, string, error) {
	if protocol == "http" {
		return GetHttp(protocol + "://" + address + ":" + strconv.Itoa(port) + path)
	} else if protocol == "https" {
		return GetHttps(protocol + "://" + address + ":" + strconv.Itoa(port) + path)
	} else {
		return nil, "", errors.New("unknown protocol")
	}
}

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
		End())

}
func GetHttps(urlStr string) (resp gorequest.Response, body string, err error) {
	belogs.Debug("GetHttps():url:", urlStr)
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, "", err
	}
	config := &tls.Config{InsecureSkipVerify: true}

	return errorsToerror(gorequest.New().Get(urlStr).
		TLSClientConfig(config).
		Timeout(DefaultTimeout*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		End())

}

func Post(protocol string, address string, port int, path string, postJson string) (gorequest.Response, string, error) {
	if protocol == "http" {
		return PostHttp(protocol+"://"+address+":"+strconv.Itoa(port)+path, postJson)
	} else if protocol == "https" {
		return PostHttps(protocol+"://"+address+":"+strconv.Itoa(port)+path, postJson)
	} else {
		return nil, "", errors.New("unknown protocol")
	}
}

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
		Send(postJson).
		End())

}
func PostHttps(urlStr string, postJson string) (resp gorequest.Response, body string, err error) {
	belogs.Debug("PostHttps():url:", urlStr, "    len(postJson):", len(postJson))
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, "", err
	}
	config := &tls.Config{InsecureSkipVerify: true}
	return errorsToerror(gorequest.New().Post(urlStr).
		TLSClientConfig(config).
		Timeout(DefaultTimeout*time.Minute).
		Set("User-Agent", DefaultUserAgent).
		Set("Referrer", url.Host).
		Send(postJson).
		End())

}
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
