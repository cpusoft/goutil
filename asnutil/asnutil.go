package asnutil

import ()

// asn is included in parent asn ;  .
// selfMin, selfMax, parentMin, parentMax
func AsnIncludeInParentAsn(selfMin, selfMax, parentMin, parentMax uint64) bool {

	if selfMin == 0 || selfMax == 0 ||
		parentMin == 0 || parentMax == 0 {
		return false
	}

	// parentMin <--- selfMin <---------> selfMax ---> parentMax
	if parentMin <= selfMin && selfMax <= parentMax {
		return true
	}
	return false
}
