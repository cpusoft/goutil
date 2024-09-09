package ip2regionutil

import (
	"fmt"
	"testing"
)

func TestSearchIp2Region(t *testing.T) {
	f := `F:\share\我的坚果云\Go\dns\research\ip2region\data\ip2region.xdb`
	ip := `114.114.114.114`
	r, e := SearchIp2Region(f, ip)
	fmt.Println(r, e)
}
