package executil

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/cpusoft/goutil/belogs"
)

func ExecCommandCombinedOutput(commandName string, params []string) (out string, err error) {
	result := exec.Command(commandName, params...)
	b, err := result.CombinedOutput()
	if err != nil {
		belogs.Error("ExecCommandCombinedOutput(): CombinedOutput fail, commandName:", commandName,
			"   params:", params, err)
		return "", err
	}
	out = string(b)
	belogs.Debug("ExecCommandCombinedOutput(): CombinedOutput , commandName:", commandName,
		"   params:", params, "  out:", out)
	return out, nil
}

func ExecCommandStdoutPipe(commandName string, params []string, fmtShow bool) (contentArray []string, err error) {

	var line string
	contentArray = make([]string, 0)
	cmd := exec.Command(commandName, params...)
	//显示运行的命令
	if fmtShow {
		fmt.Printf("exec:%s\n", strings.Join(cmd.Args[:], " "))
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		//belogs.Error("execCommand(): commandName:", commandName, "  err:", err)
		if fmtShow {
			fmt.Fprintln(os.Stderr, "error=>", err.Error())
		}
		return contentArray, err
	}

	cmd.Start()
	reader := bufio.NewReader(stdout)
	for {
		tmp, _, err2 := reader.ReadLine()
		line = string(tmp)
		//line, err2 := reader.ReadString('\n') //[]byte(osutil.GetNewLineSep())[0])
		if err2 != nil || io.EOF == err2 {
			//belogs.Error("execCommand(): ReadString(): line: ", line, "  err2:", err2)
			break
		}
		if fmtShow {
			fmt.Println(line)
		}
		//belogs.Debug("execCommand(): line:", line)
		contentArray = append(contentArray, line)
	}

	cmd.Wait()
	return contentArray, nil
}

// Deprecated: using ExecCommandStdoutPipe
func ExecCommand(commandName string, params []string, fmtShow bool) (contentArray []string, err error) {
	return ExecCommandStdoutPipe(commandName, params, fmtShow)
}
