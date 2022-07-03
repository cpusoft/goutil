package asn1node

import (
	"errors"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
)

/*

YYMMDDhhmmZ
YYMMDDhhmm+hhmm
YYMMDDhhmm-hhmm
YYMMDDhhmmssZ
YYMMDDhhmmss+hhmm
YYMMDDhhmmss-hhmm

*/

func EncodeUTCTime(t time.Time) ([]byte, error) {
	data := make([]byte, 0, 17)
	return AppendUTCTime(data, t)
}

func DecodeUTCTime(data []byte) (time.Time, error) {
	t, _, err := ParseUTCTime(data)
	return t, err
}

func AppendUTCTime(data []byte, t time.Time) ([]byte, error) {

	year, month, day := t.Date()

	shortYear := yearCollapse(year)
	if shortYear == -1 {
		belogs.Error("AppendUTCTime(): shortYear==01, bad convert time to UTCTime: invalid year(%d)", year)
		return nil, errors.New("shortYear is wrong")
	}
	data = appendTwoDigits(data, shortYear)
	data = appendTwoDigits(data, int(month))
	data = appendTwoDigits(data, day)

	hour, min, sec := t.Clock()

	data = appendTwoDigits(data, hour)
	data = appendTwoDigits(data, min)
	data = appendTwoDigits(data, sec)

	_, offset := t.Zone()
	offsetMinutes := offset / 60

	switch {
	case offsetMinutes == 0:
		return append(data, 'Z'), nil
	case offsetMinutes > 0:
		data = append(data, '+')
	case offsetMinutes < 0:
		data = append(data, '-')
	}

	if offsetMinutes < 0 {
		offsetMinutes = -offsetMinutes
	}

	data = appendTwoDigits(data, offsetMinutes/60) // hours
	data = appendTwoDigits(data, offsetMinutes%60) // mins

	return data, nil
}

func ParseUTCTime(data []byte) (time.Time, []byte, error) {

	var err error

	ds := make([]int, 6)
	data, err = parseTwoDigitsSeries(data, ds)
	if err != nil {
		return time.Time{}, nil, err
	}

	var (
		shortYear = ds[0]
		month     = time.Month(ds[1])
		day       = ds[2]
	)

	year := yearExpand(shortYear)

	var (
		hour = ds[3]
		min  = ds[4]
		sec  = ds[5]
	)

	if len(data) < 1 {
		return time.Time{}, nil, errors.New("parse UTCTime: insufficient data length")
	}

	b := data[0]
	data = data[1:]

	var negative bool
	switch b {
	case 'Z':
		t := time.Date(year, month, day, hour, min, sec, 0, time.UTC)
		return t, data, nil
	case '-':
		negative = true
	case '+':
		negative = false
	default:
		belogs.Error("ParseUTCTime(): parse UTCTime: invalid character: %q", b)
		return time.Time{}, nil, errors.New("parse UTCTime: invalid character")
	}

	ds = make([]int, 2)
	data, err = parseTwoDigitsSeries(data, ds)
	if err != nil {
		return time.Time{}, nil, err
	}
	offsetMinutes := int(ds[0])*60 + int(ds[1])
	if negative {
		offsetMinutes = -offsetMinutes
	}

	const timeInLocal = true
	if timeInLocal {
		t := time.Date(year, month, day, hour, min, sec, 0, time.UTC)
		t = t.Add(time.Minute * time.Duration(-offsetMinutes))
		t = t.In(time.Local)
		return t, data, nil
	}

	loc := time.FixedZone("", offsetMinutes*60)
	t := time.Date(year, month, day, hour, min, sec, 0, loc)
	return t, data, nil
}

func appendTwoDigits(data []byte, x int) []byte {
	var (
		lo = '0' + byte(x%10)
		hi = '0' + byte((x/10)%10)
	)
	return append(data, hi, lo)
}

func parseTwoDigits(data []byte) (int, []byte) {
	if len(data) < 2 {
		return -1, data
	}

	hi, ok := convert.ByteToDigit(data[0])
	if !ok {
		return -1, data
	}

	lo, ok := convert.ByteToDigit(data[1])
	if !ok {
		return -1, data
	}

	x := hi*10 + lo

	return x, data[2:]
}

func parseTwoDigitsSeries(data []byte, ds []int) ([]byte, error) {
	var d int
	for i := range ds {
		d, data = parseTwoDigits(data)
		if d == -1 {
			return nil, errors.New("invalid series of digits")
		}
		ds[i] = d
	}
	return data, nil
}

// year: [0..99]
func yearExpand(year int) int {
	if inInterval(year, 0, 50) {
		return year + 2000
	}
	if inInterval(year, 50, 100) {
		return year + 1900
	}
	return -1
}

// year: [1950..2049]
func yearCollapse(year int) int {
	if inInterval(year, 2000, 2050) {
		return year - 2000
	}
	if inInterval(year, 1950, 2000) {
		return year - 1900
	}
	return -1
}

// Value a is in [min..max)
func inInterval(a int, min, max int) bool {
	return (min <= a) && (a < max)
}

// Generalized Long Year, 4 nums
func ParseGeneralizedTime(data []byte) (ret time.Time, err error) {
	if len(data) < 15 {
		return time.Now(), errors.New("parseGeneralizedTime fail")
	}
	belogs.Debug("ParseGeneralizedTime():", convert.PrintBytesOneLine(data))

	year := string(data[0:4])
	month := string(data[4:6])
	day := string(data[6:8])
	hour := string(data[8:10])
	minute := string(data[10:12])
	second := string(data[12:14])
	z := string(data[14])
	tm := year + "-" + month + "-" + day + " " + hour + ":" + minute + ":" + second + z
	ret, err = time.Parse("2006-01-02 15:04:05Z", tm)
	if err != nil {
		belogs.Error("ParseGeneralizedTime(): parseGeneralizedTime fail:", err)
		return time.Now(), errors.New("parseGeneralizedTime fail")
	}
	return ret, nil
}
