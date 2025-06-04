package filetools

import (
	"encoding/json"
	"log"
	"strings"
	"testing"
)

// 测试用例
func TestReadCSVToMap(t *testing.T) {
	filename := "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\city_ip.csv"

	expectedData := map[string]string{"City": "北京市 移动", "A": "39.156.128.0"}

	actualData, err := ReadCSVToMap(filename)
	if err != nil {
		t.Errorf("Failed to read CSV: %v", err)
	}

	expectedJSON, _ := json.Marshal(expectedData)
	actualJSON, _ := json.Marshal(actualData)

	if !strings.Contains(string(actualJSON), string(expectedJSON)) {
		t.Errorf("Expected data did not match actual data.\nExpected: %s\nActual: %s", expectedJSON, actualJSON)
	} else {
		t.Logf("Actual: %s", actualJSON)
	}
}

func TestLoadYAMLAsJSON(t *testing.T) {
	// 示例调用
	jsonStr, err := LoadYAMLAsJSON("C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\cdn_cname.yml")
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
