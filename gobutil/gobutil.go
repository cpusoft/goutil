package gobutil

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"sync"
)

var (
	gobLock         sync.Mutex
	registeredTypes sync.Map // 自动注册缓存，避免重复注册
)

func autoRegister(t reflect.Type) {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if _, ok := registeredTypes.Load(t); !ok {
		gob.Register(reflect.Zero(t).Interface())
		registeredTypes.Store(t, struct{}{})
	}
}

// MarshalGob 自动注册 + 序列化
func MarshalGob(f any) []byte {
	if f == nil {
		return nil
	}
	gobLock.Lock()
	defer gobLock.Unlock()

	autoRegister(reflect.TypeOf(f))

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(f); err != nil {
		return nil
	}
	return buf.Bytes()
}

// UnmarshalGob 自动注册 + 反序列化
func UnmarshalGob(b []byte, f any) error {
	if b == nil || f == nil {
		return nil
	}
	gobLock.Lock()
	defer gobLock.Unlock()

	autoRegister(reflect.TypeOf(f))
	return gob.NewDecoder(bytes.NewReader(b)).Decode(f) // 关键：无 &
}
