package belogs

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type CustomFormatter struct{}

func (c *CustomFormatter) Format(lm *LogMsg) string {
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

func (t *TestLogger) Destroy() {
	panic("implement me")
}

func (t *TestLogger) Flush() {
	panic("implement me")
}

func (t *TestLogger) SetFormatter(f LogFormatter) {
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
	when := time.Now()
	lm := &LogMsg{
		Msg:        "message",
		FilePath:   "/User/go/beego/main.go",
		Level:      LevelWarn,
		LineNumber: 10,
		When:       when,
	}
	got := tes.ToString(lm)
	want := lm.FilePath + ":" + strconv.Itoa(lm.LineNumber) + "|" +
		when.Format(tes.WhenFormat) + levelPrefix[lm.Level] + ">> " + lm.Msg
	if got != want {
		t.Errorf("want %s, got %s", want, got)
	}
}
