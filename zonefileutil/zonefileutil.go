package zonefileutil

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	zonefile "github.com/bwesterb/go-zonefile"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/guregu/null"
)

type ZoneFileModel struct {
	zonefile *zonefile.Zonefile `json:"-"`

	Origin          string           `json:"origin"`
	Ttl             null.Int         `json:"ttl"`
	ResourceRecords []ResourceRecord `json:"resourceRecords"`
}

func (c *ZoneFileModel) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%-10s%-20s", "$ORIGIN", c.Origin) + osutil.GetNewLineSep())
	if c.Ttl.IsZero() {
		b.WriteString(osutil.GetNewLineSep())
	} else {
		b.WriteString(fmt.Sprintf("%-10s%-20s", "$TTL",
			strconv.Itoa(int(c.Ttl.ValueOrZero()))) + osutil.GetNewLineSep())
	}
	for i := range c.ResourceRecords {
		b.WriteString(c.ResourceRecords[i].String())
	}
	return b.String()
}

func newZoneFileModel(zonefile *zonefile.Zonefile) *ZoneFileModel {
	c := &ZoneFileModel{}
	c.zonefile = c.zonefile
	c.ResourceRecords = make([]ResourceRecord, 0)
	return c
}

// type is golang keyword, so use Rr***
type ResourceRecord struct {
	RrDomain string   `json:"rrDomain"`
	RrClass  string   `json:"rrClass"`
	RrType   string   `json:"rrType"`
	RrTtl    null.Int `json:"rrTtl"`
	RrValues []string `json:"rrValues"`
}

func (c ResourceRecord) String() string {
	var b strings.Builder
	b.Grow(512)
	ttl := ""
	if !c.RrTtl.IsZero() {
		ttl = strconv.Itoa(int(c.RrTtl.ValueOrZero()))
	}
	var space string
	b.WriteString(fmt.Sprintf("%-20s%-6s%-4s%-6s%-4s", c.RrDomain, ttl, c.RrClass, c.RrType, space))
	for i := range c.RrValues {
		b.WriteString(c.RrValues[i] + " ")
	}
	b.WriteString(osutil.GetNewLineSep())
	return b.String()
}

// not support $include
func LoadZoneFile(file string) (zoneFileModel *ZoneFileModel, err error) {
	// Load zonefile
	data, err := ioutil.ReadFile(file)
	if err != nil {
		belogs.Error("LoadZoneFile(): ReadFile fail:", file, err)
		return nil, err
	}
	belogs.Debug("LoadZoneFile():len(data):", file, len(data))

	zf, perr := zonefile.Load(data)
	if perr != nil {
		belogs.Error("LoadZoneFile():Load fail:", file, perr.LineNo(), perr)
		return nil, errors.New(perr.Error())
	}
	belogs.Debug("LoadZoneFile():len(zf.Entries):", file, len(zf.Entries()))

	zoneFileModel = newZoneFileModel(zf)
	for i, e := range zf.Entries() {
		belogs.Debug("LoadZoneFile(): i :", i, "  e:", e)
		if len(e.Command()) > 0 {
			belogs.Debug("LoadZoneFile(): Command:", string(e.Command()), e.Values())
			if string(e.Command()) == "$ORIGIN" && len(e.Values()) > 0 {
				zoneFileModel.Origin = string(e.Values()[0])
			} else if string(e.Command()) == "$TTL" && len(e.Values()) > 0 {
				ttlStr := string(e.Values()[0])
				belogs.Debug("LoadZoneFile(): ttlStr:", ttlStr)
				ttl, _ := strconv.Atoi(ttlStr)
				zoneFileModel.Ttl = null.IntFrom(int64(ttl))
			}
		} else {
			resourceRecord := ResourceRecord{
				RrDomain: string(e.Domain()),
				RrClass:  string(e.Class()),
				RrType:   string(e.Type()),
			}
			belogs.Debug("LoadZoneFile(): resourceRecord:", resourceRecord)

			if e.TTL() != nil {
				resourceRecord.RrTtl = null.IntFrom(int64(*e.TTL()))
			}
			belogs.Debug("LoadZoneFile(): resourceRecord.Ttl:", resourceRecord.RrTtl)

			vs := make([]string, 0)
			for j := range e.Values() {
				vs = append(vs, string(e.Values()[j]))
				belogs.Debug("LoadZoneFile(): vs:", vs)
			}
			resourceRecord.RrValues = vs
			belogs.Debug("LoadZoneFile(): resourceRecord:", jsonutil.MarshalJson(resourceRecord))
			zoneFileModel.ResourceRecords = append(zoneFileModel.ResourceRecords, resourceRecord)
		}
	}
	belogs.Info("LoadZoneFile(): zoneFileModel:", jsonutil.MarshalJson(zoneFileModel))
	return zoneFileModel, nil
}

func SaveZoneFile(zoneFileModel *ZoneFileModel, file string) (err error) {
	if err := checkZoneFileModel(zoneFileModel); err != nil {
		belogs.Error("SaveZoneFile(): checkZoneFileModel fail:", file, err)
		return err
	}
	b := zoneFileModel.String()
	belogs.Debug("SaveZoneFile(): file:", file, "  len(b):", len(b))
	return fileutil.WriteBytesToFile(file, []byte(b))
}

