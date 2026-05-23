package loggoroutineutil

import (
	"runtime"
	"strings"
)

// 返回 goroutine 摘要信息，由调用方决定如何打印
//
//	byteLen: 2*1024*1024
func GetGoroutineSummary(byteLen uint64) (total int, businessCount int, detail string) {
	total = runtime.NumGoroutine()

	// 2MB buffer 通常足够，极端场景下如果不够会自动截断
	buf := make([]byte, byteLen)
	n := runtime.Stack(buf, true)
	stacks := string(buf[:n])

	goroutines := strings.Split(stacks, "\n\n")

	var b strings.Builder
	b.Grow(4096)
	b.WriteString("Key goroutines:\n")

	biz := 0
	for _, g := range goroutines {
		if g == "" {
			continue
		}
		lines := strings.Split(g, "\n")
		if len(lines) == 0 {
			continue
		}

		header := lines[0]
		if isSystemGoroutine(header, lines) {
			continue
		}
		biz++

		frames := extractTopFrames(lines, 3)
		b.WriteString(header)
		b.WriteString("\n")
		for _, f := range frames {
			b.WriteString("  ")
			b.WriteString(f)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return total, biz, b.String()
}

// 判断是否为系统/运行时 goroutine
func isSystemGoroutine(header string, lines []string) bool {
	systemPatterns := []string{
		"runtime.goexit",
		"runtime.gcBgMarkWorker",
		"runtime.bgsweep",
		"runtime.bgscavenge",
		"runtime.forcegchelper",
		"database/sql.(*DB).connectionOpener",
		"database/sql.(*DB).connectionCleaner",
		"net/http.(*Server).Serve",
		"internal/poll.runtime_pollWait",
		"syscall.Syscall",
	}

	for _, p := range systemPatterns {
		for _, line := range lines {
			if strings.Contains(line, p) {
				return true
			}
		}
	}
	return false
}

// 提取栈顶 N 帧
func extractTopFrames(lines []string, n int) []string {
	var frames []string
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "\t") || strings.HasPrefix(line, " ") {
			frames = append(frames, strings.TrimSpace(line))
			if len(frames) >= n {
				break
			}
		}
	}
	return frames
}

// byteLen= 10*1024*1024
func GetGoroutine(byteLen uint64) string {
	buf := make([]byte, byteLen)
	stackLen := runtime.Stack(buf, true) // true = 打印所有 goroutine
	return string(buf[:stackLen])

}
