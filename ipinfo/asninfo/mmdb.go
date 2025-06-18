package asninfo

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/oschwald/maxminddb-golang"
)

// MMDBConfig 数据库配置结构
type MMDBConfig struct {
	IPv4Path             string
	IPv6Path             string
	MaxConcurrentQueries int
	CacheSize            int
	QueryTimeout         time.Duration
}

// MMDBError 自定义错误类型
type MMDBError struct {
	Code    int
	Message string
	Err     error
}

func (e *MMDBError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
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

// IsInitialized 检查是否已经有数据库连接被初始化
func IsInitialized() bool {
	return len(mmDb) > 0
}

// InitMMDBConn 初始化 MaxMind ASN 数据库连接
func (m *MMDBManager) InitMMDBConn() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	type dbInfo struct {
		filePath     string
		connectionId string
	}

	dbFiles := []dbInfo{
		{m.config.IPv4Path, "ipv4"},
		{m.config.IPv6Path, "ipv6"},
	}

	for _, db := range dbFiles {
		if _, err := os.Stat(db.filePath); os.IsNotExist(err) {
			return &MMDBError{
				Code:    1,
				Message: fmt.Sprintf("数据库文件不存在: %s", db.filePath),
				Err:     err,
			}
		}

		if _, ok := m.mmDb[db.connectionId]; ok {
			continue
		}

		conn, err := maxminddb.Open(db.filePath)
		if err != nil {
			return &MMDBError{
				Code:    2,
				Message: fmt.Sprintf("打开数据库失败 [%s]", db.filePath),
				Err:     err,
			}
		}

		m.mmDb[db.connectionId] = conn
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
			lastErr = &MMDBError{
				Code:    3,
				Message: fmt.Sprintf("关闭数据库失败 [%s]", connectionId),
				Err:     err,
			}
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
					IP: ipStr,
					Error: &MMDBError{
						Code:    4,
						Message: "查询超时",
					},
				}
			}
		}(i, ip)
	}

	wg.Wait()
	return results
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

func getIpVersion(ipString string) int {
	ipVersion := 4
	if strings.Contains(ipString, ":") {
		ipVersion = 6
	}

	return ipVersion
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
				return nil, &MMDBError{
					Code:    5,
					Message: "解析网络段失败",
					Err:     err,
				}
			}
			if record.AutonomousSystemNumber == targetASN {
				findIPs = append(findIPs, ipNet)
			}
		}

		if err := networks.Err(); err != nil {
			return nil, &MMDBError{
				Code:    6,
				Message: "遍历数据库时发生错误",
				Err:     err,
			}
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
		return 0, &MMDBError{
			Code:    7,
			Message: fmt.Sprintf("数据库未找到: %s", connectionId),
		}
	}

	count := 0
	networks := reader.Networks()
	for networks.Next() {
		count++
	}

	if err := networks.Err(); err != nil {
		return 0, &MMDBError{
			Code:    8,
			Message: "统计数据库大小时发生错误",
			Err:     err,
		}
	}

	return count, nil
}

// PrintASNInfo 打印ASN信息
func PrintASNInfo(ipInfo *ASNInfo) {
	if ipInfo == nil {
		fmt.Println("ASN信息为空")
		return
	}

	fmt.Printf("IP: %15s | 版本: %d | 找到ASN: %v | ASN: %6d | 组织: %s\n",
		ipInfo.IP,
		ipInfo.IPVersion,
		ipInfo.FoundASN,
		ipInfo.OrganisationNumber,
		ipInfo.OrganisationName,
	)
}
