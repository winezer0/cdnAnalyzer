package maputils

import (
	"github.com/winezer0/cdnAnalyzer/pkg/fileutils"
	"testing"
)

func TestSortMap(t *testing.T) {
	filePath := "C:\\Users\\WINDOWS\\Downloads\\sources.json"
	data, _ := fileutils.ReadJsonToMap(filePath)
	data = SortMapRecursively(data)
	outPath := "C:\\Users\\WINDOWS\\Downloads\\sources.new.json"
	fileutils.WriteJson(outPath, data)
}
