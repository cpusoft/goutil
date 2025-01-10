package httpclient

const (
	DefaultUserAgent     = "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36 RPSTIR2"
	DefaultTimeoutMins   = 10
	DefaultIpType        = "all"
	DefaultRangeLength   = uint64(10000000)
	RetryCount           = 3
	RetryIntervalSeconds = 5
)

type HttpClientConfig struct {
	TimeoutMins   uint64 `json:"timeoutMins"`
	TimeoutMillis uint64 `json:"timeoutMillis"`
	RetryCount    uint64 `json:"retryCount"`
	// all/ipv4/ipv6
	IpType string `json:"ipType"`
	// range
	RangeLength uint64 `json:"rangeLength"`
	VerifyHttps bool   `json:"verifyHttps"`
}

/*
// Minutes

	func SetTimeout(timeoutMins uint64) {
		if timeoutMins > 0 {
			globalHttpClientConfig.TimeoutMins = timeoutMins
		}
	}

	func ResetTimeout() {
		globalHttpClientConfig.TimeoutMins = uint64(DefaultTimeoutMins)
	}
*/
func NewHttpClientConfigWithParam(timeoutMins uint64,
	retryCount uint64, ipType string, verifyHttps bool) *HttpClientConfig {
	c := &HttpClientConfig{
		TimeoutMins: timeoutMins,
		RetryCount:  retryCount,
		IpType:      ipType,
		RangeLength: DefaultRangeLength,
		VerifyHttps: verifyHttps,
	}
	return c
}

func NewHttpClientConfig() *HttpClientConfig {
	httpClientConfig := new(HttpClientConfig)
	httpClientConfig.TimeoutMins = uint64(DefaultTimeoutMins)
	httpClientConfig.RetryCount = RetryCount
	httpClientConfig.IpType = DefaultIpType
	httpClientConfig.RangeLength = DefaultRangeLength
	httpClientConfig.VerifyHttps = true
	return httpClientConfig
}
