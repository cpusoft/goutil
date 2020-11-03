package uuidutil

import (
	belogs "github.com/cpusoft/beego/logs"
	uuid "github.com/satori/go.uuid"
)

func GetUuid() string {
	u, err := uuid.NewV4()
	if err != nil {
		belogs.Error("GetUuid(): fail", err)
		return ""
	}
	return u.String()
}
