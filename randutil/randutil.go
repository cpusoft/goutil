package randutil

import (
	"math/rand"
	"time"
)

func Intn(n uint) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(int(n))
}
