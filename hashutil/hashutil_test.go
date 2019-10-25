package hashutil

import (
	"fmt"
	"testing"
)

func TestSha256(t *testing.T) {
	s := []byte{0x01, 0x02, 0x02}
	sh := Sha256(s)
	fmt.Println(sh)
}

func TestSha256File(t *testing.T) {
	s := `G:\Download\92dQUVu3YtGADFrmxsLbJP6VLQI.mft`
	sh, err := Sha256File(s)
	fmt.Println(sh, err)
}
