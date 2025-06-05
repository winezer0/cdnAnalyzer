package fileutils

import (
	"bufio"
	"os"
)

// ReadTextToList reads a text file and returns its contents as a slice of strings, where each element is a line from the file.
func ReadTextToList(filename string) ([]string, error) {
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
