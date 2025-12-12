package xormdb

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/stringutil"
)

// ////////////////////////////////
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

// Deprecated: should using stringutil.Int64sToInString
func Int64sToInString(s []int64) string {
	return stringutil.Int64sToInString(s)
}

// /////////////////////////////////
// StringArray is a custom type for storing []string as JSON in the database
type StringArray []string

// FromDB 从数据库JSON字符串转换为[]string
func (s *StringArray) FromDb(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*s = []string{}
		return nil
	}
	// 解析JSON数组为[]string
	var arr []string
	if err := json.Unmarshal(b, &arr); err != nil {
		belogs.Error("StringArray.FromDb(): fail, b:", jsonutil.MarshalJson(b), err)
		return fmt.Errorf("unmarshal json fail: %w", err)
	}
	*s = arr
	return nil
}

// ToDB 将[]string转换为数据库JSON字符串
func (s StringArray) ToDb() ([]byte, error) {
	if len(s) == 0 {
		return []byte("[]"), nil
	}
	// 序列化为JSON字符串
	data, err := json.Marshal(s)
	if err != nil {
		belogs.Error("StringArray.ToDb(): fail, s:", jsonutil.MarshalJson(s), err)
		return nil, fmt.Errorf("Marshal json fail: %w", err)
	}
	return data, nil
}
