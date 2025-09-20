package queryip

import (
	"fmt"
	"github.com/winezer0/cdnAnalyzer/pkg/asninfo"
	"github.com/winezer0/cdnAnalyzer/pkg/iplocate/ipv4qqwry"
	"github.com/winezer0/cdnAnalyzer/pkg/iplocate/ipv6zxwry"
	"net"
)

// IpDbConfig 存储程序配置
type IpDbConfig struct {
	AsnIpvxDb    string
	Ipv4LocateDb string
	Ipv6LocateDb string
}

// DBEngines 存储所有数据库引擎实例
type DBEngines struct {
	AsnEngine  *asninfo.MMDBManager
	IPv4Engine *ipv4qqwry.Ipv4Location
	IPv6Engine *ipv6zxwry.Ipv6Location
}

// IPDbInfo 存储IP解析的中间结果
type IPDbInfo struct {
	IPv4Locations []map[string]string
	IPv6Locations []map[string]string
	IPv4AsnInfos  []asninfo.ASNInfo
	IPv6AsnInfos  []asninfo.ASNInfo
}

// InitDBEngines 初始化所有数据库引擎
func InitDBEngines(config *IpDbConfig) (*DBEngines, error) {
	// 初始化ASN数据库管理器
	asnConfig := &asninfo.MMDBConfig{
		AsnIpvxDb:            config.AsnIpvxDb,
		MaxConcurrentQueries: 100,
	}
	asnManager := asninfo.NewMMDBManager(asnConfig)
	if err := asnManager.InitMMDBConn(); err != nil {
		return nil, fmt.Errorf("初始化ASN数据库失败: %w", err)
	}

	// 初始化IPv4地理位置数据库
	ipv4Engine, err := ipv4qqwry.NewIPv4Location(config.Ipv4LocateDb)
	if err != nil {
		asnManager.Close()
		return nil, fmt.Errorf("初始化IPv4数据库失败: %w", err)
	}

	// 初始化IPv6地理位置数据库
	ipv6Engine, err := ipv6zxwry.NewIPv6Location(config.Ipv6LocateDb)
	if err != nil {
		asnManager.Close()
		ipv4Engine.Close()
		return nil, fmt.Errorf("初始化IPv6数据库失败: %w", err)
	}

	return &DBEngines{
		AsnEngine:  asnManager,
		IPv4Engine: ipv4Engine,
		IPv6Engine: ipv6Engine,
	}, nil
}

// QueryIPInfo 查询IP信息（ASN和地理位置）
func (engines *DBEngines) QueryIPInfo(ipv4s []string, ipv6s []string) (*IPDbInfo, error) {
	info := &IPDbInfo{}

	// 使用通道来并发处理IP信息
	type ipv4Result struct {
		location map[string]string
		asn      asninfo.ASNInfo
	}

	type ipv6Result struct {
		location map[string]string
		asn      asninfo.ASNInfo
	}

	ipv4Chan := make(chan ipv4Result, len(ipv4s))
	ipv6Chan := make(chan ipv6Result, len(ipv6s))

	// 并发处理IPv4信息
	for _, ipv4 := range ipv4s {
		go func(ip string) {
			// 查询位置信息
			location := engines.IPv4Engine.Find(ip)
			locationMap := map[string]string{ip: location}

			// 查询ASN信息
			asnInfo := engines.AsnEngine.FindASN(ip)

			ipv4Chan <- ipv4Result{
				location: locationMap,
				asn:      *asnInfo,
			}
		}(ipv4)
	}

	// 并发处理IPv6信息
	for _, ipv6 := range ipv6s {
		go func(ip string) {
			// 查询位置信息
			location := engines.IPv6Engine.Find(ip)
			locationMap := map[string]string{ip: location}

			// 查询ASN信息
			asnInfo := engines.AsnEngine.FindASN(ip)

			ipv6Chan <- ipv6Result{
				location: locationMap,
				asn:      *asnInfo,
			}
		}(ipv6)
	}

	// 收集IPv4结果
	for i := 0; i < len(ipv4s); i++ {
		result := <-ipv4Chan
		info.IPv4Locations = append(info.IPv4Locations, result.location)
		info.IPv4AsnInfos = append(info.IPv4AsnInfos, result.asn)
	}

	// 收集IPv6结果
	for i := 0; i < len(ipv6s); i++ {
		result := <-ipv6Chan
		info.IPv6Locations = append(info.IPv6Locations, result.location)
		info.IPv6AsnInfos = append(info.IPv6AsnInfos, result.asn)
	}

	return info, nil
}

// QuerySingleIP 查询单个IP的信息
func (engines *DBEngines) QuerySingleIP(ip string) (string, *asninfo.ASNInfo) {
	// 查询位置信息
	var location string
	if isIPv4(ip) {
		location = engines.IPv4Engine.Find(ip)
	} else {
		location = engines.IPv6Engine.Find(ip)
	}

	// 查询ASN信息
	asnInfo := engines.AsnEngine.FindASN(ip)

	return location, asnInfo
}

// Close 关闭所有数据库连接
func (engines *DBEngines) Close() error {
	var lastErr error

	// 关闭ASN数据库
	if engines.AsnEngine != nil {
		if err := engines.AsnEngine.Close(); err != nil {
			lastErr = err
		}
	}

	// 关闭IPv4数据库
	if engines.IPv4Engine != nil {
		engines.IPv4Engine.Close()
	}

	// 关闭IPv6数据库
	if engines.IPv6Engine != nil {
		engines.IPv6Engine.Close()
	}

	return lastErr
}

// isIPv4 判断是否为IPv4地址
func isIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() != nil
}
