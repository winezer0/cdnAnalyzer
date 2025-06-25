package maputils

import (
	fileutils2 "cdnAnalyzer/pkg/fileutils"
	"fmt"
	"testing"
)

func TestDeepMerge(t *testing.T) {
	//jsonA := `{
	//	"name": "Alice",
	//	"hobbies": ["reading", "coding"],
	//	"address": {
	//		"city": "Beijing",
	//		"zip": "100000"
	//	},
	//	"email": "alice@example.com"
	//}`
	//
	//jsonB := `{
	//	"name": "Bob",
	//	"hobbies": ["gaming", "reading"],
	//	"address": {
	//		"street": "Main St"
	//	},
	//	"email": "bob_new@example.com"
	//}`
	//
	//mapA, _ := ParseJSON(jsonA)
	//mapB, _ := ParseJSON(jsonB)

	mapA, _ := fileutils2.ReadJsonToMap("C:\\Users\\WINDOWS\\Downloads\\sources_china.json")
	mapB, _ := fileutils2.ReadJsonToMap("C:\\Users\\WINDOWS\\Downloads\\sources_data.json")
	merged := DeepMerge(mapA, mapB)

	// 指定输出文件路径
	filePath := "C:\\Users\\WINDOWS\\Downloads\\output.json"

	// 写入 JSON 文件
	err := fileutils2.WriteJsonFromStruct(filePath, merged)
	if err != nil {
		fmt.Println("写入失败:", err)
		return
	} else {
		fmt.Printf("数据已成功写入到文件: %s\n", filePath)
	}

	//
	//mergedJSON := AnyToJsonStr(merged)
	//fmt.Println(mergedJSON)
}
