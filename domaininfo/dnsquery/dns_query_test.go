package dnsquery

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestQueryDNS 测试单个 DNS 记录查询功能
func TestQueryDNS(t *testing.T) {
	domain := "example.com"
	resolver := "8.8.8.8"

	// 测试 A 记录
	records, err := ResolveDNS(domain, resolver, "A", 5*time.Second)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)

	// 测试 AAAA 记录
	records, err = ResolveDNS(domain, resolver, "AAAA", 5*time.Second)
	assert.NoError(t, err)
	t.Logf("AAAA Records: %v", records)

	// 测试 CNAME 记录
	records, err = ResolveDNS("www.cloudflare.com", resolver, "CNAME", 5*time.Second)
	assert.NoError(t, err)
	t.Logf("CNAME Records: %v", records)

	// 测试 TXT 记录
	records, err = ResolveDNS("example.com", resolver, "TXT", 5*time.Second)
	assert.NoError(t, err)
	t.Logf("TXT Records: %v", records)

	// 测试错误情况（非法 DNS 地址）
	_, err = ResolveDNS(domain, "invalid.dns.server", "A", 1*time.Second)
	assert.Error(t, err)
}

func TestResolveAllDNSWithResolvers(t *testing.T) {
	domains := []string{"example.com", "www.baidu.com"}
	resolvers := []string{
		"8.8.8.8",
		"1.1.1.1",
		"9.9.9.9",
	}

	start := time.Now()
	domainResolverDNSResultMap := ResolveDNSWithResolversMulti(domains, nil, resolvers, 5*time.Second, 15)
	elapsed := time.Since(start)
	t.Logf("✅ ResolveDNSWithResolversAtom 总共查询 %d 个解析器，耗时: %v", len(resolvers), elapsed)

	domainDNSResultMap := MergeDomainResolverResultMap(domainResolverDNSResultMap)
	// 现在 finalResults 是 map[string]*DNSResult 类型
	for domain, result := range domainDNSResultMap {
		OptimizeDNSResult(result)
		fmt.Printf("Domain: %s\n", domain)
		b, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(b))
	}
}
