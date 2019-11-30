package rrdputil

import (
	"fmt"
	"sort"
	"testing"
)

func TestSort(t *testing.T) {
	v2 := []NotificationDelta{{Serial: 1, Uri: "3", Hash: "333"},
		{Serial: 0, Uri: "6", Hash: "666"},
		{Serial: 3, Uri: "2", Hash: "2222"},
		{Serial: 8, Uri: "7", Hash: "7777"}}
	fmt.Println(v2)
	sort.Sort(NotificationDeltas(v2))
	fmt.Println(v2)
}
