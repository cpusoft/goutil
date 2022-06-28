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
	return strings.TrimSuffix(strings.TrimSpace(strings.ToLower(t)), ".")
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

//  https://datatracker.ietf.org/doc/rfc8765/
//  RrName: ***, or @, or ""
//
// if rrClass==ANY,                 remove all RRsets from a name in all classes: TTL = 0xFFFFFFFE, RDLEN = 0
//
// if rrClass!=ANY and rrType==ANY, remove all RRsets from a name in given class: TTL = 0xFFFFFFFE, RDLEN = 0
//
// if rrClass!=ANY and rrType!=ANY and rrData == emtpy, remove CLASS and TYPE specify the RRset: TTL = 0xFFFFFFFE, RDLEN = 0
//
// if rrClass!=ANY and rrType!=ANY and rrData != emtpy, Remove an individual RR from a name  TTL = 0xFFFFFFFF
func DelResourceRecord(zoneFileModel *ZoneFileModel, delResourceRecord *ResourceRecord) (newDelResourceRecord *ResourceRecord, err error) {
	if err := checkZoneFileModel(zoneFileModel); err != nil {
		belogs.Error("DelResourceRecord(): checkZoneFileModel fail:", zoneFileModel, err)
		return nil, err
	}
	zoneFileModel.resourceRecordMutex.Lock()
	defer zoneFileModel.resourceRecordMutex.Unlock()

	if err := CheckNameAndTypeAndValues(delResourceRecord); err != nil {
		belogs.Error("DelResourceRecord(): CheckNameAndTypeAndValues delResourceRecord fail:",
			jsonutil.MarshalJson(delResourceRecord), err)
		return nil, err
	}

	belogs.Info("DelResourceRecord():will delResourceRecord :", jsonutil.MarshalJson(delResourceRecord))
	zfRrs := make([]*ResourceRecord, 0)
	newDelResourceRecord = delResourceRecord
	if delResourceRecord.RrClass == "ANY" || (delResourceRecord.RrClass != "ANY" && delResourceRecord.RrType == "ANY") {
		newDelResourceRecord.RrTtl = null.IntFrom(DSO_REMOVE_COLLECTIVE_RESOURCE_RECORD_TTL)
		for i := range zoneFileModel.ResourceRecords {
			if zoneFileModel.ResourceRecords[i].RrName != delResourceRecord.RrName {
				zfRrs = append(zfRrs, zoneFileModel.ResourceRecords[i])
			}
		}
	} else if delResourceRecord.RrClass != "ANY" && delResourceRecord.RrType != "ANY" && len(delResourceRecord.RrValues) == 0 {
		newDelResourceRecord.RrTtl = null.IntFrom(DSO_REMOVE_COLLECTIVE_RESOURCE_RECORD_TTL)
		for i := range zoneFileModel.ResourceRecords {
			if !(zoneFileModel.ResourceRecords[i].RrName == delResourceRecord.RrName &&
				zoneFileModel.ResourceRecords[i].RrClass == delResourceRecord.RrClass &&
				zoneFileModel.ResourceRecords[i].RrType == delResourceRecord.RrType) {
				zfRrs = append(zfRrs, zoneFileModel.ResourceRecords[i])
			}
		}
	} else if delResourceRecord.RrClass != "ANY" && delResourceRecord.RrType != "ANY" && len(delResourceRecord.RrValues) != 0 {
		newDelResourceRecord.RrTtl = null.IntFrom(DSO_REMOVE_SPECIFIED_RESOURCE_RECORD_TTL)
		for i := range zoneFileModel.ResourceRecords {
			if !EqualResourceRecord(zoneFileModel.ResourceRecords[i], delResourceRecord) {
				zfRrs = append(zfRrs, zoneFileModel.ResourceRecords[i])
			}
		}
	}
	zoneFileModel.ResourceRecords = zfRrs
	belogs.Info("DelResourceRecord(): found and delete, new zoneFileModel.ResourceRecords:", jsonutil.MarshalJson(zfRrs),
		"   newDelResourceRecord:", newDelResourceRecord)
	return newDelResourceRecord, nil
}

