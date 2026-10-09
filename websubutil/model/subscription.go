package model

import (
	"time"

	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type Subscription struct {
	ID        int64         `json:"id"` // LabRpkiWebSubHub.Id
	Topic     string        `json:"topic"`
	Callback  string        `json:"callback"`
	Secret    string        `json:"secret"`
	LeaseTime time.Duration `json:"lease"`
	Expires   time.Time     `json:"expires"`
	Reason    error         `json:"-"`
}

func (c Subscription) String() string {
	m := make(map[string]string)
	m["ID"] = convert.ToString(c.ID)
	m["Topic"] = c.Topic
	m["Callback"] = c.Callback
	m["Secret"] = c.Secret
	m["LeaseTime"] = convert.ToString(c.LeaseTime)
	m["Expires"] = convert.Time2String(c.Expires)
	return jsonutil.MarshalJson(m)
}
