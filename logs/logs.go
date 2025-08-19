package logs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	logs "github.com/cpusoft/goutil/belogs"
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
	//async := conf.DefaultBool("logs::async", false)
	//fmt.Println("logLevel:", logLevel, "  logName:", logName, "  async:", async)

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
		fmt.Println("found " + path + " failed, " + err.Error())
	}
	filePath := path + string(os.PathSeparator) + logName
	//fmt.Println("log file is ", filePath)

	logConfig := make(map[string]interface{})
	logConfig["daily"] = true
	logConfig["hourly"] = false
	logConfig["filename"] = filePath // + "." + ts
	logConfig["maxlines"] = 0
	logConfig["maxfiles"] = 0
	logConfig["maxsize"] = 0
	logConfig["maxdays"] = 30
	logConfig["maxhours"] = 0
	logConfig["level"] = logLevelInt

	logConfigStr, _ := json.Marshal(logConfig)
	//fmt.Println("log:logConfigStr", string(logConfigStr))
	//logs.NewLogger(1024)
	//AdapterFile
	err = logs.SetLogger(logs.AdapterFile, string(logConfigStr))
	if err != nil {
		fmt.Println(filePath + " SetLogger failed, " + err.Error() + ",   " + string(logConfigStr))
	}
	logs.GetBeeLogger().DelLogger("console")
	//if async {
	//	logs.Async()
	//}

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
