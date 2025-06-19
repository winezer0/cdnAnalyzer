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
	maxConcurrency := 10

	// === 第一次调用：启用 EDNS ===
	t.Log("Running with EDNS enabled...")
	start := time.Now()
	resultsWithEDNS := ResolveEDNSDomainsWithCities(domains, cities, 5*time.Second, maxConcurrency, true)
	durationWithEDNS := time.Since(start)
	fmt.Printf("✅ Time taken EDNS with cnames: %v\n\n", durationWithEDNS)
	printEDNSResultMap(MergeEDNSResultsMap(resultsWithEDNS))

	// 执行批量查询
	// === 第二次调用：禁用 EDNS ===
	t.Log("Running with EDNS disabled...")
	start = time.Now()
	ednsEesultsNoCNMAES := ResolveEDNSDomainsWithCities(domains, cities, 3*time.Second, maxConcurrency, false)
	ednsDurationNoCNMAES := time.Since(start)
	fmt.Printf("✅ Time taken EDNS without cnames:  %v\n\n", ednsDurationNoCNMAES)

	printEDNSResultMap(MergeEDNSResultsMap(ednsEesultsNoCNMAES))
}
