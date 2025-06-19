package ednsquery

import "fmt"

func MergeEDNSResultMap(locationResults map[string]EDNSResult) EDNSResult {
	mr := EDNSResult{
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
	mr.CNAMEChains = keys(cnamesChainSet)
	mr.A = keys(aSet)
	mr.AAAA = keys(aaaaSet)
	mr.CNAME = keys(cnameSet)
	mr.Errors = keys(errorSet)

	return mr
}

// MergeEDNSResultsMap 合并每个域名下的所有地理位置的 EDNS 查询结果
func MergeEDNSResultsMap(results map[string]map[string]EDNSResult) map[string]EDNSResult {
	mergedResults := make(map[string]EDNSResult)

	for domain, ednsMap := range results {
		mergedResults[domain] = MergeEDNSResultMap(ednsMap)
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

func printEDNSResultMap(mergedAll map[string]EDNSResult) {
	for domain, merged := range mergedAll {
		fmt.Printf("=== Merged Result for %s ===\n", domain)
		fmt.Printf("basic Domain: %s\n", merged.Domain)
		fmt.Printf("Final Domain: %s\n", merged.FinalDomain)
		fmt.Printf("Used Locations: %v\n", merged.Locations)
		fmt.Printf("A Records: %v\n", merged.A)
		fmt.Printf("AAAA Records: %v\n", merged.AAAA)
		fmt.Printf("CNAME Records: %v\n", merged.CNAME)
		fmt.Printf("Name Servers: %v\n", merged.NameServers)
		fmt.Printf("CNAME Chain: %v\n", merged.CNAMEChains)
		fmt.Printf("Errors: %v\n", merged.Errors)
	}
}
