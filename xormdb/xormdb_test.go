package xormdb

import (
	"fmt"
	"testing"
)

func TestInt64sToInString(t *testing.T) {
	s := []int64{2, 3, 4, 5, 6, 7, 33}
	ss := Int64sToInString(s)
	fmt.Println(ss)
}
