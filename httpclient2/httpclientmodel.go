package httpclient2

const (
	DefaultUserAgent           = "RPSTIR2/1.0"
	DefaultAllTimeout          = 300
	DefaultConnectionTimeout   = 10
	DefaultTlsHandshakeTimeout = 10
	DefaultKeepalive           = 30
	DefaultIpType              = "all"
	DefaultRangeLength         = uint64(10000000)
	DefaultRetryCount          = 3
	RetryIntervalSeconds       = 5
	// need same in ginjwt.go
	JWT_HEADER_AUTHORIZATION = "Authorization"
)

type HttpClientConfig struct {
	AllTimeoutSeconds          uint64 `json:"allTimeoutSeconds"`
	TlsHandshakeTimeoutSeconds uint64 `json:"tlsHandshakeTimeoutSeconds"`
	ConnectionTimeoutSeconds   uint64 `json:"connectionTimeoutSeconds"`
	KeepAliveSeconds           uint64 `json:"keepAliveSeconds"`

	RetryCount    uint64 `json:"retryCount"`
	IpType        string `json:"ipType"`
	VerifyHttps   bool   `json:"verifyHttps"`
	RootFile      string `json:"rootFile"`
	KeyFile       string `json:"keyFile"`
	CrtFile       string `json:"crtFile"`
	ContentType   string `json:"contentType"`
	Authorization string `json:"authorization"`
}

func NewHttpClientConfig(opts ...Option) *HttpClientConfig {
	cfg := &HttpClientConfig{
		AllTimeoutSeconds:          DefaultAllTimeout,
		ConnectionTimeoutSeconds:   DefaultConnectionTimeout,
		TlsHandshakeTimeoutSeconds: DefaultTlsHandshakeTimeout,
		KeepAliveSeconds:           DefaultKeepalive,
		RetryCount:                 DefaultRetryCount,
		IpType:                     DefaultIpType,
		VerifyHttps:                true,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

/*
cfg := NewHttpClientConfig(

	WithAllTimeoutSeconds(30),
	WithTlsHandshakeTimeoutSeconds(5),
	WithConnectionTimeoutSeconds(3),
	WithRetryCount(2),
	WithIpType("ipv4"),
	WithVerifyHttps(true),
	WithContentType("application/json"),
	WithAuthorization("Bearer xxx"),

)

resp, body, err := GetHttpWithConfig("https://httpbin.org/get", cfg)
*/
type Option func(*HttpClientConfig)

func WithAllTimeoutSeconds(v uint64) Option {
	return func(c *HttpClientConfig) { c.AllTimeoutSeconds = v }
}
func WithTlsHandshakeTimeoutSeconds(v uint64) Option {
	return func(c *HttpClientConfig) { c.TlsHandshakeTimeoutSeconds = v }
}
func WithConnectionTimeoutSeconds(v uint64) Option {
	return func(c *HttpClientConfig) { c.ConnectionTimeoutSeconds = v }
}
func WithKeepAliveSeconds(v uint64) Option {
	return func(c *HttpClientConfig) { c.KeepAliveSeconds = v }
}

func WithRetryCount(v uint64) Option {
	return func(c *HttpClientConfig) { c.RetryCount = v }
}
func WithIpType(v string) Option {
	return func(c *HttpClientConfig) { c.IpType = v }
}
func WithVerifyHttps(v bool) Option {
	return func(c *HttpClientConfig) { c.VerifyHttps = v }
}
func WithRootFile(v string) Option {
	return func(c *HttpClientConfig) { c.RootFile = v }
}
func WithKeyFile(v string) Option {
	return func(c *HttpClientConfig) { c.KeyFile = v }
}
func WithCrtFile(v string) Option {
	return func(c *HttpClientConfig) { c.CrtFile = v }
}
func WithContentType(v string) Option {
	return func(c *HttpClientConfig) { c.ContentType = v }
}
func WithAuthorization(v string) Option {
	return func(c *HttpClientConfig) { c.Authorization = v }
}

/*
cfg := (&HttpClientConfig{}).
    SetAllTimeoutSeconds(30).
    SetTlsHandshakeTimeoutSeconds(5).
    SetConnectionTimeoutSeconds(3).
    SetRetryCount(2).
    SetIpType("ipv4").
    SetVerifyHttps(true).
    SetContentType("application/json").
    SetAuthorization("Bearer xxx")

resp, body, err := GetHttpWithConfig("https://httpbin.org/get", cfg)


func (c *HttpClientConfig) SetAllTimeoutSeconds(v uint64) *HttpClientConfig {
	c.AllTimeoutSeconds = v
	return c
}

func (c *HttpClientConfig) SetTlsHandshakeTimeoutSeconds(v uint64) *HttpClientConfig {
	c.TlsHandshakeTimeoutSeconds = v
	return c
}

func (c *HttpClientConfig) SetConnectionTimeoutSeconds(v uint64) *HttpClientConfig {
	c.ConnectionTimeoutSeconds = v
	return c
}

func (c *HttpClientConfig) SetRetryCount(v uint64) *HttpClientConfig {
	c.RetryCount = v
	return c
}

func (c *HttpClientConfig) SetIpType(v string) *HttpClientConfig {
	c.IpType = v
	return c
}

func (c *HttpClientConfig) SetVerifyHttps(v bool) *HttpClientConfig {
	c.VerifyHttps = v
	return c
}

func (c *HttpClientConfig) SetRootFile(v string) *HttpClientConfig {
	c.RootFile = v
	return c
}

func (c *HttpClientConfig) SetKeyFile(v string) *HttpClientConfig {
	c.KeyFile = v
	return c
}

func (c *HttpClientConfig) SetCrtFile(v string) *HttpClientConfig {
	c.CrtFile = v
	return c
}

func (c *HttpClientConfig) SetContentType(v string) *HttpClientConfig {
	c.ContentType = v
	return c
}

func (c *HttpClientConfig) SetAuthorization(v string) *HttpClientConfig {
	c.Authorization = v
	return c
}
*/
