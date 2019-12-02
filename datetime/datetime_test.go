package datetime

import (
	"fmt"
	"testing"
)

func TestXYZ(t *testing.T) {
	TIME_LAYOUT := "060102150405Z"
	tm, e := ParseTime("190601095044Z", TIME_LAYOUT)
	fmt.Println(tm, e)

}
