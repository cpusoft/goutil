package belogs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccessLog_format(t *testing.T) {
	alc := &AccessLogRecord{
		RequestTime: time.Date(2020, 9, 19, 21, 21, 21, 11, time.UTC),
	}

	res := alc.format(apacheFormat)
	println(res)
	assert.Equal(t, " - - [19/Sep/2020 09:21:21] \" 0 0\" 0.000000  ", res)

	res = alc.format(jsonFormat)
	assert.Equal(t,
		"{\"remote_addr\":\"\",\"request_time\":\"2020-09-19T21:21:21.000000011Z\",\"request_method\":\"\",\"request\":\"\",\"server_protocol\":\"\",\"host\":\"\",\"status\":0,\"body_bytes_sent\":0,\"elapsed_time\":0,\"http_referrer\":\"\",\"http_user_agent\":\"\",\"remote_user\":\"\"}\n", res)

	AccessLog(alc, jsonFormat)
}
