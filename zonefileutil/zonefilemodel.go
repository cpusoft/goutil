package zonefileutil

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"

	zonefile "github.com/bwesterb/go-zonefile"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/guregu/null"
)

type ZoneFileModel struct {
	zonefile *zonefile.Zonefile `json:"-"`

	// will have "." in the end  // lower
	Origin string `json:"origin"`
	// null.NewInt(0, false) or null.NewInt(i64, true)
	Ttl                 null.Int          `json:"ttl"`
	ResourceRecords     []*ResourceRecord `json:"resourceRecords"`
	resourceRecordMutex sync.RWMutex      `json:"-"`
	ZoneFileName        string            `json:"zoneFileName"`
}

func (c *ZoneFileModel) String() string {
	var b strings.Builder
	c.resourceRecordMutex.RLock()
	defer c.resourceRecordMutex.RUnlock()

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

func newZoneFileModel(zonefile *zonefile.Zonefile, zoneFileName string) *ZoneFileModel {
	c := &ZoneFileModel{}
	c.zonefile = zonefile
	c.ResourceRecords = make([]*ResourceRecord, 0)
	c.ZoneFileName = zoneFileName
	return c
}

// $origin must exist ;
// not support $include ;
// all ttl must be digital ;
// zoneFileName must be absolute path filename;
func LoadZoneFile(zoneFileName string) (zoneFileModel *ZoneFileModel, err error) {
	// Load zonefile
	data, err := ioutil.ReadFile(zoneFileName)
	if err != nil {
		belogs.Error("LoadZoneFile(): ReadFile fail:", zoneFileName, err)
		return nil, err
	}
	belogs.Debug("LoadZoneFile():len(data):", zoneFileName, len(data))

	zf, perr := zonefile.Load(data)
	if perr != nil {
		belogs.Error("LoadZoneFile():Load fail:", zoneFileName, perr)
		return nil, errors.New(perr.Error())
	}
	belogs.Debug("LoadZoneFile():len(zf.Entries):", zoneFileName, len(zf.Entries()))
	var lastRrName string
	zoneFileModel = newZoneFileModel(zf, zoneFileName)
	for i, e := range zf.Entries() {
		belogs.Debug("LoadZoneFile(): i :", i, "  e:", e)
		if len(e.Command()) > 0 {
			belogs.Debug("LoadZoneFile(): Command:", string(e.Command()), e.Values())
			if string(e.Command()) == "$ORIGIN" && len(e.Values()) > 0 {
				zoneFileModel.Origin = FormatRrDomain(string(e.Values()[0]))
			} else if string(e.Command()) == "$TTL" && len(e.Values()) > 0 {
				ttlStr := string(e.Values()[0])
				belogs.Debug("LoadZoneFile(): ttlStr:", ttlStr)
				ttl, _ := strconv.Atoi(ttlStr)
				zoneFileModel.Ttl = null.IntFrom(int64(ttl))
			}
		} else {
			// check Domain,if is empty, get last rrName
			rrName := string(e.Domain())
			if len(rrName) == 0 && len(lastRrName) > 0 {
				rrName = lastRrName
			} else if len(rrName) > 0 {
				lastRrName = rrName
			}
			// get ttl
			rrTtl := null.NewInt(0, false)
			if e.TTL() != nil {
				rrTtl = null.IntFrom(int64(*e.TTL()))
			}
			vs := make([]string, 0)
			for j := range e.Values() {
				vs = append(vs, string(e.Values()[j]))
				belogs.Debug("LoadZoneFile(): vs:", vs)
			}
			resourceRecord := NewResourceRecord("", rrName,
				string(e.Class()), string(e.Type()), rrTtl, vs)
			belogs.Debug("LoadZoneFile(): resourceRecord.Ttl:", resourceRecord.RrTtl)

			resourceRecord.RrValues = vs
			belogs.Debug("LoadZoneFile(): resourceRecord:", jsonutil.MarshalJson(resourceRecord))
			zoneFileModel.ResourceRecords = append(zoneFileModel.ResourceRecords, resourceRecord)
		}
	}

	// check
	if len(zoneFileModel.Origin) == 0 {
		belogs.Error("LoadZoneFile():Origin must be exist, fail:", zoneFileName)
		return nil, errors.New("Origin must be exist")
	}

	belogs.Info("LoadZoneFile(): zoneFileModel:", jsonutil.MarshalJson(zoneFileModel))
	return zoneFileModel, nil
}

// if file is empty, then save to zoneFileName in LoadZoneFile()
func SaveZoneFile(zoneFileModel *ZoneFileModel, file string) (err error) {
	belogs.Debug("SaveZoneFile(): file:", file)

	if err := checkZoneFileModel(zoneFileModel); err != nil {
		belogs.Error("SaveZoneFile(): checkZoneFileModel fail:", file, err)
		return err
	}
	b := zoneFileModel.String()
	if len(file) == 0 {
		file = zoneFileModel.ZoneFileName
	}
	belogs.Debug("SaveZoneFile(): save to file:", file, "  len(b):", len(b))
	return fileutil.WriteBytesToFile(file, []byte(b))
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
