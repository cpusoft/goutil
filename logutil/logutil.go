package logutil

import (
	"fmt"

	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/osutil"
)

// fileExtNoDot: "log" or "json"
func GetLogPathName(fileExtNoDot string) (logPathName string) {
	logName := conf.DefaultString("logs::name", "logs.+"+fileExtNoDot)

	var currentPath string
	var err error
	logPath := conf.String("logs::dir")
	if logPath == "" {
		logPath, currentPath, err = osutil.GetConfOrLogPath("log")
		if err != nil {
			fmt.Println("GetConfOrLogPath failed, " + err.Error())
		}
		if logPath == "" {
			fmt.Println("found logpath failed, use currentPath:", currentPath)
			logPath = currentPath
		}
	}
	logPathName = logPath + logName
	fmt.Println("logPathName:" + logPathName)
	return logPathName
}
