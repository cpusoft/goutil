package dnsutil

import (
	"fmt"

	"github.com/cpusoft/goutil/jsonutil"
)

// can not DnsError, must NewDnsError(**,***)
type dnsError struct {
	rCode uint8  // response DSO_RCODE_***
	msg   string // from error.Error()
}

func (c dnsError) Error() string {
	return jsonutil.MarshalJson(c)
}

func NewDnsError(rCode uint8, msg string) *dnsError {
	return &dnsError{rCode: rCode, msg: msg}
}
func (c dnsError) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf(`{"rCode":%d,"msg":"%s"}`, c.rCode, c.msg)
	return []byte(str), nil
}
func (c dnsError) GetRCode() (rCode uint8) {
	return c.rCode
}

func (c dnsError) GetMsg() string {
	return c.msg
}
