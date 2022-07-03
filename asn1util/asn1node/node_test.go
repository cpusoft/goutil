package asn1node

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

func TestParseHex(t *testing.T) {
	sigStr := `30819C3014A1123010300E04010230090307002001067C208C300B06096086480165030402013077303416106234325F697076365F6C6F612E706E6704209516DD64BE7C1725B9FCA117120E58E8D842A5206873399B3DDFFC91C4B6ACF0303F161B6234325F736572766963655F646566696E6974696F6E2E6A736F6E04200AE1394722005CD92F4C6AA024D5D6B3E2E67D629F11720D9478A633A117A1C7`
	n, err := ParseHex(sigStr)
	fmt.Println(err)
	s := jsonutil.MarshalJson(n)
	fmt.Println("Json:\n" + s)

	fmt.Println("data:\n" + convert.PrintBytesOneLine(n.Nodes[0].Data))
	fmt.Println("data:\n" + convert.PrintBytesOneLine(n.Nodes[0].Nodes[0].Data))
	fmt.Println("data:\n" + convert.PrintBytesOneLine(n.Nodes[0].Nodes[0].Nodes[0].Data))

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

	}

}
