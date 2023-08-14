package httpclient

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/parnurzeal/gorequest"
)

var globalHttpClientConfig = NewHttpClientConfig()
var RetryHttpStatus = []int{http.StatusBadRequest, http.StatusInternalServerError,
	http.StatusRequestTimeout, http.StatusBadGateway, http.StatusGatewayTimeout}

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

func CloneGLobalHttpClient() *HttpClientConfig {
	c := &HttpClientConfig{
		TimeoutMins: globalHttpClientConfig.TimeoutMins,
		RetryCount:  globalHttpClientConfig.RetryCount,
		IpType:      globalHttpClientConfig.IpType,
	}
	return c
}
