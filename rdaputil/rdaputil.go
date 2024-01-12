package rdaputil

import (
	"errors"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/openrdap/rdap"
)

// domain: google.com
func RdapDomain(domain string) (*rdap.Domain, error) {
	start := time.Now()
	belogs.Debug("RdapDomain(): domain:", domain)
	req := &rdap.Request{
		Type:  rdap.DomainRequest,
		Query: domain,
	}

	client := &rdap.Client{}
	resp, err := client.Do(req)
	if err != nil {
		belogs.Error("RdapDomain(): client.Do fail, domain:", domain, err)
		return nil, err
	}

	rdapDomain, ok := resp.Object.(*rdap.Domain)
	if !ok {
		belogs.Error("RdapDomain(): Object is not Domain fail, domain:", domain)
		return nil, errors.New("resp is not rdap.Domain")
	}
	belogs.Debug("RdapDomain(): get rdap ok, domain:", domain, "  rdapDomain:", jsonutil.MarshalJson(rdapDomain), "  time(s):", time.Since(start))
	return rdapDomain, nil
}

// asn: 2846
func RdapAsn(asn uint64) (*rdap.Autnum, error) {
	start := time.Now()
	belogs.Debug("RdapAsn(): asn:", asn)
	req := &rdap.Request{
		Type:  rdap.AutnumRequest,
		Query: convert.ToString(asn),
	}

	client := &rdap.Client{}
	resp, err := client.Do(req)
	if err != nil {
		belogs.Error("RdapAsn(): client.Do fail, asn:", asn, err)
		return nil, err
	}

	rdapAsn, ok := resp.Object.(*rdap.Autnum)
	if !ok {
		belogs.Error("RdapAsn(): Object is not Autnum fail, asn:", asn)
		return nil, errors.New("resp is not rdap.Autnum")
	}
	belogs.Debug("RdapAsn(): get rdap ok, asn:", asn, "  rdapAsn:", jsonutil.MarshalJson(rdapAsn), "  time(s):", time.Since(start))
	return rdapAsn, nil
}

// addressprefix: "8.8.8.0/24"
func RdapAddressPrefix(addressprefix string) (*rdap.IPNetwork, error) {
	start := time.Now()
	belogs.Debug("RdapAddressPrefix(): addressprefix:", addressprefix)
	req := &rdap.Request{
		Type:  rdap.IPRequest,
		Query: addressprefix,
	}

	client := &rdap.Client{}
	resp, err := client.Do(req)
	if err != nil {
		belogs.Error("RdapAddressPrefix(): client.Do fail, addressprefix:", addressprefix, err)
		return nil, err
	}

	rdapAddressPrefix, ok := resp.Object.(*rdap.IPNetwork)
	if !ok {
		belogs.Error("RdapAddressPrefix(): Object is not IPNetwork fail, addressprefix:", addressprefix)
		return nil, errors.New("resp is not rdap.IPNetwork")
	}
	belogs.Debug("RdapAddressPrefix(): get rdap ok, addressprefix:", addressprefix, "  rdapAddressPrefix:", jsonutil.MarshalJson(rdapAddressPrefix), "  time(s):", time.Since(start))
	return rdapAddressPrefix, nil
}
