package structutils

import (
	"reflect"
)

// DeepMerge 合并两个 map[string]interface{}，b 优先级更高
func DeepMerge(a, b map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// 先遍历 a 中的所有键
	for key, value := range a {
		if bValue, ok := b[key]; ok {
			// 如果两者都是 map[string]interface{}，递归合并
			aMap, aIsMap := value.(map[string]interface{})
			bMap, bIsMap := bValue.(map[string]interface{})
			if aIsMap && bIsMap {
				result[key] = DeepMerge(aMap, bMap)
				continue
			}

			// 如果两者都是字符串，则新值覆盖旧值
			if _, aIsStr := value.(string); aIsStr {
				if _, bIsStr := bValue.(string); bIsStr {
					result[key] = bValue
					continue
				}
			}

			// 如果两者都是 slice，进行去重合并
			aSlice, aIsSlice := takeAsSlice(value)
			bSlice, bIsSlice := takeAsSlice(bValue)
			if aIsSlice && bIsSlice {
				result[key] = mergeUniqueSlices(aSlice, bSlice)
				continue
			}

			// 其他情况用 b 的值覆盖 a
			result[key] = bValue
		} else {
			result[key] = value
		}
	}

	// 添加 b 中有而 a 中没有的键
	for key, value := range b {
		if _, exists := result[key]; !exists {
			result[key] = value
		}
	}

	return result
}

// 判断是否是 slice 并尝试转换为 []interface{}
func takeAsSlice(value interface{}) ([]interface{}, bool) {
	if reflect.TypeOf(value).Kind() == reflect.Slice {
		if s, ok := value.([]interface{}); ok {
			return s, true
		}
	}
	return nil, false
}

// 去重合并两个 slice
func mergeUniqueSlices(a, b []interface{}) []interface{} {
	set := make(map[interface{}]struct{})
	var result []interface{}

	for _, item := range a {
		set[item] = struct{}{}
		result = append(result, item)
	}

	for _, item := range b {
		if _, exists := set[item]; !exists {
			set[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}
