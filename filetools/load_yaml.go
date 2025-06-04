package filetools

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

// LoadYAMLAsJSON 通用函数：将 YAML 文件加载并转为 JSON 字符串
func LoadYAMLAsJSON(filePath string) (string, error) {
	// 1. 读取文件内容
	data, err := ioutil.ReadFile(filePath)
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
