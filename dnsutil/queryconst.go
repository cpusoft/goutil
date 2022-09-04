package dnsutil

const (
	// dso header: 12bytes: Id(2) + Qr/OpCode/Z/RCode(2) + ZOCOUNT(2) + PRCOUNT(2) + UPCOUNT(2) + ADCOUNT(2)
	QUERY_LENGTH_MIN = 12
)
