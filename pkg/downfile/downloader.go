package downfile

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

// DownloadError 自定义错误类型
type DownloadError struct {
	StatusCode int
	Message    string
	Type       string
}

func (e DownloadError) Error() string {
	return e.Message
}

// downloadFile 下载文件
func downloadFile(client *http.Client, downloadUrl, storePath string, keepOldFile bool) error {
	// 创建目标文件的目录（如果不存在）
	if err := os.MkdirAll(filepath.Dir(storePath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建临时文件（使用唯一名称避免冲突）
	tempFile := storePath + fmt.Sprintf(".%d.download", time.Now().UnixNano())
	out, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}

	resp, err := httpGet(client, downloadUrl, err)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 获取文件大小
	fileSize := resp.ContentLength
	fileName := filepath.Base(storePath)

	// 使用defer确保在函数退出时处理临时文件
	var downloadSuccess bool
	defer func() {
		out.Close()
		if !downloadSuccess {
			// 下载失败，删除临时文件
			os.Remove(tempFile)
		}
	}()

	// 创建进度跟踪器
	tracker := NewProgressTracker(fileSize, fileName)
	defer tracker.Close()

	// 启动进度监控协程
	go tracker.MonitorSpeed()
	go tracker.DisplayProgress()

	// 创建计数Writer
	countingWriter := tracker.GetCountingWriter(out)

	// 复制内容，支持取消
	buf := make([]byte, DownloadBufferSize)
	_, err = copyBuffer(countingWriter, resp.Body, buf)

	// 检查是否是因为速度过低取消导致的错误
	cancelReason := tracker.GetCancelReason()
	if cancelReason == ErrLowSpeed {
		return DownloadError{
			Message: fmt.Sprintf("下载已取消: 速度过低，低于最小要求 (%s/s)，网络可能存在问题",
				formatSize(int64(MinRequiredSpeed))),
			Type: ErrLowSpeed,
		}
	}

	// 检查其他错误
	if err != nil {
		return fmt.Errorf("下载内容失败: %w", err)
	}

	// 显示下载摘要
	tracker.DisplaySummary()

	// 关闭文件，确保内容写入磁盘
	if err := out.Close(); err != nil {
		return fmt.Errorf("关闭文件失败: %w", err)
	}

	// 标记下载成功，避免在defer中删除临时文件
	downloadSuccess = true

	// 处理旧文件（如果存在）
	if FileExists(storePath) {
		if keepOldFile {
			// 保留旧文件，重命名为.old
			oldFilePath := storePath + ".old"
			// 如果已经存在.old文件，先删除它
			if FileExists(oldFilePath) {
				if err := os.Remove(oldFilePath); err != nil {
					return fmt.Errorf("删除旧的备份文件失败: %w", err)
				}
			}
			// 重命名当前文件为.old
			if err := os.Rename(storePath, oldFilePath); err != nil {
				return fmt.Errorf("备份旧文件失败: %w", err)
			}
			fmt.Printf("    已备份旧文件为: %s\n", oldFilePath)
		} else {
			// 不保留旧文件，直接删除
			if err := os.Remove(storePath); err != nil {
				return fmt.Errorf("错误:删除旧文件失败: %w", err)
			}
		}
	}

	// 重命名临时文件为最终文件名
	if err := os.Rename(tempFile, storePath); err != nil {
		return fmt.Errorf("错误: 重命名临时文件失败: %w", err)
	}

	// 更新文件下载时间缓存
	if err := UpdateFileDownloadTime(storePath); err != nil {
		fmt.Printf("    错误: 更新下载缓存失败: %v\n", err)
	}

	return nil
}

func httpGet(client *http.Client, downloadUrl string, err error) (*http.Response, error) {
	// 创建HTTP请求
	req, err := http.NewRequest("GET", downloadUrl, nil)
	if err != nil {
		return nil, err
	}

	// 设置User-Agent以避免某些服务器的限制
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		// 对于404错误，返回特殊错误类型
		if resp.StatusCode == http.StatusNotFound {
			return nil, DownloadError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("资源不存在，HTTP状态码: %d (404 Not Found)", resp.StatusCode),
				Type:       ErrResourceNotFound,
			}
		}
		return nil, fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}
	return resp, nil
}

// CountingWriter 是一个包装io.Writer的结构，用于跟踪写入的字节数
type CountingWriter struct {
	Writer     io.Writer
	BytesCount *atomic.Int64
}

// Write 实现io.Writer接口，并原子地更新计数器
func (w *CountingWriter) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	if n > 0 {
		w.BytesCount.Add(int64(n))
	}
	return n, err
}

// copyBuffer 标准的数据复制
func copyBuffer(dst io.Writer, src io.Reader, buf []byte) (written int64, err error) {
	if buf == nil {
		buf = make([]byte, 32*1024)
	}

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				return
			}
			if nr != nw {
				err = io.ErrShortWrite
				return
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			return
		}
	}
}
