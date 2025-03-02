package lib_file

import (
	"fmt"
	"os"
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
