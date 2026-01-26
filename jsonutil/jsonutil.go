package jsonutil

import (
	"github.com/bytedance/sonic"
)

// str := MarshalJson(user)
func MarshalJson(f interface{}) string {
	body, err := sonic.Marshal(f)
	if err != nil {
		return ""
	}
	return string(body)
}
func MarshalJsonIndent(f interface{}) string {
	// - 第1个参数：要序列化的对象
	// - 第2个参数：前缀字符串（每行开头添加的字符串，通常为空）
	// - 第3个参数：缩进字符串（通常用"\t"或"  "）
	body, err := sonic.MarshalIndent(f, "", "  ")
	if err != nil {
		return ""
	}
	return string(body)
}

/*
var user1 = User{}
UnmarshalJson(body1, &user1)
*/
func UnmarshalJson(str string, f interface{}) error {

	return sonic.Unmarshal([]byte(str), &f)
}
