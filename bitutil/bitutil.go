package bitutil

/*
// 0x01(0000 0001) --> left 2 --> 0x(0000 0111)
// rfc3779 max : usine |, right bits are one
func LeftAndFillOne(bits uint8) (a uint8) {
	a = 1
	for i := uint8(0); i < bits; i++ {
		a = a | a<<1
	}
	return a
}

// 0x01(1111 1111) --> left 6 --> 0x(1100 0000)
// rfc3779 min : usine &, right bits are zero
func LeftAndFillZero(bits uint8) (a uint8) {
	a = 0xff
	for i := uint8(0); i < bits; i++ {
		a = a << 1
	}
	return a
}
*/

// Shift0x00LeftFillOne 将 0x01 左移指定位数并填充右侧为 1
// 例如: bits=2 -> 0x03 (0000 0011)
// 修复点：
// 1. 增加 bits 边界校验，限制 0 <= bits <= 8
// 2. 处理 bits=0 的特殊情况（返回 0，符合位操作逻辑）
// 3. 保留原函数核心计算逻辑
func Shift0x00LeftFillOne(bits uint8) (a byte) {
	// 边界校验：uint8 最多 8 位
	if bits == 0 {
		return 0 // 0 位移位返回 0，符合位操作常识
	}
	if bits > 8 {
		bits = 8 // 超过 8 位按 8 位处理，避免溢出
	}

	// 保留原函数核心逻辑
	a = 1
	for i := uint8(0); i < bits-1; i++ {
		a = a | a<<1
	}
	return a
}

// Shift0xffLeftFillZero 将 0xff 左移指定位数并填充右侧为 0
// 例如: bits=6 -> 0xc0 (1100 0000)
// 修复点：
// 1. 增加 bits 边界校验，限制 0 <= bits <= 8
// 2. 处理 bits>8 的情况（返回 0，符合 uint8 移位特性但更可控）
// 3. 保留原函数核心计算逻辑
func Shift0xffLeftFillZero(bits uint8) (a byte) {
	// 边界校验
	if bits == 0 {
		return 0xff // 0 位移位返回原值，符合位操作常识
	}
	if bits > 8 {
		return 0 // 超过 8 位返回 0，避免无意义的移位
	}

	// 保留原函数核心逻辑
	a = 0xff
	a = a << bits
	return a
}
