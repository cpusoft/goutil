package dnsutil

import (
	"errors"
	"io/ioutil"
	"strconv"

	"github.com/bwesterb/go-zonefile"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/guregu/null"
)

// $origin must exist ;
// not support $include ;
// all ttl must be digital ;
// zoneFileName must be absolute path filename;
func LoadFromZoneFile(zoneFileName string) (originModel *OriginModel, err error) {
	// Load zonefile
	data, err := ioutil.ReadFile(zoneFileName)
	if err != nil {
		belogs.Error("LoadFromZoneFile(): ReadFile fail:", zoneFileName, err)
		return nil, err
	}
	belogs.Debug("LoadFromZoneFile():len(data):", zoneFileName, len(data))

	zf, perr := zonefile.Load(data)
	if perr != nil {
		belogs.Error("LoadFromZoneFile():Load fail:", zoneFileName, perr)
		return nil, errors.New(perr.Error())
	}
	belogs.Debug("LoadFromZoneFile():len(zf.Entries):", zoneFileName, len(zf.Entries()))
	var lastRrName string
	originModel = NewOriginModel()
	for i, e := range zf.Entries() {
		belogs.Debug("LoadFromZoneFile(): i :", i, "  e:", e)
		if len(e.Command()) > 0 {
			belogs.Debug("LoadFromZoneFile(): Command:", string(e.Command()), e.Values())
			if string(e.Command()) == "$ORIGIN" && len(e.Values()) > 0 {
				originModel.Origin = FormatOrigin(string(e.Values()[0]))
			} else if string(e.Command()) == "$TTL" && len(e.Values()) > 0 {
				ttlStr := string(e.Values()[0])
				belogs.Debug("LoadFromZoneFile(): ttlStr:", ttlStr)
				ttl, _ := strconv.Atoi(ttlStr)
				originModel.Ttl = null.IntFrom(int64(ttl))
			}
		} else {
			// check Domain,if is empty, get last rrName
			rrName := FormatRrName(string(e.Domain()))
			if len(rrName) == 0 && len(lastRrName) > 0 {
				rrName = lastRrName
			} else if len(rrName) > 0 {
				lastRrName = rrName
			}
			belogs.Debug("LoadFromZoneFile(): rrName:", rrName)

			// get ttl
			rrTtl := null.NewInt(0, false)
			if e.TTL() != nil {
				ttl := int64(*e.TTL())
				if ttl > DSO_ADD_RECOURCE_RECORD_MAX_TTL {
					belogs.Error("LoadFromZoneFile(): ttl is bigger than DSO_ADD_RECOURCE_RECORD_MAX_TTL:", ttl, DSO_ADD_RECOURCE_RECORD_MAX_TTL)
					return nil, errors.New("ttl is bigger than DSO_ADD_RECOURCE_RECORD_MAX_TTL")
				}
				rrTtl = null.IntFrom(ttl)
			}
			belogs.Debug("LoadFromZoneFile(): rrTtl:", rrTtl)

			var rrData string
			for j := range e.Values() {
				rrData += (string(e.Values()[j]) + " ")
			}
			belogs.Debug("LoadFromZoneFile(): rrData:", rrData)

			rrModel := NewRrModel(originModel.Origin, rrName,
				string(e.Type()), string(e.Class()), rrTtl, rrData)
			belogs.Debug("LoadFromZoneFile(): rrModel:", jsonutil.MarshalJson(rrModel))
			originModel.RrModels = append(originModel.RrModels, rrModel)
		}
	}

	// check
	if len(originModel.Origin) == 0 {
		belogs.Error("LoadFromZoneFile():Origin must be exist, fail:", zoneFileName)
		return nil, errors.New("Origin must be exist")
	}

	belogs.Info("LoadFromZoneFile(): originModel:", jsonutil.MarshalJson(originModel))
	return originModel, nil
}

func SaveToZoneFile(originModel *OriginModel, file string) (err error) {
	belogs.Debug("SaveToZoneFile(): file:", file)
	b := originModel.String()
	belogs.Debug("SaveToZoneFile(): save to file:", file, "  len(b):", len(b))
	return fileutil.WriteBytesToFile(file, []byte(b))
}
