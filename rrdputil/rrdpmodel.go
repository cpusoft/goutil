package rrdputil

import (
	xml "encoding/xml"

	"github.com/cpusoft/goutil/jsonutil"
)

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

// support sort: from  bigger to smaller
//
//	v := []NotificationDelta{{Serial: 10, Uri: "3", Hash: "333"},{Serial: 9, Uri: "6", Hash: "666"},{Serial: 8, Uri: "2", Hash: "2222"},{Serial: 7, Uri: "7", Hash: "7777"}}
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
	return v[i].Serial > v[j].Serial
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

type RrdpFile struct {
	FilePath string `json:"filePath"`
	FileName string `json:"fileName"`
	// add /del
	SyncType string `json:"syncType"`
	//snapshoturl or deltaurl
	SourceUrl string `json:"sourceUrl"`
}
