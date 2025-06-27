package downfile

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LoadConfig 加载配置文件
func LoadConfig(filename string) (DownConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config DownConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析YAML失败: %w", err)
	}

	return config, nil
}

// FileExists 检查文件是否存在
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// ConvertGitHubURL 转换GitHub URL为原始内容URL
func ConvertGitHubURL(url string) string {
	// 只转换blob URL，不转换releases下载链接
	if strings.Contains(url, "/releases/") {
		return url
	}
	return strings.Replace(strings.Replace(url, "github.com", "raw.githubusercontent.com", 1), "/blob/", "/", 1)
}

// formatDuration 格式化时间为易读格式
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)

	if d.Hours() >= 1 {
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%d小时%d分%d秒", h, m, s)
	} else if d.Minutes() >= 1 {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%d分%d秒", m, s)
	}
	return fmt.Sprintf("%d秒", int(d.Seconds()))
}

// 格式化大小为易读格式
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func MakeDirs(path string, isFile bool) error {
	dir := path
	if isFile {
		dir = filepath.Dir(path)
	}
	_, err := os.Stat(dir)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, os.ModePerm)
	}
	return err
}
