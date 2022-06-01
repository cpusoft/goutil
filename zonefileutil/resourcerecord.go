package zonefileutil

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/guregu/null"
)

// type is golang keyword, so use Rr***
type ResourceRecord struct {
	// will have "." in the end  // lower
	RrDomain string `json:"rrDomain,omitempty"`
	// will remove the end "." // remove origin, only hostname // lower
	RrName string `json:"rrName,omitempty"`
	// upper
	RrClass string `json:"rrClass,omitempty"`
	// upper
	RrType string `json:"rrType,omitempty"`
	// null.NewInt(0, false) or null.NewInt(i64, true)
	RrTtl    null.Int `json:"rrTtl,omitempty"`
	RrValues []string `json:"rrValues,omitempty"`
}

//
func NewResourceRecord(rrDomain, rrName, rrClass, rrType string,
	rrTtl null.Int, rrValues []string) (resourceRecord *ResourceRecord) {
	resourceRecord = &ResourceRecord{
		RrDomain: FormatRrDomain(rrDomain),
		RrName:   FormatRrName(rrName),
		RrClass:  FormatRrClassOrRrType(rrClass),
		RrType:   FormatRrClassOrRrType(rrType),
		RrTtl:    rrTtl,
		RrValues: rrValues,
	}
	return
}

func FormatResourceRecord(oldResourceRecord *ResourceRecord) (newResourceRecord *ResourceRecord) {
	newResourceRecord = &ResourceRecord{
		RrDomain: FormatRrDomain(oldResourceRecord.RrDomain),
		RrName:   FormatRrName(oldResourceRecord.RrName),
		RrClass:  FormatRrClassOrRrType(oldResourceRecord.RrClass),
		RrType:   FormatRrClassOrRrType(oldResourceRecord.RrType),
		RrTtl:    oldResourceRecord.RrTtl,
		RrValues: oldResourceRecord.RrValues,
	}
	return newResourceRecord
}

func FormatRrClassOrRrType(t string) string {
	return strings.TrimSpace(strings.ToUpper(t))
}

// //  remove origin, only hostname, and will remove the end "." // lower
func FormatRrName(t string) string {
	return strings.TrimRight(strings.TrimSpace(strings.ToLower(t)), ".")
}

// will have "." in the end  // lower
func FormatRrDomain(t string) string {
	s := strings.TrimSpace(strings.ToLower(t))
	// should have "." as end
	if !strings.HasSuffix(s, ".") {
		s += "."
	}
	return s
}

func (c *ResourceRecord) String() string {
	var b strings.Builder
	b.Grow(128)

	ttl := ""
	if !c.RrTtl.IsZero() {
		ttl = strconv.Itoa(int(c.RrTtl.ValueOrZero()))
	}
	var space string
	b.WriteString(fmt.Sprintf("%-20s%-6s%-4s%-6s%-4s", c.RrName, ttl, c.RrClass, c.RrType, space))
	for i := range c.RrValues {
		b.WriteString(c.RrValues[i] + " ")
	}
	b.WriteString(osutil.GetNewLineSep())
	return b.String()
}

// resourceRecord should have RrName and RrType and RrValues
// domain: ***, or @, or ""
func DelResourceRecord(zoneFileModel *ZoneFileModel, oldResourceRecord *ResourceRecord) (err error) {
	if err := checkZoneFileModel(zoneFileModel); err != nil {
		belogs.Error("DelResourceRecord(): checkZoneFileModel fail:", err)
		return err
	}
	if err := CheckResourceRecord(oldResourceRecord); err != nil {
		belogs.Error("DelResourceRecord(): CheckResourceRecord oldResourceRecord fail:", err)
		return err
	}

	belogs.Debug("DelResourceRecord(): oldResourceRecord :", jsonutil.MarshalJson(oldResourceRecord))
	rr := make([]*ResourceRecord, 0)
	zoneFileModel.resourceRecordMutex.Lock()
	defer zoneFileModel.resourceRecordMutex.Unlock()
	for i := range zoneFileModel.ResourceRecords {
		if !equalResourceRecord(zoneFileModel.ResourceRecords[i], oldResourceRecord) {
			rr = append(rr, zoneFileModel.ResourceRecords[i])
		}
	}
	zoneFileModel.ResourceRecords = rr
	belogs.Info("DelResourceRecord(): found and delete, new zoneFileModel.ResourceRecords:", jsonutil.MarshalJson(rr))
	return nil
}

// oldResourceRecord/newResourceRecord: should have Domain and Type and Values
func UpdateResourceRecord(zoneFileModel *ZoneFileModel, oldResourceRecord, newResourceRecord *ResourceRecord) (err error) {
	if err := checkZoneFileModel(zoneFileModel); err != nil {
		belogs.Error("UpdateResourceRecord(): checkZoneFileModel fail:", err)
		return err
	}
	if err := CheckResourceRecord(oldResourceRecord); err != nil {
		belogs.Error("UpdateResourceRecord(): CheckResourceRecord oldResourceRecord fail:", err)
		return err
	}
	if err := CheckResourceRecord(newResourceRecord); err != nil {
		belogs.Error("UpdateResourceRecord(): CheckResourceRecord newResourceRecord fail:", err)
		return err
	}

	belogs.Debug("UpdateResourceRecord():  oldResourceRecord :", jsonutil.MarshalJson(oldResourceRecord),
		"  newResourceRecord :", jsonutil.MarshalJson(newResourceRecord))

	zoneFileModel.resourceRecordMutex.Lock()
	defer zoneFileModel.resourceRecordMutex.Unlock()
	for i := range zoneFileModel.ResourceRecords {
		if equalResourceRecord(zoneFileModel.ResourceRecords[i], oldResourceRecord) {
			zoneFileModel.ResourceRecords[i] = newResourceRecord
			belogs.Info("UpdateResourceRecord(): found and update ,new zoneFileModel.ResourceRecords:", jsonutil.MarshalJson(zoneFileModel.ResourceRecords))
			return nil
		}
	}
	return nil
}

