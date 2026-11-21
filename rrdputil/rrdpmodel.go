package rrdputil

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	xml "encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/orisano/gosax/xmlb"
)

// ///////////////////////////////////////////////
// notification
// ///////////////////////////////////////////////
type NotificationModel struct {
	XMLName   xml.Name             `xml:"notification"`
	Xmlns     string               `xml:"xmlns,attr"`
	Version   string               `xml:"version,attr"`
	SessionId string               `xml:"session_id,attr"`
	Serial    uint64               `xml:"serial,attr"`
	Snapshot  NotificationSnapshot `xml:"snapshot"`
	Deltas    []NotificationDelta  `xml:"delta"`

	//map[serial]serial: just save exist serial ,for check
	NotificationUrl string            `xml:"-"`
	MapSerialDeltas map[uint64]uint64 `xml:"-"`
	MaxSerial       uint64            `xml:"-"`
	MinSerial       uint64            `xml:"-"`
}

type NotificationSnapshot struct {
	XMLName xml.Name `xml:"snapshot" json:"snapshot"`
	Uri     string   `xml:"uri,attr" json:"uri"`
	Hash    string   `xml:"hash,attr" json:"-"`
}
type NotificationDelta struct {
	XMLName xml.Name `xml:"delta" json:"delta"`
	Serial  uint64   `xml:"serial,attr" json:"serial"`
	Uri     string   `xml:"uri,attr" json:"uri"`
	Hash    string   `xml:"hash,attr" json:"-"`
}

// support sort: from  bigger to smaller: from newer to older
//
//	v := []NotificationDelta{{Serial: 10, Uri: "3", Hash: "333"},{Serial: 9, Uri: "6", Hash: "666"},{Serial: 8, Uri: "2", Hash: "2222"},{Serial: 7, Uri: "7", Hash: "7777"}}
//	sort.Sort(NotificationDeltasSort(v))
type NotificationDeltasSort []NotificationDelta

func (v NotificationDeltasSort) Len() int {
	return len(v)
}

func (v NotificationDeltasSort) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// default comparison.
func (v NotificationDeltasSort) Less(i, j int) bool {
	return v[i].Serial > v[j].Serial
}

// ///////////////////////////////////////
// snapshot
// ///////////////////////////////////////////////////
type SnapshotModel struct {
	XMLName          xml.Name          `xml:"snapshot" json:"snapshot"`
	Xmlns            string            `xml:"xmlns,attr" json:"xmlns"`
	Version          string            `xml:"version,attr" json:"version"`
	SessionId        string            `xml:"session_id,attr" json:"session_id"`
	Serial           uint64            `xml:"serial,attr" json:"serial"`
	SnapshotPublishs []SnapshotPublish `xml:"publish"  json:"publish"`
	// to check
	Hash        string `xml:"-"`
	SnapshotUrl string `xml:"-"`
}

func (c SnapshotModel) String() string {
	m := make(map[string]interface{})
	m["snapshotUrl"] = c.SnapshotUrl
	m["sessionId"] = c.SessionId
	m["serial"] = c.Serial
	m["len(snapshotPublishs)"] = len(c.SnapshotPublishs)
	return jsonutil.MarshalJson(m)
}

type SnapshotPublish struct {
	XMLName xml.Name `xml:"publish" json:"publish"`
	Xmlns   string   `xml:"xmlns,attr" json:"xmlns"`
	Uri     string   `xml:"uri,attr" json:"uri"`
	Base64  string   `xml:",chardata" json:"-"`
}

// ////////////////////////////////////////////////////
// delta
// ////////////////////////////////////////////////////
type DeltaModel struct {
	XMLName        xml.Name        `xml:"delta" json:"delta"`
	Xmlns          string          `xml:"xmlns,attr" json:"xmlns"`
	Version        string          `xml:"version,attr" json:"version"`
	SessionId      string          `xml:"session_id,attr" json:"session_id"`
	Serial         uint64          `xml:"serial,attr" json:"serial"`
	DeltaPublishs  []DeltaPublish  `xml:"publish" json:"publish"`
	DeltaWithdraws []DeltaWithdraw `xml:"withdraw" json:"withdraw"`

	// to check
	Hash     string `xml:"-"  json:"hash"`
	DeltaUrl string `xml:"-"  json:"deltaUrl"`
}

