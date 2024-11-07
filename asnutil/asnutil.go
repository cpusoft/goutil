package asnutil

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/cpusoft/goutil/belogs"
)

// asn is included in parent asn ;  .
// selfAsn, selfMin, selfMax,parentAsn, parentMin, parentMax
func AsnIncludeInParentAsn(selfAsn, selfMin, selfMax, parentAsn, parentMin, parentMax uint64) bool {

	if selfAsn == 0 && selfMin == 0 && selfMax == 0 &&
		parentAsn == 0 && parentMin == 0 && parentMax == 0 {
		return false
	}

	if selfAsn == parentAsn {
		return true
	}
	if parentMin <= selfAsn && selfAsn <= parentMax {
		return true
	}

	// parentMin <--- selfMin <---------> selfMax ---> parentMax
	if parentMin <= selfMin && selfMax <= parentMax {
		return true
	}
	return false
}

func GetAsnOwnerByCymru(asn int) (string, error) {
	//get asn owner
	if asn == 0 {
		return "", errors.New("asn is zero")
	}

	lookupUrl := fmt.Sprintf(`AS%d.asn.cymru.com`, asn)
	belogs.Debug("GetAsnOwnerByCymru():asn lookupUrl:", asn, lookupUrl)
	txt, err := net.LookupTXT(lookupUrl)
	if err != nil {
		belogs.Error("GetAsnOwnerByCymru(): lookupUrl fail:", lookupUrl, asn, err)
		return "", err
	}
	if len(txt) == 0 {
		belogs.Info("GetAsnOwnerByCymru(): lookupUrl txt is empty:", asn, txt)
		return "", nil
	}
	// dig short AS266087.asn.cymru.com TXT
	// AS266087.asn.cymru.com. 76104   IN      TXT     "266087 | BR | lacnic | 2017-03-13 | Orbitel Telecomunicacoes e Informatica Ltda, BR"
	belogs.Debug("getAsnOwner(): lookupUrl,txt:", lookupUrl, txt)
	splits := strings.Split(txt[0], "|")
	if len(splits) < 3 {
		belogs.Error("getAsnOwner(): Split fail:", asn, txt)
		return "", nil
	}

	return strings.TrimSpace(splits[len(splits)-1] + "," + splits[2]), nil
}
