package filetools

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"
)

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
	if IsEmptyFile(filename) || writeMode == "w" || writeMode == "w+" {
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
