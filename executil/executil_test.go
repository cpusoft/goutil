package executil

import (
	"fmt"
	"testing"
)

func TestExecCommand(t *testing.T) {

	params := []string{"/C", "dir", "/a"}
	fmtShow := true

	ss, err := ExecCommandStdoutPipe("cmd", params, fmtShow)

	fmt.Println(err)
	for i := range ss {
		fmt.Println(ss[i])
	}

}

func TestExecCommandCombinedOutput(t *testing.T) {
	params := []string{"/C", "dir", "/a"}
	out, err := ExecCommandCombinedOutput("cmd", params)
	fmt.Println("out", out)
	fmt.Println(err)

}
