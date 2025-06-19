package ednsquery

// MergedResult 存放最后格式化的结果
// MergedResult 存放最后格式化的结果
type MergedResult struct {
	Domain      string   // 原始域名
	FinalDomain string   // 最终解析出的域名（CNAME 链尾部）
	NameServers []string // 权威 DNS 列表
	CNAMEs      []string // CNAME 链条
	A           []string // A 记录
	AAAA        []string // AAAA 记录
	CNAME       []string // CNAME 记录
	Errors      []string // 错误信息
	Locations   []string // 所有参与查询的 location（如 Beijing@8.8.8.8）
}

func MergeToMergedResult(locationResults map[string]EDNSResult) MergedResult {
	mr := MergedResult{
		Locations: make([]string, 0, len(locationResults)),
	}

	aSet := make(map[string]struct{})
	aaaaSet := make(map[string]struct{})
	cnameSet := make(map[string]struct{})
	nsSet := make(map[string]struct{})
	cnamesChainSet := make(map[string]struct{})
	errorSet := make(map[string]struct{})

	first := true

	for location, res := range locationResults {
		// 收集 location
		mr.Locations = append(mr.Locations, location)

		if first {
			mr.Domain = res.Domain
			mr.FinalDomain = res.FinalDomain
			first = false
		}

		// 合并 NameServers
		for _, ns := range res.NameServers {
			nsSet[ns] = struct{}{}
		}

		// 合并 CNAMEChains 链
		for _, cname := range res.CNAMEChains {
			cnamesChainSet[cname] = struct{}{}
		}

		// 合并 A 记录
		for _, a := range res.A {
			aSet[a] = struct{}{}
		}

		// 合并 AAAA 记录
		for _, aaaa := range res.AAAA {
			aaaaSet[aaaa] = struct{}{}
		}

		// 合并 CNAME 记录
		for _, cname := range res.CNAME {
			cnameSet[cname] = struct{}{}
		}

		// 合并错误信息
		for _, err := range res.Errors {
			errorSet[err] = struct{}{}
		}
	}

	// 对 Locations 去重（可选）
	locationMap := make(map[string]struct{})
	for _, loc := range mr.Locations {
		locationMap[loc] = struct{}{}
	}
	mr.Locations = keys(locationMap)

	// 合并其他字段
	mr.NameServers = keys(nsSet)
	mr.CNAMEs = keys(cnamesChainSet)
	mr.A = keys(aSet)
	mr.AAAA = keys(aaaaSet)
	mr.CNAME = keys(cnameSet)
	mr.Errors = keys(errorSet)

	return mr
}

// MergeEDNSDomainsResults 合并每个域名下的所有地理位置的 EDNS 查询结果
func MergeEDNSDomainsResults(results map[string]map[string]EDNSResult) map[string]MergedResult {
	mergedResults := make(map[string]MergedResult)

	for domain, ednsMap := range results {
		mergedResults[domain] = MergeToMergedResult(ednsMap)
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
