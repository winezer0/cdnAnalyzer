package ednsquery

import (
	"fmt"
	"testing"
	"time"
)

func TestEDNSQuery(t *testing.T) {
	domain := "www.baidu.com" // 替换为你要测试的域名
	eDNSQueryResults := ResolveEDNS(domain, "175.1.238.1", "8.8.8.8:53", 5*time.Second)

	t.Logf("Query results for %s:", domain)
	t.Logf("  A:    %v", eDNSQueryResults.A)
	t.Logf("  CNAME: %v", eDNSQueryResults.CNAME)
	t.Logf("  Errors: %v", eDNSQueryResults.Errors)
}

func TestResolveEDNSDomainsWithCities(t *testing.T) {
	// 测试域名列表
	domains := []string{
		"www.baidu.com",
		"www.taobao.com",
	}

	// 模拟几个城市位置（用于 EDNS_SUBNET 的 IP）
	cities := []map[string]string{
		{"City": "Beijing", "IP": "1.1.1.1"},
		{"City": "Shanghai", "IP": "8.8.8.8"},
		{"City": "Guangzhou", "IP": "114.114.114.114"},
	}

	// 设置超时时间和最大并发数
	timeout := 3 * time.Second
	maxConcurrency := 10

	// 执行批量查询
	results := ResolveEDNSDomainsWithCities(domains, cities, timeout, maxConcurrency, true)

	//// 打印结果
	//for domain, ednsMap := range results {
	//	fmt.Printf("=== Results for Domain: %s ===\n", domain)
	//	for location, res := range ednsMap {
	//		fmt.Printf("Location: %s\n", location)
	//		fmt.Printf("  Original Domain: %s\n", res.Domain)
	//		fmt.Printf("  Final Domain: %s\n", res.FinalDomain)
	//		fmt.Printf("  NameServers: %v\n", res.NameServers)
	//		fmt.Printf("  CNAMEChains: %v\n", res.CNAMEChains)
	//		fmt.Printf("  A Records: %v\n", res.A)
	//		fmt.Printf("  AAAA Records: %v\n", res.AAAA)
	//		fmt.Printf("  CNAME Records: %v\n", res.CNAME)
	//		if len(res.Errors) > 0 {
	//			fmt.Printf("  Errors: %v\n", res.Errors)
	//		}
	//		fmt.Println()
	//	}
	//	fmt.Println("==================================")
	//}

	mergedAll := MergeEDNSResultsMap(results)
	for domain, merged := range mergedAll {
		fmt.Printf("=== Merged Result for %s ===\n", domain)
		fmt.Printf("basic Domain: %s\n", merged.Domain)
		fmt.Printf("Final Domain: %s\n", merged.FinalDomain)
		fmt.Printf("Used Locations: %v\n", merged.Locations)
		fmt.Printf("A Records: %v\n", merged.A)
		fmt.Printf("AAAA Records: %v\n", merged.AAAA)
		fmt.Printf("CNAME Records: %v\n", merged.CNAME)
		fmt.Printf("Name Servers: %v\n", merged.NameServers)
		fmt.Printf("CNAME Chain: %v\n", merged.CNAMEs)
		if len(merged.Errors) > 0 {
			fmt.Printf("Errors: %v\n", merged.Errors)
		}
		fmt.Println()
	}
}
