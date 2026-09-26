package model

import "time"

const (
	ModeSubscribe   = "subscribe"
	ModeUnsubscribe = "unsubscribe"
	ModeDenied      = "denied"
	ModePublish     = "publish"
)

type PublishResponse struct {
	Send               bool      `json:"send"`
	PublishTime        time.Time `json:"publishTime"`
	SerailNumbers      []uint32  `json:"serialNumbers"`
	ResponseStatusCode int       `json:"responseStatusCode"`
	ResponseTime       time.Time `json:"responseTime"`
	ResponseBody       string    `json:"responseBody"`
	ErrorMsg           string    `json:"errorMsg"`
}
