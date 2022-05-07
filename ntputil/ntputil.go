package ntputil

import (
	"errors"
	"time"

	"github.com/beevik/ntp"
	"github.com/cpusoft/goutil/belogs"
)

func GetNtpTime() (time.Time, error) {
	// https://www.ntppool.org/zone/cn
	ntpServers := []string{"0.cn.pool.ntp.org", "1.cn.pool.ntp.org", "2.cn.pool.ntp.org", "3.cn.pool.ntp.org",
		"0.pool.ntp.org", "1.pool.ntp.org", "2.pool.ntp.org", "3.pool.ntp.org"}
	for i := range ntpServers {
		tm, err := ntp.Time(ntpServers[i])
		if err != nil {
			belogs.Error("GetNtpTime(): Time fail: ", ntpServers[i], err)
			continue
		} else {
			belogs.Debug("GetNtpTime(): Time : ", ntpServers[i], tm)
			return tm, nil
		}
	}
	return time.Now(), errors.New("no ntp server is availble")
}

func GetFormatNtpTime(format string) (string, error) {
	tm, err := GetNtpTime()
	if err != nil {
		return "", err
	}
	if len(format) == 0 {
		format = "2006-01-02 15:04:05 MST"
	}
	return tm.Local().Format(format), nil
}
