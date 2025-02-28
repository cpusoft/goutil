package belogs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	apacheFormatPattern = "%s - - [%s] \"%s %d %d\" %f %s %s"
	apacheFormat        = "APACHE_FORMAT"
	jsonFormat          = "JSON_FORMAT"
	customFormat        = "CUSTOM_FORMAT"
)

var CustomAccessLogFunc func(r *AccessLogRecord) string

// AccessLogRecord is astruct for holding access log data.
type AccessLogRecord struct {
	RemoteAddr     string        `json:"remote_addr"`
	RequestTime    time.Time     `json:"request_time"`
	RequestMethod  string        `json:"request_method"`
	Request        string        `json:"request"`
	ServerProtocol string        `json:"server_protocol"`
	Host           string        `json:"host"`
	Status         int           `json:"status"`
	BodyBytesSent  int64         `json:"body_bytes_sent"`
	ElapsedTime    time.Duration `json:"elapsed_time"`
	HTTPReferrer   string        `json:"http_referrer"`
	HTTPUserAgent  string        `json:"http_user_agent"`
	RemoteUser     string        `json:"remote_user"`
}

func (r *AccessLogRecord) json() ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	disableEscapeHTML(encoder)

	err := encoder.Encode(r)
	return buffer.Bytes(), err
}

func disableEscapeHTML(i interface{}) {
	if e, ok := i.(interface {
		SetEscapeHTML(bool)
	}); ok {
		e.SetEscapeHTML(false)
	}
}

// AccessLog - Format and print access log.
func AccessLog(r *AccessLogRecord, format string) {
	msg := r.format(format)
	lm := &LogMsg{
		Msg:   strings.TrimSpace(msg),
		When:  time.Now(),
		Level: levelLoggerImpl,
	}
	beeLogger.writeMsg(lm)
}

func (r *AccessLogRecord) format(format string) string {
	msg := ""
	switch format {
	case apacheFormat:
		timeFormatted := r.RequestTime.Format("02/Jan/2006 03:04:05")
		msg = fmt.Sprintf(apacheFormatPattern, r.RemoteAddr, timeFormatted, r.Request, r.Status, r.BodyBytesSent,
			r.ElapsedTime.Seconds(), r.HTTPReferrer, r.HTTPUserAgent)
	case customFormat:
		if CustomAccessLogFunc != nil {
			return CustomAccessLogFunc(r)
		}
		fallthrough
	case jsonFormat:
		fallthrough
	default:
		jsonData, err := r.json()
		if err != nil {
			msg = fmt.Sprintf(`{"Error": "%s"}`, err)
		} else {
			msg = string(jsonData)
		}
	}
	return msg
}
