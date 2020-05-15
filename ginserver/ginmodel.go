package ginserver

import ()

type GinUserModel struct {
	Id                uint64 `json:"id"`
	FullName          string `json:"fullName"`
	Phone             string `json:"phone"`
	Email             string `json:"email"`
	Wechat            string `json:"wechat"`
	Twitter           string `json:"twitter"`
	Company           string `json:"company"`
	AsnAddressPrefixs string `json:"asnAddressPrefixs"`
	Status            string `json:"status"`

	RoleId    uint64 `json:"roleId"`
	RoleTitle string `json:"roleTitle"`
}
