package belogs

import (
	"fmt"
	"path"
	"time"
)

type LogMsg struct {
	Level               int
	Msg                 string
	When                time.Time
	FilePath            string
	LineNumber          int
	Args                []interface{}
	Prefix              string
	enableFullFilePath  bool
	enableFuncCallDepth bool
}

// OldStyleFormat you should never invoke this
func (lm *LogMsg) OldStyleFormat() string {
	msg := lm.Msg

	if len(lm.Args) > 0 {
		msg = fmt.Sprintf(lm.Msg, lm.Args...)
	}

	msg = lm.Prefix + " " + msg

	if lm.enableFuncCallDepth {
		filePath := lm.FilePath
		if !lm.enableFullFilePath {
			_, filePath = path.Split(filePath)
		}
		msg = fmt.Sprintf("[%s:%d] %s", filePath, lm.LineNumber, msg)
	}

	msg = levelPrefix[lm.Level] + " " + msg
	return msg
}
