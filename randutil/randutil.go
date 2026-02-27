package randutil

import (
	"math/rand"
	"sync"
	"time"
)

// 包级别的全局随机数生成器，仅初始化一次，保证随机性和性能
var (
	rng   *rand.Rand
	once  sync.Once  // 保证rng只初始化一次
	mutex sync.Mutex // 保证并发安全
)

// 初始化随机数生成器，仅执行一次
func initRng() {
	source := rand.NewSource(time.Now().UnixNano())
	rng = rand.New(source)
}

// Intn 返回 [0, n) 范围内的随机整数，保持原参数/返回值逻辑
func Intn(n uint) int {
	// 确保随机数生成器只初始化一次
	once.Do(initRng)

	// 加锁保证并发安全
	mutex.Lock()
	defer mutex.Unlock()

	return rng.Intn(int(n))
}

// IntRange 返回 [min, min+n) 范围内的随机整数，保持原参数/返回值逻辑
func IntRange(min, n uint) int {
	return int(min) + Intn(n)
}
