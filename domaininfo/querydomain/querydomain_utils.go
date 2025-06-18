package querydomain

import (
	"cdnCheck/domaininfo/dnsquery"
	"cdnCheck/domaininfo/ednsquery"
	"cdnCheck/maputils"
	"fmt"
)

func MergeEDNSToDNS(ednsResult ednsquery.EDNSResult, dnsResult dnsquery.DNSResult) dnsquery.DNSResult {
	// 合并 CNAME
	dnsResult.A = maputils.UniqueMergeSlices(dnsResult.A, ednsResult.A)
	dnsResult.AAAA = maputils.UniqueMergeSlices(dnsResult.AAAA, ednsResult.AAAA)
	dnsResult.CNAME = maputils.UniqueMergeSlices(dnsResult.CNAME, ednsResult.CNAME)

	// 如果 Error 为 nil，先初始化为空 map
	if dnsResult.Error == nil {
		dnsResult.Error = make(map[string]string)
	}

	// 合并 Errors 到 Error map
	for i, err := range ednsResult.Errors {
		key := fmt.Sprintf("edns_error_%d", i)
		dnsResult.Error[key] = err
	}

	return dnsResult
}

// MergeEDNSResults 合并去重多个 ECS结果
func MergeEDNSResults(ednsResults map[string]ednsquery.EDNSResult) ednsquery.EDNSResult {
	merged := ednsquery.EDNSResult{}

	for _, res := range ednsResults {
		merged.A = maputils.UniqueMergeSlices(merged.A, res.A)
		merged.AAAA = maputils.UniqueMergeSlices(merged.AAAA, res.AAAA)
		merged.CNAME = maputils.UniqueMergeSlices(merged.CNAME, res.CNAME)
		merged.Errors = maputils.UniqueMergeSlices(merged.Errors, res.Errors)
	}

	return merged
}
