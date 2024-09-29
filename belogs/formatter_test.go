package belogs

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type CustomFormatter struct{}

func (*CustomFormatter) Format(lm *LogMsg) string {
	return "hello, msg: " + lm.Msg
}

type TestLogger struct {
	Formatter string `json:"formatter"`
	Expected  string
	formatter LogFormatter
}

func (t *TestLogger) Init(config string) error {
	er := json.Unmarshal([]byte(config), t)
	t.formatter, _ = GetFormatter(t.Formatter)
	return er
}

func (t *TestLogger) WriteMsg(lm *LogMsg) error {
	msg := t.formatter.Format(lm)
	if msg != t.Expected {
		return errors.New("not equal")
	}
	return nil
}

func (*TestLogger) Destroy() {
	panic("implement me")
}

func (*TestLogger) Flush() {
	panic("implement me")
}

func (*TestLogger) SetFormatter(_ LogFormatter) {
	panic("implement me")
}

func TestCustomFormatter(t *testing.T) {
	RegisterFormatter("custom", &CustomFormatter{})
	tl := &TestLogger{
		Expected: "hello, msg: world",
	}
	assert.Nil(t, tl.Init(`{"formatter": "custom"}`))
	assert.Nil(t, tl.WriteMsg(&LogMsg{
		Msg: "world",
	}))
}

func TestPatternLogFormatter(t *testing.T) {
	tes := &PatternLogFormatter{
		Pattern:    "%F:%n|%w%t>> %m",
		WhenFormat: "2006-01-02",
	}
	when, _ := time.Parse(tes.WhenFormat, "2022-04-17")
	testCases := []struct {
		msg  *LogMsg
		want string
	}{
		{
			msg: &LogMsg{
				Msg:        "hello %s",
				FilePath:   "/User/go/beego/main.go",
				Level:      LevelWarn,
				LineNumber: 10,
				When:       when,
				Args:       []interface{}{"world"},
			},
			want: "/User/go/beego/main.go:10|2022-04-17[W]>> hello world",
		},
		{
			msg: &LogMsg{
				Msg:        "hello",
				FilePath:   "/User/go/beego/main.go",
				Level:      LevelWarn,
				LineNumber: 10,
				When:       when,
			},
			want: "/User/go/beego/main.go:10|2022-04-17[W]>> hello",
		},
		{
			msg:  &LogMsg{},
			want: ":0|0001-01-01[M]>> ",
		},
	}

	for _, tc := range testCases {
		got := tes.ToString(tc.msg)
		assert.Equal(t, tc.want, got)
	}
}
