package errorutil

import (
	belogs "github.com/astaxie/beego/logs"
)

func LogErrAndPanic(msg string, err error, willPanic bool) {
	if err != nil {
		belogs.Error(msg, err.Error())
		if willPanic {
			panic(err)
		}
	}
}
