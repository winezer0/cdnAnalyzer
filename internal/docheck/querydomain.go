package docheck

import (
	"github.com/winezer0/cdninfo/internal/analyzer"
	"github.com/winezer0/cdninfo/pkg/classify"
	"github.com/winezer0/cdninfo/pkg/domaininfo/dnsquery"
	"github.com/winezer0/cdninfo/pkg/domaininfo/querydomain"
	"github.com/winezer0/xutils/logging"
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
