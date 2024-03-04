package randutil

import (
	"fmt"
	"testing"
)

func TestIntn(t *testing.T) {
	a := Intn(2)
	b := fmt.Sprintf("%s?%d", "http://www.aaa.com/", Intn(1))
	fmt.Println(a, b)
}
