package uuidutil

import (
	belogs "github.com/beego/beego/v2/core/logs"
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
