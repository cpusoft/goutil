package asn1cert

import (
	"encoding/asn1"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

type RoaFileAsn1 struct {
	SignedDataOid asn1.ObjectIdentifier `json:"signedDataOid"`
	SignedDatas   []asn1.RawValue       `json:"signedDatas" asn1:"optional,explicit,default:0,tag:0"`
}

type OctetString []byte
type RoaOctectStringAsn1 struct {
	RoaOid          asn1.ObjectIdentifier
	RoaOctectString OctetString `asn1:"tag:0,explicit,optional"`
}

// asID as in rfc6482
type RoaAsn1 struct {
	RoaAsIdAsn1       RoaAsIdAsn1        `json:"asID"`
	RoaIpAddressAsn1s []RoaIpAddressAsn1 `json:"ipAddrBlocks"`
}
type RoaAsIdAsn1 int64
type RoaIpAddressAsn1 struct {
	RoaAddressFamily      []byte                 `json:"roaAddressFamily"`
	RoaPrefixAddressAsn1s []RoaPrefixAddressAsn1 `json:"addresses"`
}
type RoaPrefixAddressAsn1 struct {
	RoaPrefixAddressAsn1 asn1.BitString `json:"prefixAddress"`
	RoaMaxLengthAsn1     int            `asn1:"optional,default:-1" json:"maxLength"`
}

// data: asn1.FullBytes
func ParseToRouteOriginAttestation(data []byte) (roaAsn1 RoaAsn1, err error) {
	belogs.Debug("ParseToRouteOriginAttestation(): len(data):", len(data))
	var roaOctectStringAsn1 RoaOctectStringAsn1
	_, err = asn1.Unmarshal(data, &roaOctectStringAsn1)
	if err != nil {
		belogs.Error("ParseToRouteOriginAttestation(): Unmarshal roaAsn1 fail:", err)
		return
	}
	belogs.Debug("ParseToRouteOriginAttestation(): roaAsn1:", jsonutil.MarshalJson(roaAsn1))

	_, err = asn1.Unmarshal([]byte(roaOctectStringAsn1.RoaOctectString), &roaAsn1)
	if err != nil {
		belogs.Error("ParseToRouteOriginAttestation(): Unmarshal roaAsn1 fail:", err)
		return
	}
	belogs.Debug("ParseToRouteOriginAttestation():roaAsn1:", jsonutil.MarshalJson(roaAsn1))
	return
}
