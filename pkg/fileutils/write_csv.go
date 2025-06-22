package fileutils

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"
)

// WriteCSV 将 []map[string]interface{} 或 []map[string]string 写入 CSV 文件
// 支持各种基本类型的自动转换为字符串
func WriteCSV(filename string, data interface{}, forceQuote bool) error {
	// 检查数据类型并转换
	var mapData []map[string]interface{}

	switch typedData := data.(type) {
	case []map[string]string:
		// 将 []map[string]string 转换为 []map[string]interface{}
		mapData = make([]map[string]interface{}, len(typedData))
		for i, row := range typedData {
			mapData[i] = make(map[string]interface{})
			for k, v := range row {
				mapData[i][k] = v
			}
		}
	case []map[string]interface{}:
		mapData = typedData
	default:
		return fmt.Errorf("unsupported data type: %T, expected []map[string]string or []map[string]interface{}", data)
	}

	if len(mapData) == 0 {
		return fmt.Errorf("data is empty")
	}

	// 收集所有唯一的表头（键）
	headersMap := make(map[string]struct{})
	for _, row := range mapData {
		for key := range row {
			headersMap[key] = struct{}{}
		}
	}

	// 将表头映射转换为排序后的切片
	headers := make([]string, 0, len(headersMap))
	for key := range headersMap {
		headers = append(headers, key)
	}
	sort.Strings(headers) // 可选：按字母顺序排序表头

	// 创建 CSV 文件
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(headers); err != nil {
		return err
	}

	// 写入行
	for _, rowMap := range mapData {
		record := make([]string, len(headers))
		for i, header := range headers {
			value, exists := rowMap[header]
			if !exists {
				record[i] = ""
				continue
			}
			record[i] = quoteIfNeeded(anyToString(value), forceQuote)
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return nil
}

// anyToString 将任意类型转换为字符串
func anyToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.Format(time.RFC3339)
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// quoteIfNeeded 在需要时为字符串添加引号
func quoteIfNeeded(s string, forceQuote bool) string {
	if forceQuote {
		return `"` + s + `"`
	}
	return s
}
