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
	RrType string `json:"rrType,omitempty"`
	// upper
	RrClass string `json:"rrClass,omitempty"`
	// null.NewInt(0, false) or null.NewInt(i64, true)
	RrTtl    null.Int `json:"rrTtl,omitempty"`
	RrValues []string `json:"rrValues,omitempty"`
}

//
func NewResourceRecord(rrDomain, rrName, rrType, rrClass string,
	rrTtl null.Int, rrValues []string) (resourceRecord *ResourceRecord) {
	resourceRecord = &ResourceRecord{
		RrDomain: FormatRrDomain(rrDomain),
		RrName:   FormatRrName(rrName),
		RrType:   FormatRrClassOrRrType(rrType),
		RrClass:  FormatRrClassOrRrType(rrClass),
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

// not include Domain:
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
		belogs.Error("DelResourceRecord(): checkZoneFileModel fail:", zoneFileModel, err)
		return err
	}
	if err := CheckResourceRecord(oldResourceRecord); err != nil {
		belogs.Error("DelResourceRecord(): CheckResourceRecord oldResourceRecord fail:", oldResourceRecord, err)
		return err
	}

	belogs.Info("DelResourceRecord(): oldResourceRecord :", jsonutil.MarshalJson(oldResourceRecord))
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
		belogs.Error("UpdateResourceRecord(): checkZoneFileModel fail:", zoneFileModel, err)
		return err
	}
	if err := CheckResourceRecord(oldResourceRecord); err != nil {
		belogs.Error("UpdateResourceRecord(): CheckResourceRecord oldResourceRecord fail:", oldResourceRecord, err)
		return err
	}
	if err := CheckAddOrUpdateResourceRecord(newResourceRecord, true); err != nil {
		belogs.Error("UpdateResourceRecord(): CheckResourceRecord newResourceRecord fail:", newResourceRecord, err)
		return err
	}
	// rrdomain
	if len(newResourceRecord.RrDomain) == 0 {
		newResourceRecord.RrDomain = newResourceRecord.RrName + "." + zoneFileModel.Origin
	}
	belogs.Info("UpdateResourceRecord():  oldResourceRecord :", jsonutil.MarshalJson(oldResourceRecord),
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
		belogs.Error("AddResourceRecord(): checkZoneFileModel fail:", zoneFileModel, err)
		return err
	}
	// if afterResourceRecord==nil, then need check RrName
	needRrName := (afterResourceRecord == nil)
	if err := CheckAddOrUpdateResourceRecord(newResourceRecord, needRrName); err != nil {
		belogs.Error("AddResourceRecord(): CheckAddOrUpdateResourceRecord newResourceRecord fail:", newResourceRecord, "   needRrName:", needRrName, err)
		return err
	}
	// rrdomain
	if len(newResourceRecord.RrDomain) == 0 {
		newResourceRecord.RrDomain = newResourceRecord.RrName + "." + zoneFileModel.Origin
	}
	belogs.Info("AddResourceRecord():  afterResourceRecord :", afterResourceRecord,
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
func QueryResourceRecords(zoneFileModel *ZoneFileModel, queryResourceRecord *ResourceRecord) (resourceRecords []*ResourceRecord, err error) {

	if err := checkZoneFileModel(zoneFileModel); err != nil {
		belogs.Error("QueryResourceRecords(): checkZoneFileModel fail:", zoneFileModel, err)
		return nil, err
	}
	if err := CheckResourceRecord(queryResourceRecord); err != nil {
		belogs.Error("QueryResourceRecords(): CheckResourceRecord queryResourceRecord fail:", queryResourceRecord, err)
		return nil, err
	}

	rrName := queryResourceRecord.RrName
	if len(rrName) == 0 {
		rrName = "@"
	}
	rrType := queryResourceRecord.RrType
	if len(rrType) == 0 {
		rrType = "ANY"
	}
	belogs.Info("QueryResourceRecords(): queryResourceRecord:", jsonutil.MarshalJson(queryResourceRecord))

	resourceRecords = make([]*ResourceRecord, 0)
	zoneFileModel.resourceRecordMutex.RLock()
	defer zoneFileModel.resourceRecordMutex.RUnlock()
	for i := range zoneFileModel.ResourceRecords {
		if zoneFileModel.ResourceRecords[i].RrName == rrName {
			rrTmp := deepcopyResourceRecord(zoneFileModel.ResourceRecords[i])
			if rrTmp.RrTtl.IsZero() {
				rrTmp.RrTtl = zoneFileModel.Ttl
			}
			if rrType == "ANY" {
				resourceRecords = append(resourceRecords, rrTmp)
			} else {
				if rrType == rrTmp.RrType {
					resourceRecords = append(resourceRecords, rrTmp)
				}
			}
		}
	}
	belogs.Info("QueryResourceRecords():queryResourceRecord:", jsonutil.MarshalJson(queryResourceRecord),
		"   resourceRecords :", jsonutil.MarshalJson(resourceRecords))
	return resourceRecords, nil
}

func CheckResourceRecord(resourceRecord *ResourceRecord) error {
	if resourceRecord == nil {
		belogs.Error("CheckResourceRecord():resourceRecord is nil, fail:")
		return errors.New("resourceRecord is nill")
	}
	if len(resourceRecord.RrName) == 0 && len(resourceRecord.RrType) == 0 &&
		len(resourceRecord.RrValues) == 0 {
		belogs.Error("CheckResourceRecord():rrName,rrType and rrValues are all empty, fail:")
		return errors.New("rrName,rrType and rrValues are all empty")
	}
	return nil
}

func CheckAddOrUpdateResourceRecord(resourceRecord *ResourceRecord, needRrName bool) error {
	if resourceRecord == nil {
		belogs.Error("CheckAddOrUpdateResourceRecord():resourceRecord is nil, fail:")
		return errors.New("resourceRecord is nill")
	}
	if needRrName && len(resourceRecord.RrName) == 0 {
		belogs.Error("CheckAddOrUpdateResourceRecord():rrName is empty, fail:")
		return errors.New("rrName,rrType and rrValues are all empty")
	}
	if len(resourceRecord.RrType) == 0 ||
		len(resourceRecord.RrValues) == 0 {
		belogs.Error("CheckAddOrUpdateResourceRecord():rrType or rrValues is empty, fail:")
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

func deepcopyResourceRecord(resourceRecord *ResourceRecord) (newResourceRecord *ResourceRecord) {
	newResourceRecord = &ResourceRecord{
		RrDomain: resourceRecord.RrDomain,
		RrName:   resourceRecord.RrName,
		RrType:   resourceRecord.RrType,
		RrClass:  resourceRecord.RrClass,
		RrTtl:    resourceRecord.RrTtl,
	}
	newResourceRecord.RrValues = make([]string, 0)
	copy(newResourceRecord.RrValues, resourceRecord.RrValues)
	return newResourceRecord
}
