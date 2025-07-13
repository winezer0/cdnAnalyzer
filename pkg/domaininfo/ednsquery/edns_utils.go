package ednsquery

import "github.com/winezer0/cdnAnalyzer/pkg/logging"

func mergeCityEDNSResultMap(cityEDNSResultMap CityEDNSResultMap) *EDNSResult {
	if len(cityEDNSResultMap) == 0 {
		return &EDNSResult{}
	}

	mr := EDNSResult{
		Locations:   make([]string, 0, len(cityEDNSResultMap)),
		A:           nil,
		AAAA:        nil,
		CNAME:       nil,
		NameServers: nil,
		CNAMEChains: nil,
		Errors:      nil,
	}

	aSet := make(map[string]struct{})
	aaaaSet := make(map[string]struct{})
	cnameSet := make(map[string]struct{})
	nsSet := make(map[string]struct{})
	cnamesChainSet := make(map[string]struct{})
	errorSet := make(map[string]struct{})
	locationSet := make(map[string]struct{})

	var first = true
	for location, res := range cityEDNSResultMap {
		// 记录 location
		locationSet[location] = struct{}{}

		// 设置 Domain 和 FinalDomain（只取第一个）
		if first {
			mr.Domain = res.Domain
			mr.FinalDomain = res.FinalDomain
			first = false
		}

		// 合并 NameServers
		addStringsToSet(res.NameServers, nsSet)

		// 合并 CNAME 链条
		addStringsToSet(res.CNAMEChains, cnamesChainSet)

		// 合并 A/AAAA/CNAME/错误
		addStringsToSet(res.A, aSet)
		addStringsToSet(res.AAAA, aaaaSet)
		addStringsToSet(res.CNAME, cnameSet)
		addStringsToSet(res.Errors, errorSet)
	}

	// 转换 set 到 slice
	mr.Locations = keys(locationSet)
	mr.NameServers = keys(nsSet)
	mr.CNAMEChains = keys(cnamesChainSet)
	mr.A = keys(aSet)
	mr.AAAA = keys(aaaaSet)
	mr.CNAME = keys(cnameSet)
	mr.Errors = keys(errorSet)

	return &mr
}

// MergeDomainCityEDNSResultMap 合并每个域名下的所有地理位置的 EDNS 查询结果
func MergeDomainCityEDNSResultMap(results DomainCityEDNSResultMap) DomainEDNSResultMap {
	mergedResults := make(DomainEDNSResultMap)
	for domain, ednsMap := range results {
		mergedResults[domain] = mergeCityEDNSResultMap(ednsMap)
	}
	return mergedResults
}

// keys 将 map 的 key 转换为 slice（用于去重）
func keys(m map[string]struct{}) []string {
	list := make([]string, 0, len(m))
	for k := range m {
		list = append(list, k)
	}
	return list
}

// addStringsToSet 将字符串切片加入 map
func addStringsToSet(items []string, set map[string]struct{}) {
	for _, item := range items {
		set[item] = struct{}{}
	}
}

func printEDNSResultMap(mergedAll map[string]*EDNSResult) {
	for domain, merged := range mergedAll {
		logging.Debugf("=== Merged DNSResult for %s ===", domain)
		logging.Debugf("basic Domain: %s", merged.Domain)
		logging.Debugf("Final Domain: %s", merged.FinalDomain)
		logging.Debugf("Used Locations: %v", merged.Locations)
		logging.Debugf("A Records: %v", merged.A)
		logging.Debugf("AAAA Records: %v", merged.AAAA)
		logging.Debugf("CNAME Records: %v", merged.CNAME)
		logging.Debugf("Name Servers: %v", merged.NameServers)
		logging.Debugf("CNAME Chain: %v", merged.CNAMEChains)
		logging.Debugf("Errors: %v", merged.Errors)
	}
}
