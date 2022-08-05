package zonefileutil

/*
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
func LoadZoneFile(zoneFileName string) (originModel *dnsutil.OriginModel, err error) {
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
	originModel = dnsutil.NewOriginModel()
	for i, e := range zf.Entries() {
		belogs.Debug("LoadZoneFile(): i :", i, "  e:", e)
		if len(e.Command()) > 0 {
			belogs.Debug("LoadZoneFile(): Command:", string(e.Command()), e.Values())
			if string(e.Command()) == "$ORIGIN" && len(e.Values()) > 0 {
				originModel.Origin = dnsutil.FormatOrigin(string(e.Values()[0]))
			} else if string(e.Command()) == "$TTL" && len(e.Values()) > 0 {
				ttlStr := string(e.Values()[0])
				belogs.Debug("LoadZoneFile(): ttlStr:", ttlStr)
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
			belogs.Debug("LoadZoneFile(): rrName:", rrName)

			// get ttl
			rrTtl := null.NewInt(0, false)
			if e.TTL() != nil {
				ttl := int64(*e.TTL())
				if ttl > dnsutil.DSO_ADD_RECOURCE_RECORD_MAX_TTL {
					belogs.Error("LoadZoneFile(): ttl is bigger than DSO_ADD_RECOURCE_RECORD_MAX_TTL:", ttl, dnsutil.DSO_ADD_RECOURCE_RECORD_MAX_TTL)
					return nil, errors.New("ttl is bigger than DSO_ADD_RECOURCE_RECORD_MAX_TTL")
				}
				rrTtl = null.IntFrom(ttl)
			}
			belogs.Debug("LoadZoneFile(): rrTtl:", rrTtl)

			var rrData string
			for j := range e.Values() {
				rrData += (string(e.Values()[j]) + " ")
			}
			belogs.Debug("LoadZoneFile(): rrData:", rrData)

			rrModel := dnsutil.NewRrModel(originModel.Origin, rrName,
				string(e.Type()), string(e.Class()), rrTtl, rrData)
			belogs.Debug("LoadZoneFile(): rrModel:", jsonutil.MarshalJson(rrModel))
			originModel.RrModels = append(originModel.RrModels, rrModel)
		}
	}

	// check
	if len(originModel.Origin) == 0 {
		belogs.Error("LoadZoneFile():Origin must be exist, fail:", zoneFileName)
		return nil, errors.New("Origin must be exist")
	}

	belogs.Info("LoadZoneFile(): originModel:", jsonutil.MarshalJson(originModel))
	return originModel, nil
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
*/
