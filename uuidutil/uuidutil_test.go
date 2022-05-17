package uuidutil

import (
	"fmt"
	"testing"
)

func TestGetUuid(t *testing.T) {
	u := GetUuid()
	fmt.Println(u)
}
