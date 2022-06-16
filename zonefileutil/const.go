package zonefileutil

const (
	DNS_OPCODE_DSO  = 6
	DNS_QR_REQUEST  = 0
	DNS_QR_RESPONSE = 1

	//DNS_TYPE_ALL   = 0 // not actually type, means all
	DNS_TYPE_A     = 1
	DNS_TYPE_NS    = 2
	DNS_TYPE_CNAME = 5
	DNS_TYPE_SOA   = 6
	DNS_TYPE_PTR   = 12
	DNS_TYPE_MX    = 15
	DNS_TYPE_TXT   = 16
	DNS_TYPE_AAAA  = 28
	DNS_TYPE_SRV   = 33
	DNS_TYPE_ANY   = 255

	//DNS_CLASS_ALL = 0 // not actually class, means all
	DNS_CLASS_IN  = 1
	DNS_CLASS_ANY = 255
)

var DnsIntTypes map[int]string = map[int]string{
	//	DNS_TYPE_ALL:   "ALL",
	DNS_TYPE_A:     "A",
	DNS_TYPE_NS:    "NS",
	DNS_TYPE_CNAME: "CNAME",
	DNS_TYPE_SOA:   "SOA",
	DNS_TYPE_PTR:   "PTR",
	DNS_TYPE_MX:    "MX",
	DNS_TYPE_TXT:   "TXT",
	DNS_TYPE_AAAA:  "AAAA",
	DNS_TYPE_SRV:   "SRV",
	DNS_TYPE_ANY:   "ANY",
}
var DnsStrTypes map[string]int = map[string]int{
	//	DNS_TYPE_ALL:   "ALL",
	"A":     DNS_TYPE_A,
	"NS":    DNS_TYPE_NS,
	"CNAME": DNS_TYPE_CNAME,
	"SOA":   DNS_TYPE_SOA,
	"PTR":   DNS_TYPE_PTR,
	"MX":    DNS_TYPE_MX,
	"TXT":   DNS_TYPE_TXT,
	"AAAA":  DNS_TYPE_AAAA,
	"SRV":   DNS_TYPE_SRV,
	"ANY":   DNS_TYPE_ANY,
}

var DnsIntClasses map[int]string = map[int]string{
	//DNS_CLASS_ALL: "ALL",
	DNS_CLASS_IN:  "IN",
	DNS_CLASS_ANY: "ANY",
}
var DnsStrClasses map[string]int = map[string]int{
	//DNS_CLASS_ALL: "ALL",
	"IN":  DNS_CLASS_IN,
	"ANY": DNS_CLASS_ANY,
}
