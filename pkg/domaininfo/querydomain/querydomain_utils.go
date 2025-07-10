package querydomain

import (
	"fmt"
	"github.com/winezer0/cdnAnalyzer/pkg/domaininfo/dnsquery"
	"github.com/winezer0/cdnAnalyzer/pkg/domaininfo/ednsquery"
	"github.com/winezer0/cdnAnalyzer/pkg/maputils"
)

func MergeEDNSMapToDNSMap(dnsMap dnsquery.DomainDNSResultMap, ednsMap ednsquery.DomainEDNSResultMap) dnsquery.DomainDNSResultMap {
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
		dnsquery.OptimizeDNSResult(dnsResult)
	}
	return dnsMap
}
