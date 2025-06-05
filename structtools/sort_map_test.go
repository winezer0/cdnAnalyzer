package structutils

import (
	"cdnCheck/fileutils"
	"testing"
)

func TestSortMap(t *testing.T) {
	filePath := "C:\\Users\\WINDOWS\\Downloads\\sources.json"
	data, _ := fileutils.ReadJsonToMap(filePath)
	data = SortMapRecursively(data)
	outPath := "C:\\Users\\WINDOWS\\Downloads\\sources.new.json"
	fileutils.WriteJsonFromStruct(outPath, data)
}
