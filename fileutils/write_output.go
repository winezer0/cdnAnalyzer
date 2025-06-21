package fileutils

import (
	"cdnCheck/maputils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func WriteOutputToFile(data interface{}, outputType, outputFile, targetFile string) error {
	var err error

	switch strings.ToLower(outputType) {
	case "csv":
		if outputFile == "" {
			outputFile = getDefaultFilename(targetFile, ".csv")
		}
		sliceToMaps, mapErr := maputils.ConvertStructsToMaps(data)
		if mapErr == nil {
			err = WriteCSV(outputFile, sliceToMaps, true)
		} else {
			err = WriteText(outputFile, maputils.AnyToTxtStr(data))
		}

	case "json":
		if outputFile == "" {
			outputFile = getDefaultFilename(targetFile, ".json")
		}
		err = os.WriteFile(outputFile, maputils.AnyToJsonBytes(data), 0644)

	case "txt":
		if outputFile == "" {
			outputFile = getDefaultFilename(targetFile, ".txt")
		}
		err = WriteText(outputFile, maputils.AnyToTxtStr(data))

	default:
		fmt.Println(maputils.AnyToTxtStr(data))
		return nil
	}

	if err != nil {
		return fmt.Errorf("写入 %s 失败: %w", outputType, err)
	}
	fmt.Printf("结果已写入%s文件: %s\n", outputType, outputFile)
	return nil
}

// 辅助函数
func getDefaultFilename(base, suffix string) string {
	if base == "" {
		return ""
	}
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	return name + ".check" + suffix
}
