package fileutils

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
)

func WriteJson(filePath string, v interface{}) error {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 序列化失败: %w", err)
	}

	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("文件写入失败: %w", err)
	}

	return nil
}

func WriteSortedJson(filePath string, v interface{}) error {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 序列化失败: %w", err)
	}

	jsonBytes, err = sortJsonBytes(jsonBytes, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 格式化异常: %w", err)
	}

	err = os.WriteFile(filePath, jsonBytes, 0644)
	if err != nil {
		return fmt.Errorf("文件写入失败: %w", err)
	}

	return nil
}

// sortJSONRecursive 递归处理任何类型，对 map[string]interface{} 按 key 排序
func sortJSONRecursive(data interface{}) interface{} {
	// 反射获取类型
	val := reflect.ValueOf(data)

	// 如果是 nil 或者零值，直接返回
	if !val.IsValid() {
		return data
	}

	switch val.Kind() {
	case reflect.Map:
		// 处理 map[string]interface{}
		result := make(map[string]interface{})
		keys := make([]string, 0)

		for _, keyVal := range val.MapKeys() {
			key := keyVal.String()
			keys = append(keys, key)
		}
		sort.Strings(keys) // 按 key 排序

		for _, key := range keys {
			value := val.MapIndex(reflect.ValueOf(key)).Interface()
			result[key] = sortJSONRecursive(value)
		}
		return result

	case reflect.Slice, reflect.Array:
		// 处理 slice/array，递归每个元素
		var result []interface{}
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i).Interface()
			result = append(result, sortJSONRecursive(elem))
		}
		return result

	default:
		// 其他基本类型（int, string, bool, etc.）直接返回
		return data
	}
}

// sortJsonBytes 主函数：接收原始 JSON 数据字节，返回排序后的 JSON 字节数组和字符串
func sortJsonBytes(data []byte, prefix, indent string) ([]byte, error) {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	sorted := sortJSONRecursive(raw)

	// 使用 json.MarshalIndent 保留格式
	sortedJSONBytes, err := json.MarshalIndent(sorted, prefix, indent)
	if err != nil {
		return nil, err
	}

	return sortedJSONBytes, nil
}
