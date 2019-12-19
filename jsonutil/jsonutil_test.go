package jsonutil

import (
	"container/list"
	"encoding/json"
	"fmt"
	"testing"
	"time"
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
type MyTime time.Time

const (
	timeFormart = "2006-01-02 15:04:05"
)

func (t *MyTime) UnmarshalJSON(data []byte) (err error) {

	now, err := time.ParseInLocation(`"`+timeFormart+`"`, string(data), time.Local)
	*t = MyTime(now)
	return
}

func (t MyTime) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(timeFormart)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, timeFormart)
	b = append(b, '"')
	return b, nil
}
func (t MyTime) String() string {
	return time.Time(t).Format(timeFormart)
}

type SyncLogRtrState struct {
	StartTime MyTime `json:"startTime"`
	EndTime   MyTime `json:"endTime"`
}

func TestTimeJson(t *testing.T) {

	s := ""
	var syncLogRtrState = SyncLogRtrState{}
	fmt.Println("after Unmarshal: ", MarshalJson(syncLogRtrState))

	UnmarshalJson(s, &syncLogRtrState)
	fmt.Println("after Unmarshal: ", syncLogRtrState)

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
