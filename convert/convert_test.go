package convert

import (
	"fmt"
	"reflect"
	"testing"
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

	z = bb

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
