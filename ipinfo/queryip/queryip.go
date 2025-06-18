package queryip

import (
	"cdnCheck/cdncheck"
	"cdnCheck/ipinfo/asninfo"
	"cdnCheck/ipinfo/ipv4info"
	"cdnCheck/ipinfo/ipv6info"
	"cdnCheck/maputils"
	"cdnCheck/models"
	"fmt"
)

// IpDbConfig 存储程序配置
type IpDbConfig struct {
	AsnIpv4Db    string
	AsnIpv6Db    string
	Ipv4LocateDb string
	Ipv6LocateDb string
}

// IPDbInfo 存储IP解析的中间结果
type IPDbInfo struct {
	IPv4Locations []map[string]string
	IPv6Locations []map[string]string
	IPv4AsnInfos  []asninfo.ASNInfo
	IPv6AsnInfos  []asninfo.ASNInfo
}

// DBEngines 存储所有数据库引擎实例
type DBEngines struct {
	IPv6Engine *ipv6info.Ipv6Location
}

// IPProcessor IP信息查询处理器
type IPProcessor struct {
	Engines *DBEngines
	Config  *IpDbConfig
}

// NewIPProcessor 创建新的IP处理器
func NewIPProcessor(engines *DBEngines, config *IpDbConfig) *IPProcessor {
	return &IPProcessor{
		Engines: engines,
		Config:  config,
	}
}

// InitDBEngines 初始化所有数据库引擎
func InitDBEngines(config *IpDbConfig) (*DBEngines, error) {
	// 初始化ASN数据库
	if err := asninfo.InitMMDBConn(config.AsnIpv4Db, config.AsnIpv6Db); err != nil {
		return nil, fmt.Errorf("初始化ASN数据库失败: %w", err)
	}

	// 初始化IPv4数据库
	if err := ipv4info.LoadDBFile(config.Ipv4LocateDb); err != nil {
		return nil, fmt.Errorf("初始化IPv4数据库失败: %w", err)
	}

	// 初始化IPv6数据库
	ipv6Engine, err := ipv6info.NewIPv6Location(config.Ipv6LocateDb)
	if err != nil {
		return nil, fmt.Errorf("初始化IPv6数据库失败: %w", err)
	}

	return &DBEngines{
		IPv6Engine: ipv6Engine,
	}, nil
}

// parseIPInfo 解析IP信息并返回中间结果
func parseIPInfo(ipv4s, ipv6s []string, engines *DBEngines) (*IPDbInfo, error) {
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
			location, _ := ipv4info.QueryIP(ip)
			locationToStr := ipv4info.LocationToStr(*location)
			locationMap := map[string]string{ip: locationToStr}

			// 查询ASN信息
			asnInfo := asninfo.FindASN(ip)

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
			asnInfo := asninfo.FindASN(ip)

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

// ProcessIPInfo 处理单个IP信息
func (p *IPProcessor) ProcessIPInfo(ipv4s, ipv6s []string) (*IPDbInfo, error) {
	return parseIPInfo(ipv4s, ipv6s, p.Engines)
}

// ProcessCheckInfo 处理单个CheckInfo的IP信息
func (p *IPProcessor) ProcessCheckInfo(checkInfo *models.CheckInfo) error {
	ipInfo, err := p.ProcessIPInfo(checkInfo.A, checkInfo.AAAA)
	if err != nil {
		return fmt.Errorf("处理IP信息失败: %w", err)
	}

	// 更新CheckInfo的IP信息
	checkInfo.Ipv6Asn = ipInfo.IPv6AsnInfos
	checkInfo.Ipv4Asn = ipInfo.IPv4AsnInfos
	checkInfo.Ipv4Locate = ipInfo.IPv4Locations
	checkInfo.Ipv6Locate = ipInfo.IPv6Locations

	return nil
}

// ProcessCheckInfos 批量处理CheckInfo列表
func (p *IPProcessor) ProcessCheckInfos(checkInfos []*models.CheckInfo) error {
	for _, checkInfo := range checkInfos {
		if err := p.ProcessCheckInfo(checkInfo); err != nil {
			fmt.Printf("处理CheckInfo失败: %v\n", err)
			continue
		}
	}
	return nil
}

// ProcessCDNInfo 处理CDN相关信息
func (p *IPProcessor) ProcessCDNInfo(checkInfo *models.CheckInfo, sourceData *models.CDNData) error {
	cnameList := checkInfo.CNAME
	// 合并 A 和 AAAA 记录
	ipList := maputils.UniqueMergeSlices(checkInfo.A, checkInfo.AAAA)
	// 合并 asn记录列表
	asnList := maputils.UniqueMergeAnySlices(
		asninfo.GetUniqueOrgNumbers(checkInfo.Ipv4Asn),
		asninfo.GetUniqueOrgNumbers(checkInfo.Ipv6Asn),
	)
	// 合并 IP定位列表
	ipLocateList := maputils.UniqueMergeSlices(
		maputils.GetMapsUniqueValues(checkInfo.Ipv4Locate),
		maputils.GetMapsUniqueValues(checkInfo.Ipv6Locate),
	)

	// 判断IP解析结果数量是否在CDN内
	checkInfo.IpSizeIsCdn, checkInfo.IpSize = cdncheck.IpsSizeIsCdn(ipList, 3)

	// 检查结果中的 CDN | WAF | CLOUD 情况
	checkInfo.IsCdn, checkInfo.CdnCompany = cdncheck.CheckCategory(sourceData.CDN, ipList, asnList, cnameList, ipLocateList)
	checkInfo.IsWaf, checkInfo.WafCompany = cdncheck.CheckCategory(sourceData.WAF, ipList, asnList, cnameList, ipLocateList)
	checkInfo.IsCloud, checkInfo.CloudCompany = cdncheck.CheckCategory(sourceData.CLOUD, ipList, asnList, cnameList, ipLocateList)

	return nil
}

// ProcessCDNInfos 批量处理CDN信息
func (p *IPProcessor) ProcessCDNInfos(checkInfos []*models.CheckInfo, sourceData *models.CDNData) error {
	for _, checkInfo := range checkInfos {
		if err := p.ProcessCDNInfo(checkInfo, sourceData); err != nil {
			fmt.Printf("处理CDN信息失败: %v\n", err)
			continue
		}
	}
	return nil
}
