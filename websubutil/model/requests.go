package model

import (
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

// SubscribeRequest represents a form request for a subscribe.
type SubscribeRequest struct {
	Mode               string `form:"hub.mode" validate:"required"`
	Callback           string `form:"hub.callback" validate:"required"`
	CallbackRemoteAddr string `json:"callbackRemoteAddr"`
	Topic              string `form:"hub.topic" validate:"required"`
	Secret             string `form:"hub.secret" validate:"max=200"`
	LeaseSeconds       int    `form:"hub.lease_seconds" validate:""`
}

func (c SubscribeRequest) String() string {
	m := make(map[string]string)
	m["mode"] = c.Mode
	m["callback"] = c.Callback
	m["callbackRemoteAddr"] = c.CallbackRemoteAddr
	m["topic"] = c.Topic
	m["leaseSeconds"] = convert.ToString(c.LeaseSeconds)
	return jsonutil.MarshalJson(m)

}

// UnsubscribeRequest represents a form request for an unsubscribe.
type UnsubscribeRequest struct {
	Mode     string `form:"hub.mode" validate:"required"`
	Callback string `form:"hub.callback" validate:"required,url"`
	Topic    string `form:"hub.topic" validate:"required"`
}

// PublishRequest represents a form request for a publish.
type PublishRequest struct {
	Topic string `form:"hub.topic" validate:"required"`
}
