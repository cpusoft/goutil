package xmlutil

import (
	xml "encoding/xml"
)

//str := MarshalJson(user)
func MarshalXml(f interface{}) string {
	body, err := xml.Marshal(f)
	if err != nil {
		return ""
	}
	return string(body)
}

/*
  var user1 = User{}
  UnmarshalXml(body1, &user1)
*/
func UnmarshalXml(str string, f interface{}) error {

	return xml.Unmarshal([]byte(str), &f)
}
