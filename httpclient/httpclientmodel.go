package httpclient

import "time"

const (
	DefaultUserAgent     = "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36 RPSTIR2"
	DefaultTimeout       = 10
	RetryCount           = 3
	RetryIntervalSeconds = 5
)

type HttpClientConfig struct {
	Timeout    time.Duration `json:"timeout"`
	RetryCount int           `json:"retryCount"`
}
