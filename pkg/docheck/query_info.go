package docheck

import (
	"cdnAnalyzer/pkg/analyzer"
	"cdnAnalyzer/pkg/classify"
	"cdnAnalyzer/pkg/domaininfo/querydomain"
	"cdnAnalyzer/pkg/ipinfo/queryip"
	"fmt"
	"os"
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
			checkInfo = querydomain.PopulateDNSResult(domainEntry, result)
		} else {
			fmt.Printf("No DNS result for domain: %s\n", domainEntry.FMT)
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
		fmt.Printf("初始化数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer ipEngines.Close()

	//对 checkInfos 中的A/AAAA记录进行IP信息查询，并赋值回去
	for _, checkInfo := range checkInfos {
		if len(checkInfo.A) > 0 || len(checkInfo.AAAA) > 0 {
			ipInfo, err := ipEngines.QueryIPInfo(checkInfo.A, checkInfo.AAAA)
			if err != nil {
				fmt.Printf("查询IP信息失败: %v\n", err)
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
