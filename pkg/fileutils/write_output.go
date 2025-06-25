package fileutils

import (
	maputils2 "cdnAnalyzer/pkg/maputils"
	"fmt"
	"os"
	"strings"
)

func WriteOutputToFile(data interface{}, outputType, outputFile string) error {
	var err error

	switch strings.ToLower(outputType) {
	case "csv":
		if outputFile == "" {
			outputFile = "result.csv"
		}
		sliceToMaps, mapErr := maputils2.ConvertStructsToMaps(data)
		if mapErr == nil {
			err = WriteCSV(outputFile, sliceToMaps, true)
		} else {
			err = WriteText(outputFile, maputils2.AnyToTxtStr(data))
		}

	case "json":
		if outputFile == "" {
			outputFile = "result.json"
		}
		err = os.WriteFile(outputFile, maputils2.AnyToJsonBytes(data), 0644)

	case "txt":
		if outputFile == "" {
			outputFile = "result.txt"
		}
		err = WriteText(outputFile, maputils2.AnyToTxtStr(data))

	default:
		fmt.Println(maputils2.AnyToTxtStr(data))
		return nil
	}

	if err != nil {
		return fmt.Errorf("写入 %s 失败: %w", outputType, err)
	}
	fmt.Printf("结果已写入%s文件: %s\n", outputType, outputFile)
	return nil
}
