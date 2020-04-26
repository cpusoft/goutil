package errorutil

import (
	"errors"

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

//
func CloseErrorChanToError(errChan chan error) error {
	close(errChan)
	errStr := ""
	for {
		if value, ok := <-errChan; ok {
			errStr = errStr + ";" + value.Error()
		} else {
			break //表示channel已经被关闭，退出循环
		}
	}
	if len(errStr) > 0 {
		return errors.New(errStr)
	}
	return nil

}
