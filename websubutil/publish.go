package websubutil

import (
	"bytes"
	"crypto/hmac"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/websubutil/model"
	"github.com/jpillora/backoff"
)

func Notify(client *http.Client, job PublishJob) (publishResponse model.PublishResponse, err error) {
	b := &backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    10 * time.Minute,
		Factor: 2,
		Jitter: false,
	}

	var attempts int
	for {
		// 每次重试都重新创建 req，避免 Body 被耗尽
		req, err := http.NewRequest(http.MethodPost, job.Subscription.Callback, bytes.NewReader(job.Data))
		if err != nil {
			belogs.Error("Notify(): NewRequest fail", err)
			publishResponse.ErrorMsg = err.Error()
			return publishResponse, err
		}

		if job.Subscription.Secret != "" {
			mac := hmac.New(NewHasher(job.Hub.Hasher), []byte(job.Subscription.Secret))
			mac.Write(job.Data)
			req.Header.Set("X-Hub-Signature", job.Hub.Hasher+"="+hex.EncodeToString(mac.Sum(nil)))
		}

		req.Header.Set("Content-Type", job.ContentType)
		req.Header.Set("Link", fmt.Sprintf("<%s>; rel=\"hub\", <%s>; rel=\"self\"", job.Hub.URL, job.Subscription.Topic))
		publishResponse.PublishTime = time.Now()
		res, err := client.Do(req)
		publishResponse.ResponseTime = time.Now()
		publishResponse.ResponseStatusCode = res.StatusCode
		if err == nil {
			bodyBytes, err := io.ReadAll(res.Body)
			if err == nil {
				belogs.Error("Notify(): io.ReadAll(res.Body) fail", err)
			} else {
				publishResponse.ResponseBody = string(bodyBytes)
			}
			res.Body.Close()

			if res.StatusCode >= 200 && res.StatusCode <= 299 {
				publishResponse.Send = true
			} else if res.StatusCode == http.StatusGone {
				publishResponse.Send = false
			}
			return publishResponse, nil
		}
		belogs.Error("Notify(): client.Do fail, url:", job.Subscription.Callback, err)
		publishResponse.ErrorMsg = err.Error()
		attempts++
		if attempts >= 3 {
			break
		}
		<-time.After(b.Duration())
	}

	return publishResponse, errors.New("failed to publish after 3 attempts")
}