// newResourceRecord.RrName: ***, or @, or ""
//
// Class/Type cannot ANY,'update' must be individual RR
//
// oldResourceRecord/newResourceRecord: should have same Domain and Type(CANNOT "ANY"), just different Values to update
func UpdateResourceRecord(zoneFileModel *ZoneFileModel, oldResourceRecord, newResourceRecord *ResourceRecord) (err error) {
	if err := checkZoneFileModel(zoneFileModel); err != nil {
		belogs.Error("UpdateResourceRecord(): checkZoneFileModel fail:", zoneFileModel, err)
		return err
	}
	if err := CheckNameAndTypeAndValues(oldResourceRecord); err != nil {
		belogs.Error("UpdateResourceRecord(): CheckNameAndTypeAndValues oldResourceRecord fail:", oldResourceRecord, err)
		return err
	}
	if err := CheckDomainOrNameAndTypeAndValues(newResourceRecord, true); err != nil {
		belogs.Error("UpdateResourceRecord(): CheckDomainOrNameAndTypeAndValues newResourceRecord fail:", newResourceRecord, err)
		return err
	}
	if err = CheckClassTypeShouldNoAny(oldResourceRecord); err != nil {
		belogs.Error("AddResourceRecord(): oldResourceRecord RrClass or RrType cannot be ANY, oldResourceRecord:", oldResourceRecord, err)
		return err
	}
	if err = CheckClassTypeShouldNoAny(newResourceRecord); err != nil {
		belogs.Error("AddResourceRecord(): newResourceRecord RrClass or RrType cannot be ANY, newResourceRecord:", newResourceRecord, err)
		return err
	}

	// rrdomain
	if len(newResourceRecord.RrDomain) == 0 {
		newResourceRecord.RrDomain = newResourceRecord.RrName + "." + zoneFileModel.Origin
	}

	if oldResourceRecord.RrName != newResourceRecord.RrName ||
		oldResourceRecord.RrType != newResourceRecord.RrType {
		belogs.Error("UpdateResourceRecord(): oldRr's rrName or rrType is not equal to newRr's RrName or rrType, fail:",
			"  oldRr:", jsonutil.MarshalJson(oldResourceRecord), " newRr", jsonutil.MarshalJson(newResourceRecord))
		return errors.New("OldRr's rrName and rrType all should  be equal to newRr")
	}

	belogs.Info("UpdateResourceRecord():  oldResourceRecord :", jsonutil.MarshalJson(oldResourceRecord),
		"  newResourceRecord :", jsonutil.MarshalJson(newResourceRecord))

	zoneFileModel.resourceRecordMutex.Lock()
	defer zoneFileModel.resourceRecordMutex.Unlock()
	for i := range zoneFileModel.ResourceRecords {
		if EqualResourceRecord(zoneFileModel.ResourceRecords[i], oldResourceRecord) {
			zoneFileModel.ResourceRecords[i] = newResourceRecord
			belogs.Info("UpdateResourceRecord(): found and update ,new zoneFileModel.ResourceRecords:", jsonutil.MarshalJson(zoneFileModel.ResourceRecords))
			return nil
		}
	}
	return nil
}

