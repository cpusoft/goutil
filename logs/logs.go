package logs

import (
	"encoding/json"
	"fmt"
	belogs "github.com/astaxie/beego/logs"
	"os"
	"time"

	conf "github.com/cpusoft/goutil/conf"
	osutil "github.com/cpusoft/goutil/osutil"
)

/*
	LevelEmergency = iota
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInformational
	LevelDebug
*/
func init() {

	logLevel := conf.String("logs::level")
	logName := conf.String("logs::name")
	//fmt.Println("log", logLevel, logName)

	var logLevelInt int = belogs.LevelInformational
	switch logLevel {
	case "LevelEmergency":
		logLevelInt = belogs.LevelEmergency
	case "LevelAlert":
		logLevelInt = belogs.LevelAlert
	case "LevelCritical":
		logLevelInt = belogs.LevelCritical
	case "LevelError":
		logLevelInt = belogs.LevelError
	case "LevelWarning":
		logLevelInt = belogs.LevelWarning
	case "LevelNotice":
		logLevelInt = belogs.LevelNotice
	case "LevelInformational":
		logLevelInt = belogs.LevelInformational
	case "LevelDebug":
		logLevelInt = belogs.LevelDebug
	}
	ts := time.Now().Format("2006-01-02")

	logConfig := make(map[string]interface{})
	logConfig["filename"] = osutil.GetParentPath() + string(os.PathSeparator) + "log" + string(os.PathSeparator) + logName + "." + ts
	logConfig["level"] = logLevelInt
	logConfigStr, _ := json.Marshal(logConfig)
	//fmt.Println("log:logConfigStr", string(logConfigStr))
	belogs.SetLogger(belogs.AdapterFile, string(logConfigStr))

}

func LogDebugBytes(title string, buf []byte) {

	belogs.Debug(title)

	data_lines := make([]string, (len(buf)/30)+1)
	for i, b := range buf {
		data_lines[i/30] += fmt.Sprintf("%02x ", b)
	}

	for i := 0; i < len(data_lines); i++ {
		belogs.Debug(data_lines[i])
	}
}
