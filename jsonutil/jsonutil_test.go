package jsonutil

import (
	"container/list"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/stretchr/testify/assert"
)

// -------------------------- 测试用数据结构 --------------------------
// 基础测试结构体
type TestUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age,omitempty"` // 空值忽略
}

// 循环引用结构体（测试序列化失败场景）
type CycleRefStruct struct {
	Data string          `json:"data"`
	Next *CycleRefStruct `json:"next"`
}

// -------------------------- 核心测试用例 --------------------------

// TestMarshalJson 测试JSON序列化（字符串返回）
func TestMarshalJson(t *testing.T) {
	// 场景1：正常结构体序列化
	user := TestUser{ID: 1001, Name: "张三", Age: 25}
	expected := `{"id":1001,"name":"张三","age":25}`
	assert.Equal(t, expected, MarshalJson(user))

	// 场景2：空对象序列化（返回合法空JSON）
	emptyUser := TestUser{}
	expectedEmpty := `{"id":0,"name":""}`
	assert.Equal(t, expectedEmpty, MarshalJson(emptyUser))

	// 场景3：不支持的类型（chan）序列化（失败返回空字符串）
	ch := make(chan int)
	assert.Equal(t, "", MarshalJson(ch))

	// 场景4：循环引用序列化（失败返回空字符串）
	var cycle CycleRefStruct
	cycle.Data = "test"
	cycle.Next = &cycle // 构建循环引用
	assert.Equal(t, "", MarshalJson(cycle))

	// 场景5：nil入参序列化（返回空字符串）
	assert.Equal(t, "", MarshalJson(nil))
}

// TestMarshalJsonBytes 测试JSON序列化（字节数组返回）
func TestMarshalJsonBytes(t *testing.T) {
	// 场景1：正常结构体序列化
	user := TestUser{ID: 1002, Name: "李四"}
	expected := []byte(`{"id":1002,"name":"李四"}`)
	assert.Equal(t, expected, MarshalJsonBytes(user))

	// 场景2：不支持的类型（func）序列化（失败返回nil）
	fn := func() {}
	assert.Nil(t, MarshalJsonBytes(fn))

	// 场景3：空切片序列化（返回合法JSON）
	emptySlice := []int{}
	assert.Equal(t, []byte("[]"), MarshalJsonBytes(emptySlice))
}

// TestMarshalJsonIndent 测试带缩进的JSON序列化
func TestMarshalJsonIndent(t *testing.T) {
	// 场景1：正常结构体缩进序列化
	user := TestUser{ID: 1003, Name: "王五"}
	expected := `{
  "id": 1003,
  "name": "王五"
}`
	assert.Equal(t, expected, MarshalJsonIndent(user))

	// 场景2：不支持的类型（循环引用）序列化（失败返回空字符串）
	var cycle CycleRefStruct
	cycle.Next = &cycle
	assert.Equal(t, "", MarshalJsonIndent(cycle))
}

// TestUnmarshalJson 测试JSON字符串反序列化
func TestUnmarshalJson(t *testing.T) {
	// 场景1：正常反序列化到结构体
	jsonStr := `{"id":1004,"name":"赵六","age":30}`
	var user TestUser
	err := UnmarshalJson(jsonStr, &user)
	assert.NoError(t, err)
	assert.Equal(t, 1004, user.ID)
	assert.Equal(t, "赵六", user.Name)
	assert.Equal(t, 30, user.Age)

	// 场景2：传入nil目标对象（返回错误）
	var nilUser *TestUser = nil
	err = UnmarshalJson(jsonStr, nilUser)
	assert.Error(t, err)
	assert.Equal(t, "UnmarshalJsonBytes(): target object is nil", err.Error())

	// 场景3：空字符串反序列化（返回错误）
	err = UnmarshalJson("", &user)
	assert.Error(t, err)
	assert.Equal(t, "UnmarshalJsonBytes(): input data is empty", err.Error())

	// 场景4：全空白字符的JSON（返回sonic原生错误，漏洞点）
	blankJson := "   \t\n"
	err = UnmarshalJson(blankJson, &user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "syntax error") // 原生解析错误，不友好

	// 场景5：非法JSON格式（多余逗号）
	invalidJson := `{"id":1005,"name":"钱七",}`
	err = UnmarshalJson(invalidJson, &user)
	assert.Error(t, err)
}

