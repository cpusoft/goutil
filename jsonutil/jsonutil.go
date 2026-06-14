package jsonutil

import (
	"bytes"
	"errors"
	"sync"

	"github.com/bytedance/sonic"
)

// str := MarshalJson(user)
func MarshalJson(f interface{}) string {
	body := MarshalJsonBytes(f)
	if body == nil {
		return ""
	}
	return string(body)
}

/*
	func MarshalJsonBytes(f interface{}) []byte {
		body, err := sonic.Marshal(f)
		if err != nil {
			return nil
		}
		return body
	}
*/
var encBufPool = sync.Pool{
	New: func() interface{} {
		// 预分配 64KB，覆盖大部分场景
		return bytes.NewBuffer(make([]byte, 0, 64*1024))
	},
}

func MarshalJsonBytes(v interface{}) []byte {
	if v == nil {
		return nil
	}

	buf := encBufPool.Get().(*bytes.Buffer)
	buf.Reset()

	enc := sonic.ConfigStd.NewEncoder(buf)
	if err := enc.Encode(v); err != nil {
		encBufPool.Put(buf)
		return nil
	}

	data := buf.Bytes()
	if n := len(data); n > 0 && data[n-1] == '\n' {
		data = data[:n-1]
	}

	// 必须显式拷贝，不能返回 buf.Bytes() 的引用
	out := make([]byte, len(data))
	copy(out, data)

	encBufPool.Put(buf)
	return out
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
	if f == nil {
		return errors.New("UnmarshalJson(): target object is nil")
	}
	if len(str) == 0 {
		return errors.New("UnmarshalJson(): input data is empty")
	}
	return sonic.UnmarshalString(str, f) //UnmarshalJson([]byte(str), f)
}
func UnmarshalJsonBytes(data []byte, f interface{}) error {
	if f == nil {
		return errors.New("UnmarshalJsonBytes(): target object is nil")
	}
	if len(data) == 0 {
		return errors.New("UnmarshalJsonBytes(): input data is empty")
	}
	return sonic.Unmarshal(data, f)
}
