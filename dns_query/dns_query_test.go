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

func TestSyncQueryAllDNS(t *testing.T) {
	domain := "www.baidu.com"
	dnsServer := "8.8.8.8:53"
	timeout := 3 * time.Second

	result := SyncQueryAllDNS(domain, dnsServer, timeout)
	b, _ := json.MarshalIndent(result, "", "  ")
	t.Logf("QueryAllDNS result for %s @ %s:\n%s", domain, dnsServer, string(b))

	if len(result.A) == 0 && len(result.CNAME) == 0 && len(result.NS) == 0 {
		t.Errorf("QueryAllDNS failed, got empty result: %+v", result)
	}
}

func TestSyncPoolQueryAllDNS(t *testing.T) {
	domain := "www.baidu.com"
	dnsServer := "8.8.8.8:53"
	timeout := 3 * time.Second

	result := SyncPoolQueryAllDNS(domain, dnsServer, timeout, 3)
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

	resolvers, err := LoadFilesToList(tmpFile)
	if err != nil {
		t.Fatalf("LoadFilesToList error: %v", err)
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

	result := QueryAllDNSWithMultiResolvers(domain, resolvers, timeout, 5)
	b, _ := json.MarshalIndent(result, "", "  ")
	t.Logf("QueryAllDNSWithMultiResolvers result for %s:\n%s", domain, string(b))

	if len(result.A) == 0 && len(result.CNAME) == 0 && len(result.NS) == 0 {
		t.Errorf("QueryAllDNSWithMultiResolvers failed, got empty result: %+v", result)
	}
}

func TestLookupNSServer(t *testing.T) {
	type args struct {
		domain    string
		dnsServer string
		timeout   time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid subdomain",
			args: args{
				domain:    "www.baidu.com",
				dnsServer: "8.8.8.8:53",
				timeout:   5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "Valid TLD domain",
			args: args{
				domain:    "example.com",
				dnsServer: "1.1.1.1:53",
				timeout:   5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "Invalid domain",
			args: args{
				domain:    "invalid.nonexistent.tld",
				dnsServer: "8.8.8.8:53",
				timeout:   3 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LookupNSServers(tt.args.domain, tt.args.dnsServer, tt.args.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookupNSServers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) == 0 {
				t.Errorf("LookupNSServers() returned empty NS list but no error")
			}
			t.Logf("NS Records for %s: %+v", tt.args.domain, got)
		})
	}
}

func TestLookupCNAMEChain(t *testing.T) {
	tests := []struct {
		name      string
		domain    string
		server    string
		timeout   time.Duration
		wantChain []string
	}{
		{
			name:      "No CNAME",
			domain:    "example.com",
			server:    "8.8.8.8:53",
			timeout:   3 * time.Second,
			wantChain: []string{"example.com"},
		},
		{
			name:      "Simple CNAME Chain",
			domain:    "www.baidu.com",
			server:    "8.8.8.8:53",
			timeout:   3 * time.Second,
			wantChain: []string{"www.baidu.com", "www.a.shifen.com."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain, _ := LookupCNAMEChain(tt.domain, tt.server, tt.timeout)
			t.Logf("CNAME chain for %s: %v", tt.domain, chain)

			if len(chain) != len(tt.wantChain) {
				t.Errorf("Expected length %d, got %d", len(tt.wantChain), len(chain))
				return
			}

			for i := range chain {
				if chain[i] != tt.wantChain[i] {
					t.Errorf("chain[%d] = %q, want %q", i, chain[i], tt.wantChain[i])
				}
			}
		})
	}
}
