package maputils

import (
	fileutils2 "github.com/winezer0/cdnAnalyzer/pkg/fileutils"
	"testing"
)

func TestSortMap(t *testing.T) {
	filePath := "C:\\Users\\WINDOWS\\Downloads\\sources.json"
	data, _ := fileutils2.ReadJsonToMap(filePath)
	data = SortMapRecursively(data)
	outPath := "C:\\Users\\WINDOWS\\Downloads\\sources.new.json"
	fileutils2.WriteJson(outPath, data)
}
