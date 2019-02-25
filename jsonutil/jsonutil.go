package jsonutil

import (
	"encoding/json"
)

//str := MarshalJson(user)
func MarshalJson(f interface{}) string {
	body, err := json.Marshal(f)
	if err != nil {
		return ""
	}
	return string(body)
}

/*
  var user1 = User{}
  UnmarshalJson(body1, &user1)
*/
func UnmarshalJson(str string, f interface{}) {

	json.Unmarshal([]byte(str), &f)
}
