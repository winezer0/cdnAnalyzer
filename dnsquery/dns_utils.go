package dnsquery

import (
	"cdnCheck/filetools"
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
		merged.A = filetools.UniqueMergeSlices(merged.A, r.A)
		merged.AAAA = filetools.UniqueMergeSlices(merged.AAAA, r.AAAA)
		merged.CNAME = filetools.UniqueMergeSlices(merged.CNAME, r.CNAME)
		merged.NS = filetools.UniqueMergeSlices(merged.NS, r.NS)
		merged.MX = filetools.UniqueMergeSlices(merged.MX, r.MX)
		merged.TXT = filetools.UniqueMergeSlices(merged.TXT, r.TXT)
	}
	return merged
}
