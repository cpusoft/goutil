package rrdputil

import (
	xml "encoding/xml"
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
	MapSerialDeltas map[uint64]uint64 `xml:"-"`
	MaxSerail       uint64            `xml:"-"`
	MinSerail       uint64            `xml:"-"`
}

type NotificationSnapshot struct {
	XMLName xml.Name `xml:"snapshot"`
	Uri     string   `xml:"uri,attr"`
	Hash    string   `xml:"hash,attr"`
}
type NotificationDelta struct {
	XMLName xml.Name `xml:"delta"`
	Serial  uint64   `xml:"serial,attr"`
	Uri     string   `xml:"uri,attr"`
	Hash    string   `xml:"hash,attr"`
}

// support sort: from smaller to bigger
//	v := []NotificationDelta{{Serial: 1, Uri: "3", Hash: "333"},{Serial: 0, Uri: "6", Hash: "666"},{Serial: 3, Uri: "2", Hash: "2222"},{Serial: 8, Uri: "7", Hash: "7777"}}
//	sort.Sort(NotificationDeltas(v))
type NotificationDeltas []NotificationDelta

func (v NotificationDeltas) Len() int {
	return len(v)
}

func (v NotificationDeltas) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// default comparison.
func (v NotificationDeltas) Less(i, j int) bool {
	return v[i].Serial < v[j].Serial
}

type SnapshotModel struct {
	XMLName          xml.Name          `xml:"snapshot"`
	Xmlns            string            `xml:"xmlns,attr"`
	Version          string            `xml:"version,attr"`
	SessionId        string            `xml:"session_id,attr"`
	Serial           uint64            `xml:"serial,attr"`
	SnapshotPublishs []SnapshotPublish `xml:"publish"`
	// to check
	Hash string `xml:"-"`
}

type SnapshotPublish struct {
	XMLName xml.Name `xml:"publish"`
	Uri     string   `xml:"uri,attr"`
	Base64  string   `xml:",chardata"`
}

type DeltaModel struct {
	XMLName        xml.Name        `xml:"delta"`
	Xmlns          string          `xml:"xmlns,attr"`
	Version        string          `xml:"version,attr"`
	SessionId      string          `xml:"session_id,attr"`
	Serial         uint64          `xml:"serial,attr"`
	DeltaPublishs  []DeltaPublish  `xml:"publish"`
	DeltaWithdraws []DeltaWithdraw `xml:"withdraw"`

	// to check
	Hash string `xml:"-"`
}

type DeltaPublish struct {
	XMLName xml.Name `xml:"publish"`
	Uri     string   `xml:"uri,attr"`
	Hash    string   `xml:"hash,attr"`
	Base64  string   `xml:",chardata"`
}
type DeltaWithdraw struct {
	XMLName xml.Name `xml:"withdraw"`
	Uri     string   `xml:"uri,attr"`
	Hash    string   `xml:"hash,attr"`
}

type RrdpFile struct {
	FilePath string
	FileName string
	// add /del
	SyncType string
}
