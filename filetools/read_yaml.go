package filetools

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"os"
)

// ReadYamlAsJson 通用函数：将 YAML 文件加载并转为 JSON 字符串
func ReadYamlAsJson(filePath string) (string, error) {
	// 1. 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// 2. 解析为通用 map 结构
	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return "", err
	}

	// 3. 转为 JSON 格式的字节流（带缩进）
	jsonData, err := json.MarshalIndent(yamlData, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
