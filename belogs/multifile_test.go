package belogs

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestFiles_1(t *testing.T) {
	log := NewLogger(10000)
	log.SetLogger("multifile", `{"filename":"test.log","separate":["emergency", "alert", "critical", "error", "warning", "notice", "info", "debug"]}`)
	log.Debug("debug")
	log.Informational("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("error")
	log.Alert("alert")
	log.Critical("critical")
	log.Emergency("emergency")
	fns := []string{""}
	fns = append(fns, levelNames[0:]...)
	name := "test"
	suffix := ".log"
	for _, fn := range fns {

		file := name + suffix
		if fn != "" {
			file = name + "." + fn + suffix
		}
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		b := bufio.NewReader(f)
		lineNum := 0
		lastLine := ""
		for {
			line, _, err := b.ReadLine()
			if err != nil {
				break
			}
			if len(line) > 0 {
				lastLine = string(line)
				lineNum++
			}
		}
		expected := 1
		if fn == "" {
			expected = LevelDebug + 1
		}
		if lineNum != expected {
			t.Fatal(file, "has", lineNum, "lines not "+strconv.Itoa(expected)+" lines")
		}
		if lineNum == 1 {
			if !strings.Contains(lastLine, fn) {
				t.Fatal(file + " " + lastLine + " not contains the log msg " + fn)
			}
		}
		os.Remove(file)
	}
}
