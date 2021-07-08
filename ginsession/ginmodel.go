package ginsession

type GinUserModel struct {
	// user Id
	Id uint64 `json:"id" xorm:"id"`

	// role Id
	RoleId uint64 `json:"roleId" xorm:"roleId"`
}
