package dnsutil

const (
	DSO_MESSAGE                 = 0
	DSO_TYPE_RESERVED           = 0
	DSO_TYPE_KEEPALIVE          = 1
	DSO_TYPE_RETRY_DELAY        = 2
	DSO_TYPE_ENCRYPTION_PADDING = 3
	DSO_TYPE_SUBSCRIBE          = 0x40
	DSO_TYPE_PUSH               = 0x41
	DSO_TYPE_UNSUBSCRIBE        = 0x42
	DSO_TYPE_RECONFIRM          = 0x43

	DSO_TYPE_KEEPALIVE_LENGTH     = 8         // 2*32bit-->2*4byte
	DSO_TYPE_RETRY_DELAY_LENGTH   = 4         // 32bit-->4byte
	DSO_TYPE_SUBSCRIBE_MIN_LENGTH = 2 + 2 + 2 // byte(name)+2*uint16(type+class)
	DSO_TYPE_PUSH_MIN_LENGTH      = 10        // type(2)+class(2)+ttl(4)+rdlen(2)
	DSO_TYPE_RECONFIRM_MIN_LENGTH = 8         // type(2)+class(2)+ttl(4)
	DSO_TYPE_UNSUBSCRIBE_LENGTH   = 2         // 16bit

	DSO_ERROR_IGNORE        = -1
	DSO_ERROR_CLOSE_CONNECT = -2

	// dso header: 4bytes:DSO-TYPE(2)+DSO-LENGTH(2)
	DSO_LENGTH_MIN     = 4
	DSO_TLV_LENGTH_MIN = 2

	DSO_SESSION_STATE_DISCONNECTED          = "session_disconnected"
	DSO_SESSION_STATE_CONNECTING            = "session_connecting"
	DSO_SESSION_STATE_CONNECTED_SESSIONLESS = "session_connected_sessionless"
	DSO_SESSION_STATE_ESTABLISHING_SESSION  = "session_establishing_session"
	DSO_SESSION_STATE_ESTABLISHED_SESSION   = "session_established_session"

	// RFC8490  6.2.DSO Session Timeouts
	// 15s
	DSO_DEFAULT_INACTIVITY_TIMEOUT_SECONDS = 15
	DSO_DEFAULT_KEEPALIVE_INTERVAL_SECONDS = 15

	DSO_MIN_INACTIVITY_TIMEOUT_SECONDS = 10
	DSO_MIN_KEEPALIVE_INTERVAL_SECONDS = 10

	DSO_MAX_INACTIVITY_TIMEOUT_SECONDS = 0xffffffff
	DSO_MAX_KEEPALIVE_INTERVAL_SECONDS = 0xffffffff

	DSO_ADD_RECOURCE_RECORD_MAX_TTL        = 0x7FFFFFFF
	DSO_DEL_SPECIFIED_RESOURCE_RECORD_TTL  = 0xFFFFFFFF
	DSO_DEL_COLLECTIVE_RESOURCE_RECORD_TTL = 0xFFFFFFFE
)

var DsoIntTypes map[uint8]string = map[uint8]string{
	DSO_TYPE_KEEPALIVE:          "keepalive",
	DSO_TYPE_RETRY_DELAY:        "retrydelay",
	DSO_TYPE_ENCRYPTION_PADDING: "encryptionpadding",
	DSO_TYPE_SUBSCRIBE:          "subscribe",
	DSO_TYPE_PUSH:               "push",
	DSO_TYPE_UNSUBSCRIBE:        "unsubscribe",
	DSO_TYPE_RECONFIRM:          "reconfirm",
}
