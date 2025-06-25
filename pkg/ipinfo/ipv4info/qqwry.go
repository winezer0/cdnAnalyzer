package ipv4info

import (
	"cdnAnalyzer/pkg/fileutils"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/zu1k/nali/pkg/wry"
)

// Ipv4Location IPv4地理位置数据库管理器
type Ipv4Location struct {
	wry.IPDB[uint32]
	mu sync.RWMutex // 添加读写锁保护并发访问
}

// NewIPv4Location 从文件路径创建新的IPv4地理位置数据库管理器
func NewIPv4Location(filePath string) (*Ipv4Location, error) {
	if filePath == "" {
		return nil, fmt.Errorf("IP数据库[%v]文件路径为空", filePath)
	}

	fileData, err := fileutils.ReadFileBytes(filePath)
	if err != nil {
		return nil, err
	}

	if !checkIPv4File(fileData) {
		return nil, fmt.Errorf("IP数据库[%v]内容存在错误", filePath)
	}

	header := fileData[0:8]
	start := binary.LittleEndian.Uint32(header[:4])
	end := binary.LittleEndian.Uint32(header[4:])

	return &Ipv4Location{
		IPDB: wry.IPDB[uint32]{
			Data:     fileData,
			OffLen:   3,
			IPLen:    4,
			IPCnt:    (end-start)/7 + 1,
			IdxStart: start,
			IdxEnd:   end,
		},
	}, nil
}

// find 内部查询方法，返回详细结果
func (db *Ipv4Location) find(query string) (result *wry.Result, err error) {
	// 验证IP地址
	ip := net.ParseIP(query)
	if ip == nil {
		return nil, errors.New("无效的IPv4地址")
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return nil, errors.New("无效的IPv4地址")
	}

	// 转换为uint32进行查询
	ip4uint := binary.BigEndian.Uint32(ip4)

	// 搜索索引
	offset := db.SearchIndexV4(ip4uint)
	if offset <= 0 {
		return nil, errors.New("查询无效")
	}

	// 解析结果
	reader := wry.NewReader(db.Data)
	reader.Parse(offset + 4)
	return reader.Result.DecodeGBK(), nil
}

// Find 查询IPv4地址的地理位置信息
func (db *Ipv4Location) Find(query string) string {
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
func (db *Ipv4Location) BatchFind(queries []string) map[string]string {
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

// formatLocationResult 清理和格式化地理位置结果
func formatLocationResult(country string) string {
	if country == "" {
		return ""
	}

	// 替换特殊字符
	result := strings.ReplaceAll(country, "–", " ")
	result = strings.ReplaceAll(result, "\t", " ")

	// 去除首尾空格
	result = strings.TrimSpace(result)

	return result
}

// checkIPv4File 检查IPv4数据库文件的有效性
func checkIPv4File(data []byte) bool {
	// 检查最小长度
	if len(data) < 8 {
		return false
	}

	// 解析头部信息
	header := data[0:8]
	start := binary.LittleEndian.Uint32(header[:4])
	end := binary.LittleEndian.Uint32(header[4:])

	// 验证索引范围
	if start >= end {
		return false
	}

	// 验证数据完整性
	if uint32(len(data)) < end+7 {
		return false
	}

	return true
}

// GetDatabaseInfo 获取数据库信息
func (db *Ipv4Location) GetDatabaseInfo() map[string]interface{} {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return map[string]interface{}{
		"ip_count":    db.IPCnt,
		"index_start": db.IdxStart,
		"index_end":   db.IdxEnd,
		"data_size":   len(db.Data),
	}
}

// Close 关闭数据库连接（清理资源）
func (db *Ipv4Location) Close() {
	db.mu.Lock()
	defer db.mu.Unlock()

	// 清理数据
	db.Data = nil
}
