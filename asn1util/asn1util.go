package asn1util

import (
	"bytes"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/fileutil"
)

var oid = map[string]string{
	"2.5.4.3":                    "CN",
	"2.5.4.4":                    "SN",
	"2.5.4.5":                    "serialNumber",
	"2.5.4.6":                    "C",
	"2.5.4.7":                    "L",
	"2.5.4.8":                    "ST",
	"2.5.4.9":                    "streetAddress",
	"2.5.4.10":                   "O",
	"2.5.4.11":                   "OU",
	"2.5.4.12":                   "title",
	"2.5.4.17":                   "postalCode",
	"2.5.4.42":                   "GN",
	"2.5.4.43":                   "initials",
	"2.5.4.44":                   "generationQualifier",
	"2.5.4.46":                   "dnQualifier",
	"2.5.4.65":                   "pseudonym",
	"0.9.2342.19200300.100.1.25": "DC",
	"1.2.840.113549.1.9.1":       "emailAddress",
	"0.9.2342.19200300.100.1.1":  "userid",
	"2.5.29.20":                  "CRL Number",
}

func GetDNFromName(namespace pkix.Name, sep string) (string, error) {
	return GetDNFromRDNSeq(namespace.ToRDNSequence(), sep)
}

func GetDNFromRDNSeq(rdns pkix.RDNSequence, sep string) (string, error) {
	// 使用 strings.Builder 优化字符串拼接性能
	var sb strings.Builder
	first := true
	for _, s := range rdns {
		for _, i := range s {
			if !first {
				sb.WriteString(sep)
			}
			first = false

			// 核心修复：先统一获取 OID 对应的名称（无论 Value 类型）
			oidStr := i.Type.String()
			fieldName, ok := oid[oidStr]
			if !ok {
				fieldName = oidStr
			}

			// 再根据 Value 类型拼接值
			switch v := i.Value.(type) {
			case string:
				sb.WriteString(fmt.Sprintf("%s=%s", fieldName, v))
			default:
				sb.WriteString(fmt.Sprintf("%s=%v", fieldName, v))
			}
		}
	}
	return sb.String(), nil
}

func DecodeString(data []byte) (ret string) {
	if len(data) == 0 {
		return ""
	}
	return string(data)
}

func DecodeUTF8String(data []byte) (ret string) {
	if len(data) == 0 {
		return ""
	}
	return string(data)
}

func DecodeIA5String(data []byte) (ret string) {
	if len(data) == 0 {
		return ""
	}
	return string(data)
}

func DecodeBool(data byte) (ret bool) {
	return data != 0x00
}

// UTC is short Year, 2 nums
func DecodeUTCTime(data []byte) (ret string, err error) {
	// 精确检查长度，避免越界访问
	if len(data) < 13 {
		return "", errors.New("DecodeUTCTime fail")
	}
	// 严格校验：只处理标准13字节的UTCTime，多余字节忽略（符合ASN.1规范）
	if len(data) > 13 {
		data = data[:13] // 截断到标准长度，避免多余数据干扰
	}
	year := "20" + string(data[0:2])
	month := string(data[2:4])
	day := string(data[4:6])
	hour := string(data[6:8])
	minute := string(data[8:10])
	second := string(data[10:12])
	z := string(data[12])
	return fmt.Sprintf("%s-%s-%s %s:%s:%s%s", year, month, day, hour, minute, second, z), nil
}

// Generalized Long Year, 4 nums
func DecodeGeneralizedTime(data []byte) (ret string, err error) {
	// 第一层防护：基础长度校验
	if len(data) < 15 {
		return "", fmt.Errorf("DecodeGeneralizedTime fail: invalid length %d (minimum 15 required)", len(data))
	}
	// 严格校验：只处理标准15字节的GeneralizedTime
	if len(data) > 15 {
		data = data[:15]
	}

	year := string(data[0:4])
	month := string(data[4:6])
	day := string(data[6:8])
	hour := string(data[8:10])
	minute := string(data[10:12])
	second := string(data[12:14])
	z := string(data[14])
	return fmt.Sprintf("%s-%s-%s %s:%s:%s%s", year, month, day, hour, minute, second, z), nil
}

func DecodeOid(data []byte) (ret string) {
	if len(data) == 0 {
		return ""
	}

	oids := make([]uint32, 0, len(data)+2) // 预分配容量优化
	//the first byte using: first_arc* 40+second_arc
	//the later , when highest bit is 1, will add to next to calc
	// https://msdn.microsoft.com/en-us/library/windows/desktop/bb540809(v=vs.85).aspx
	f := uint32(data[0])
	var first, second uint32
	if f < 80 {
		first = f / 40
		second = f % 40
	} else {
		first = 2
		second = f - 80
	}
	oids = append(oids, first, second)

	var tmp uint32
	for i := 1; i < len(data); i++ { // 从第二个字节开始遍历，避免重复处理
		f = uint32(data[i])
		if f >= 0x80 {
			// 检查溢出风险
			if tmp > math.MaxUint32>>7 {
				belogs.Warn("DecodeOid(): potential overflow when parsing OID byte")
				tmp = 0
				continue
			}
			tmp = tmp<<7 + (f & 0x7f)
		} else {
			oidVal := tmp<<7 + (f & 0x7f)
			oids = append(oids, oidVal)
			tmp = 0
		}
	}

	var buffer bytes.Buffer
	for i, o := range oids {
		if o == 0 && i > 1 { // 跳过除前两位外的0值
			continue
		}
		if buffer.Len() > 0 {
			buffer.WriteByte('.')
		}
		buffer.WriteString(fmt.Sprint(o))
	}

	result := buffer.String()
	belogs.Debug("DecodeOid(): oid:", result)
	return result
}

