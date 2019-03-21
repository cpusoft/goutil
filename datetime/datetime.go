package datetime

import (
	"time"
)

// return YYYY-MM-DD HH24:MS:SS Now
func Now() string {
	return ToString(time.Now())
}

// return YYYY-MM-DD HH24:MS:SS
func ToString(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
