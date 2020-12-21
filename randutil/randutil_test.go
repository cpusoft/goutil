package randutil

import (
	"fmt"
	"testing"
)

func TestIntn(t *testing.T) {
	a := Intn(999999)
	b := fmt.Sprintf("%s?%d", "http://www.aaa.com/", Intn(999))
	fmt.Println(a, b)
}
