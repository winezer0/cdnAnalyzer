package asninfo

import (
	"fmt"
	"sync"
	"time"
)

// BatchQueryOptions 批量查询选项
type BatchQueryOptions struct {
	Timeout    time.Duration
	MaxWorkers int
}

// BatchQueryResult 批量查询结果
type BatchQueryResult struct {
	IP     string
	Result *ASNInfo
	Error  error
}

// BatchFindASN 批量查询ASN信息
func (m *MMDBManager) BatchFindASN(ips []string, opts *BatchQueryOptions) []BatchQueryResult {
	if opts == nil {
		opts = &BatchQueryOptions{
			Timeout:    m.config.QueryTimeout,
			MaxWorkers: m.config.MaxConcurrentQueries,
		}
	}

	results := make([]BatchQueryResult, len(ips))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, opts.MaxWorkers)

	for i, ip := range ips {
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量

		go func(index int, ipStr string) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量

			// 使用带超时的查询
			done := make(chan struct{})
			var result *ASNInfo
			var err error

			go func() {
				result = m.FindASN(ipStr)
				close(done)
			}()

			select {
			case <-done:
				results[index] = BatchQueryResult{
					IP:     ipStr,
					Result: result,
					Error:  err,
				}
			case <-time.After(opts.Timeout):
				results[index] = BatchQueryResult{
					IP:    ipStr,
					Error: fmt.Errorf("查询超时"),
				}
			}
		}(i, ip)
	}

	wg.Wait()
	return results
}
