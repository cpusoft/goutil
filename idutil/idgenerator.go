package idutil

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"
)

// IDGenerator 结构体
type IDGeneratorInc struct {
	mu sync.Mutex
	id int
}

// NewIDGenerator 创建一个新的ID生成器
func NewIDGeneratorInc() *IDGeneratorInc {
	return &IDGeneratorInc{id: 0}
}

// 自增
func (gen *IDGeneratorInc) Generate() int {
	gen.mu.Lock()
	defer gen.mu.Unlock()
	gen.id++
	return gen.id
}

// IDGeneratorBaseTime 结构体
type IDGeneratorBaseTime struct {
	mu       sync.Mutex
	lastTime int64
	sequence int
}

// NewIDGenerator 创建一个新的ID生成器
func NewIDGeneratorBaseTime() *IDGeneratorBaseTime {
	return &IDGeneratorBaseTime{
		lastTime: time.Now().UnixNano() / int64(time.Millisecond),
		sequence: 0,
	}
}

// Generate 生成一个新的ID
func (gen *IDGeneratorBaseTime) Generate() int64 {
	gen.mu.Lock()
	defer gen.mu.Unlock()

	now := time.Now().UnixNano() / int64(time.Millisecond)

	if now == gen.lastTime {
		gen.sequence++
	} else {
		gen.lastTime = now
		gen.sequence = 0
	}

	id := (now << 20) | int64(gen.sequence)
	return id
}

// 定义全局的 snowflake 节点
var node *snowflake.Node

// 初始化节点
func InitSnowflake(nodeID int64) error {
	var err error
	node, err = snowflake.NewNode(nodeID)
	if err != nil {
		return fmt.Errorf("failed to initialize snowflake node: %w", err)
	}
	return nil
}

// 生成自增 ID 的方法
func GenerateID() snowflake.ID {
	return node.Generate()
}
