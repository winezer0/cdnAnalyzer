package maputils

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// StructToMapString 将任意结构体实例转换为 map[string]string 非原始类型字段会尝试 JSON 序列化
func StructToMapString(s interface{}) (map[string]string, error) {
	result := make(map[string]string)

	v := reflect.ValueOf(s).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		value := v.Field(i).Interface()

		switch val := value.(type) {
		case string, bool, int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64:
			result[jsonTag] = fmt.Sprintf("%v", val)
		default:
			if data, err := json.Marshal(val); err == nil {
				result[jsonTag] = string(data)
			} else {
				return nil, fmt.Errorf("字段 %s 序列化失败: %v", jsonTag, err)
			}
		}
	}

	return result, nil
}

// ConvertStructSliceToMaps 将任意结构体切片转换为 []map[string]string
func ConvertStructSliceToMaps(slice interface{}) ([]map[string]string, error) {
	sliceVal := reflect.ValueOf(slice)
	if sliceVal.Kind() != reflect.Slice {
		return nil, fmt.Errorf("输入必须是切片")
	}

	var results []map[string]string

	for i := 0; i < sliceVal.Len(); i++ {
		elem := sliceVal.Index(i)
		if elem.Kind() == reflect.Ptr && elem.IsNil() {
			continue
		}
		structMap, err := StructToMapString(elem.Interface())
		if err != nil {
			return nil, err
		}
		results = append(results, structMap)
	}
	return results, nil
}

// GetMapsUniqueValues 提取maps结构中所有键值对的值
func GetMapsUniqueValues(maps []map[string]string) []string {
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