// newResourceRecord: should have Domain and Type and Values
//
// afterResourceRecord: maybe nil, some rr must after the specified record, eg: ignored domain
//
// if afterResourceRecord Domain and Type and Values all are empty , newResourceRecord will add in the end
func AddResourceRecord(zoneFileModel *ZoneFileModel, afterResourceRecord, newResourceRecord *ResourceRecord) (err error) {
	if err := checkZoneFileModel(zoneFileModel); err != nil {
		belogs.Error("AddResourceRecord(): checkZoneFileModel fail:", zoneFileModel, err)
		return err
	}
	// if afterResourceRecord==nil, then need check RrName
	needRrName := (afterResourceRecord == nil)
	if err := CheckDomainOrNameAndTypeAndValues(newResourceRecord, needRrName); err != nil {
		belogs.Error("AddResourceRecord(): CheckDomainOrNameAndTypeAndValues newResourceRecord fail:", newResourceRecord, "   needRrName:", needRrName, err)
		return err
	}
	if err = CheckClassTypeShouldNoAny(newResourceRecord); err != nil {
		belogs.Error("AddResourceRecord(): RrClass or RrType cannot be ANY, newResourceRecord:", newResourceRecord, err)
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
	if afterResourceRecord != nil && CheckNameAndTypeAndValues(afterResourceRecord) == nil { // afterResourceRecord Domain and Type and Values are not all empty
		rr := make([]*ResourceRecord, 0)
		for i := range zoneFileModel.ResourceRecords {
			rr = append(rr, zoneFileModel.ResourceRecords[i])
			if EqualResourceRecord(zoneFileModel.ResourceRecords[i], afterResourceRecord) {
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
	if err := CheckNameAndTypeAndValues(queryResourceRecord); err != nil {
		belogs.Error("QueryResourceRecords(): CheckNameAndTypeAndValues queryResourceRecord fail:", queryResourceRecord, err)
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
				belogs.Debug("QueryResourceRecords():rrType is (ANY):", rrType, ", will add rrTmp:", jsonutil.MarshalJson(rrTmp))
			} else {
				if rrType == rrTmp.RrType {
					resourceRecords = append(resourceRecords, rrTmp)
					belogs.Debug("QueryResourceRecords():rrType is ", rrType, ", will add rrTmp:", jsonutil.MarshalJson(rrTmp))
				}
			}
		}
	}
	belogs.Info("QueryResourceRecords():queryResourceRecord:", jsonutil.MarshalJson(queryResourceRecord),
		"   resourceRecords :", jsonutil.MarshalJson(resourceRecords))
	return resourceRecords, nil
}

func CheckNameAndTypeAndValues(resourceRecord *ResourceRecord) error {
	if resourceRecord == nil {
		belogs.Error("CheckNameAndTypeAndValues():resourceRecord is nil, fail:")
		return errors.New("resourceRecord is nill")
	}
	if len(resourceRecord.RrName) == 0 && len(resourceRecord.RrType) == 0 &&
		len(resourceRecord.RrValues) == 0 {
		belogs.Error("CheckNameAndTypeAndValues():rrName,rrType and rrValues are all empty, fail:")
		return errors.New("rrName,rrType and rrValues are all empty")
	}
	return nil
}

// check rrDomain/rrType/rrValues/
// if needRrName, check rrName
func CheckClassTypeShouldNoAny(resourceRecord *ResourceRecord) error {
	if resourceRecord == nil {
		belogs.Error("CheckClassTypeNoAny():resourceRecord is nil, fail:")
		return errors.New("resourceRecord is nill")
	}
	if resourceRecord.RrClass == "ANY" || resourceRecord.RrType == "ANY" {
		belogs.Error("CheckClassTypeNoAny():Class or Type is ANY, fail:", jsonutil.MarshalJson(resourceRecord))
		return errors.New("Class or Type is ANY")
	}
	return nil
}

// check rrDomain/rrType/rrValues/
// if needRrName, check rrName
func CheckDomainOrNameAndTypeAndValues(resourceRecord *ResourceRecord, needRrName bool) error {
	if resourceRecord == nil {
		belogs.Error("CheckDomainOrNameAndTypeAndValues():resourceRecord is nil, fail:")
		return errors.New("resourceRecord is nill")
	}
	if len(resourceRecord.RrDomain) == 0 {
		belogs.Error("CheckDomainOrNameAndTypeAndValues():rrDomain is empty, fail:")
		return errors.New("rrDomain is empty")
	}

	if needRrName && len(resourceRecord.RrName) == 0 {
		belogs.Error("CheckDomainOrNameAndTypeAndValues():rrName is empty, fail:")
		return errors.New("rrName is empty")
	}

	if len(resourceRecord.RrType) == 0 {
		belogs.Error("CheckDomainOrNameAndTypeAndValues():rrType is empty, fail:")
		return errors.New("rrType is empty")
	}

	if len(resourceRecord.RrValues) == 0 {
		belogs.Error("CheckDomainOrNameAndTypeAndValues(): rrValues is empty, fail:")
		return errors.New("rrValues is empty")
	}
	return nil
}

func EqualResourceRecord(leftResourceRecord, rightResourceRecord *ResourceRecord) bool {
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
	newResourceRecord.RrValues = make([]string, len(resourceRecord.RrValues))
	copy(newResourceRecord.RrValues, resourceRecord.RrValues)
	belogs.Debug("deepcopyResourceRecord(): newResourceRecord:", jsonutil.MarshalJson(newResourceRecord))
	return newResourceRecord
}

// get rrkey:
func GetResourceRecordKey(resourceRecord *ResourceRecord) string {
	if resourceRecord == nil {
		return ""
	}
	rrKey := resourceRecord.RrDomain + "#" + resourceRecord.RrType
	belogs.Info("GetResourceRecordKey():rrKey:", rrKey)
	return rrKey
}

func GetResourceRecordAnyKey(resourceRecord *ResourceRecord) string {
	if resourceRecord == nil {
		return ""
	}
	rrKey := resourceRecord.RrDomain + "#" + DnsIntTypes[DNS_TYPE_ANY]
	belogs.Info("getResourceRecordAnyKey():rrKey:", rrKey)
	return rrKey
}
