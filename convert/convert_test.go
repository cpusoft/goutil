package convert

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestInterface2Map(t *testing.T) {
	m := make(map[string]string, 0)
	m["aa"] = "11"
	m["bb"] = "bb"
	fmt.Println(m)

	var z interface{}
	z = m
	fmt.Println(z)

	m1, _ := Interface2Map(z)
	fmt.Println(m1)

	v1 := reflect.ValueOf(m1)
	fmt.Println("v1:", v1.Kind())

	bb := []byte{0x01, 0x02, 0x03}
	s, err := GetInterfaceType(bb)
	fmt.Println("type:", s, err)

}

func TestBytes2String(t *testing.T) {
	byt := []byte{0x01, 0x02, 0x03, 0xa1, 0xfc, 0x7c}
	s := Bytes2String(byt)
	fmt.Println(s)
}

func TestBytes2StringSection(t *testing.T) {
	byt := []byte{0x01, 0x02, 0x03, 0xa1, 0xfc, 0x7c, 0x01, 0x02, 0x03, 0xa1, 0xfc, 0x7c}
	s := Bytes2StringSection(byt, 8)
	fmt.Println(s)
}

type User struct {
	Id         int         `json:"id"`
	Username   string      `json:"username"`
	Password   string      `json:"password"`
	StateItems []StateItem `json:"stateItems"`
}

type StateItem struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

func TestStruct2Map(t *testing.T) {
	s1 := StateItem{"state", "valid"}
	s2 := StateItem{"error", "errorssss"}
	s3 := StateItem{"warning", "warningsss"}
	ss := make([]StateItem, 0)
	ss = append(ss, s1)
	ss = append(ss, s2)
	ss = append(ss, s3)
	user := User{Id: 5, Username: "zhangsan", Password: "password", StateItems: ss}

	data := Struct2Map(user)
	fmt.Printf("%v", data)
}

func TestTime2String(t *testing.T) {
	tt := time.Now()
	s := Time2String(tt)
	fmt.Println(s)

	s = Time2StringZone(tt)
	fmt.Println(s)
}
