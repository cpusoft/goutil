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

// compare==1: src==dst;;;
// compare==1: src>dst: src is after dst;;;;
// compare==-1: src<dst: src is before dst;;;
// first: compare notafter
// second: compare notbefore
func CompareTimeRange(srcNotBefore, srcNotAfter,
	dstNotBefore, dstNotAfter time.Time) (compare int) {
	if srcNotAfter.After(dstNotAfter) {
		return 1
	}
	if srcNotAfter.Before(dstNotAfter) {
		return -1
	}
	if srcNotBefore.After(dstNotBefore) {
		return 1
	}
	if srcNotBefore.Before(dstNotBefore) {
		return -1
	}
	return 0

}
