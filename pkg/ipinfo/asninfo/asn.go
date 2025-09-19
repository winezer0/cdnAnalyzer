package asninfo

import (
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/oschwald/maxminddb-golang"
)

// MMDBConfig 数据库配置结构
type MMDBConfig struct {
	// 使用统一的数据库路径，支持同时包含IPv4和IPv6数据的数据库
	UnifiedDBPath        string
	MaxConcurrentQueries int
}

// MMDBManager 数据库管理器
type MMDBManager struct {
	config    *MMDBConfig
	mmDb      *maxminddb.Reader
	queryChan chan struct{}
	mu        sync.RWMutex
}

// NewMMDBManager 创建新的数据库管理器
func NewMMDBManager(config *MMDBConfig) *MMDBManager {
	if config.MaxConcurrentQueries <= 0 {
		config.MaxConcurrentQueries = 100
	}
	return &MMDBManager{
		config:    config,
		queryChan: make(chan struct{}, config.MaxConcurrentQueries),
	}
}

type ASNInfo struct {
	IP                 string `json:"ip"`
	IPVersion          int    `json:"ip_version"`
	FoundASN           bool   `json:"found_asn"`
	OrganisationNumber uint32 `json:"as_number"`
	OrganisationName   string `json:"as_organisation"`
}

// ASNRecord 定义结构体映射MMDB中的ASN数据结构
type ASNRecord struct {
	AutonomousSystemNumber       uint32 `maxminddb:"autonomous_system_number"`
	AutonomousSystemOrganization string `maxminddb:"autonomous_system_organization"`
}

func NewASNInfo(ipString string, ipVersion int) *ASNInfo {
	return &ASNInfo{ipString, ipVersion, false, 0, ""}
}

// InitMMDBConn 初始化 MaxMind ASN 数据库连接
func (m *MMDBManager) InitMMDBConn() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已存在连接，跳过
	if m.mmDb != nil {
		return nil
	}

	// 检查文件是否存在
	if _, err := os.Stat(m.config.UnifiedDBPath); os.IsNotExist(err) {
		return fmt.Errorf("数据库文件不存在: %s", m.config.UnifiedDBPath)
	}

	// 打开数据库
	conn, err := maxminddb.Open(m.config.UnifiedDBPath)
	if err != nil {
		return fmt.Errorf("打开数据库失败 [%s]: %v", m.config.UnifiedDBPath, err)
	}

	// 存入实例
	m.mmDb = conn
	return nil
}

// Close 关闭数据库连接
func (m *MMDBManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.mmDb != nil {
		if err := m.mmDb.Close(); err != nil {
			return fmt.Errorf("关闭数据库失败: %v", err)
		}
		m.mmDb = nil
	}
	return nil
}

// IsInitialized 检查数据库是否已初始化
func (m *MMDBManager) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mmDb != nil
}

// FindASN 查询单个IP的ASN信息
func (m *MMDBManager) FindASN(ipStr string) *ASNInfo {
	// 获取查询许可
	select {
	case m.queryChan <- struct{}{}:
		defer func() { <-m.queryChan }()
	default:
		return &ASNInfo{
			IP:        ipStr,
			IPVersion: getIpVersion(ipStr),
			FoundASN:  false,
		}
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return &ASNInfo{
			IP:        ipStr,
			IPVersion: getIpVersion(ipStr),
			FoundASN:  false,
		}
	}

	ipVersion := getIpVersion(ipStr)
	asnInfo := NewASNInfo(ipStr, ipVersion)

	m.mu.RLock()
	reader := m.mmDb
	m.mu.RUnlock()

	// 如果数据库未初始化，返回空结果
	if reader == nil {
		return asnInfo
	}

	var asnRecord ASNRecord
	if err := reader.Lookup(ip, &asnRecord); err != nil {
		return asnInfo
	}

	if asnRecord.AutonomousSystemNumber > 0 {
		asnInfo.OrganisationNumber = asnRecord.AutonomousSystemNumber
		asnInfo.OrganisationName = asnRecord.AutonomousSystemOrganization
		asnInfo.FoundASN = true
	}

	return asnInfo
}

// ASNToIPRanges 通过ASN号反查所有IP段
func (m *MMDBManager) ASNToIPRanges(targetASN uint32) ([]*net.IPNet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.mmDb == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var findIPs []*net.IPNet

	networks := m.mmDb.Networks()
	for networks.Next() {
		var record ASNRecord
		ipNet, err := networks.Network(&record)
		if err != nil {
			return nil, fmt.Errorf("解析网络段失败: %v", err)
		}
		if record.AutonomousSystemNumber == targetASN {
			findIPs = append(findIPs, ipNet)
		}
	}

	if err := networks.Err(); err != nil {
		return nil, fmt.Errorf("遍历数据库时发生错误: %v", err)
	}

	return findIPs, nil
}

// CountMMDBSize 统计数据库大小
func (m *MMDBManager) CountMMDBSize() (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.mmDb == nil {
		return 0, fmt.Errorf("数据库未初始化")
	}

	count := 0
	networks := m.mmDb.Networks()
	for networks.Next() {
		count++
	}

	if err := networks.Err(); err != nil {
		return 0, fmt.Errorf("统计数据库大小时发生错误: %v", err)
	}

	return count, nil
}

// BatchFindASN 批量查询多个IP的ASN信息
func (m *MMDBManager) BatchFindASN(ips []string) []*ASNInfo {
	results := make([]*ASNInfo, len(ips))

	for i, ip := range ips {
		results[i] = m.FindASN(ip)
	}

	return results
}