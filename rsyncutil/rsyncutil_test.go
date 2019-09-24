package rsyncutil

import (
	"fmt"
	"testing"
)

func TestRsyncLocalIncludeFileExt(t *testing.T) {
	srcPath := "/root/rpki/repo/repo"
	destPath := "/root/rpki/repo/repo-1"
	includeFileExt = "*.roa"
	err := RsyncLocalIncludeFileExt(srcPath, destPath, includeFileExt)
	fmt.Println(err)
}