// resourceRecord should have RrDomain and RrType and RrValues
// domain: ***, or @, or ""
func DelResourceRecord(zoneFileModel *ZoneFileModel, oldResourceRecord ResourceRecord) (err error) {
	if err := checkZoneFileModel(zoneFileModel); err != nil {
		belogs.Error("DelResourceRecord(): checkZoneFileModel fail:", err)
		return err
	}
	if err := checkResourceRecord(oldResourceRecord); err != nil {
		belogs.Error("DelResourceRecord(): checkResourceRecord oldResourceRecord fail:", err)
		return err
	}

	belogs.Debug("DelResourceRecord(): oldResourceRecord :", jsonutil.MarshalJson(oldResourceRecord))
	rr := make([]ResourceRecord, 0)
	for i := range zoneFileModel.ResourceRecords {
		if !(zoneFileModel.ResourceRecords[i].RrDomain == oldResourceRecord.RrDomain &&
			zoneFileModel.ResourceRecords[i].RrType == oldResourceRecord.RrType &&
			jsonutil.MarshalJson(zoneFileModel.ResourceRecords[i].RrValues) == jsonutil.MarshalJson(oldResourceRecord.RrValues)) {
			rr = append(rr, zoneFileModel.ResourceRecords[i])
			break
		}
	}
	zoneFileModel.ResourceRecords = rr
	belogs.Info("DelResourceRecord(): found and delete ,new zoneFileModel.ResourceRecords:", jsonutil.MarshalJson(rr))
	return nil
}

// oldResourceRecord/newResourceRecord: should have Domain and Type and Values
func UpdateResourceRecord(zoneFileModel *ZoneFileModel, oldResourceRecord, newResourceRecord ResourceRecord) (err error) {
	if err := checkZoneFileModel(zoneFileModel); err != nil {
		belogs.Error("UpdateResourceRecord(): checkZoneFileModel fail:", err)
		return err
	}
	if err := checkResourceRecord(oldResourceRecord); err != nil {
		belogs.Error("UpdateResourceRecord(): checkResourceRecord oldResourceRecord fail:", err)
		return err
	}
	if err := checkResourceRecord(newResourceRecord); err != nil {
		belogs.Error("UpdateResourceRecord(): checkResourceRecord newResourceRecord fail:", err)
		return err
	}

	belogs.Debug("UpdateResourceRecord():  oldResourceRecord :", jsonutil.MarshalJson(oldResourceRecord),
		"  newResourceRecord :", jsonutil.MarshalJson(newResourceRecord))

	for i := range zoneFileModel.ResourceRecords {
		if zoneFileModel.ResourceRecords[i].RrDomain == oldResourceRecord.RrDomain &&
			zoneFileModel.ResourceRecords[i].RrType == oldResourceRecord.RrType &&
			jsonutil.MarshalJson(zoneFileModel.ResourceRecords[i].RrValues) == jsonutil.MarshalJson(oldResourceRecord.RrValues) {
			zoneFileModel.ResourceRecords[i] = newResourceRecord
			belogs.Info("UpdateResourceRecord(): found and update ,new zoneFileModel.ResourceRecords:", jsonutil.MarshalJson(zoneFileModel.ResourceRecords))
			return nil
		}
	}
	return nil
}

// newResourceRecord: should have Domain and Type and Values
// afterResourceRecord: some rr must after the specified record, eg: ignored domain
// if afterResourceRecord Domain and Type and Values all are empty , newResourceRecord will add in the end
func AddResourceRecord(zoneFileModel *ZoneFileModel, afterResourceRecord, newResourceRecord ResourceRecord) (err error) {
	if err := checkZoneFileModel(zoneFileModel); err != nil {
		belogs.Error("AddResourceRecord(): checkZoneFileModel fail:", err)
		return err
	}
	if err := checkResourceRecord(newResourceRecord); err != nil {
		belogs.Error("AddResourceRecord(): checkResourceRecord newResourceRecord fail:", err)
		return err
	}
	belogs.Debug("AddResourceRecord():  afterResourceRecord :", jsonutil.MarshalJson(afterResourceRecord),
		"   newResourceRecord :", jsonutil.MarshalJson(afterResourceRecord))
	if checkResourceRecord(afterResourceRecord) == nil { // afterResourceRecord Domain and Type and Values are not all empty
		rr := make([]ResourceRecord, 0)
		for i := range zoneFileModel.ResourceRecords {
			rr = append(rr, zoneFileModel.ResourceRecords[i])
			if zoneFileModel.ResourceRecords[i].RrDomain == afterResourceRecord.RrDomain &&
				zoneFileModel.ResourceRecords[i].RrType == afterResourceRecord.RrType &&
				jsonutil.MarshalJson(zoneFileModel.ResourceRecords[i].RrValues) == jsonutil.MarshalJson(afterResourceRecord.RrValues) {
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

func checkZoneFileModel(zoneFileModel *ZoneFileModel) error {
	if zoneFileModel == nil {
		belogs.Error("checkZoneFileModel():zoneFileModel is nil, fail:")
		return errors.New("zoneFileModel is nil")
	}
	if len(zoneFileModel.Origin) == 0 {
		belogs.Error("checkZoneFileModel():Origin is empty, fail:")
		return errors.New("Origin is empty")
	}
	return nil
}

func checkResourceRecord(resourceRecord ResourceRecord) error {
	if len(resourceRecord.RrDomain) == 0 && len(resourceRecord.RrType) == 0 &&
		len(resourceRecord.RrValues) == 0 {
		belogs.Error("checkResourceRecord():rrDomain,rrType and rrValues are all empty, fail:")
		return errors.New("rrDomain,rrType and rrValues are all empty")
	}
	return nil
}
