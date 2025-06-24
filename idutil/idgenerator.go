package idutil

import (
	"math/rand"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/cpusoft/goutil/belogs"
)

/*
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
*/
func GenerateSnowflakInt64(nodeId int64) (int64, error) {
	return GenerateInt64(nodeId)
}

// Deprecated: Use GenerateSnowflakInt64 instead.
func GenerateInt64(nodeId int64) (int64, error) {
	node, err := snowflake.NewNode(nodeId)
	if err != nil {
		belogs.Error("GenerateInt64(): nodeId:", nodeId, err)
		return 0, err
	}
	return node.Generate().Int64(), nil
}

func GenerateSnowflakeString(nodeId int64) (string, error) {
	return GenerateString(nodeId)
}

// Deprecated: Use GenerateSnowflakeString instead.
func GenerateString(nodeId int64) (string, error) {
	node, err := snowflake.NewNode(nodeId)
	if err != nil {
		belogs.Error("GenerateString(): nodeId:", nodeId, err)
		return "", err
	}
	return node.Generate().String(), nil
}

// GenerateSequentialUint64UUID generates a uint64 based on timestamp and randomness.
func GenerateSequentialUint64UUID() uint64 {
	ts := uint64(time.Now().UnixNano())  // Current time in nanoseconds
	r := uint64(rand.Int63())            // Random 63-bit number
	return (ts << 32) | (r & 0xFFFFFFFF) // Combine time and random bits
}
