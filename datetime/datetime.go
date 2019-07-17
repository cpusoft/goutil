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

//const TIME_LAYOUT = "2006-01-02 15:04:05"
//tm, e := time.Parse("060102150405Z", "190601095044Z")
func ParseTime(s string, timeLayout string) (tm time.Time, err error) {
	return time.Parse(timeLayout, s)
}
