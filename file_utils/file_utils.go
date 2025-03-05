package file_utils

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// IsFileEmpty checks if a file is empty or does not exist.
// Returns true if the file does not exist or is empty, otherwise returns false.
func IsFileEmpty(filename string) bool {
	// Get file info
	fileInfo, err := os.Stat(filename)
	if os.IsNotExist(err) || fileInfo.Size() == 0 {
		return true
	}
	return false
}

func getTimesFileName(outputFile string) string {
	// 自定义时间格式
	timeInfo := time.Now().Format("20060102_150405")
	return fmt.Sprintf("output.%s.csv", timeInfo)
}

// WriteListToFile writes a slice of strings to a file, with each element written as a separate line.
func WriteListToFile(filename string, data []string) error {
	// Create or truncate the file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Create a writer and write each line to the file
	writer := bufio.NewWriter(file)
	for _, line := range data {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write line: %w", err)
		}
	}

	// Flush any buffered data to the underlying file
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
}

func WriteCSVFromMapSlice(filename string, data []map[string]string, forceQuote bool) error {
	if len(data) == 0 {
		return fmt.Errorf("data is empty")
	}

	// Collect all unique headers (keys)
	headersMap := make(map[string]struct{})
	for _, row := range data {
		for key := range row {
			headersMap[key] = struct{}{}
		}
	}

	// Convert headers map to a sorted slice
	headers := make([]string, 0, len(headersMap))
	for key := range headersMap {
		headers = append(headers, key)
	}
	sort.Strings(headers) // Optional: Sort headers alphabetically

	// Create the CSV file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write rows
	for _, rowMap := range data {
		record := make([]string, len(headers))
		for i, header := range headers {
			record[i] = quoteIfNeeded(rowMap[header], forceQuote)
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return nil
}

// quoteIfNeeded ensures that a string is quoted if it contains commas, double quotes, or newlines.
func quoteIfNeeded(s string, forceQuote bool) string {
	if !forceQuote {
		return s
	}

	if len(s) == 0 {
		return `""`
	}

	needsQuotes := false
	for _, r := range s {
		switch r {
		case '"', ',', '\n', '\r':
			needsQuotes = true
		}
		if needsQuotes {
			break
		}
	}
	if needsQuotes {
		// Escape any existing double quotes and wrap the string in double quotes.
		s = `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

// WriteCSV writes a slice of maps to a CSV file with optional forced quoting and write mode.
func WriteCSV(filename string, data []map[string]string, forceQuote bool, writeMode string, headers []string) error {
	if len(data) == 0 {
		return fmt.Errorf("data is empty")
	}

	// Collect all unique headers (keys)
	headersMap := make(map[string]struct{})
	for _, row := range data {
		for key := range row {
			headersMap[key] = struct{}{}
		}
	}

	// Convert headers map to a sorted slice
	if headers != nil || len(headers) == 0 {
		headers := make([]string, 0, len(headersMap))
		for key := range headersMap {
			headers = append(headers, key)
		}
		sort.Strings(headers) // Optional: Sort headers alphabetically
	}

	// Define the file open mode based on the input parameter
	var flag int
	switch strings.ToLower(writeMode) {
	case "w":
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case "a":
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	case "w+":
		flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	case "a+":
		flag = os.O_RDWR | os.O_CREATE | os.O_APPEND
	default:
		return fmt.Errorf("invalid write mode: %s", writeMode)
	}

	// Create or open the CSV file
	file, err := os.OpenFile(filename, flag, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// If the file is being overwritten (not appending), write the headers
	if IsFileEmpty(filename) || writeMode == "w" || writeMode == "w+" {
		if err := writer.Write(headers); err != nil {
			return err
		}
	}

	// Write rows
	for _, rowMap := range data {
		record := make([]string, len(headers))
		for i, header := range headers {
			record[i] = quoteIfNeeded(rowMap[header], forceQuote)
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return nil
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
