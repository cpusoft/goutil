package dnsutil

import (
	"github.com/cpusoft/goutil/jsonutil"
)

// can not DnsError, must NewDnsError(**,***)
type dnsError struct {
	Msg string `json:"msg"` // from error.Error()

	Id                uint16 `json:"id"` // from dns id/messageId
	OpCode            uint8  `json:"opCode"`
	RCode             uint8  `json:"rCode"`             // response DSO_RCODE_***
	NextConnectPolicy int    `json:"nextConnectPolicy"` //	tcptlsutil.NEXT_CONNECT_POLICY_***
}

func (c dnsError) Error() string {
	return jsonutil.MarshalJson(c)
}

func NewDnsError(msg string, id uint16, opCode uint8, rCode uint8, nextConnectPolicy int) *dnsError {
	return &dnsError{
		Msg:               msg,
		Id:                id,
		OpCode:            opCode,
		RCode:             rCode,
		NextConnectPolicy: nextConnectPolicy,
	}
}

/*
func (c dnsError) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf(`{"rCode":%d,"msg":"%s"}`, c.rCode, c.msg)
	return []byte(str), nil
}


func GetRCode(err error) (rCode uint8) {
	if e, ok := err.(dsoError); ok {
		return e.rCode
	}
	return dnsutil.DNS_RCODE_NOERROR
}
func GetMsg(err error) string {
	if e, ok := err.(dsoError); ok {
		return e.msg
	}
	return ""
}
func GetNextConnectPolicy(err error) int {
	if e, ok := err.(dsoError); ok {
		return e.nextConnectPolicy
	}
	return tcptlsutil.NEXT_CONNECT_POLICY_KEEP
}
*/
