package zonefileutil

import (
	"fmt"
	"testing"

	"github.com/guregu/null"
)

func TestLoadZoneFile(t *testing.T) {
	file := `mydomain.com.zone`
	zf, err := LoadZoneFile(file)
	fmt.Println("now:\n"+zf.String(), err)

	//DelResourceRecord(zf, rr)
	//fmt.Println("del\n"+zf.String(), err)

	afterV := []string{"101.228.10.127"}
	afterR := ResourceRecord{RrDomain: "test", RrType: "A", RrValues: afterV}
	newV := []string{"101.228.10.128"}
	newR := ResourceRecord{RrDomain: "", RrType: "A", RrTtl: null.IntFrom(600), RrValues: newV}
	AddResourceRecord(zf, afterR, newR)
	fmt.Println("add\n"+zf.String(), err)

	oldV := []string{"101.228.10.127"}
	oldR := ResourceRecord{RrDomain: "test", RrType: "A", RrValues: oldV}
	newV1 := []string{"101.228.10.129"}
	newR1 := ResourceRecord{RrDomain: "test", RrType: "A", RrTtl: null.IntFrom(500), RrValues: newV1}
	UpdateResourceRecord(zf, oldR, newR1)
	fmt.Println("update\n", zf.String(), err)
}
