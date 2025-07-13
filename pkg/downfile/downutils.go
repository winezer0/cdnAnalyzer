package downfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
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

func GetItemFilePath(filename, downloadDir string) string {
	storePath := filename
	if !filepath.IsAbs(filename) {
		storePath = filepath.Join(downloadDir, filename)
	}
	return storePath
}

// FilterEnableItems 仅保留 enable=true 的配置项
func FilterEnableItems(items []DownItem) []DownItem {
	var enabledItems []DownItem
	if len(items) > 0 {
		for _, item := range items {
			if item.Enable {
				enabledItems = append(enabledItems, item)
			}
		}
	}
	return enabledItems
}

// CleanupIncompleteDownloads 清理下载目录下未完成的下载文件（以 .download 结尾的文件）
func CleanupIncompleteDownloads(downloadDir string) error {
	// 检查下载目录是否存在
	if _, err := os.Stat(downloadDir); os.IsNotExist(err) {
		return nil // 目录不存在，直接返回
	}
	// 查找所有 .download 文件
	files, err := FindFilesBySuffix(downloadDir, ".download")
	if err != nil {
		return fmt.Errorf("查找未完成下载文件失败: %w", err)
	}
	// 删除每个文件
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			fmt.Fprintf(os.Stderr, "警告: 删除未完成下载文件失败 %s: %v\n", file, err)
		}
	}
	return nil
}

// FindFilesBySuffix 递归查找指定目录下所有以后缀 suffix 结尾的文件
func FindFilesBySuffix(root, suffix string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), suffix) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func DownloadFileSimple(url string, proxy string, storePath string) error {
	// 创建HTTP客户端配置
	clientConfig := &ClientConfig{
		ConnectTimeout: 30,
		IdleTimeout:    30,
		ProxyURL:       proxy,
	}
	httpClient, err := CreateHTTPClient(clientConfig)
	if err != nil {
		return err
	}
	err = downloadFile(httpClient, url, storePath, false)
	if err != nil {
		return err
	}
	return nil
}
