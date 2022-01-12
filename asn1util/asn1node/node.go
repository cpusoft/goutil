package asn1node

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"
	"unicode/utf8"

	"github.com/cpusoft/goutil/asn1util/asn1base"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
)

var (
	ErrNodeIsConstructed    = errors.New("node is constructed")
	ErrNodeIsNotConstructed = errors.New("node is not constructed")
)

/*
change from
https://github.com/gitchander/asn1

*/

type Node struct {
	class       int
	tag         int
	constructed bool // isCompound

	Data  []byte      `json:"data"`            // Primitive:   (isCompound = false)
	Value interface{} `json:"value,omitempty"` // Primitive:  int/bool/string/time... (isCompound = false)
	Nodes []*Node     `json:"nodes,omitempty"` // Constructed: (isCompound = true)
}

func NewNode(class int, tag int) *Node {
	return &Node{
		class: class,
		tag:   tag,
	}
}

func CheckNode(n *Node, class int, tag int) error {
	if n.class != class {
		belogs.Error("CheckNode(): class fail, class: n.class %d != class %d", n.class, class)
		return errors.New("class not equal")
	}
	if n.tag != tag {
		belogs.Error("CheckNode(): tag fail, tag: n.tag %d != tag %d", n.tag, tag)
		return errors.New("tag not equal")
	}
	return nil
}

func (n *Node) GetTag() int {
	return n.tag
}

func (n *Node) GetClass() int {
	return n.class
}

func (n *Node) getHeader() Header {
	return Header{
		Class:      n.class,
		Tag:        n.tag,
		IsCompound: n.constructed,
	}
}

func (n *Node) IsPrimitive() bool {
	return !(n.constructed)
}

func (n *Node) IsConstructed() bool {
	return (n.constructed)
}

func (n *Node) setHeader(h Header) error {
	*n = Node{
		class:       h.Class,
		tag:         h.Tag,
		constructed: h.IsCompound,
	}
	return nil
}

func (n *Node) checkHeader(h Header) error {
	k := n.getHeader()
	if !EqualHeaders(k, h) {
		return errors.New("der: invalid header")
	}
	return nil
}

