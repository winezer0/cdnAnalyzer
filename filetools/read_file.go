package filetools

import (
	"bufio"
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

// ReadFileToList reads a text file and returns its contents as a slice of strings, where each element is a line from the file.
func ReadFileToList(filename string) ([]string, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	// Read the file line by line
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Check for any errors that occurred during the scan
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
