package asn1node

import (
	"encoding/hex"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/jsonutil"
)

/*
call ParseFile --> n *Node
--> print jsonUtil.MarshalJson(n)
--> compare to asn1 struct
--> get every value

{"nodes":[{"nodes":[{"nodes":[{"nodes":[{"nodes":[{"value":"Ag=="},{"nodes":[{"value":"ACABBnwgjA=="}]}]}]}]}]},{"nodes":[{"value":"2.16.840.1.101.3.4.2.1"}]},{"nodes":[{"nodes":[{"value":"b42_ipv6_loa.png"},{"value":"lRbdZL58FyW5/KEXEg5Y6NhCpSBoczmbPd/8kcS2rPA="}]},{"nodes":[{"value":"b42_service_definition.json"},{"value":"CuE5RyIAXNkvTGqgJNXWs+LmfWKfEXINlHimM6EXocc="}]}]}]}

	ipNode := n.Nodes[0].Nodes[0].Nodes[0].Nodes[0]
	ipFamliy := ipNode.Nodes[0]
	ipAddress := ipNode.Nodes[1].Nodes[0]
	oidNode := n.Nodes[1].Nodes[0]
	fileHashNode := n.Nodes[2]
	fmt.Println(jsonutil.MarshalJson(ipFamliy))
	fmt.Println(jsonutil.MarshalJson(ipAddress))

	fmt.Println(jsonutil.MarshalJson(oidNode))
	for _, child := range fileHashNode.Nodes {
		fmt.Println(jsonutil.MarshalJson(child.Nodes[0]))
		fmt.Println(jsonutil.MarshalJson(child.Nodes[1]))
		nameAndHashStr := jsonutil.MarshalJson(child.Nodes)
		fmt.Println(nameAndHashStr)
		nameAndHash := make([]NameAndHash, 0)
		err = jsonutil.UnmarshalJson(nameAndHashStr, &nameAndHash)
		fmt.Println(nameAndHash, err)
	}

then -->
{"value":"Ag=="}
{"value":"ACABBnwgjA=="}
{"value":"2.16.840.1.101.3.4.2.1"}
{"value":"b42_ipv6_loa.png"}
{"value":"lRbdZL58FyW5/KEXEg5Y6NhCpSBoczmbPd/8kcS2rPA="}
[{"value":"b42_ipv6_loa.png"},{"value":"lRbdZL58FyW5/KEXEg5Y6NhCpSBoczmbPd/8kcS2rPA="}]
[{b42_ipv6_loa.png} {lRbdZL58FyW5/KEXEg5Y6NhCpSBoczmbPd/8kcS2rPA=}] <nil>
{"value":"b42_service_definition.json"}
{"value":"CuE5RyIAXNkvTGqgJNXWs+LmfWKfEXINlHimM6EXocc="}
[{"value":"b42_service_definition.json"},{"value":"CuE5RyIAXNkvTGqgJNXWs+LmfWKfEXINlHimM6EXocc="}]
[{b42_service_definition.json} {CuE5RyIAXNkvTGqgJNXWs+LmfWKfEXINlHimM6EXocc=}] <nil>

*/
func ParseFile(fileName string) (n *Node, err error) {
	belogs.Debug("ParseFile(): fileName:", fileName)

	data, err := fileutil.ReadFileToBytes(fileName)
	if err != nil {
		belogs.Error("ParseFile(): fileName:", fileName, err)
		return nil, err
	}
	belogs.Debug("ParseFile(): fileName:", fileName, "  len(data):", len(data), err)
	return ParseBytes(data)
}

func ParseHex(hexStr string) (n *Node, err error) {
	belogs.Debug("ParseHex(): hex:", hexStr)

	data, err := hex.DecodeString(hexStr)
	if err != nil {
		belogs.Error("ParseHex(): hex:", hexStr, err)
		return nil, err
	}
	belogs.Debug("ParseHex(): hex:", hexStr, "  len(data):", len(data), err)
	return ParseBytes(data)
}

func ParseBytes(data []byte) (n *Node, err error) {
	if len(data) == 0 {
		belogs.Error("ParseBytes(): data is emtpy:", len(data))
		return nil, errors.New("data is empty")
	}
	n = new(Node)
	rest, err := DecodeNode(data, n)
	if err != nil {
		belogs.Error("ParseBytes():  DecodeNode fail:", len(data), err)
		return nil, err
	}
	belogs.Info("ParseBytes(): len(data):", len(data), "  n:", jsonutil.MarshalJson(n),
		"   len(rest):", len(rest))
	return n, nil
}
