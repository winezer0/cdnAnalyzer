package dnsquery

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestQueryDNS 测试单个 DNS 记录查询功能
func TestQueryDNS(t *testing.T) {
	domain := "example.com"
	resolver := "8.8.8.8"

	// 测试 A 记录
	records, err := QueryDNS(domain, resolver, "A", 5*time.Second)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)

	// 测试 AAAA 记录
	records, err = QueryDNS(domain, resolver, "AAAA", 5*time.Second)
	assert.NoError(t, err)
	t.Logf("AAAA Records: %v", records)

	// 测试 CNAME 记录
	records, err = QueryDNS("www.cloudflare.com", resolver, "CNAME", 5*time.Second)
	assert.NoError(t, err)
	t.Logf("CNAME Records: %v", records)

	// 测试 TXT 记录
	records, err = QueryDNS("example.com", resolver, "TXT", 5*time.Second)
	assert.NoError(t, err)
	t.Logf("TXT Records: %v", records)

	// 测试错误情况（非法 DNS 地址）
	_, err = QueryDNS(domain, "invalid.dns.server", "A", 1*time.Second)
	assert.Error(t, err)
}

func TestResolveAllDNSWithResolvers(t *testing.T) {
	domain := "example.com"
	resolvers := []string{
		"8.8.8.8",
		"1.1.1.1",
		"9.9.9.9",
	}

	start1 := time.Now()
	ResolveDNSWithResolversAtom(domain, nil, resolvers, 5*time.Second, 15)
	elapsed1 := time.Since(start1)
	t.Logf("✅ ResolveDNSWithResolversAtom 总共查询 %d 个解析器，耗时: %v", len(resolvers), elapsed1)

	start2 := time.Now()
	ctx := context.Background()
	results := ResolveDNSWithResolvers(ctx, domain, nil, resolvers, 5*time.Second, 5)
	elapsed := time.Since(start2)

	t.Logf("✅ ResolveDNSWithResolvers 总共查询 %d 个解析器，耗时: %v", len(resolvers), elapsed)
	t.Log("----------------------------")

	foundAnyRecord := false
	successCount := 0

	for _, res := range results {
		t.Logf("📡 Resolver: %s", res.Resolver)

		if len(res.Result.Error) > 0 {
			t.Log("  ❌ Errors:")
			for typ, err := range res.Result.Error {
				t.Logf("    - [%s]: %s", typ, err)
			}
		}

		hasRecords := false

		if len(res.Result.A) > 0 {
			t.Log("  📡 A Records:")
			for _, ip := range res.Result.A {
				t.Logf("    - %s", ip)
			}
			hasRecords = true
		}

		if len(res.Result.AAAA) > 0 {
			t.Log("  🌐 AAAA Records:")
			for _, ip := range res.Result.AAAA {
				t.Logf("    - %s", ip)
			}
			hasRecords = true
		}

		if len(res.Result.CNAME) > 0 {
			t.Log("  🔁 CNAME Records:")
			for _, cname := range res.Result.CNAME {
				t.Logf("    - %s", cname)
			}
			hasRecords = true
		}

		if len(res.Result.NS) > 0 {
			t.Log("  📦 NS Records:")
			for _, ns := range res.Result.NS {
				t.Logf("    - %s", ns)
			}
			hasRecords = true
		}

		if len(res.Result.MX) > 0 {
			t.Log("  📨 MX Records:")
			for _, mx := range res.Result.MX {
				t.Logf("    - %s", mx)
			}
			hasRecords = true
		}

		if len(res.Result.TXT) > 0 {
			t.Log("  📄 TXT Records:")
			for _, txt := range res.Result.TXT {
				t.Logf("    - %q", txt)
			}
			hasRecords = true
		}

		if hasRecords {
			successCount++
			foundAnyRecord = true
		} else {
			t.Log("No useful records returned")
		}

		t.Log("----------------------------")
	}

	assert.True(t, foundAnyRecord, "至少应有一个 resolver 返回了有效记录")
	assert.Equal(t, len(resolvers), len(results))
	assert.GreaterOrEqual(t, successCount, 1, "至少一个 resolver 成功返回记录")
}
