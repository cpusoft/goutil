package urlutil

import (
	"fmt"
	"testing"
)

func TestHost1(t *testing.T) {
	url := `//1.2.3.4:33`
	u, err := Host(url)
	fmt.Println(u, err)
}
func TestHostAndPathFile(t *testing.T) {
	url := `rsync://rpki.apnic.net:999/member_repository/A91270E6/75648ECED63511E896631322C4F9AE02/dVNRzYJvKfhxtLyVlPTpSNvnc-k.mft?aa=bbb`
	u, err := HostAndPathFile(url)
	fmt.Println(u, err)
}
func TestHostAndPathAndFile(t *testing.T) {
	url := `rsync://rpki.apnic.net:999/member_repository/A91270E6/75648ECED63511E896631322C4F9AE02/dVNRzYJvKfhxtLyVlPTpSNvnc-k.mft?aa=bbb`
	host, path, file, err := HostAndPathAndFile(url)
	fmt.Println(host, path, file, err)
}

func TestHost(t *testing.T) {
	url := `rsync://rpki.apnic.net:999/member_repository/A91270E6/75648ECED63511E896631322C4F9AE02/dVNRzYJvKfhxtLyVlPTpSNvnc-k.mft?aa=bbb`
	u, err := Host(url)
	fmt.Println(u, err)
}

func TestPath(t *testing.T) {
	url := `rsync://rpki.apnic.net:999/member_repository/A91270E6/75648ECED63511E896631322C4F9AE02/dVNRzYJvKfhxtLyVlPTpSNvnc-k.mft?aa=bbb`
	u, err := Path(url)
	fmt.Println(u, err)
}
