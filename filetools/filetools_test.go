package filetools

import (
	"encoding/json"
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