// newResourceRecord: should have Domain and Type and Values
// afterResourceRecord: maybe nil, some rr must after the specified record, eg: ignored domain
// if afterResourceRecord Domain and Type and Values all are empty , newResourceRecord will add in the end
func AddResourceRecord(zoneFileModel *ZoneFileModel, afterResourceRecord, newResourceRecord *ResourceRecord) (err error) {
	if err := checkZoneFileModel(zoneFileModel); err != nil {
		belogs.Error("AddResourceRecord(): checkZoneFileModel fail:", err)
		return err
	}
	// not check afterResourceRecord
	if err := CheckResourceRecord(newResourceRecord); err != nil {
		belogs.Error("AddResourceRecord(): CheckResourceRecord newResourceRecord fail:", err)
		return err
	}
	belogs.Debug("AddResourceRecord():  afterResourceRecord :", afterResourceRecord,
		"   newResourceRecord :", jsonutil.MarshalJson(newResourceRecord))
	zoneFileModel.resourceRecordMutex.Lock()
	defer zoneFileModel.resourceRecordMutex.Unlock()
	if afterResourceRecord != nil && CheckResourceRecord(afterResourceRecord) == nil { // afterResourceRecord Domain and Type and Values are not all empty
		rr := make([]*ResourceRecord, 0)
		for i := range zoneFileModel.ResourceRecords {
			rr = append(rr, zoneFileModel.ResourceRecords[i])
			if equalResourceRecord(zoneFileModel.ResourceRecords[i], afterResourceRecord) {
				if len(newResourceRecord.RrName) == 0 {
					newResourceRecord.RrName = afterResourceRecord.RrName
				}
				rr = append(rr, newResourceRecord) // add newResourceRecord after afterResourceRecord
			}
		}
		zoneFileModel.ResourceRecords = rr
	} else {
		zoneFileModel.ResourceRecords = append(zoneFileModel.ResourceRecords, newResourceRecord)
	}
	belogs.Info("AddResourceRecord(): add ,new zoneFileModel.ResourceRecords :", jsonutil.MarshalJson(zoneFileModel.ResourceRecords))
	return nil
}

// rrName: ==hostname, or empty --> @,
// rrType: ==***, or "any" /"all" / "" --> all
func QueryResourceRecords(zoneFileModel *ZoneFileModel, queryResourceRecord *ResourceRecord) (resourceRecords []*ResourceRecord) {
	belogs.Debug("QueryResourceRecords(): queryResourceRecord:", jsonutil.MarshalJson(queryResourceRecord))
	rrName := queryResourceRecord.RrName
	if len(rrName) == 0 {
		rrName = "@"
	}
	rrType := queryResourceRecord.RrType
	if len(rrType) == 0 || rrType == "ANY" {
		rrType = "ALL"
	}
	belogs.Debug("QueryResourceRecords(): trim and lower/upper, rrName:", rrName, "    rrType:", rrType)

	resourceRecords = make([]*ResourceRecord, 0)
	zoneFileModel.resourceRecordMutex.RLock()
	defer zoneFileModel.resourceRecordMutex.RUnlock()
	for i := range zoneFileModel.ResourceRecords {
		if zoneFileModel.ResourceRecords[i].RrName == rrName {
			if rrType == "ALL" {
				resourceRecords = append(resourceRecords, zoneFileModel.ResourceRecords[i])
			} else {
				if zoneFileModel.ResourceRecords[i].RrType == rrType {
					resourceRecords = append(resourceRecords, zoneFileModel.ResourceRecords[i])
				}
			}
		}
	}
	belogs.Info("QueryResourceRecords(): trim and lower/upper, rrName:", rrName,
		"    rrType:", rrType, "   resourceRecords :", jsonutil.MarshalJson(resourceRecords))
	return resourceRecords
}

func CheckResourceRecord(resourceRecord *ResourceRecord) error {
	if resourceRecord == nil {
		return errors.New("resourceRecord is nill")
	}
	if len(resourceRecord.RrName) == 0 && len(resourceRecord.RrType) == 0 &&
		len(resourceRecord.RrValues) == 0 {
		belogs.Error("CheckResourceRecord():rrName,rrType and rrValues are all empty, fail:")
		return errors.New("rrName,rrType and rrValues are all empty")
	}
	return nil
}

func equalResourceRecord(leftResourceRecord, rightResourceRecord *ResourceRecord) bool {
	if leftResourceRecord == nil || rightResourceRecord == nil {
		return false
	}
	if leftResourceRecord.RrName == rightResourceRecord.RrName &&
		leftResourceRecord.RrType == rightResourceRecord.RrType &&
		jsonutil.MarshalJson(leftResourceRecord.RrValues) == jsonutil.MarshalJson(rightResourceRecord.RrValues) {
		return true
	}
	return false
}
