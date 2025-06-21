package maputils

import (
	"fmt"
)

// GetMapValue 获取 map指定键的值
func GetMapValue(m map[string]string, key string) string {
	val, ok := m[key]
	if !ok {
		panic(fmt.Sprintf("entry missing key: %s", key))
	}
	return val
}

// GetMapsValuesUnique 提取maps结构中所有键值对的值
func GetMapsValuesUnique(maps []map[string]string) []string {
	seen := make(map[string]bool)
	var values []string

	for _, m := range maps {
		for _, v := range m {
			if !seen[v] {
				seen[v] = true
				values = append(values, v)
			}
		}
	}
	return values
}
