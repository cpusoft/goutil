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
	return fmt.Sprint(*c)
}

func (c *ArrayFlags) Set(value string) error {
	*c = append(*c, value)
	return nil
}
