package ginserver

import ()

type GinUserModel struct {
	// user Id
	Id uint64 `json:"id"`

	// role Id
	RoleId uint64 `json:"roleId"`
}
