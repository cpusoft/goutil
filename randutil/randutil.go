package randutil

import (
	"math/rand"
	"time"
)

func Intn(n uint) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(int(n))
}

func IntRange(min, n uint) int {
	rand.Seed(time.Now().Unix())
	return int(min) + rand.Intn(int(n))
}
