package fileutils

import (
	"encoding/json"
	"log"
	"testing"
)

func TestLoadYAMLAsJSON(t *testing.T) {
	// 示例调用
	jsonStr, err := ReadYamlAsJson("C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\cdn_cname.yml")
	if err != nil {
		log.Fatalf("Error loading YAML: %v", err)
	}

	// 输出 JSON 内容
	//fmt.Println(jsonStr)

	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		panic(err)
	}

	TraverseJSON(data, "  ")
}
