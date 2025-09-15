package zaplogs

import (
	"context"
	"net/http"
	"testing"

	"go.uber.org/zap"
)

func TestZapLogs(t *testing.T) {

	simpleHttpGet("www.sogo.com")
	simpleHttpGet("http://www.sogo.com")
}

func simpleHttpGet(url string) {
	defer DeferSync()
	userInfo := make(map[string]string)
	userInfo["userId"] = "1"
	userInfo["userName"] = "userName1"
	userInfo["ownerId"] = "ownerId1"
	opInfos := make(map[string]string)
	opInfos["opLogId"] = "1"
	opInfos["opUserId"] = "opUserId1"
	cc := CustomClaims{
		// 可根据需要自行添加字段
		UserInfos: userInfo,
		TraceId:   "traceId",
		OpInfos:   opInfos,
	}

	cxt := context.WithValue(context.Background(), "CustomClaims", cc)

	DebugJw(cxt, "Trying to hit GET request for", "url", url)
	resp, err := http.Get(url)
	if err != nil {
		ErrorJw(cxt, "Error fetching URL:", zap.String("url", url), zap.Errors("err", []error{err}))
	} else {
		InfoJw(cxt, "Success! statusCode", "status", resp.Status, "url", url)
		resp.Body.Close()
	}
}
