package ip2region

import (
	"fmt"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/winezer0/cdnAnalyzer/pkg/iplocate"
	"sync"
)

// IP2Region 实现了统一的IP信息查询接口
type IP2Region struct {
	Version  *xdb.Version
	Searcher *xdb.Searcher
	mu       sync.RWMutex
	DbPath   string
}

// 确保 IP2Region 实现了 ipinfo.IPInfo 接口
var _ iplocate.IPInfo = (*IP2Region)(nil)

// NewIP2Region 创建一个新的IP2Region实例
// Version: xdb.IPv4 或 xdb.IPv6
// DbPath: 对应的xdb文件路径
func NewIP2Region(version *xdb.Version, dbPath string) (*IP2Region, error) {
	// 验证xdb文件
	if err := xdb.VerifyFromFile(dbPath); err != nil {
		return nil, fmt.Errorf("xdb文件验证失败: %w", err)
	}

	// 创建searcher
	searcher, err := xdb.NewWithFileOnly(version, dbPath)
	if err != nil {
		return nil, fmt.Errorf("创建searcher失败: %w", err)
	}

	return &IP2Region{
		Version:  version,
		DbPath:   dbPath,
		Searcher: searcher,
	}, nil
}

// NewIP2RegionWithVectorIndex 创建一个使用VectorIndex缓存的IP2Region实例
// Version: xdb.IPv4 或 xdb.IPv6
// DbPath: 对应的xdb文件路径
// vectorIndex: 预加载的VectorIndex缓存
func NewIP2RegionWithVectorIndex(version *xdb.Version, dbPath string, vectorIndex []byte) (*IP2Region, error) {
	// 验证xdb文件
	if err := xdb.VerifyFromFile(dbPath); err != nil {
		return nil, fmt.Errorf("xdb文件验证失败: %w", err)
	}

	// 创建searcher
	searcher, err := xdb.NewWithVectorIndex(version, dbPath, vectorIndex)
	if err != nil {
		return nil, fmt.Errorf("创建searcher失败: %w", err)
	}

	return &IP2Region{
		Version:  version,
		DbPath:   dbPath,
		Searcher: searcher,
	}, nil
}

// Find 查询单个IP地址的地理位置信息
func (ipr *IP2Region) Find(query string) string {
	ipr.mu.RLock()
	defer ipr.mu.RUnlock()

	if ipr.Searcher == nil {
		return ""
	}

	region, err := ipr.Searcher.SearchByStr(query)
	if err != nil {
		return ""
	}

	return region
}

// BatchFind 批量查询多个IP地址
func (ipr *IP2Region) BatchFind(queries []string) map[string]string {
	ipr.mu.RLock()
	defer ipr.mu.RUnlock()

	results := make(map[string]string, len(queries))

	for _, query := range queries {
		if ipr.Searcher == nil {
			results[query] = ""
			continue
		}

		region, err := ipr.Searcher.SearchByStr(query)
		if err != nil {
			results[query] = ""
		} else {
			results[query] = region
		}
	}

	return results
}

// GetDatabaseInfo 获取数据库信息
func (ipr *IP2Region) GetDatabaseInfo() map[string]interface{} {
	ipr.mu.RLock()
	defer ipr.mu.RUnlock()

	return map[string]interface{}{
		"Version": ipr.Version,
		"db_path": ipr.DbPath,
	}
}

// Close 关闭数据库连接（清理资源）
func (ipr *IP2Region) Close() {
	ipr.mu.Lock()
	defer ipr.mu.Unlock()

	if ipr.Searcher != nil {
		ipr.Searcher.Close()
		ipr.Searcher = nil
	}
}

// LoadVectorIndexFromFile 从文件加载VectorIndex缓存
func LoadVectorIndexFromFile(dbPath string) ([]byte, error) {
	return xdb.LoadVectorIndexFromFile(dbPath)
}
