package filetools

import (
	"encoding/csv"
	"os"
)

func ReadCSVToMap(filename string) ([]map[string]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// 假设第一行是列名
	headers := records[0]
	var result []map[string]string

	for _, record := range records[1:] { // 跳过头部
		row := make(map[string]string)
		for i, header := range headers {
			row[header] = record[i]
		}
		result = append(result, row)
	}

	return result, nil
}
