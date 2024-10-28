package opensslutil

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/cpusoft/goutil/belogs"
)

func TestOpenssl(t *testing.T) {
	cmd := exec.Command("openssl", "version")
	ldLibraryPath := `/home/openssl/openssl/lib64`
	path := `/home/openssl/openssl/bin:$PATH`
	if len(ldLibraryPath) > 0 && len(path) > 0 {
		cmd.Env = append(os.Environ(), "LD_LIBRARY_PATH="+ldLibraryPath)
		cmd.Env = append(os.Environ(), "PATH="+path)
		belogs.Debug("GetResultsByOpensslX509(): ldLibraryPath:", ldLibraryPath, "  path:", path)
	}
	output, err := cmd.CombinedOutput()
	fmt.Println(output, err)
}
