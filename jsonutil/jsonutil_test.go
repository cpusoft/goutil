package jsonutil

import (
	"container/list"
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

	body1a := MarshallJsonIndent(user)
	fmt.Println(body1a)

	ll := list.New()
	ll.PushBack(user)

	bodylist := MarshallJsonIndent(&ll)
	fmt.Println(bodylist)

	users := make([]User, 0)
	users = append(users, user)
	bodys := MarshallJsonIndent(users)
	fmt.Println(bodys)

	var user1 = User{}
	UnmarshalJson(body1, &user1)
	fmt.Println("after Unmarshal: ", user1)

	var us = UserSimple{}
	UnmarshalJson(body1, &us)
	fmt.Println("after Unmarshal: ", us)
}
