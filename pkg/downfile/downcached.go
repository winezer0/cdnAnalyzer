package downfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DownloadCache 下载缓存结构
type DownloadCache struct {
	Files map[string]time.Time `json:"files"` // 文件路径 -> 最后下载时间
}

// GetCacheFilePath 获取缓存文件路径
func GetCacheFilePath() string {
	// 缓存文件保存在用户主目录下
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// 如果无法获取主目录，则使用当前目录
		return CacheFileName
	}
	return filepath.Join(homeDir, CacheFileName)
}

// LoadDownloadCache 加载下载缓存
func LoadDownloadCache() *DownloadCache {
	cacheFilePath := GetCacheFilePath()
	cache := &DownloadCache{
		Files: make(map[string]time.Time),
	}

	// 如果缓存文件不存在，返回空缓存
	if !FileExists(cacheFilePath) {
		return cache
	}

	// 读取缓存文件
	data, err := os.ReadFile(cacheFilePath)
	if err != nil {
		fmt.Printf("警告: 读取缓存文件失败: %v\n", err)
		return cache
	}

	// 解析JSON
	if err := json.Unmarshal(data, cache); err != nil {
		fmt.Printf("警告: 解析缓存文件失败: %v\n", err)
		return &DownloadCache{
			Files: make(map[string]time.Time),
		}
	}

	return cache
}

// SaveDownloadCache 保存下载缓存
func SaveDownloadCache(cache *DownloadCache) error {
	cacheFilePath := GetCacheFilePath()

	// 将缓存转换为JSON
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化缓存失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(cacheFilePath, data, 0644); err != nil {
		return fmt.Errorf("写入缓存文件失败: %w", err)
	}

	return nil
}

// UpdateFileDownloadTime 更新文件下载时间
func UpdateFileDownloadTime(filePath string) error {
	// 规范化文件路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %w", err)
	}

	// 加载缓存
	cache := LoadDownloadCache()

	// 更新文件下载时间
	cache.Files[absPath] = time.Now()

	// 保存缓存
	return SaveDownloadCache(cache)
}

// CleanupExpiredCache 清理过期缓存记录
func CleanupExpiredCache() {
	cache := LoadDownloadCache()
	now := time.Now()
	changed := false

	// 检查每个文件记录
	for path, lastDownload := range cache.Files {
		// 如果文件不存在或者时间超过7天，从缓存中删除
		if !FileExists(path) || now.Sub(lastDownload).Hours() > CacheExpireHours { // 7天 = 168小时
			delete(cache.Files, path)
			changed = true
		}
	}

	// 如果有变化，保存缓存
	if changed {
		SaveDownloadCache(cache)
	}
}

// NeedsUpdate 检查文件是否需要更新
func NeedsUpdate(filePath string) bool {
	// 如果文件不存在，需要下载
	if !FileExists(filePath) {
		return true
	}

	// 规范化文件路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		// 如果无法获取绝对路径，保守起见返回需要更新
		return true
	}

	// 加载缓存
	cache := LoadDownloadCache()

	// 获取文件最后下载时间
	lastDownload, exists := cache.Files[absPath]
	if !exists {
		// 如果没有记录，需要更新
		return true
	}

	// 检查是否超过缓存过期时间
	return time.Since(lastDownload).Hours() > CacheExpireHours
}
