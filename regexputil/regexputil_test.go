package regexputil

import (
	"fmt"
	"testing"
)

func TestIsHex(t *testing.T) {
	ssss := `10d0c9f4328576d51cc73c042cfc15e9b3d6378`
	b, err := IsHex(ssss)
	fmt.Println(b, err)
}

func TestCheckRpkiFileName(t *testing.T) {
	ssss := `ZoN_KCuLgQ_XLZREqsT884kSssE.aaa`
	b := CheckRpkiFileName(ssss)
	fmt.Println(b)
}
