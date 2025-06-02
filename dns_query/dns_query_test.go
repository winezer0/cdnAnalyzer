package dns_query

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestQueryAllDNS(t *testing.T) {
	domain := "www.baidu.com"
	dnsServer := "8.8.8.8:53"
	timeout := 3 * time.Second

	result := QueryAllDNS(domain, dnsServer, timeout)
	b, _ := json.MarshalIndent(result, "", "  ")
	t.Logf("QueryAllDNS result for %s @ %s:\n%s", domain, dnsServer, string(b))

	if len(result.A) == 0 && len(result.CNAME) == 0 && len(result.NS) == 0 {
		t.Errorf("QueryAllDNS failed, got empty result: %+v", result)
	}
}

func TestLoadResolvers(t *testing.T) {
	// 创建一个临时的 resolvers.txt
	tmpFile := "test_resolvers.txt"
	content := "8.8.8.8\n1.1.1.1:53\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test resolvers file: %v", err)
	}
	defer os.Remove(tmpFile)

	resolvers, err := LoadResolvers(tmpFile)
	if err != nil {
		t.Fatalf("LoadResolvers error: %v", err)
	}
	t.Logf("Loaded resolvers: %v", resolvers)
	if len(resolvers) != 2 {
		t.Errorf("Expected 2 resolvers, got %d", len(resolvers))
	}
	if resolvers[0] != "8.8.8.8:53" || resolvers[1] != "1.1.1.1:53" {
		t.Errorf("Unexpected resolvers: %v", resolvers)
	}

	t.Logf("resolvers: %s", resolvers)
}

func TestQueryAllDNSWithMultiResolvers(t *testing.T) {
	domain := "www.baidu.com"
	timeout := 3 * time.Second
	resolvers := []string{"8.8.8.8:53", "1.1.1.1:53", "8.8.4.4:53", "9.9.9.9:53", "114.114.114.114:53", "223.5.5.5:53"}

	result := QueryAllDNSWithMultiResolvers(domain, resolvers, timeout)
	b, _ := json.MarshalIndent(result, "", "  ")
	t.Logf("QueryAllDNSWithMultiResolvers result for %s:\n%s", domain, string(b))

	if len(result.A) == 0 && len(result.CNAME) == 0 && len(result.NS) == 0 {
		t.Errorf("QueryAllDNSWithMultiResolvers failed, got empty result: %+v", result)
	}
}
