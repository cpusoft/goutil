package datetime

import (
	"fmt"
	"testing"
	"time"
)

func TestXYZ(t *testing.T) {
	TIME_LAYOUT := "060102150405Z"
	tm, e := ParseTime("190601095044Z", TIME_LAYOUT)
	fmt.Println(tm, e)

}

func TestAddDataByDuration(t *testing.T) {
	newT, err := AddDataByDuration(time.Now(), "12m")
	fmt.Println(newT, err)

	newT, err = AddDataByDuration(time.Now(), "30d")
	fmt.Println(newT, err)

	newT, err = AddDataByDuration(time.Now(), "1y")
	fmt.Println(newT, err)
}
