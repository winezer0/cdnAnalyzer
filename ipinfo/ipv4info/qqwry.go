package ipv4info

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/ipipdotnet/ipdb-go"
)

// QQWryConfig 数据库配置
type QQWryConfig struct {
	FilePath             string
	MaxConcurrentQueries int
	QueryTimeout         time.Duration
}

// QQWryManager 数据库管理器
type QQWryManager struct {
	config   *QQWryConfig
	data     []byte
	dataLen  uint32
	ipdbCity *ipdb.City
	dataType int
	mu       sync.RWMutex
}

// NewQQWryManager 创建新的数据库管理器
func NewQQWryManager(config *QQWryConfig) *QQWryManager {
	if config.MaxConcurrentQueries <= 0 {
		config.MaxConcurrentQueries = 100
	}
	if config.QueryTimeout == 0 {
		config.QueryTimeout = 5 * time.Second
	}

	return &QQWryManager{
		config: config,
	}
}

const (
	dataTypeDat  = 0
	dataTypeIpdb = 1
)

const (
	indexLen      = 7
	redirectMode1 = 0x01
	redirectMode2 = 0x02
)

type Location struct {
	Country  string // 国家
	Province string // 省份
	City     string // 城市
	District string // 区县
	ISP      string // 运营商
	IP       string // IP地址
}

// IsLoaded 检查数据库是否已加载
func (m *QQWryManager) IsLoaded() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	switch m.dataType {
	case dataTypeIpdb:
		return m.ipdbCity != nil
	case dataTypeDat:
		return m.data != nil && m.dataLen > 0
	default:
		return false
	}
}

// LoadDBFile 从文件加载IP数据库
func (m *QQWryManager) LoadDBFile(filepath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return errors.New("file does not exist: " + filepath)
	}
	if !info.Mode().IsRegular() {
		return errors.New("not a regular file: " + filepath)
	}
	if info.Size() == 0 {
		return errors.New("file is empty: " + filepath)
	}

	body, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	return m.LoadDBData(body)
}

// LoadDBData 从内存加载IP数据库
func (m *QQWryManager) LoadDBData(database []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if string(database[6:11]) == "build" {
		m.dataType = dataTypeIpdb
		loadCity, err := ipdb.NewCityFromBytes(database)
		if err != nil {
			return err
		}
		m.ipdbCity = loadCity
		return nil
	}
	m.data = database
	m.dataLen = uint32(len(m.data))
	return nil
}

// QueryIP 查询IP信息
func (m *QQWryManager) QueryIP(ip string) (location *Location, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	switch m.dataType {
	case dataTypeDat:
		return m.QueryIPByDat(ip)
	case dataTypeIpdb:
		return m.QueryIPByIpdb(ip)
	default:
		return nil, errors.New("data type not support")
	}
}

// Cleanup 清理资源
func (m *QQWryManager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = nil
	m.dataLen = 0
	m.ipdbCity = nil
}

// QueryIPByIpdb 从ipdb查询IP，仅加载ipdb格式数据库时使用
func QueryIPByIpdb(ip string) (location *Location, err error) {
	ret, err := ipdbCity.Find(ip, "CN")
	if err != nil {
		return
	}
	location = SplitCZResult(ret[0], ret[1], ip)
	return location, nil
}

// QueryIPByDat 从内存查询IP
func QueryIPByDat(ip string) (location *Location, err error) {
	// Implementation of QueryIPByDat function
	return nil, errors.New("QueryIPByDat not implemented")
}
