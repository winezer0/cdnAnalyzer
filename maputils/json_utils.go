package maputils

import (
	"encoding/json"
	"fmt"
)

// TraverseJSON 遍历 JSON 的所有键和值，支持嵌套结构
func TraverseJSON(data interface{}, indent string) {
	switch value := data.(type) {
	case map[string]interface{}:
		for key, val := range value {
			fmt.Printf("%sKey: %q\n", indent, key)
			TraverseJSON(val, indent+"  ") // 缩进显示层级
		}
	case []interface{}:
		for i, val := range value {
			fmt.Printf("%sIndex: %d\n", indent, i)
			TraverseJSON(val, indent+"  ")
		}
	default:
		fmt.Printf("%sValue: %#v\n", indent, value)
	}
}

// ParseJSON 解析 JSON 字符串为 map[string]interface{}
func ParseJSON(data string) (map[string]interface{}, error) {
	var v map[string]interface{}
	err := json.Unmarshal([]byte(data), &v)
	return v, err
}

// AnyToJsonStr 将任意 map 转换为格式化的 JSON 字符串（用于输出）
func AnyToJsonStr(v interface{}) string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}

// AnyToJsonBytes  将任意 map 转换为格式化的 JSON 字符串（用于输出）
func AnyToJsonBytes(v interface{}) []byte {
	data, _ := json.MarshalIndent(v, "", "  ")
	return data
}
