package executil

import (
	"fmt"
	"strings"
	"testing"
)

func TestExecCommand(t *testing.T) {

	params := []string{"ca", "-gencrl", "-verbose", "-out", "/home/rpki/gencerts/ripencc/subcert/tmp/test.crl", "", "-cert", "/home/rpki/gencerts/ripencc/subca/ripencc_subca.pem", "-keyfile", "/home/rpki/gencerts/ripencc/subca/ripencc_subca.key", "-config", "/home/rpki/gencerts/ripencc/subcert/crl.cnf", "-crl_lastupdate", "241011010203Z", "-crl_nextupdate", "341011010203Z"}
	fmtShow := true

	ss, err := ExecCommandStdoutPipe("openssl", params, fmtShow)

	fmt.Println(err)
	for i := range ss {
		fmt.Println(ss[i])
	}

}

func TestExecCommandCombinedOutput(t *testing.T) {
	p := `ca -gencrl -verbose  -out /home/rpki/gencerts/ripencc/subcert/tmp/test.crl  -cert /home/rpki/gencerts/ripencc/subca/ripencc_subca.pem -keyfile /home/rpki/gencerts/ripencc/subca/ripencc_subca.key -config /home/rpki/gencerts/ripencc/subcert/crl.cnf -crl_lastupdate 241011010203Z -crl_nextupdate 341011010203Z`
	params := strings.Split(p, " ")
	out, err := ExecCommandCombinedOutput("openssl", params)
	fmt.Println("out", out)
	fmt.Println(err)

}
