package xormdb

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

// SQL Null types utils
func SqlNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func SqlNullInt(s int64) sql.NullInt64 {
	if s < 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{
		Int64: s,
		Valid: true,
	}
}

// StringArray 自定义类型：[]string <-> JSON字符串（数据库存储）
type StringArray []string

// FromDB 从数据库JSON字符串转换为[]string
func (s *StringArray) FromDb(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*s = []string{}
		return nil
	}
	var arr []string
	if err := json.Unmarshal(b, &arr); err != nil {
		belogs.Error("StringArray.FromDb(): fail, b:", jsonutil.MarshalJson(b), err)
		return fmt.Errorf("unmarshal json fail: %w", err)
	}
	*s = arr
	return nil
}

// ToDb 将[]string转换为数据库JSON字符串
func (s StringArray) ToDb() ([]byte, error) {
	if len(s) == 0 {
		return []byte("[]"), nil
	}
	data, err := json.Marshal(s)
	if err != nil {
		belogs.Error("StringArray.ToDb(): fail, s:", jsonutil.MarshalJson(s), err)
		return nil, fmt.Errorf("marshal json fail: %w", err)
	}
	return data, nil
}
