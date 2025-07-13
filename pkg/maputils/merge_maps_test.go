package maputils

import (
	"github.com/winezer0/cdnAnalyzer/pkg/fileutils"
	"github.com/winezer0/cdnAnalyzer/pkg/logging"
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

	mapA, _ := fileutils.ReadJsonToMap("C:\\Users\\WINDOWS\\Downloads\\sources_china.json")
	mapB, _ := fileutils.ReadJsonToMap("C:\\Users\\WINDOWS\\Downloads\\sources_foreign.json")
	merged := DeepMerge(mapA, mapB)

	// 指定输出文件路径
	filePath := "C:\\Users\\WINDOWS\\Downloads\\output.json"

	// 写入 JSON 文件
	err := fileutils.WriteJson(filePath, merged)
	if err != nil {
		logging.Errorf("数据写入文件失败:", err)
		return
	} else {
		logging.Debugf("数据写入文件成功: %s\n", filePath)
	}
}
