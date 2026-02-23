package stringutil

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"github.com/cpusoft/goutil/convert"
)

// ContainInSlice 判断字符串是否存在于切片中
// 保持原有逻辑，仅优化代码可读性
func ContainInSlice(slice []string, one string) bool {
	if len(slice) == 0 || len(one) == 0 {
		return false
	}
	for _, s := range slice {
		if s == one {
			return true
		}
	}
	return false
}

// TrimNewLine 移除字符串中的换行符（\r 和 \n）
func TrimNewLine(str string) (s string) {
	s = strings.ReplaceAll(str, "\r", "") // 推荐使用 ReplaceAll（Go 1.12+），语义与 Replace 一致
	s = strings.ReplaceAll(s, "\n", "")
	return s
}

// TrimSpace 移除字符串中的空格和制表符
func TrimSpace(str string) (s string) {
	s = strings.ReplaceAll(str, "\t", "")
	s = strings.ReplaceAll(s, " ", "")
	return s
}

// TrimSpaceAndNewLine 移除空格、制表符和换行符
func TrimSpaceAndNewLine(str string) (s string) {
	s = TrimSpace(str)
	return TrimNewLine(s)
}

// TrimSuffixAll 递归移除所有后缀（修复拼写错误+无限递归问题）
// 原逻辑：移除字符串末尾所有指定的 trim 后缀（如 "test///" 移除 "/" 得到 "test"）
func TrimSuffixAll(str, trim string) (s string) {
	// 空值直接返回，避免无效递归
	if len(str) == 0 || len(trim) == 0 {
		return str
	}
	s = str
	// 循环移除后缀（替代递归，避免栈溢出）
	for strings.HasSuffix(s, trim) {
		s = strings.TrimSuffix(s, trim)
	}
	return s
}

// GetValueFromJointStr 从拼接字符串中提取指定 key 的值
// 例：line="a=1&b=2", key="a", separator="&" → 返回 "1"
func GetValueFromJointStr(line, key, separator string) string {
	if len(line) == 0 || len(key) == 0 || len(separator) == 0 {
		return ""
	}
	split := strings.Split(line, separator)
	for _, part := range split {
		if strings.HasPrefix(part, key+"=") {
			return strings.TrimPrefix(part, key+"=") // 替代 Replace，语义更清晰
		}
	}
	return ""
}

// OmitString 截断过长的字符串，保留前 end 个字符
// 修复 uint64 转 int 的溢出风险，以及索引越界问题
func OmitString(str string, end uint64) string {
	strLen := uint64(utf8.RuneCountInString(str)) // 按字符数计算，而非字节数（符合字符串截断语义）
	if strLen == 0 {
		return ""
	}
	// 确保 end 不超过字符串长度，且转换为 int 时不溢出
	var maxLen int
	if end > strLen {
		maxLen = int(strLen)
	} else {
		// 限制 int 最大值，避免 uint64 转 int 溢出
		if end > uint64(int(^uint(0)>>1)) {
			maxLen = int(^uint(0) >> 1)
		} else {
			maxLen = int(end)
		}
	}
	// 按字符数截断（避免截断多字节字符）
	return string([]rune(str)[:maxLen])
}

// Int64sToInString 将 int64 切片转为 "(1,2,3)" 格式的字符串
func Int64sToInString(s []int64) string {
	if len(s) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	buffer.WriteString("(")
	for i := 0; i < len(s); i++ {
		buffer.WriteString(convert.ToString(s[i]))
		if i < len(s)-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString(")")
	return buffer.String()
}

// StringsToInString 将字符串切片转为 "("a","b","c")" 格式的字符串
func StringsToInString(s []string) string {
	if len(s) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	buffer.WriteString("(")
	for i := 0; i < len(s); i++ {
		buffer.WriteString("\"" + s[i] + "\"")
		if i < len(s)-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString(")")
	return buffer.String()
}

// StringsToInSqlString 将字符串切片转为 '(a','b','c')' 格式的 SQL IN 语句字符串
// 修复 SQL 注入漏洞：转义字符串中的单引号
func StringsToInSqlString(s []string) string {
	if len(s) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	buffer.WriteString("(")
	for i := 0; i < len(s); i++ {
		// 转义单引号：将 ' 替换为 ''（SQL 标准转义方式）
		escapedStr := strings.ReplaceAll(s[i], "'", "''")
		buffer.WriteString("'" + escapedStr + "'")
		if i < len(s)-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString(")")
	return buffer.String()
}