func (c DeltaModel) String() string {
	m := make(map[string]interface{})
	m["deltaUrl"] = c.DeltaUrl
	m["sessionId"] = c.SessionId
	m["serial"] = c.Serial
	m["len(deltaPublishs)"] = len(c.DeltaPublishs)
	m["len(deltaWithdraws)"] = len(c.DeltaWithdraws)
	return jsonutil.MarshalJson(m)
}

// support sort: from   smaller to bigger
//
//	v := []NotificationDelta{{Serial: 1, Uri: "3", Hash: "333"},{Serial: 2, Uri: "6", Hash: "666"},{Serial: 3, Uri: "2", Hash: "2222"},{Serial: 4, Uri: "7", Hash: "7777"}}
//	sort.Sort(DeltaModelsSort(v))
type DeltaModelsSort []DeltaModel

func (v DeltaModelsSort) Len() int {
	return len(v)
}

func (v DeltaModelsSort) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// default comparison.
func (v DeltaModelsSort) Less(i, j int) bool {
	return v[i].Serial < v[j].Serial
}

type DeltaPublish struct {
	XMLName xml.Name `xml:"publish" json:"publish"`
	Uri     string   `xml:"uri,attr" json:"uri"`
	Hash    string   `xml:"hash,attr" json:"-"`
	Base64  string   `xml:",chardata" json:"-"`
	//Base64 string `xml:",innerxml" json:"base64"`
}
type DeltaWithdraw struct {
	XMLName xml.Name `xml:"withdraw" json:"withdraw"`
	Uri     string   `xml:"uri,attr" json:"uri"`
	Hash    string   `xml:"hash,attr" json:"-"`
}

// ///////////////////////////////////
// save rrdp file
// /////////////////////////////////
type RrdpFile struct {
	FilePath string `json:"filePath"`
	FileName string `json:"fileName"`
	// every file's uri: SnapshotPublish.Uri/DeltaPublish.Uri/DeltaWithdraw.Uri
	FileUri string `json:"fileUri"`
	// update(add) /del
	SyncType string `json:"syncType"`
	// snapshot's url, or delta's url
	SourceUrl string `json:"sourceUrl"`
	Serial    uint64 `json:"serial"`

	SnapshotPublish *SnapshotPublish `json:"snapshotPublish"`
	// deltaPUblish or deltaWithdraw
	DeltaPublish  *DeltaPublish  `json:"deltaPublish"`
	DeltaWithdraw *DeltaWithdraw `json:"deltaWithdraw"`
}

// //////////////////////////////////////
// unxml snapshot.xml
// ///////////////////////////////////////
// SnapshotSAXParser 流式解析器（专为你的 XML 结构封装）

// ParseBytes 解析内存中的 []byte XML 数据
func parseXmlToSnapshotModel(xmlData []byte) (*SnapshotModel, error) {
	start := time.Now()
	buf := make([]byte, 0, 4096)
	dec := xmlb.NewDecoder(bytes.NewReader(xmlData), buf)

	snapshot := &SnapshotModel{}
	var currentPub *SnapshotPublish
	var charBuf strings.Builder
	inPublish := false

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			belogs.Error("parseXmlToSnapshotModel(): Token fail", err)
			return nil, err
		}

		switch tok.Type() {
		case xmlb.StartElement:
			elem, _ := tok.StartElement()

			switch elem.Name.Local {
			case "snapshot":
				for _, attr := range elem.Attr {
					switch attr.Name.Local {
					case "xmlns":
						snapshot.Xmlns = attr.Value
					case "version":
						snapshot.Version = attr.Value
					case "session_id":
						snapshot.SessionId = attr.Value
					case "serial":
						fmt.Sscanf(attr.Value, "%d", &snapshot.Serial)
					}
				}

			case "publish":
				inPublish = true
				currentPub = &SnapshotPublish{}
				for _, attr := range elem.Attr {
					if attr.Name.Local == "uri" {
						currentPub.Uri = attr.Value
					}
				}
				charBuf.Reset()
			}

		case xmlb.EndElement:
			elem := tok.EndElement()
			if elem.Name.Local == "publish" && currentPub != nil {
				currentPub.Base64 = strings.TrimSpace(charBuf.String())
				snapshot.SnapshotPublishs = append(snapshot.SnapshotPublishs, *currentPub)
				currentPub = nil
				inPublish = false
			}

		case xmlb.CharData:
			if inPublish {
				data, _ := tok.CharData()
				charBuf.Write(data)
			}
		}
	}
	belogs.Debug("parseXmlToSnapshotModel(): len(xml)", len(xmlData), "time(s)", time.Since(start))
	return snapshot, nil
}

