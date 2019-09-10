package jsonutil

import (
	"encoding/json"
	"fmt"
	"testing"
)

type User struct {
	Id    int
	Name  string
	Age   int
	Class string
}

type UserSimple struct {
	Name  string
	Class string
}

func TestJson(t *testing.T) {
	user := User{
		Id:    1,
		Name:  "wang",
		Age:   22,
		Class: "class1",
	}
	body, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("user: ", string(body))

	body1 := MarshalJson(user)
	fmt.Println(body1)

	body1a := MarshallJsonInent(user)
	fmt.Println(body1a)

	var user1 = User{}
	UnmarshalJson(body1, &user1)
	fmt.Println("after Unmarshal: ", user1)

	var us = UserSimple{}
	UnmarshalJson(body1, &us)
	fmt.Println("after Unmarshal: ", us)
}
