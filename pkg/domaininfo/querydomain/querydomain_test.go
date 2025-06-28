package querydomain

import (
	"github.com/winezer0/cdnAnalyzer/pkg/classify"
	"testing"
	"time"
)

func TestDNSProcessor_Process(t *testing.T) {
	start := time.Now() // 记录开始时间

	// 构造测试域名
	targets := []string{
		"www.baidu.com",
		"www.taobao.com",
	}
	entries := classify.ClassifyTargets(targets).DomainEntries

	// 构造DNS和EDNS配置
	config := &DNSQueryConfig{
		Resolvers: []string{"8.8.8.8", "1.1.1.1"},
		CityMap: []map[string]string{
			{"City": "Beijing", "IP": "1.1.1.1"},
			{"City": "Shanghai", "IP": "8.8.8.8"},
		},
		Timeout:            5 * time.Second,
		MaxDNSConcurrency:  5,
		MaxEDNSConcurrency: 5,
		QueryEDNSCNAMES:    true,
		QueryEDNSUseSysNS:  false,
	}

	processor := NewDNSProcessor(config, &entries)
	resultMap := processor.Process()
	if resultMap == nil {
		t.Fatal("Process returned nil")
	}

	for domain, result := range *resultMap {
		t.Logf("Domain: %s", domain)
		t.Logf("A: %v", result.A)
		t.Logf("AAAA: %v", result.AAAA)
		t.Logf("CNAME: %v", result.CNAME)
		t.Logf("NS: %v", result.NS)
		t.Logf("MX: %v", result.MX)
		t.Logf("TXT: %v", result.TXT)
		t.Logf("Error: %v", result.Error)
	}

	elapsed := time.Since(start)
	t.Logf("Process耗时: %v", elapsed)
}

func TestDNSProcessor_FastProcess(t *testing.T) {
	start := time.Now() // 记录开始时间

	// 构造测试域名
	targets := []string{
		"www.baidu.com",
		"www.taobao.com",
	}
	entries := classify.ClassifyTargets(targets).DomainEntries

	// 构造DNS和EDNS配置
	config := &DNSQueryConfig{
		Resolvers: []string{"8.8.8.8", "1.1.1.1"},
		CityMap: []map[string]string{
			{"City": "Beijing", "IP": "1.1.1.1"},
			{"City": "Shanghai", "IP": "8.8.8.8"},
		},
		Timeout:            5 * time.Second,
		MaxDNSConcurrency:  5,
		MaxEDNSConcurrency: 5,
		QueryEDNSCNAMES:    true,
		QueryEDNSUseSysNS:  false,
	}

	processor := NewDNSProcessor(config, &entries)
	resultMap := processor.FastProcess()
	if resultMap == nil {
		t.Fatal("FastProcess returned nil")
	}

	for domain, result := range *resultMap {
		t.Logf("Domain: %s", domain)
		t.Logf("A: %v", result.A)
		t.Logf("AAAA: %v", result.AAAA)
		t.Logf("CNAME: %v", result.CNAME)
		t.Logf("NS: %v", result.NS)
		t.Logf("MX: %v", result.MX)
		t.Logf("TXT: %v", result.TXT)
		t.Logf("Error: %v", result.Error)
	}

	elapsed := time.Since(start)
	t.Logf("FastProcess耗时: %v", elapsed)
}
