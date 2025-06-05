package fileutils

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
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

// ReadYamlToMap 读取 YAML 文件并将其解析为 map[string]interface{}
func ReadYamlToMap(filePath string) (map[string]interface{}, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在: %s", filePath)
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	// 解析 YAML 到 map
	var result map[string]interface{}
	err = yaml.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("解析 YAML 失败: %v", err)
	}

	return result, nil
}

// ReadYamlToStruct 将 YAML 文件解析到指定的结构体或 map 中
func ReadYamlToStruct(filePath string, out interface{}) error {
	// 1. 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filePath)
	}

	// 2. 读取文件内容
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	// 3. 使用 yaml.Unmarshal 解析数据到 out
	err = yaml.Unmarshal(data, out)
	if err != nil {
		return fmt.Errorf("解析 YAML 失败: %v", err)
	}

	return nil
}
