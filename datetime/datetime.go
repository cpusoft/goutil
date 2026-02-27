package datetime

import (
	"errors"
	"strconv"
	"time"

	timeconv "github.com/Andrew-M-C/go.timeconv"
	"github.com/cpusoft/goutil/belogs"
)

// 定义常用时间格式常量，避免硬编码
const (
	TimeLayoutDefault = "2006-01-02 15:04:05" // 标准时间格式 YYYY-MM-DD HH24:MI:SS
)

// Now 返回当前时间的标准格式字符串（YYYY-MM-DD HH24:MI:SS）
func Now() string {
	return ToString(time.Now())
}

// ToString 将time.Time转换为标准格式字符串（YYYY-MM-DD HH24:MI:SS）
// 入参为零值时间时返回空字符串，避免返回无效的0001-01-01 00:00:00
func ToString(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(TimeLayoutDefault)
}

// ParseTime 解析字符串为time.Time，封装错误日志输出
// timeLayout: 时间格式，推荐使用本包的TimeLayoutDefault常量
// s: 待解析的时间字符串
func ParseTime(s string, timeLayout string) (tm time.Time, err error) {
	if s == "" {
		err = errors.New("time string is empty")
		belogs.Error("ParseTime(): empty input", err)
		return
	}
	if timeLayout == "" {
		err = errors.New("time layout is empty")
		belogs.Error("ParseTime(): empty layout", err)
		return
	}
	tm, err = time.Parse(timeLayout, s)
	if err != nil {
		belogs.Error("ParseTime(): parse fail, s:", s, " layout:", timeLayout, err)
	}
	return
}

/* close

// CompareTimeRange 比较两个时间范围的关系
// 返回值说明：
//
//	1: src时间范围整体在dst之后
//
// -1: src时间范围整体在dst之前
//
//	0: 两个时间范围有重叠（包含完全相等）
//
// 参数说明：
//
//	srcNotBefore: 源时间范围起始点
//	srcNotAfter: 源时间范围结束点
//	dstNotBefore: 目标时间范围起始点
//	dstNotAfter: 目标时间范围结束点
func CompareTimeRange(srcNotBefore, srcNotAfter, dstNotBefore, dstNotAfter time.Time) int {
	// 先校验入参合法性（起始时间不能晚于结束时间）
	if srcNotBefore.After(srcNotAfter) {
		belogs.Info("CompareTimeRange(): src time range is invalid, notBefore after notAfter")
		return 0
	}
	if dstNotBefore.After(dstNotAfter) {
		belogs.Info("CompareTimeRange(): dst time range is invalid, notBefore after notAfter")
		return 0
	}

	// src整体在dst之后：src的起始点 > dst的结束点
	if srcNotBefore.After(dstNotAfter) {
		return 1
	}
	// src整体在dst之前：src的结束点 < dst的起始点
	if srcNotAfter.Before(dstNotBefore) {
		return -1
	}
	// 其余情况均为有重叠
	return 0
}
*/
// AddDateByDuration 按指定时长（年/月/日）增减时间
// duration格式：1y（1年）、-2m（减2月）、3d（3天），仅支持y/m/d后缀
// 返回值：增减后的时间，错误信息（入参非法时返回）
func AddDateByDuration(t time.Time, duration string) (newT time.Time, err error) {
	// 初始化返回值为原时间，避免错误时返回零值
	newT = t

	// 校验入参
	if duration == "" {
		err = errors.New("duration is empty")
		belogs.Error("AddDateByDuration(): empty duration", err)
		return
	}
	if t.IsZero() {
		err = errors.New("input time is zero value")
		belogs.Error("AddDateByDuration(): zero time input", err)
		return
	}

	belogs.Debug("AddDateByDuration(): t:", t, "  duration:", duration)

	var years, months, days int
	// 提取数字部分和单位
	var numStr, unit string
	for i, c := range duration {
		if (c >= '0' && c <= '9') || c == '-' {
			numStr += string(c)
		} else {
			unit = duration[i:]
			break
		}
	}

	// 校验单位合法性
	if unit == "" {
		err = errors.New("duration missing unit (y/m/d)")
		belogs.Error("AddDateByDuration(): no unit in duration, duration:", duration, err)
		return
	}
	if unit != "y" && unit != "m" && unit != "d" {
		err = errors.New("invalid unit: " + unit + ", only y/m/d are supported")
		belogs.Error("AddDateByDuration(): invalid unit", err)
		return
	}

	// 解析数字
	num, err := strconv.Atoi(numStr)
	if err != nil {
		err = errors.New("parse number fail: " + err.Error())
		belogs.Error("AddDateByDuration(): atoi fail, duration:", duration, err)
		return
	}

	// 赋值对应字段
	switch unit {
	case "y":
		years = num
	case "m":
		months = num
	case "d":
		days = num
	}

	belogs.Debug("AddDateByDuration(): years:", years, "  months:", months, "  days:", days, "  duration:", duration)
	newT = timeconv.AddDate(t, years, months, days)
	belogs.Debug("AddDateByDuration(): ok, newT:", newT, "   t:", t, "  duration:", duration)
	return
}
