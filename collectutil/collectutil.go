package collectutil

import (
	"sort"
	"strings"
)

func MapValuesToSlice[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func BuildCompositeKey(columns map[string]string) string {
	var parts []string

	keys := make([]string, 0, len(columns))
	for k := range columns {
		keys = append(keys, k)
	}
	sort.Strings(keys) // 对键进行字典序排序

	// 构建 "column:value" 的字符串形式
	for _, key := range keys {
		parts = append(parts, key+":"+columns[key])
	}

	// 使用 ":" 将所有部分拼接成最终的复合键
	return strings.Join(parts, ":")
}
