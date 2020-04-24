package executil

import (
	"fmt"
	"testing"
)

func TestExecCommand(t *testing.T) {

	params := []string{"/C", "dir", "/a"}
	fmtShow := true

	ss, err := ExecCommand("cmd", params, fmtShow)

	fmt.Println(err)
	for i := range ss {
		fmt.Println(ss[i])
	}

}
