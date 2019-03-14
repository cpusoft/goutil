package datetime

import (
	"time"
)

func Now() string {
	return ToString(time.Now())
}

func ToString(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
