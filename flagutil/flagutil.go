package flagutil

import "fmt"

/*
var addrs ArrayFlags
flag.Var(&addrs, "addr", "Addresses")
flag.Parse()
./main --addr 192.168.0.55 --addr 192.168.0.56
*/

type ArrayFlags []string

func (c *ArrayFlags) String() string {
	// 防护nil指针，避免解引用panic
	if c == nil {
		return ""
	}
	return fmt.Sprint(*c)
}

func (c *ArrayFlags) Set(value string) error {
	if c == nil {
		return fmt.Errorf("ArrayFlags pointer cannot be nil") // 返回error但不修改返回值类型
	}
	*c = append(*c, value)
	return nil
}
