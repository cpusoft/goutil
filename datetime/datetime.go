package datetime

import (
	"strconv"
	"strings"
	"time"

	"github.com/Andrew-M-C/go.timeconv"
	"github.com/cpusoft/goutil/belogs"
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

// duration: 1y/1m/1d or -1y/-1m/-1d
func AddDataByDuration(t time.Time, duration string) (newT time.Time, err error) {
	belogs.Debug("AddDataByDuration(): t:", t, "  duration:", duration)
	var years, months, days int
	if strings.HasSuffix(duration, "y") {
		y := strings.TrimSuffix(duration, "y")
		years, err = strconv.Atoi(y)
	} else if strings.HasSuffix(duration, "m") {
		m := strings.TrimSuffix(duration, "m")
		months, err = strconv.Atoi(m)
	} else if strings.HasSuffix(duration, "d") {
		d := strings.TrimSuffix(duration, "d")
		days, err = strconv.Atoi(d)
	}
	if err != nil {
		belogs.Error("AddDataByDuration(): atoi fail, duration:", duration, err)
		return newT, err
	}
	belogs.Debug("AddDataByDuration(): years:", years, "  months:", months, "  days:", days, "  duration:", duration)
	newT = timeconv.AddDate(t, years, months, days)
	belogs.Debug("AddDataByDuration(): ok, newT:", newT, "   t:", t, "  duration:", duration)
	return newT, nil
}
