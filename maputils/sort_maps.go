package maputils

import (
	"fmt"
	"reflect"
	"sort"
)

// SortStructRecursive 对任意 JSON 数据结构进行深度排序（map 按 key 排序，slice 按值排序）
func SortStructRecursive(data interface{}) interface{} {
	// 判断类型
	val := reflect.ValueOf(data)

	switch val.Kind() {
	case reflect.Map:
		// 处理 map 类型
		result := make(map[string]interface{})
		keys := val.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, k := range keys {
			v := val.MapIndex(k)
			sortedValue := SortStructRecursive(v.Interface())
			result[k.String()] = sortedValue
		}
		return result

	case reflect.Slice:
		// 处理 slice 类型
		length := val.Len()
		slice := make([]interface{}, length)
		for i := 0; i < length; i++ {
			slice[i] = SortStructRecursive(val.Index(i).Interface())
		}

		// 只对基本类型 slice 进行排序（比如 string、int、float 等）
		if length > 0 && isSortableElement(slice[0]) {
			sort.Slice(slice, func(i, j int) bool {
				return fmt.Sprintf("%v", slice[i]) < fmt.Sprintf("%v", slice[j])
			})
		}

		return slice

	default:
		// 基本类型直接返回
		return data
	}
}

// isSortableElement 判断是否是可排序的基本类型
func isSortableElement(v interface{}) bool {
	switch v.(type) {
	case string, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, bool:
		return true
	default:
		return false
	}
}

// SortMapRecursively 按 key 排序 map，并递归处理嵌套结构
func SortMapRecursively(data map[string]interface{}) map[string]interface{} {
	result := SortStructRecursive(data)
	return result.(map[string]interface{})
}
