package asn1cert

import (
	"encoding/asn1"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

type RoaFileAsn1 struct {
	ContentType asn1.ObjectIdentifier
	Seqs        []asn1.RawValue `asn1:"optional,explicit,default:0,tag:0""`
}

type OctetString []byte
type RoaAsn1 struct {
	RoaOid          asn1.ObjectIdentifier
	RoaOctectString OctetString `asn1:"tag:0,explicit,optional"`
}

// asID as in rfc6482
type RouteOriginAttestation struct {
	AsId         AsId                 `json:"asID"`
	IpAddrBlocks []RoaIpAddressFamily `json:"ipAddrBlocks"`
}
type AsId int64
type RoaIpAddressFamily struct {
	AddressFamily []byte         `json:"addressFamily"`
	Addresses     []RoaIpAddress `json:"addresses"`
}
type RoaIpAddress struct {
	Address   asn1.BitString `json:"address"`
	MaxLength int            `asn1:"optional,default:-1" json:"maxLength"`
}

// data: asn1.FullBytes
func ParseToRouteOriginAttestation(data []byte) (routeOriginAttestation RouteOriginAttestation, err error) {
	belogs.Debug("ParseToRouteOriginAttestation(): len(data):", len(data))
	var roaAsn1 RoaAsn1
	_, err = asn1.Unmarshal(data, &roaAsn1)
	if err != nil {
		belogs.Error("ParseToRouteOriginAttestation(): Unmarshal roaAsn1 fail:", err)
		return
	}
	belogs.Debug("ParseToRouteOriginAttestation(): roaAsn1:", jsonutil.MarshalJson(roaAsn1))

	_, err = asn1.Unmarshal([]byte(roaAsn1.RoaOctectString), &routeOriginAttestation)
	if err != nil {
		belogs.Error("ParseToRouteOriginAttestation(): Unmarshal routeOriginAttestation fail:", err)
		return
	}
	belogs.Debug("ParseToRouteOriginAttestation():routeOriginAttestation:", jsonutil.MarshalJson(routeOriginAttestation))
	return
}
