package dnsutil

import (
	"fmt"

	"github.com/cpusoft/goutil/belogs"
)

// must be lower, so must NewDnsError(**,***), not allow dnsError{***}
type dnsError struct {
	msg string // from error.Error()

	id                uint16 // from dns id/messageId
	opCode            uint8
	rCode             uint8 // response DSO_RCODE_***
	nextConnectPolicy int   //	tcptlsutil.NEXT_CONNECT_POLICY_***
}

func (c dnsError) Error() string {
	return fmt.Sprintf(`{"msg":"%s","id":%d,"opCode":%d,"rCode":%d,"nextConnectPolicy":%d}`,
		c.msg, c.id, c.opCode, c.rCode, c.nextConnectPolicy)
}

/*
	func (c dnsError) MarshalJSON() ([]byte, error) {
		fmt.Sprintf(`{"msg":"%s","id":%d,"opCode":%d,"rCode":%d,"nextConnectPolicy":%d}`,
			c.msg, c.id, c.opCode, c.rCode, c.nextConnectPolicy)
		belogs.Debug("MarshalJSON(): dnsError:", str)
		return []byte(str), nil
	}
*/
func NewDnsError(msg string, id uint16, opCode uint8, rCode uint8, nextConnectPolicy int) error {
	return dnsError{
		msg:               msg,
		id:                id,
		opCode:            opCode,
		rCode:             rCode,
		nextConnectPolicy: nextConnectPolicy,
	}
}
func SetDnsErrorMsg(err error, msg string) {
	if e, ok := err.(dnsError); ok {
		belogs.Debug("SetDnsErrorMsg(): msg:", msg)
		e.msg = msg
	}

}

func GetDnsErrorMsg(err error) string {
	if e, ok := err.(dnsError); ok {
		return e.msg
	}
	return ""
}
func SetDnsErrorId(err error, id uint16) {
	if e, ok := err.(dnsError); ok {
		e.id = id
	}
}
func GetDnsErrorId(err error) uint16 {
	if e, ok := err.(dnsError); ok {
		return e.id
	}
	return 0
}
func SetDnsErrorOpCode(err error, opCode uint8) {
	if e, ok := err.(dnsError); ok {
		e.opCode = opCode
	}
}
func GetDnsErrorOpCode(err error) (rCode uint8) {
	if e, ok := err.(dnsError); ok {
		return e.opCode
	}
	return DNS_OPCODE_QUERY
}
func SetRCode(err error, rCode uint8) {
	if e, ok := err.(dnsError); ok {
		e.rCode = rCode
	}
}
func GetDnsErrorRCode(err error) (rCode uint8) {
	if e, ok := err.(dnsError); ok {
		return e.rCode
	}
	return DNS_RCODE_NOERROR
}
func SetDnsErrorNextConnectPolicy(err error, nextConnectPolicy int) {
	if e, ok := err.(dnsError); ok {
		e.nextConnectPolicy = nextConnectPolicy
	}
}
func GetDnsErrorNextConnectPolicy(err error) int {
	if e, ok := err.(dnsError); ok {
		return e.nextConnectPolicy
	}
	return 0 //tcptlsutil.NEXT_CONNECT_POLICY_KEEP
}
