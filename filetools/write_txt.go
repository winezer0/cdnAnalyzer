package filetools

import (
	"bufio"
	"fmt"
	"os"
)

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
