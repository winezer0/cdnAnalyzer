package fileutils

import (
	"fmt"
	"io"
	"os"
)

// ReadFileBytes 加载数据库文件
func ReadFileBytes(filePath string) ([]byte, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("IP数据库文件[%v]不存在", filePath)
	}

	// 打开文件
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0400)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}
	defer file.Close()

	// 读取文件内容
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	return fileData, nil
}
