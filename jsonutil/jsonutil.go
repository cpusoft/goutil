package jsonutil

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"sync"
)

//str := MarshalJson(user)
func MarshalJson(f interface{}) string {
	body, err := json.Marshal(f)
	if err != nil {
		return ""
	}
	return string(body)
}

func MarshallJsonIndent(f interface{}) string {
	body, err := json.Marshal(f)
	if err != nil {
		return ""
	}
	var out bytes.Buffer
	err = json.Indent(&out, body, "", "    ")
	if err != nil {
		return ""
	}
	return out.String()
}

/*
  var user1 = User{}
  UnmarshalJson(body1, &user1)
*/
func UnmarshalJson(str string, f interface{}) error {

	return json.Unmarshal([]byte(str), &f)
}

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
