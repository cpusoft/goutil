package belogs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogMsg_OldStyleFormat(t *testing.T) {
	lg := &LogMsg{
		Level:      LevelDebug,
		Msg:        "Hello, world",
		When:       time.Date(2020, 9, 19, 20, 12, 37, 9, time.UTC),
		FilePath:   "/user/home/main.go",
		LineNumber: 13,
		Prefix:     "Cus",
	}
	res := lg.OldStyleFormat()
	assert.Equal(t, "[D] Cus Hello, world", res)

	lg.enableFuncCallDepth = true
	res = lg.OldStyleFormat()
	assert.Equal(t, "[D] [main.go:13] Cus Hello, world", res)

	lg.enableFullFilePath = true

	res = lg.OldStyleFormat()
	assert.Equal(t, "[D] [/user/home/main.go:13] Cus Hello, world", res)

	lg.Msg = "hello, %s"
	lg.Args = []interface{}{"world"}
	assert.Equal(t, "[D] [/user/home/main.go:13] Cus hello, world", lg.OldStyleFormat())
}
