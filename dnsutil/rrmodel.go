package dnsutil

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/osutil"
	"github.com/guregu/null"
)

// for mysql, zonefile
type RrModel struct {
	Id       uint64 `json:"id" xorm:"id int"`
	OriginId uint64 `json:"originId" xorm:"originId int"`
	// will have "." in the end
	Origin string `json:"origin"` // lower

	// will remove the end "." // remove origin, only hostname
	RrName  string `json:"rrName" xorm:"rrName varchar"`   // lower
	RrType  string `json:"rrType" xorm:"rrType varchar"`   // upper
	RrClass string `json:"rrClass" xorm:"rrClass varchar"` // upper
	// null.NewInt(0, false) or null.NewInt(i64, true)
	RrTtl  null.Int `json:"rrTtl" xorm:"rrTtl int"`
	RrData string   `json:"rrData" xorm:"rrData varchar"`

	UpdateTime time.Time `json:"updateTime" xorm:"updateTime datetime"`
}

//
func NewRrModel(origin, rrName, rrType, rrClass string,
	rrTtl null.Int, rrData string) (rrModel *RrModel) {
	rrModel = &RrModel{
		Origin:  origin,
		RrName:  FormatRrName(rrName),
		RrType:  FormatRrClassOrRrType(rrType),
		RrClass: FormatRrClassOrRrType(rrClass),
		RrTtl:   rrTtl,
		RrData:  rrData,
	}
	return rrModel
}

// get rrkey:
func GetRrKey(rrModel *RrModel) string {
	if rrModel == nil {
		return ""
	}
	rrKey := rrModel.RrName + "." + rrModel.Origin + "#" + rrModel.RrType
	belogs.Debug("GetRrKey():rrKey:", rrKey)
	return rrKey
}

func FormatRrClassOrRrType(t string) string {
	return strings.TrimSpace(strings.ToUpper(t))
}

// //  remove origin, only hostname, and will remove the end "." // lower
func FormatRrName(t string) string {
	return strings.TrimSuffix(strings.TrimSpace(strings.ToLower(t)), ".")
}

// not include Domain:
func (rrModel *RrModel) String() string {
	var b strings.Builder
	b.Grow(128)

	ttl := ""
	if !rrModel.RrTtl.IsZero() {
		ttl = strconv.Itoa(int(rrModel.RrTtl.ValueOrZero()))
	}
	var space string
	b.WriteString(fmt.Sprintf("%-20s%-6s%-4s%-6s%-4s", rrModel.RrName, ttl, rrModel.RrClass, rrModel.RrType, space))
	b.WriteString(rrModel.RrData)
	b.WriteString(osutil.GetNewLineSep())
	return b.String()
}

// get rrAnyKey:
func GetRrAnyKey(rrModel *RrModel) string {
	if rrModel == nil {
		return ""
	}
	rrAnyKey := rrModel.RrName + "." + rrModel.Origin + "#" + DnsIntTypes[DNS_TYPE_ANY]
	belogs.Debug("GetRrAnyKey():rrAnyKey:", rrAnyKey)
	return rrAnyKey
}

// get rrDelKey
func GetRrDelKey(rrModel *RrModel) string {
	if rrModel == nil {
		return ""
	}
	rrDelKey := rrModel.RrName + "." + rrModel.Origin + "#" + DNS_RR_DEL_KEY
	belogs.Debug("GetRrDelKey():rrDelKey:", rrDelKey)
	return rrDelKey
}

func IsDelRr(rrModel *RrModel) bool {
	belogs.Debug("IsDelRr():rrTtl:", rrModel.RrTtl)
	if rrModel.RrTtl.ValueOrZero() == DSO_DEL_SPECIFIED_RESOURCE_RECORD_TTL ||
		rrModel.RrTtl.ValueOrZero() == DSO_DEL_COLLECTIVE_RESOURCE_RECORD_TTL {
		return true
	}
	return false
}

type RrDataAModel struct {
}
type RrDataNsModel struct {
}
type RrDataCNameModel struct {
}
type RrDataSoaModel struct {
}
type RrDataPtrModel struct {
}
type RrDataMxModel struct {
}
type RrDataTxtModel struct {
}

type RrDataAaaaModel struct {
}
type RrDataSrvModel struct {
}
