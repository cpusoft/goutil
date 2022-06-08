package zonefileutil

import (
	"fmt"
	"testing"
)

func TestDeepCopy(t *testing.T) {
	v := []string{"101.228.10.127", "101.228.10.128", "101.228.10.129"}
	r := ResourceRecord{RrName: "test", RrType: "A", RrValues: v}

	nr := deepcopyResourceRecord(&r)
	fmt.Println(nr)

}
