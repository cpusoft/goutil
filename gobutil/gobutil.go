package gobutil

import (
	"bytes"
	"encoding/gob"
)

func MarshalGob(f any) []byte {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(f)
	if err != nil {
		return nil
	}
	return buf.Bytes()
}
func UnmarshalGob(b []byte, f any) error {
	return gob.NewDecoder(bytes.NewReader(b)).Decode(&f)
}
