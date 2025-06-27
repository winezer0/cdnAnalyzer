package downfile

import (
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

// ProgressTracker 下载进度跟踪器
type ProgressTracker struct {
	BytesCount   *atomic.Int64 // 已下载字节数
	FileSize     int64         // 文件总大小
	StartTime    time.Time     // 下载开始时间
	LastUpdate   time.Time     // 上次更新时间
	LastSize     int64         // 上次记录的大小
	Speed        float64       // 当前下载速度
	Done         chan struct{} // 完成信号
	Name         string        // 下载的文件名
	Cancel       func()        // 用于取消下载的函数
	CancelReason atomic.Value  // 取消原因
}

// NewProgressTracker 创建新的进度跟踪器
func NewProgressTracker(fileSize int64, name string) *ProgressTracker {
	now := time.Now()
	var counter atomic.Int64

	tracker := &ProgressTracker{
		BytesCount: &counter,
		FileSize:   fileSize,
		StartTime:  now,
		LastUpdate: now,
		LastSize:   0,
		Speed:      0,
		Done:       make(chan struct{}),
		Name:       name,
		Cancel:     func() {}, // 默认空函数
	}

	// 初始化取消原因为空字符串
	tracker.CancelReason.Store("")

	return tracker
}

// Close 关闭进度跟踪器
func (pt *ProgressTracker) Close() {
	close(pt.Done)
}

// GetCountingWriter 获取计数Writer
func (pt *ProgressTracker) GetCountingWriter(w io.Writer) io.Writer {
	return &CountingWriter{
		Writer:     w,
		BytesCount: pt.BytesCount,
	}
}

// MonitorSpeed 监控下载速度
func (pt *ProgressTracker) MonitorSpeed() {
	speedCheckTicker := time.NewTicker(SpeedCheckInterval * time.Second)
	defer speedCheckTicker.Stop()

	for {
		select {
		case <-speedCheckTicker.C:
			// 检查当前下载速度是否低于最小要求
			if pt.Speed < MinRequiredSpeed {
				// 提示用户当前速度过低并取消下载
				fmt.Printf("\r    下载已取消: 速度过低 (%s/s)，低于最小要求 (%s/s)，网络可能存在问题\n",
					formatSize(int64(pt.Speed)),
					formatSize(int64(MinRequiredSpeed)))

				// 记录取消原因
				pt.CancelReason.Store(ErrLowSpeed)

				// 取消下载
				pt.Cancel()
				return
			}
		case <-pt.Done:
			return
		}
	}
}

// DisplayProgress 显示下载进度
func (pt *ProgressTracker) DisplayProgress() {
	updateInterval := time.Duration(ProgressUpdateInterval) * time.Millisecond

	if pt.FileSize > 0 {
		// 已知文件大小的情况
		ticker := time.NewTicker(updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				pt.updateSpeed()
				pt.displayKnownSizeProgress()
			case <-pt.Done:
				return
			}
		}
	} else {
		// 未知文件大小的情况
		ticker := time.NewTicker(updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				pt.updateSpeed()
				pt.displayUnknownSizeProgress()
			case <-pt.Done:
				return
			}
		}
	}
}

// updateSpeed 更新下载速度
func (pt *ProgressTracker) updateSpeed() {
	// 使用原子变量读取当前下载大小
	currentSize := pt.BytesCount.Load()

	// 计算下载速度 (bytes/second)
	currentTime := time.Now()
	timeElapsed := currentTime.Sub(pt.LastUpdate).Seconds()
	if timeElapsed > 0 {
		instantSpeed := float64(currentSize-pt.LastSize) / timeElapsed

		// 平滑速度计算 (指数移动平均)
		if pt.Speed == 0 {
			pt.Speed = instantSpeed
		} else {
			pt.Speed = 0.7*pt.Speed + 0.3*instantSpeed
		}

		pt.LastSize = currentSize
		pt.LastUpdate = currentTime
	}
}

// displayKnownSizeProgress 显示已知文件大小的下载进度
func (pt *ProgressTracker) displayKnownSizeProgress() {
	currentSize := pt.BytesCount.Load()
	progress := float64(currentSize) / float64(pt.FileSize) * 100
	speedStr := formatSize(int64(pt.Speed)) + "/s"

	if pt.Speed > MinValidSpeed {
		// 只有当速度大于最小有效值时才计算剩余时间
		remainingBytes := pt.FileSize - currentSize
		remainingSeconds := float64(remainingBytes) / pt.Speed
		// 限制最大预估时间为24小时，避免不合理的估计
		if remainingSeconds > 86400 { // 24小时 = 86400秒
			remainingSeconds = 86400
		}
		remainingTime := time.Duration(remainingSeconds) * time.Second

		fmt.Printf("\r    下载进度: %.1f%% (%s/%s) 速度: %s 剩余时间: %s",
			progress,
			formatSize(currentSize),
			formatSize(pt.FileSize),
			speedStr,
			formatDuration(remainingTime))
	} else if pt.Speed > 0 {
		// 速度极低但不为0，显示速度但不显示剩余时间
		fmt.Printf("\r    下载进度: %.1f%% (%s/%s) 速度: %s 剩余时间: 未知",
			progress,
			formatSize(currentSize),
			formatSize(pt.FileSize),
			speedStr)
	} else {
		// 速度为0，等待恢复
		fmt.Printf("\r    下载进度: %.1f%% (%s/%s) 等待数据传输...",
			progress,
			formatSize(currentSize),
			formatSize(pt.FileSize))
	}
}

// displayUnknownSizeProgress 显示未知文件大小的下载进度
func (pt *ProgressTracker) displayUnknownSizeProgress() {
	currentSize := pt.BytesCount.Load()
	speedStr := formatSize(int64(pt.Speed)) + "/s"

	if pt.Speed > MinValidSpeed {
		fmt.Printf("\r    已下载: %s 速度: %s",
			formatSize(currentSize),
			speedStr)
	} else {
		fmt.Printf("\r    已下载: %s 等待数据传输...",
			formatSize(currentSize))
	}
}

// DisplaySummary 显示下载摘要
func (pt *ProgressTracker) DisplaySummary() {
	// 如果是因为取消而终止的下载，不显示下载摘要
	if pt.GetCancelReason() != "" {
		return
	}

	// 清除进度条行
	fmt.Print("\r                                                                                          \r")

	// 显示总下载时间和平均速度
	totalTime := time.Since(pt.StartTime)
	totalBytes := pt.BytesCount.Load()
	avgSpeed := float64(totalBytes) / totalTime.Seconds()
	fmt.Printf("    下载完成: 总大小 %s, 用时 %s, 平均速度 %s/s\n",
		formatSize(totalBytes),
		formatDuration(totalTime),
		formatSize(int64(avgSpeed)))
}

// GetCancelReason 获取下载取消的原因
func (pt *ProgressTracker) GetCancelReason() string {
	return pt.CancelReason.Load().(string)
}

// SetCancelReason 设置下载取消的原因
func (pt *ProgressTracker) SetCancelReason(reason string) {
	pt.CancelReason.Store(reason)
}
