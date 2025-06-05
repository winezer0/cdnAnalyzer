package fileutils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// WriteJsonFromMap 将 map 写入 JSON 文件
func WriteJsonFromMap(filePath string, data map[string]interface{}) error {
	// 将 map 转换为格式化的 JSON 数据
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 序列化失败: %w", err)
	}

	// 写入文件
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("文件写入失败: %w", err)
	}

	return nil
}

func WriteJsonFromStruct(filePath string, v interface{}) error {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 序列化失败: %w", err)
	}

	err = ioutil.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("文件写入失败: %w", err)
	}

	return nil
}
