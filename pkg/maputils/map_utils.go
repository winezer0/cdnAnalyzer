package maputils

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
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

// ConvertMapsToStructs 使用 JSON 序列化/反序列化的方式将  json map 数据结构转为任意的数据类型
func ConvertMapsToStructs(source interface{}, target interface{}) error {
	if source == nil {
		return errors.New("source is nil")
	}
	if target == nil {
		return errors.New("target is nil")
	}

	data, err := json.Marshal(source)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	err = json.Unmarshal(data, target)
	if err != nil {
		return fmt.Errorf("unmarshal failed: %w", err)
	}

	return nil
}

// ConvertStructsToMaps 将任意结构体切片转换为 map 切片
func ConvertStructsToMaps(data interface{}) ([]map[string]interface{}, error) {
	//使用 JSON 序列化和反序列化实现结构体到 map 的转换
	// 检查输入是否为切片
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		return nil, fmt.Errorf("输入必须是切片类型，而不是 %T", data)
	}

	// 如果切片为空，返回空结果
	if val.Len() == 0 {
		return []map[string]interface{}{}, nil
	}

	// 将结构体切片转换为 JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("序列化为 JSON 失败: %w", err)
	}

	// 将 JSON 解析为 map 切片
	var result []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("从 JSON 反序列化失败: %w", err)
	}

	return result, nil
}

// AnyToTxtStr 将任意数据转换为文本格式（已优化性能）
func AnyToTxtStr(data interface{}) string {
	//根据传入的 data 类型进行判断，并采取不同的方式将其转换为字符串：
	//[]string（字符串切片） -> 每个元素一行输出
	//[]map[string]interface{}（map 切片）->	每个 map 项逐 key 输出，用 --- 分隔
	//其他类型（结构体、slice、基本类型等） -> 使用 json.MarshalIndent 转为格式化 JSON 字符串

	// 创建 strings.Builder 实例
	var sb strings.Builder

	// 处理字符串切片
	if strSlice, ok := data.([]string); ok {
		for _, s := range strSlice {
			sb.WriteString(s)
			sb.WriteString("\n")
		}
		return sb.String()
	}

	// 处理 map 切片
	if mapSlice, ok := data.([]map[string]interface{}); ok {
		for _, m := range mapSlice {
			for k, v := range m {
				sb.WriteString(fmt.Sprintf("%s: %v\n", k, v))
			}
			sb.WriteString("---\n")
		}
		return sb.String()
	}

	// 其他类型转为 JSON 字符串
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error converting to text: %v", err)
	}
	return string(jsonBytes)
}
