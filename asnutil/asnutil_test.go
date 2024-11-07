package asnutil

import (
	"fmt"
	"testing"
)

func TestGetAsnOwnerByCymru(t *testing.T) {
	r, err := GetAsnOwnerByCymru(265699)
	fmt.Println(r, err)
}
