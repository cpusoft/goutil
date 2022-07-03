package belogs

import (
	"testing"
	"time"
)

func TestFormatHeader_0(t *testing.T) {
	tm := time.Now()
	if tm.Year() >= 2100 {
		t.FailNow()
	}
	dur := time.Second
	for {
		if tm.Year() >= 2100 {
			break
		}
		h, _, _ := formatTimeHeader(tm)
		if tm.Format("2006/01/02 15:04:05.000 ") != string(h) {
			t.Log(tm)
			t.FailNow()
		}
		tm = tm.Add(dur)
		dur *= 2
	}
}

func TestFormatHeader_1(t *testing.T) {
	tm := time.Now()
	year := tm.Year()
	dur := time.Second
	for {
		if tm.Year() >= year+1 {
			break
		}
		h, _, _ := formatTimeHeader(tm)
		if tm.Format("2006/01/02 15:04:05.000 ") != string(h) {
			t.Log(tm)
			t.FailNow()
		}
		tm = tm.Add(dur)
	}
}
