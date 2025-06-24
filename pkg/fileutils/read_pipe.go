package fileutils

import (
	"bufio"
	"os"
	"strings"
)

// ReadPipeToList reads from standard input (pipe) and returns each line as a string slice.
func ReadPipeToList() ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
