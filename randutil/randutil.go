package randutil

import (
	"math/rand"
	"time"
)

func Intn(n uint) int {
	//rand.Seed(time.Now().Unix())
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)
	return rng.Intn(int(n))
}

func IntRange(min, n uint) int {
	return int(min) + Intn(n)
}
