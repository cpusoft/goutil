package ntputil

import (
	"errors"
	"time"

	"github.com/beevik/ntp"
	"github.com/cpusoft/goutil/belogs"
)

// GetNtpTime 从多个NTP服务器获取网络时间（仅单次尝试，依赖ntp库内置超时）
// 核心功能：返回第一个可用的NTP时间，全部失败则返回本地时间和错误
func GetNtpTime() (time.Time, error) {
	// 国内及通用NTP服务器池（与原代码一致）
	ntpServers := []string{"0.cn.pool.ntp.org", "1.cn.pool.ntp.org", "2.cn.pool.ntp.org", "3.cn.pool.ntp.org",
		"0.pool.ntp.org", "1.pool.ntp.org", "2.pool.ntp.org", "3.pool.ntp.org"}

	// 遍历每个NTP服务器，仅单次尝试（移除重试）
	for _, server := range ntpServers {
		tm, err := ntp.Time(server) // 依赖ntp库内置超时，无手动超时/重试
		if err != nil {
			belogs.Error("GetNtpTime(): Time fail: ", server, err)
			continue
		}
		belogs.Debug("GetNtpTime(): Time : ", server, tm)
		return tm, nil
	}

	// 修正拼写错误：availble → available
	return time.Now(), errors.New("no ntp server is available")
}

// GetFormatNtpTime 获取格式化的NTP时间（核心逻辑与原代码完全一致）
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
