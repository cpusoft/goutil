package rsyncutil

import (
	"fmt"
	"testing"
)

// destpath=G:\Download\cert\rsync
// logpath=G:\Download\cert\log
func TestRsyncToLogFile(t *testing.T) {
	rsyncUrl := "http://rpki.apnic.net/repository/"
	destPath := "/tmp/cer/"
	logPath := "/tmp/log/"
	rsyncDestPath, rsyncLogFile, err := RsyncToLogFile(rsyncUrl, destPath, logPath)
	fmt.Println(rsyncDestPath, rsyncLogFile, err)
}

func TestRsyncTestConnect(t *testing.T) {
	err := RsyncTestConnect("rsync://rpki-repo.as207960.net/repo")
	fmt.Println(err)

}
