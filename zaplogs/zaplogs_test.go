package zaplogs

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

func TestZapLogs(t *testing.T) {

	simpleHttpGet("www.sogo.com")
	simpleHttpGet("http://www.sogo.com")
}

func simpleHttpGet(url string) {
	start := time.Now()
	defer DeferSync()
	infos := make(map[string]interface{})
	infos["userId"] = "1"
	infos["userName"] = "userName1"
	infos["ownerId"] = "ownerId1"
	infos["opLogId"] = "1"
	infos["opUserId"] = "opUserId1"
	infos["traceId"] = "traceId1"
	cc := CustomClaims{
		Infos: infos,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)), //过期时间
			NotBefore: jwt.NewNumericDate(time.Now()),                    //生效时间（立即生效）
			IssuedAt:  jwt.NewNumericDate(time.Now()),                    //签发时间
		},
	}

	ctx := context.WithValue(context.Background(), JWT_CTX_CustomClaims_Infos, cc)

	DebugArgs(ctx, "Trying to hit GET request for", "url", url, "NOW")
	resp, err := http.Get(url)
	if err != nil {
		ErrorFields(ctx, "Error fetching URL:", zap.String("url", url), zap.Errors("err", []error{err}))
	} else {
		InfoArgs(ctx, "Success! statusCode", "status", resp.Status, "url", url, "time", time.Since(start))
		resp.Body.Close()
	}
}
