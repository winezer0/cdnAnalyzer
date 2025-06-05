package fileutils

import (
	"bufio"
	"fmt"
	"os"
)

// WriteTextFromList writes a slice of strings to a file, with each element written as a separate line.
func WriteTextFromList(filename string, data []string) error {
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

// WriteTextFromStruct 将任意数据写入文本文件
func WriteTextFromStruct(filePath string, data interface{}) error {
	// 将任意数据转换为字符串形式
	content := fmt.Sprintf("%v", data)

	// 写入文件
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("文件写入失败: %w", err)
	}

	return nil
}
