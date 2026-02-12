package jsonutil

import (
	"encoding/hex"
	"sync"
)

// []byte: show in string
type PrintableBytes []byte

func (c PrintableBytes) MarshalText() ([]byte, error) {
	str := string(c)
	return []byte(str), nil
}

func (i *PrintableBytes) UnmarshalText(b []byte) error {
	str := string(b)
	*i = []byte(str)
	return nil
}

// []byte: show in hex
type HexBytes []byte

func (c HexBytes) MarshalText() ([]byte, error) {
	str := hex.EncodeToString(c)
	return []byte(str), nil
}

func (i *HexBytes) UnmarshalText(b []byte) error {

	b, err := hex.DecodeString(string(b))
	if err != nil {
		return err
	}
	*i = b
	return nil
}

// sync.Map does not support JSON printing
// need convert to normal map[string]interface{} to JSON printing
type JsonSyncMap struct {
	sync.Map
}

func (c JsonSyncMap) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	c.Range(func(key, value interface{}) bool {
		m[key.(string)] = value
		return true
	})
	return []byte(MarshalJson(m)), nil
}
