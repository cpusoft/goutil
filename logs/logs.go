package logs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	logs "github.com/beego/beego/v2/core/logs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/osutil"
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
	// get process file name as log name
	logName := filepath.Base(os.Args[0])
	if logName != "" {
		logName = strings.Split(logName, ".")[0] + ".log"
	} else {
		logName = conf.String("logs::name")
	}
	async := conf.DefaultBool("logs::async", false)
	//fmt.Println("log", logLevel, logName)

	var logLevelInt int = logs.LevelInformational
	switch logLevel {
	case "LevelEmergency":
		logLevelInt = logs.LevelEmergency
	case "LevelAlert":
		logLevelInt = logs.LevelAlert
	case "LevelCritical":
		logLevelInt = logs.LevelCritical
	case "LevelError":
		logLevelInt = logs.LevelError
	case "LevelWarning":
		logLevelInt = logs.LevelWarning
	case "LevelNotice":
		logLevelInt = logs.LevelNotice
	case "LevelInformational":
		logLevelInt = logs.LevelInformational
	case "LevelDebug":
		logLevelInt = logs.LevelDebug
	}
	//ts := time.Now().Format("2006-01-02")

	//
	path, err := osutil.GetCurrentOrParentAbsolutePath("log")
	if err != nil {
		panic("found " + path + " failed, " + err.Error())
	}
	filePath := path + string(os.PathSeparator) + logName
	fmt.Println("log file is ", filePath)

	logConfig := make(map[string]interface{})
	logConfig["filename"] = filePath // + "." + ts
	logConfig["level"] = logLevelInt
	// no max lines
	logConfig["maxlines"] = 0
	logConfig["maxsize"] = 0
	logConfig["daily"] = true
	logConfig["maxdays"] = 30

	logConfigStr, _ := json.Marshal(logConfig)
	fmt.Println("log:logConfigStr", string(logConfigStr))
	logs.NewLogger(1000000)
	logs.SetLogger(logs.AdapterFile, string(logConfigStr))
	if async {
		logs.Async()
	}

}

func LogDebugBytes(title string, buf []byte) {

	logs.Debug(title)

	dataLines := make([]string, (len(buf)/30)+1)
	for i, b := range buf {
		dataLines[i/30] += fmt.Sprintf("%02x ", b)
	}

	for i := 0; i < len(dataLines); i++ {
		logs.Debug(dataLines[i])
	}
}