// TestUnmarshalJsonBytes 测试JSON字节数组反序列化
func TestUnmarshalJsonBytes(t *testing.T) {
	// 场景1：正常反序列化
	jsonBytes := []byte(`{"id":1006,"name":"孙八"}`)
	var user TestUser
	err := UnmarshalJsonBytes(jsonBytes, &user)
	assert.NoError(t, err)
	assert.Equal(t, 1006, user.ID)

	// 场景2：nil目标对象（返回错误）
	err = UnmarshalJsonBytes(jsonBytes, nil)
	assert.Error(t, err)
	assert.Equal(t, "UnmarshalJsonBytes(): target object is nil", err.Error())

	// 场景3：空字节数组（返回错误）
	emptyBytes := []byte{}
	err = UnmarshalJsonBytes(emptyBytes, &user)
	assert.Error(t, err)
	assert.Equal(t, "UnmarshalJsonBytes(): input data is empty", err.Error())

	// 场景4：类型不匹配（字符串ID赋值给int字段）
	wrongTypeJson := []byte(`{"id":"1007","name":"周九"}`)
	err = UnmarshalJsonBytes(wrongTypeJson, &user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "type mismatch")

	// 场景5：多余字段（忽略，正常返回）
	extraFieldJson := []byte(`{"id":1008,"name":"吴十","age":40,"extra":"无用字段"}`)
	err = UnmarshalJsonBytes(extraFieldJson, &user)
	assert.NoError(t, err)
	assert.Equal(t, 1008, user.ID)
}

// TestEdgeCases 测试边缘场景
func TestEdgeCases(t *testing.T) {
	// 场景1：omitempty字段（Age为0时不序列化）
	user := TestUser{ID: 1009, Name: "郑十一", Age: 0}
	assert.Equal(t, `{"id":1009,"name":"郑十一"}`, MarshalJson(user))

	// 场景2：反序列化缺失字段（赋默认值）
	jsonStr := `{"id":1010,"name":"王十二"}`
	var user2 TestUser
	err := UnmarshalJson(jsonStr, &user2)
	assert.NoError(t, err)
	assert.Equal(t, 0, user2.Age) // 缺失字段赋默认值

	// 场景3：序列化nil指针（返回null）
	var nilPtr *TestUser = nil
	assert.Equal(t, "null", MarshalJson(nilPtr))

	// 场景4：反序列化null值（结构体指针赋nil）
	nullJson := `null`
	var user3 *TestUser
	err = UnmarshalJson(nullJson, &user3)
	assert.NoError(t, err)
	assert.Nil(t, user3)
}

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
	//timeFormart = "2006-01-02 15:04:05"
	timeFormart = time.RFC3339
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
	var syncLogRtrState = SyncLogRtrState{
		StartTime: MyTime(time.Now()),
		EndTime:   MyTime(time.Now()),
	}
	fmt.Println("after MarshalJson: ", MarshalJson(syncLogRtrState))

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

	body1a := MarshalJsonIndent(user)
	fmt.Println(body1a)

	ll := list.New()
	ll.PushBack(user)

	bodylist := MarshalJsonIndent(&ll)
	fmt.Println(bodylist)

	users := make([]User, 0)
	users = append(users, user)
	bodys := MarshalJsonIndent(users)
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

func TestSonicJson(t *testing.T) {
	user := User{
		Id:    1,
		Name:  "wang",
		Age:   22,
		Class: "class1",
	}
	output, err := sonic.Marshal(&user)
	fmt.Println("sonic json: ", output, err)
	var user1 = User{}
	err = sonic.Unmarshal(output, &user1)
	fmt.Println("after Unmarshal sonic json: ", user1)
}
