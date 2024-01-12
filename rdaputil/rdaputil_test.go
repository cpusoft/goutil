package rdaputil

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestRdap(t *testing.T) {
	d, err := RdapDomain("baidu.com")
	fmt.Println(jsonutil.MarshalJson(d), err)

	a, err := RdapAsn(uint64(2846))
	fmt.Println(jsonutil.MarshalJson(a), err)

	i, err := RdapAddressPrefix("8.8.8.0/24")
	fmt.Println(jsonutil.MarshalJson(i), err)

}
