package downfile

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ProcessDownItems 处理配置组
func ProcessDownItems(client *http.Client, items []DownItem, downloadDir string, forceUpdate bool, keepOld bool, retries int) int {
	successCount := 0
	for _, item := range items {
		// 组合最终文件路径 // 不是绝对路径，才拼接下载目录
		storePath := GetItemFilePath(item.FileName, downloadDir)

		// 检查文件是否存在以及是否需要更新
		fileExists := FileExists(storePath)
		needsUpdate := forceUpdate || !fileExists || (item.KeepUpdated && NeedsUpdate(storePath))

		if fileExists && !needsUpdate {
			fmt.Printf("  文件 %s 已存在且不需要更新，跳过下载\n", item.FileName)
			successCount++
			continue
		}

		//创建目录并存储结果
		err := MakeDirs(storePath, true)
		if err != nil {
			fmt.Printf("  目录[%s]初始化失败:%v\n", item.FileName, err)
			continue
		}
		fmt.Printf("  开始下载 %s...\n", item.Module)

		success := false
		resourceNotFound := false

		// 尝试从每个URL下载
		for _, url := range item.DownloadURLs {
			// 处理GitHub URL
			downloadURL := url
			if strings.Contains(url, "github.com") && strings.Contains(url, "/blob/") {
				downloadURL = ConvertGitHubURL(url)
				fmt.Printf("    转换GitHub URL: %s -> %s\n", url, downloadURL)
			}

			// 尝试下载，支持重试
			for attempt := 1; attempt <= retries; attempt++ {
				if attempt > 1 {
					fmt.Printf("    第 %d 次重试下载...\n", attempt)
				} else {
					fmt.Printf("    尝试从 %s 下载...\n", downloadURL)
				}

				// 使用普通的HTTP请求
				if err := downloadFile(client, downloadURL, storePath, keepOld); err != nil {
					// 检查是否是404错误
					var downloadErr DownloadError
					fmt.Printf("    下载失败: %v\n", err)

					if errors.As(err, &downloadErr) && downloadErr.Type == ErrResourceNotFound {
						fmt.Printf("    资源不存在 (404)，请检查配置中的URL是否正确\n")
						resourceNotFound = true
						break // 404错误不需要重试
					}

					// 如果不是最后一次尝试，则等待后重试
					if attempt < retries {
						waitTime := time.Duration(attempt) * 2 * time.Second
						fmt.Printf("    等待 %v 后重试...\n", waitTime)
						time.Sleep(waitTime)
						continue
					}
					break // 所有重试都失败
				} else {
					fmt.Printf("    成功下载 %s 到 %s\n", item.Module, storePath)
					successCount++
					success = true
					break // 下载成功，不需要继续重试
				}
			}

			if success || resourceNotFound {
				break // 当前URL下载成功或资源不存在，不需要尝试下一个URL
			}
		}

		if !success {
			if resourceNotFound {
				fmt.Printf("  警告: %s 的资源不存在，请检查配置文件中的URL\n", item.Module)
			} else {
				fmt.Printf("  错误: 所有下载源都失败，无法下载 %s\n", item.Module)
			}
		}
	}
	return successCount
}
