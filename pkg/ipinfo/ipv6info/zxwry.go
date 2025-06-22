package ipv6info

// forked from https://github.com/zu1k/nali
// ipv6db数据使用http://ip.zxinc.org的免费离线数据（更新到2021年）

import (
	"cdnCheck/pkg/fileutils"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/zu1k/nali/pkg/wry"
)

// Ipv6Location IPv6地理位置数据库管理器
type Ipv6Location struct {
	wry.IPDB[uint64]
	mu sync.RWMutex // 添加读写锁保护并发访问
}

// NewIPv6Location 从文件路径创建新的IPv6地理位置数据库管理器
func NewIPv6Location(filePath string) (*Ipv6Location, error) {
	if filePath == "" {
		return nil, fmt.Errorf("IP数据库[%v]文件路径为空", filePath)
	}

	fileData, err := fileutils.ReadFileBytes(filePath)
	if err != nil {
		return nil, err
	}

	if !checkIPv6File(fileData) {
		return nil, fmt.Errorf("IP数据库[%v]内容存在错误", filePath)
	}

	header := fileData[:24]
	offLen := header[6]
	ipLen := header[7]

	start := binary.LittleEndian.Uint64(header[16:24])
	counts := binary.LittleEndian.Uint64(header[8:16])
	end := start + counts*11

	return &Ipv6Location{
		IPDB: wry.IPDB[uint64]{
			Data:     fileData,
			OffLen:   offLen,
			IPLen:    ipLen,
			IPCnt:    counts,
			IdxStart: start,
			IdxEnd:   end,
		},
	}, nil
}

// find 内部查询方法，返回详细结果
func (db *Ipv6Location) find(query string) (result *wry.Result, err error) {
	// 验证IP地址
	ip := net.ParseIP(query)
	if ip == nil {
		return nil, errors.New("无效的IPv6地址")
	}

	ip6 := ip.To16()
	if ip6 == nil {
		return nil, errors.New("无效的IPv6地址")
	}

	// 取前8字节进行查询
	ip6 = ip6[:8]
	ipu64 := binary.BigEndian.Uint64(ip6)

	// 搜索索引
	offset := db.SearchIndexV6(ipu64)
	if offset <= 0 {
		return nil, errors.New("查询无效")
	}

	// 解析结果
	reader := wry.NewReader(db.Data)
	reader.Parse(offset)
	return &reader.Result, nil
}

// Find 查询IPv6地址的地理位置信息
func (db *Ipv6Location) Find(query string) string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	result, err := db.find(query)
	if err != nil || result == nil {
		return ""
	}

	// 清理和格式化结果
	return formatLocationResult(result.Country)
}

// BatchFind 批量查询多个IP地址
func (db *Ipv6Location) BatchFind(queries []string) map[string]string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	results := make(map[string]string, len(queries))

	for _, query := range queries {
		result, err := db.find(query)
		if err != nil || result == nil {
			results[query] = ""
		} else {
			results[query] = formatLocationResult(result.Country)
		}
	}

	return results
}

// GetDatabaseInfo 获取数据库信息
func (db *Ipv6Location) GetDatabaseInfo() map[string]interface{} {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return map[string]interface{}{
		"ip_count":    db.IPCnt,
		"index_start": db.IdxStart,
		"index_end":   db.IdxEnd,
		"data_size":   len(db.Data),
		"off_len":     db.OffLen,
		"ip_len":      db.IPLen,
	}
}

// Close 关闭数据库连接（清理资源）
func (db *Ipv6Location) Close() {
	db.mu.Lock()
	defer db.mu.Unlock()

	// 清理数据
	db.Data = nil
}

// checkIPv6File 检查IPv6数据库文件的有效性
func checkIPv6File(data []byte) bool {
	// 检查最小长度
	if len(data) < 4 {
		return false
	}

	// 检查文件标识
	if string(data[:4]) != "IPDB" {
		return false
	}

	// 检查头部长度
	if len(data) < 24 {
		return false
	}

	// 解析头部信息
	header := data[:24]
	start := binary.LittleEndian.Uint64(header[16:24])
	counts := binary.LittleEndian.Uint64(header[8:16])
	end := start + counts*11

	// 验证索引范围
	if start >= end {
		return false
	}

	// 验证数据完整性
	if uint64(len(data)) < end {
		return false
	}

	return true
}

// formatLocationResult 清理和格式化地理位置结果
func formatLocationResult(country string) string {
	if country == "" {
		return ""
	}

	// 替换特殊字符
	result := strings.ReplaceAll(country, "\t", " ")
	result = strings.ReplaceAll(result, "–", " ")

	// 去除首尾空格
	result = strings.TrimSpace(result)

	return result
}