func syncByRsyncAgain(uri string) (newBase64, newHash string, err error) {
	f, err := os.CreateTemp("", "rsync-*.tmp")
	if err != nil {
		belogs.Error("syncByRsyncAgain(): syncByRsyncAgain fail, rsyncUrl is :", uri, " err is :", err)
		return "", "", err
	}
	tmpPath := f.Name()
	// 关闭 fd，留给 rsync 写入
	_ = f.Close()

	// 退出清理
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	// rsync 直接写入这个临时文件
	// 在这里读取、校验 tmpPath
	belogs.Debug("syncByRsyncAgain(): Command: rsync", "-Lirzts", "--del", "--timeout=120", "--contimeout=120",
		"--no-motd", "-4", uri, tmpPath)
	cmd := exec.Command("rsync", "-Lrzts", "--del", "--timeout=120", "--contimeout=120", "--no-motd", "-4", uri, tmpPath)
	// if success, the len(output) will be zero
	output, err := cmd.CombinedOutput()
	if err != nil {
		belogs.Error("syncByRsyncAgain(): exec.Command fail, rsyncUrl is :", uri, "   output is ", string(output), " err is :", err)
		// some err detail in output
		err = errors.New(string(output) + ", " + err.Error())
		return "", "", err
	}
	belogs.Debug("syncByRsyncAgain(): rsync ok, save to tmpPath:", tmpPath, "  output:", string(output))
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", "", err
	}
	newBase64 = base64.StdEncoding.EncodeToString(data)
	h := sha256.New()
	h.Write(data)

	// sum 是 [32]byte 原始哈希
	newHash = hex.EncodeToString(h.Sum(nil))
	return newBase64, newHash, nil
}

func checkIsValidRrdp(uri, base64, hash string, isDelta bool) error {
	if uri == "" {
		belogs.Error("checkIsValidRrdp(): uri is empty")
		return errors.New("uri is empty")
	}
	if base64 == "" {
		belogs.Error("checkIsValidRrdp(): base64 is empty", "uri", uri)
		return errors.New("base64 is empty")
	}
	data, hit, err := DecodeBase64(base64)
	validB64 := err == nil

	h := sha256.New()
	h.Write(data)
	newHash := hex.EncodeToString(h.Sum(nil))
	hashMatch := newHash == hash
	if validB64 && hashMatch || hit {
		belogs.Info("checkIsValidRrdp(): check rrdp is ok,  uri:", uri)
		return nil
	}
	belogs.Error("checkIsValidRrdp(): check rrdp fail, should use rsync download again", "uri", uri)
	return errors.New("check rrdp fail, should use rsync download again")
}

func DecodeBase64(src string) ([]byte, bool, error) {
	segments := strings.Split(src, "/")
	var hit bool
	for _, seg := range segments {
		if len(seg) > 0 {
			hit = true
			break
		}
	}

	b, err := base64.StdEncoding.DecodeString(src)
	if err == nil {
		return b, hit, nil
	}
	b, err = base64.RawStdEncoding.DecodeString(src)
	if err == nil {
		return b, hit, nil
	}
	b, err = base64.URLEncoding.DecodeString(src)
	if err == nil {
		return b, hit, nil
	}
	b, err = base64.RawURLEncoding.DecodeString(src)
	if err == nil {
		return b, hit, nil
	}
	return nil, hit, err
}
