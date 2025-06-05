package dnsquery

import (
	"cdnCheck/structtools"
	"context"
	"fmt"
	"net"
	"strings"
)

// GetSystemDefaultAddress 获取系统默认的本地dns服务器
func GetSystemDefaultAddress() (addr string) {
	resolver := net.Resolver{}
	resolver.PreferGo = true
	resolver.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		addr = address
		return nil, fmt.Errorf(address)
	}
	_, _ = resolver.LookupHost(context.Background(), "Addr")
	return addr
}

// DnsServerAddPort 为dns服务器IP补充53端口
func DnsServerAddPort(s string) string {
	s = strings.TrimSuffix(strings.ToLower(s), ".")
	if strings.Contains(s, ":") {
		return s
	}
	return s + ":53"
}

func HasPort(addr string) bool {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return true
		}
		if addr[i] < '0' || addr[i] > '9' {
			break
		}
	}
	return false
}

// MergeDNSResults 合并去重多个 DNSResult
func MergeDNSResults(results []DNSResult) DNSResult {
	merged := DNSResult{}
	for _, r := range results {
		merged.A = structutils.UniqueMergeSlices(merged.A, r.A)
		merged.AAAA = structutils.UniqueMergeSlices(merged.AAAA, r.AAAA)
		merged.CNAME = structutils.UniqueMergeSlices(merged.CNAME, r.CNAME)
		merged.NS = structutils.UniqueMergeSlices(merged.NS, r.NS)
		merged.MX = structutils.UniqueMergeSlices(merged.MX, r.MX)
		merged.TXT = structutils.UniqueMergeSlices(merged.TXT, r.TXT)
	}
	return merged
}

// MergeEDNSResults 合并去重多个 ECS结果
func MergeEDNSResults(results map[string]EDNSResult) EDNSResult {
	merged := EDNSResult{}

	for _, res := range results {
		merged.A = structutils.UniqueMergeSlices(merged.A, res.A)
		merged.CNAME = structutils.UniqueMergeSlices(merged.CNAME, res.CNAME)
		merged.Errors = structutils.UniqueMergeSlices(merged.Errors, res.Errors)
	}

	return merged
}

func MergeEDNSToDNS(edns EDNSResult, dns DNSResult) DNSResult {
	// 将 A 分类合并到 A 或 AAAA
	for _, ip := range edns.A {
		if isIPv4(ip) {
			dns.A = structutils.UniqueMergeSlices(dns.A, []string{ip})
		} else {
			dns.AAAA = structutils.UniqueMergeSlices(dns.AAAA, []string{ip})
		}
	}

	// 合并 CNAME
	dns.CNAME = structutils.UniqueMergeSlices(dns.CNAME, edns.CNAME)

	// 如果 Error 为 nil，先初始化为空 map
	if dns.Error == nil {
		dns.Error = make(map[string]string)
	}

	// 合并 Errors 到 Error map
	for i, err := range edns.Errors {
		key := fmt.Sprintf("edns_error_%d", i)
		dns.Error[key] = err
	}

	return dns
}

// 辅助函数：判断是否是 IPv4
func isIPv4(ip string) bool {
	parts := strings.Split(ip, ".")
	return len(parts) == 4
}
