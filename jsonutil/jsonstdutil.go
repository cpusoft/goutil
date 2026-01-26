package jsonutil

import (
	"bytes"
	"encoding/json"
)

// str := MarshalJson(user)
func StdMarshalJson(f interface{}) string {
	body, err := json.Marshal(f)
	if err != nil {
		return ""
	}
	return string(body)
}
func StdMarshalJsonIndent(f interface{}) string {
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

// Deprecated
func StdMarshallJsonIndent(f interface{}) string {
	return MarshalJsonIndent(f)
}

/*
var user1 = User{}
UnmarshalJson(body1, &user1)
*/
func StdUnmarshalJson(str string, f interface{}) error {

	return json.Unmarshal([]byte(str), &f)
}
