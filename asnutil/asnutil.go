package asnutil

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
