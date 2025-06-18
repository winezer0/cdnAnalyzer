package asninfo

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/oschwald/maxminddb-golang"
)

// MMDBConfig 数据库配置结构
type MMDBConfig struct {
	IPv4Path             string
	IPv6Path             string
	MaxConcurrentQueries int
	QueryTimeout         time.Duration
}

// MMDBManager 数据库管理器
type MMDBManager struct {
	config    *MMDBConfig
	mmDb      map[string]*maxminddb.Reader
	queryChan chan struct{}
	mu        sync.RWMutex
}

// NewMMDBManager 创建新的数据库管理器
func NewMMDBManager(config *MMDBConfig) *MMDBManager {
	if config.MaxConcurrentQueries <= 0 {
		config.MaxConcurrentQueries = 100
	}
	if config.QueryTimeout == 0 {
		config.QueryTimeout = 5 * time.Second
	}

	return &MMDBManager{
		config:    config,
		mmDb:      make(map[string]*maxminddb.Reader),
		queryChan: make(chan struct{}, config.MaxConcurrentQueries),
	}
}

var mmDb = map[string]*maxminddb.Reader{}

type ASNInfo struct {
	IP                 string `json:"ip"`
	IPVersion          int    `json:"ip_version"`
	FoundASN           bool   `json:"found_asn"`
	OrganisationNumber uint64 `json:"as_number"`
	OrganisationName   string `json:"as_organisation"`
}

type ASNRecord struct {
	AutonomousSystemNumber uint64 `maxminddb:"autonomous_system_number"`
	AutonomousSystemOrg    string `maxminddb:"autonomous_system_organization"`
}

func NewASNInfo(ipString string, ipVersion int) *ASNInfo {
	return &ASNInfo{ipString, ipVersion, false, 0, ""}
}

// InitMMDBConn 初始化 MaxMind ASN 数据库连接
func (m *MMDBManager) InitMMDBConn() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 定义要加载的数据库连接信息
	dbPaths := map[string]string{
		"ipv4": m.config.IPv4Path,
		"ipv6": m.config.IPv6Path,
	}

	for connId, dbPath := range dbPaths {
		// 如果已存在连接，跳过
		if _, ok := m.mmDb[connId]; ok {
			continue
		}

		// 检查文件是否存在
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			return fmt.Errorf("数据库文件不存在: %s", dbPath)
		}

		// 打开数据库
		conn, err := maxminddb.Open(dbPath)
		if err != nil {
			return fmt.Errorf("打开数据库失败 [%s]: %v", dbPath, err)
		}

		// 存入 map
		m.mmDb[connId] = conn
	}

	return nil
}

// CloseMMDBConn 关闭数据库连接
func (m *MMDBManager) CloseMMDBConn() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for connectionId, conn := range m.mmDb {
		if err := conn.Close(); err != nil {
			lastErr = fmt.Errorf("关闭数据库失败 [%s]", connectionId)
			continue
		}
		delete(m.mmDb, connectionId)
	}
	return lastErr
}

// IsInitialized 检查数据库是否已初始化
func (m *MMDBManager) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.mmDb) > 0
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
	connectionId := "ipv" + strconv.Itoa(ipVersion)

	m.mu.RLock()
	reader, ok := m.mmDb[connectionId]
	m.mu.RUnlock()

	if !ok {
		return asnInfo
	}

	var asnRecord ASNRecord
	if err := reader.Lookup(ip, &asnRecord); err != nil {
		return asnInfo
	}

	if asnRecord.AutonomousSystemNumber > 0 {
		asnInfo.OrganisationNumber = asnRecord.AutonomousSystemNumber
		asnInfo.OrganisationName = asnRecord.AutonomousSystemOrg
		asnInfo.FoundASN = true
	}

	return asnInfo
}

// ASNToIPRanges 通过ASN号反查所有IP段
func (m *MMDBManager) ASNToIPRanges(targetASN uint64) ([]*net.IPNet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var findIPs []*net.IPNet
	connectionIds := []string{"ipv4", "ipv6"}

	for _, connectionId := range connectionIds {
		reader, ok := m.mmDb[connectionId]
		if !ok {
			continue
		}

		networks := reader.Networks()
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
	}

	return findIPs, nil
}

// CountMMDBSize 统计数据库大小
func (m *MMDBManager) CountMMDBSize(connectionId string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	reader, ok := m.mmDb[connectionId]
	if !ok {
		return 0, fmt.Errorf("数据库未找到: %s", connectionId)
	}

	count := 0
	networks := reader.Networks()
	for networks.Next() {
		count++
	}

	if err := networks.Err(); err != nil {
		return 0, fmt.Errorf("统计数据库大小时发生错误: %v", err)
	}

	return count, nil
}