// return byte directly
func DecodeBytes(data []byte) (ret string) {
	if len(data) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, b := range data {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(fmt.Sprintf("%02x", b))
	}
	return sb.String()
}

func EncodeInteger(val uint64) []byte {
	if val == 0 {
		return []byte{0x00}
	}

	var out bytes.Buffer
	found := false
	shift := uint(56)
	mask := uint64(0xFF00000000000000)

	for mask > 0 {
		if !found && (val&mask != 0) {
			found = true
		}
		if found || (shift == 0) {
			out.WriteByte(byte((val & mask) >> shift))
		}
		shift -= 8
		mask >>= 8
	}
	return out.Bytes()
}

func DecodeInteger(data []byte) (ret uint64) {
	if len(data) == 0 {
		return 0
	}
	// 检查溢出风险
	if len(data) > 8 {
		belogs.Warn("DecodeInteger(): input too long, potential overflow")
		return 0
	}

	for _, i := range data {
		// 溢出检查
		if ret > (math.MaxUint64-uint64(i))/256 {
			belogs.Warn("DecodeInteger(): integer overflow detected")
			return 0
		}
		ret = ret*256 + uint64(i)
	}
	return ret
}

// FiniteLen get length
func DecodeFiniteLen(data []byte) (datalen uint64, datapos uint64, err error) {
	// 1. 基础长度校验：至少需要2字节（类型+长度标识）
	if len(data) < 2 {
		return 0, 0, errors.New("data too short for finite length decoding")
	}

	// 2. 解析长度标识字节
	lenByte := data[1]
	datalen = uint64(lenByte)
	datapos = 2

	// 3. 处理长长度格式（最高位为1）
	if lenByte&0x80 != 0 {
		// 提取长度字段的字节数（去掉最高位）
		lenFieldLen := lenByte & 0x7F
		// 校验1：长度字段的字节数不能为0，也不能超过数据长度（类型1字节 + 长度标识1字节 + 长度字段N字节）
		if lenFieldLen == 0 || int(1+1+lenFieldLen) > len(data) {
			return 0, 0, fmt.Errorf("data too short for length field: need %d bytes for length, have %d", 1+1+lenFieldLen, len(data))
		}

		// 校验2：长度字段的字节数不能超过8（uint64最大8字节）
		if lenFieldLen > 8 {
			return 0, 0, fmt.Errorf("length field too long: %d bytes (max 8)", lenFieldLen)
		}

		// 解析长度字段的实际值
		lenFieldData := data[2 : 2+lenFieldLen]
		datalen = DecodeInteger(lenFieldData)
		datapos = 2 + uint64(lenFieldLen)

		// 校验3：解析出的长度不能超过剩余数据长度（可选，根据业务需求，仅当需要验证内容完整性时启用）
		if datapos+datalen > uint64(len(data)) {
			return 0, 0, fmt.Errorf("data too short for content: need %d bytes (pos %d + len %d), have %d", datapos+datalen, datapos, datalen, len(data))
		}
	}

	belogs.Debug("DecodeFiniteLen():return datalen: ", datalen, " datapos:", datapos)
	return datalen, datapos, nil
}

// InfiniteLen just care about the 0x00 0x00
func DecodeInfiniteLen(data []byte) (datalen uint64, datapos uint64, err error) {
	if len(data) < 2 {
		return 0, 0, errors.New("data too short for infinite length decoding")
	}
	endbytes := []byte{0x00, 0x00}
	pos := bytes.Index(data, endbytes)
	if pos < 0 {
		return uint64(len(data)), 2, errors.New("end bytes (0x00 0x00) not found")
	}
	datalen = uint64(pos)
	datapos = 2
	return datalen, datapos, nil
}

// FiniteLen will get length, but InfiniteLen using 0x00 0x00 to get length
func DecodeFiniteAndInfiniteLen(data []byte) (datalen uint64, datapos uint64, err error) {
	if len(data) < 1 {
		return 0, 0, errors.New("empty data for length decoding")
	}

	data0Len := data[1]
	belogs.Debug("DecodeFiniteAndInfiniteLen():again seq0Len:", data0Len)
	if data0Len == byte(0x80) {
		datalen, datapos, err = DecodeInfiniteLen(data)
	} else {
		datalen, datapos, err = DecodeFiniteLen(data)
	}
	belogs.Debug("DecodeFiniteAndInfiniteLen():datalen:", datalen, " datapos:", datapos)
	return datalen, datapos, err
}

