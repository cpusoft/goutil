
package belogs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJLWriter_Format(t *testing.T) {
	lg := &LogMsg{
		Level:      LevelDebug,
		Msg:        "Hello, world",
		When:       time.Date(2020, 9, 19, 20, 12, 37, 9, time.UTC),
		FilePath:   "/user/home/main.go",
		LineNumber: 13,
		Prefix:     "Cus",
	}
	jl := newJLWriter().(*JLWriter)
	res := jl.Format(lg)
	assert.Equal(t, "2020-09-19 20:12:37 [D] Cus Hello, world", res)
}
