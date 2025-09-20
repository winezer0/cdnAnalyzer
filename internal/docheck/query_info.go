package docheck

import (
	"github.com/winezer0/cdnAnalyzer/internal/analyzer"
	"github.com/winezer0/cdnAnalyzer/pkg/classify"
	"github.com/winezer0/cdnAnalyzer/pkg/domaininfo/dnsquery"
	"github.com/winezer0/cdnAnalyzer/pkg/domaininfo/querydomain"
	"github.com/winezer0/cdnAnalyzer/pkg/logging"
	"github.com/winezer0/cdnAnalyzer/pkg/queryip"
)

// QueryDomainInfo 进行域名信息解析
func QueryDomainInfo(dnsConfig *querydomain.DNSQueryConfig, domainEntries []classify.TargetEntry) []*analyzer.CheckInfo {
	// 创建DNS处理器并执行查询
	dnsProcessor := querydomain.NewDNSProcessor(dnsConfig, &domainEntries)
	dnsResult := dnsProcessor.Process()

	//将dns查询结果合并到 CheckInfo 中去
	var checkInfos []*analyzer.CheckInfo
	for _, domainEntry := range domainEntries {
		var checkInfo *analyzer.CheckInfo
		//当存在dns查询结果时,补充
		if result, ok := (*dnsResult)[domainEntry.FMT]; ok && result != nil {
			checkInfo = PopulateDNSResult(domainEntry, result)
		} else {
			logging.Warnf("No DNS result for domain: %s", domainEntry.FMT)
			checkInfo = analyzer.NewDomainCheckInfo(domainEntry.RAW, domainEntry.FMT, domainEntry.FromUrl)
		}
		checkInfos = append(checkInfos, checkInfo)
	}
	return checkInfos
}

// QueryIPInfo 进行IP信息查询
func QueryIPInfo(ipDbConfig *queryip.IpDbConfig, checkInfos []*analyzer.CheckInfo) []*analyzer.CheckInfo {
	// 初始化IP数据库引擎
	ipEngines, err := queryip.InitDBEngines(ipDbConfig)
	if err != nil {
		logging.Fatalf("初始化数据库失败: %v", err)
	}
	defer ipEngines.Close()

	//对 checkInfos 中的A/AAAA记录进行IP信息查询，并赋值回去
	for _, checkInfo := range checkInfos {
		if len(checkInfo.A) > 0 || len(checkInfo.AAAA) > 0 {
			ipInfo, err := ipEngines.QueryIPInfo(checkInfo.A, checkInfo.AAAA)
			if err != nil {
				logging.Warnf("查询IP信息失败: %v", err)
			} else {
				checkInfo.Ipv4Locate = ipInfo.IPv4Locations
				checkInfo.Ipv4Asn = ipInfo.IPv4AsnInfos
				checkInfo.Ipv6Locate = ipInfo.IPv6Locations
				checkInfo.Ipv6Asn = ipInfo.IPv6AsnInfos
			}
		}
	}

	return checkInfos
}

// PopulateDNSResult 将 DNS 查询结果填充到 CheckInfo 中
func PopulateDNSResult(domainEntry classify.TargetEntry, query *dnsquery.DNSResult) *analyzer.CheckInfo {
	dnsResult := analyzer.NewDomainCheckInfo(domainEntry.RAW, domainEntry.FMT, domainEntry.FromUrl)

	// 逐个复制 DNS 记录
	dnsResult.A = append(dnsResult.A, query.A...)
	dnsResult.AAAA = append(dnsResult.AAAA, query.AAAA...)
	dnsResult.CNAME = append(dnsResult.CNAME, query.CNAME...)
	dnsResult.NS = append(dnsResult.NS, query.NS...)
	dnsResult.MX = append(dnsResult.MX, query.MX...)
	dnsResult.TXT = append(dnsResult.TXT, query.TXT...)

	return dnsResult
}