// found the location of 0x00 0x00
func IndexEndOfBytes(oldb []byte, tagType uint8, hierarchyFor00 int, topHierarchyFor00 int) (int, error) {
	belogs.Debug("IndexEndOfBytes():len(oldb):", len(oldb), "  tagType", tagType,
		"   hierarchyFor00:", hierarchyFor00, "     topHierarchyFor00:", topHierarchyFor00)

	if len(oldb) <= 2 {
		return -1, fmt.Errorf("bytes is too short (length: %d)", len(oldb))
	}

	//0x30 80, 0011 0000
	//0xa0 80, 1010 0000
	//0x24 80, 0010 0100
	var pos int
	endbytes := []byte{0x00, 0x00}
	var TypeConstructed uint8 = 32 // xx1xxxxxb  0011 0010
	var TypeLastIndex byte = 0xa0  // see certpacket

	if oldb[0] == TypeLastIndex || (tagType == TypeConstructed && hierarchyFor00 < topHierarchyFor00) {
		pos = bytes.LastIndex(oldb, endbytes)
		belogs.Debug("IndexEndOfBytes(): LastIndex  pos:", pos)
	} else {
		// may be more 0x00 0x00 are together, found the latest 0x00 0x00
		pos = bytes.Index(oldb, endbytes)
		belogs.Debug("IndexEndOfBytes():Index initial pos:", pos)

		// 修复循环条件，避免越界
		for pos > 0 && len(oldb) > pos+2*len(endbytes) {
			if pos+4 > len(oldb) {
				break
			}
			if bytes.Equal(oldb[pos+2:pos+4], endbytes) {
				pos += 2
				belogs.Debug("IndexEndOfBytes():for Index  pos:", pos)
			} else {
				break
			}
		}
		belogs.Debug("IndexEndOfBytes(): Index  pos:", pos)
	}

	// 正确处理未找到的情况
	if pos < 0 {
		return len(oldb), fmt.Errorf("end bytes (0x00 0x00) not found in data")
	}
	return pos, nil
}

func GetTopHierarchyFor00(oldb []byte) int {
	if len(oldb) == 0 {
		return 0
	}

	top := 0
	endbytes := []byte{0x00, 0x00}
	pos := bytes.LastIndex(oldb, endbytes)

	// 修复循环条件，避免无限循环
	for pos > 0 && len(oldb) >= pos+len(endbytes) {
		oldb = oldb[:pos]
		top += 1
		pos = bytes.LastIndex(oldb, endbytes)
	}

	belogs.Debug("GetTopHierarchyFor00(): top:", top)
	return top
}

func TrimSuffix00(oldByte []byte, cerEndIndex int) (b []byte, i int) {
	if len(oldByte) == 0 || cerEndIndex <= 0 {
		return oldByte, cerEndIndex
	}

	// 第一步：修剪后缀连续的 0x00 0x00
	null2Bytes := []byte{0x00, 0x00}
	for len(oldByte) >= len(null2Bytes) && bytes.HasSuffix(oldByte, null2Bytes) {
		oldByte = oldByte[:len(oldByte)-len(null2Bytes)]
		cerEndIndex -= len(null2Bytes)
	}

	// 第二步：修剪后缀单个的 0x00
	null1Byte := []byte{0x00}
	for len(oldByte) >= len(null1Byte) && bytes.HasSuffix(oldByte, null1Byte) {
		oldByte = oldByte[:len(oldByte)-len(null1Byte)]
		cerEndIndex -= len(null1Byte)
	}

	return oldByte, cerEndIndex
}

func TrimPrefix00(olddb []byte) []byte {
	if len(olddb) == 0 {
		return olddb
	}

	// 第一步：修剪前缀连续的 0x00 0x00
	null2Bytes := []byte{0x00, 0x00}
	for len(olddb) >= len(null2Bytes) && bytes.HasPrefix(olddb, null2Bytes) {
		olddb = olddb[len(null2Bytes):]
	}

	// 第二步：修剪前缀单个的 0x00
	null1Byte := []byte{0x00}
	for len(olddb) >= len(null1Byte) && bytes.HasPrefix(olddb, null1Byte) {
		olddb = olddb[len(null1Byte):]
	}

	return olddb
}

// ExtKeyUsagesToInts 将 x509.ExtKeyUsage 切片转换为 int 切片
func ExtKeyUsagesToInts(exts []x509.ExtKeyUsage) []int {
	if len(exts) == 0 {
		return []int{}
	}
	ints := make([]int, len(exts))
	for i, ext := range exts {
		// 核心修复：直接转换，无额外加减
		ints[i] = int(ext)
	}
	return ints
}

// deprecated
func ReadFileAndDecodeBase64(file string) (fileByte []byte, fileDecodeBase64Byte []byte, err error) {
	return fileutil.ReadFileAndDecodeCertBase64(file)
}