func encodeValue(n *Node) ([]byte, error) {
	if !n.constructed {
		return convert.CloneBytes(n.Data), nil
	}
	return encodeNodes(n.Nodes)
}
func EncodeNode(data []byte, n *Node) (rest []byte, err error) {

	header := n.getHeader()
	data, err = EncodeHeader(data, &header)
	if err != nil {
		return nil, err
	}

	value, err := encodeValue(n)
	if err != nil {
		return nil, err
	}

	length := len(value)
	data, err = EncodeLength(data, length)
	if err != nil {
		return nil, err
	}

	data = append(data, value...)
	return data, err
}
func encodeNodes(ns []*Node) (data []byte, err error) {
	for _, n := range ns {
		if n == nil {
			continue
		}
		data, err = EncodeNode(data, n)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func DecodeNode(data []byte, n *Node) (rest []byte, err error) {

	var header Header
	data, err = DecodeHeader(data, &header)
	if err != nil {
		if err == io.EOF {
			belogs.Debug("DecodeNode(): DecodeHeader is end:", data, err)
		} else {
			belogs.Error("DecodeNode(): DecodeHeader fail:", data, err)
		}
		return nil, err
	}
	err = n.setHeader(header)
	if err != nil {
		belogs.Error("DecodeNode(): setHeader fail:", err)
		return nil, err
	}

	var length int
	data, err = DecodeLength(data, &length)
	if err != nil {
		belogs.Error("DecodeNode(): DecodeLength fail:", err)
		return nil, err
	}
	if len(data) < length {
		belogs.Error("DecodeNode():len(data) < length fail:")
		return nil, errors.New("insufficient data length")
	}

	err = decodeValue(data[:length], n)
	if err != nil {
		belogs.Error("DecodeNode(): decodeValue fail:", err)
		return nil, err
	}

	rest = data[length:]
	return rest, nil
}

func decodeValue(data []byte, n *Node) error {

	if !n.constructed {
		var err error
		n.Data = convert.CloneBytes(data)
		switch n.tag {
		case TAG_END_OF_CONTENT:
			n.Value = nil
			belogs.Debug("decodeValue(): TAG_END_OF_CONTENT ", "  tag:", n.tag, "   data:", convert.PrintBytesOneLine(n.Data), n.Value)
		case TAG_BOOLEAN:
			n.Value, err = n.GetBool()
			belogs.Debug("decodeValue(): TAG_BOOLEAN ", "  tag:", n.tag, "   data:", convert.PrintBytesOneLine(n.Data), n.Value)
		case TAG_INTEGER:
			n.Value, err = n.GetBigInt()

			b := new(big.Int).SetBytes(data)
			n.Value = *b
			belogs.Debug("decodeValue(): TAG_INTEGER ", "  tag:", n.tag, "   data:", convert.PrintBytesOneLine(n.Data), n.Value)

		case TAG_BIT_STRING:
			n.Value = n.Data
			belogs.Debug("decodeValue(): TAG_BIT_STRING ", "  tag:", n.tag, "   len(data):", len(n.Data))
		case TAG_OCTET_STRING:
			n.Value = n.Data
			belogs.Debug("decodeValue(): TAG_OCTET_STRING ", "  tag:", n.tag, "   len(data):", len(n.Data))
		case TAG_NULL:
			n.Value = nil
			belogs.Debug("decodeValue(): TAG_NULL ", "  tag:", n.tag, "   data:", convert.PrintBytesOneLine(n.Data), n.Value)
		case TAG_OID:
			n.Value, err = n.GetOid()
			belogs.Debug("decodeValue(): TAG_OID ", "  tag:", n.tag, "   data:", convert.PrintBytesOneLine(n.Data), n.Value)
		case TAG_REAL:
			n.Value, err = n.GetReal()
			belogs.Debug("decodeValue(): TAG_REAL ", "  tag:", n.tag, "   data:", convert.PrintBytesOneLine(n.Data), n.Value)
		case TAG_ENUMERATED:
			n.Value, err = n.GetEnumerated()
			belogs.Debug("decodeValue(): TAG_ENUMERATED ", "  tag:", n.tag, "   data:", convert.PrintBytesOneLine(n.Data), n.Value)
		case TAG_UTF8_STRING:
			n.Value, err = n.GetString()
			belogs.Debug("decodeValue(): TAG_UTF8_STRING ", "  tag:", n.tag, "   data:", convert.PrintBytesOneLine(n.Data), n.Value)
		case TAG_TIME:
			n.Value, err = n.GetUTCTime()
			belogs.Debug("decodeValue(): TAG_TIME ", "  tag:", n.tag, "   data:", convert.PrintBytesOneLine(n.Data), n.Value)
		case TAG_NUMBERIC_STRING:
			n.Value, err = n.GetString()
			belogs.Debug("decodeValue(): TAG_NUMBERIC_STRING ", "  tag:", n.tag, "   data:", convert.PrintBytesOneLine(n.Data), n.Value)
		case TAG_PRINTABLE_STRING:
			n.Value, err = n.GetString()
			belogs.Debug("decodeValue(): TAG_PRINTABLE_STRING ", "  tag:", n.tag, "   data:", convert.PrintBytesOneLine(n.Data), n.Value)
		case TAG_T61_STRING:
			n.Value, err = n.GetString()
			belogs.Debug("decodeValue(): TAG_T61_STRING ", "  tag:", n.tag, "   data:", convert.PrintBytesOneLine(n.Data), n.Value)
		case TAG_VIDEOTEX_STRING:
			n.Value, err = n.GetString()
			belogs.Debug("decodeValue(): TAG_VIDEOTEX_STRING ", "  tag:", n.tag, "   data:", convert.PrintBytesOneLine(n.Data), n.Value)
		case TAG_IA5_STRING:
			n.Value, err = n.GetString()
			belogs.Debug("decodeValue(): TAG_IA5_STRING ", "  tag:", n.tag, "   data:", convert.PrintBytesOneLine(n.Data), n.Value)
		case TAG_UTC_TIME:
			n.Value, err = n.GetUTCTime()
			belogs.Debug("decodeValue(): TAG_UTC_TIME ", "  tag:", n.tag, "   data:", convert.PrintBytesOneLine(n.Data), n.Value)
		case TAG_GENERALIZED_TIME:
			n.Value, err = n.GetGeneralizedTime()
			belogs.Debug("decodeValue(): TAG_GENERALIZED_TIME ", "  tag:", n.tag, "   data:", convert.PrintBytesOneLine(n.Data), n.Value)
		case TAG_BMP_STRING:
			n.Value, err = n.GetString()
			belogs.Debug("decodeValue(): TAG_BMP_STRING ", "  tag:", n.tag, "   data:", convert.PrintBytesOneLine(n.Data), n.Value)
		default:
			err = errors.New("tag is not supported")
		}
		if err != nil {
			belogs.Debug("decodeValue(): fail, tag:", n.tag, "  data:", convert.PrintBytesOneLine(n.Data), err)
			return err
		}
		return nil
	} else {
		n.Data = convert.CloneBytes(data)
		ns, err := decodeNodes(data)
		if err != nil {
			return err
		}
		n.Nodes = ns

		return nil
	}
}
func decodeNodes(data []byte) (ns []*Node, err error) {
	for {
		child := new(Node)
		data, err = DecodeNode(data, child)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		ns = append(ns, child)
	}
	return ns, nil
}

//----------------------------------------------------------------------------

func (n *Node) SetNodes(ns []*Node) {
	n.constructed = true
	n.Nodes = ns
}

func (n *Node) GetNodes() ([]*Node, error) {
	if !n.constructed {
		return nil, ErrNodeIsNotConstructed
	}
	return n.Nodes, nil
}

func (n *Node) SetBool(b bool) {
	n.constructed = false
	n.Data = BoolEncode(b)
}

func (n *Node) GetBool() (bool, error) {
	if n.constructed {
		return false, ErrNodeIsConstructed
	}
	return BoolDecode(n.Data)
}

func (n *Node) SetInt(i int64) {
	n.constructed = false
	n.Data = IntEncode(i)
}

func (n *Node) GetInt() (int64, error) {
	if n.constructed {
		return 0, ErrNodeIsConstructed
	}
	return IntDecode(n.Data)
}

func (n *Node) SetUint(u uint64) {
	n.constructed = false
	n.Data = UintEncode(u)
}

func (n *Node) GetUint() (uint64, error) {
	if n.constructed {
		return 0, ErrNodeIsConstructed
	}
	return UintDecode(n.Data)
}

func (n *Node) SetBytes(bs []byte) {
	n.constructed = false
	n.Data = bs
}

func (n *Node) GetBytes() ([]byte, error) {
	if n.constructed {
		return nil, ErrNodeIsConstructed
	}
	return n.Data, nil
}

func (n *Node) SetString(s string) {
	n.constructed = false
	n.Data = []byte(s)
}

func (n *Node) GetString() (string, error) {
	if n.constructed {
		return "", ErrNodeIsConstructed
	}
	if !utf8.Valid(n.Data) {
		return "", errors.New("invalid utf8 string")
		//return "", errors.New("data is not utf-8 string")
	}
	return string(n.Data), nil
}

func (n *Node) SetUTCTime(t time.Time) error {
	data, err := EncodeUTCTime(t)
	if err != nil {
		return err
	}
	n.constructed = false
	n.Data = data
	return nil
}

func (n *Node) GetUTCTime() (time.Time, error) {
	if n.constructed {
		return time.Time{}, ErrNodeIsConstructed
	}
	return DecodeUTCTime(n.Data)
}

func (n *Node) GetOid() (string, error) {
	if n.constructed {
		return "", ErrNodeIsConstructed
	}
	oids := make([]uint32, len(n.Data)+2)
	//the first byte using: first_arc* 40+second_arc
	//the later , when highest bit is 1, will add to next to calc
	// https://msdn.microsoft.com/en-us/library/windows/desktop/bb540809(v=vs.85).aspx
	f := uint32(n.Data[0])
	if f < 80 {
		oids[0] = f / 40
		oids[1] = f % 40
	} else {
		oids[0] = 2
		oids[1] = f - 80
	}
	var tmp uint32
	for i := 2; i <= len(n.Data); i++ {
		f = uint32(n.Data[i-1])
		if f >= 0x80 {
			tmp = tmp<<7 + (f & 0x7f)
		} else {
			oids[i] = tmp<<7 + (f & 0x7f)
			tmp = 0
		}
	}
	var buffer bytes.Buffer
	for i := 0; i < len(oids); i++ {
		if oids[i] == 0 {
			continue
		}
		buffer.WriteString(fmt.Sprint(oids[i]) + ".")
	}
	return buffer.String()[0 : len(buffer.String())-1], nil
}

func (n *Node) GetReal() (float64, error) {
	//https://github.com/guidoreina/asn1/blob/58d422657c0378218587c89647b771348e2f7d07/asn1/ber/common.cpp
	return 0, errors.New("not supported")
}

func (n *Node) GetEnumerated() (int64, error) {
	return n.GetInt()
}

func (n *Node) GetGeneralizedTime() (time.Time, error) {
	return ParseGeneralizedTime(n.Data)
}

func (n *Node) GetBigInt() (*big.Int, error) {
	return asn1base.ParseBigInt(n.Data)
}
