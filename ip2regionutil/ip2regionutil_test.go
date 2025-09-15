package ip2regionutil

import (
	"fmt"
	"testing"
)

func TestSearchIp2Region(t *testing.T) {
	f := `./ip2region_v4.xdb`
	ip := `93.184.215.14 `
	r, e := SearchIp2Region(f, ip)
	fmt.Println(r, e)
}
