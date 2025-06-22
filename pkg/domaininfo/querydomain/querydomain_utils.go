package querydomain

import (
	"cdnCheck/pkg/classify"
	dnsquery2 "cdnCheck/pkg/domaininfo/dnsquery"
	"cdnCheck/pkg/domaininfo/ednsquery"
	"cdnCheck/pkg/maputils"
	"cdnCheck/pkg/models"
	"fmt"
)

func MergeEDNSMapToDNSMap(dnsMap dnsquery2.DomainDNSResultMap, ednsMap ednsquery.DomainEDNSResultMap) dnsquery2.DomainDNSResultMap {
	for domain, ednsResult := range ednsMap {
		dnsResult, ok := dnsMap[domain]
		if !ok || ednsResult == nil {
			continue
		}
		// 合并 A 记录
		dnsResult.A = maputils.UniqueMergeSlices(dnsResult.A, ednsResult.A)
		// 合并 AAAA 记录
		dnsResult.AAAA = maputils.UniqueMergeSlices(dnsResult.AAAA, ednsResult.AAAA)
		// 合并 CNAME 记录
		dnsResult.CNAME = maputils.UniqueMergeSlices(dnsResult.CNAME, ednsResult.CNAME)
		// 合并 Errors
		if dnsResult.Error == nil {
			dnsResult.Error = make(map[string]string)
		}
		for i, err := range ednsResult.Errors {
			key := fmt.Sprintf("edns_error_%d", i)
			dnsResult.Error[key] = err
		}

		//优化整理
		dnsquery2.OptimizeDNSResult(dnsResult)
	}
	return dnsMap
}

// PopulateDNSResult 将 DNS 查询结果填充到 CheckInfo 中
func PopulateDNSResult(domainEntry classify.TargetEntry, query *dnsquery2.DNSResult) *models.CheckInfo {
	dnsResult := models.NewDomainCheckInfo(domainEntry.RAW, domainEntry.FMT, domainEntry.FromUrl)

	// 逐个复制 DNS 记录
	dnsResult.A = append(dnsResult.A, query.A...)
	dnsResult.AAAA = append(dnsResult.AAAA, query.AAAA...)
	dnsResult.CNAME = append(dnsResult.CNAME, query.CNAME...)
	dnsResult.NS = append(dnsResult.NS, query.NS...)
	dnsResult.MX = append(dnsResult.MX, query.MX...)
	dnsResult.TXT = append(dnsResult.TXT, query.TXT...)

	return dnsResult
}
