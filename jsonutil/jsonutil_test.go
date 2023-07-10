package jsonutil

import (
	"container/list"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

type Tt struct {
	Title string
	C     HexBytes
}

func TestHexBytes(t *testing.T) {

	tt := Tt{
		Title: "sss",
		C:     []byte{0x01, 0x02, 0x0f, 0xab},
	}
	str := MarshalJson(tt)
	fmt.Println(str)

	tt1 := Tt{}
	UnmarshalJson(str, &tt1)
	fmt.Println(tt1)
}

type Bs struct {
	Pbs PrintableBytes
	S   string
	N   []byte
}

func TestPrintableBytes(t *testing.T) {
	s := `测试类型`
	bs := Bs{Pbs: []byte(s), S: s, N: []byte(s)}
	str := MarshalJson(bs)
	fmt.Println(str)

	bs2 := Bs{}
	UnmarshalJson(str, &bs2)
	fmt.Println(bs2)
}

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

type TcpTlsMsg struct {
	// common
	MsgType   uint64      `json:"msgType"`
	MsgResult chan string `json:"-"`

	// for close
	ConnKey string `json:"connKey,omitempty"`

	// for send data //
	// NEXT_CONNECT_CLOSE_POLICY_NO  NEXT_CONNECT_CLOSE_POLICY_GRACEFUL  NEXT_CONNECT_CLOSE_POLICY_FORCIBLE
	NextConnectClosePolicy int `json:"nextConnectClosePolicy,omitempty"`
	//NEXT_RW_POLICY_ALL,NEXT_RW_POLICY_WAIT_READ,NEXT_RW_POLICY_WAIT_WRITE
	NextRwPolicy int            `json:"nextRwPolicy,omitempty"`
	SendData     PrintableBytes `json:"sendData,omitempty"`
}

func TestTcptlsMsg(t *testing.T) {
	sendData := []byte{0x01, 0x02, 0x03}
	tcpTlsMsg := &TcpTlsMsg{
		MsgType:                1,
		NextConnectClosePolicy: 2,
		NextRwPolicy:           3,
		SendData:               sendData,
	}
	fmt.Println("sendMessageModel(): tcptlsclient, will send tcpTlsMsg:",
		MarshalJson(*tcpTlsMsg))

}

func TestJsonSyncMap(t *testing.T) {
	failUrls := JsonSyncMap{}
	snapshotFailUrls := make(map[string]string)
	snapshotFailUrls["1_https://rpki.telecentras.lt/"] = "1"
	snapshotFailUrls["2_https://rrdp.ripe.net/"] = "2"
	snapshotFailUrls["3_https://ca.rg.net"] = "3"
	snapshotFailUrls["4_https://google.com"] = "4"
	for k, v := range snapshotFailUrls {
		split := strings.Split(k, "_")
		url := split[1]
		failUrls.Store(url, v)
	}
	fmt.Println(failUrls)
	fmt.Println(MarshalJson(failUrls))
}
