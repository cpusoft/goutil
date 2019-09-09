package executil

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func ExecCommand(commandName string, params []string, ftmShow bool) (contentArray []string, err error) {

	var line string
	contentArray = make([]string, 0)
	cmd := exec.Command(commandName, params...)
	//显示运行的命令
	if ftmShow {
		fmt.Printf("exec:%s\n", strings.Join(cmd.Args[:], " "))
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		//belogs.Error("execCommand(): commandName:", commandName, "  err:", err)
		if ftmShow {
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
		if ftmShow {
			fmt.Println(line)
		}
		//belogs.Debug("execCommand(): line:", line)
		contentArray = append(contentArray, line)
	}

	cmd.Wait()
	return contentArray, nil
}
