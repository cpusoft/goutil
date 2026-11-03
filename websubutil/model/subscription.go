package model

import (
	"time"

	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type Subscription struct {
	ID                 int64         `json:"id"` // LabRpkiWebSubHub.Id
	Topic              string        `json:"topic"`
	Callback           string        `json:"callback"`
	CallbackRemoteAddr string        `json:"callbackRemoteAddr"` // IP:Port
	Secret             string        `json:"secret"`
	LeaseTime          time.Duration `json:"lease"`
	Expires            time.Time     `json:"expires"`
	Reason             error         `json:"-"`
}

func (c Subscription) String() string {
	m := make(map[string]string)
	m["id"] = convert.ToString(c.ID)
	m["topic"] = c.Topic
	m["callback"] = c.Callback
	m["callbackRemoteAddr"] = c.CallbackRemoteAddr
	m["leaseTime"] = convert.ToString(c.LeaseTime)
	m["expires"] = convert.Time2String(c.Expires)
	return jsonutil.MarshalJson(m)
}
