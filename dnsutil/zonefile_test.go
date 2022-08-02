package dnsutil

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestLoadFromZoneFile(t *testing.T) {
	f := `mydomain.com.zone`
	o, err := LoadFromZoneFile(f)
	fmt.Println(jsonutil.MarshalJson(o), err)

	f1 := `tmp.zone`
	err = SaveToZoneFile(o, f1)
	fmt.Println(jsonutil.MarshalJson(o), err)
}
