package dnsquery

import (
	"cdnCheck/maputils"
	"net"
	"regexp"
	"strings"
)

// DnsServerAddPort 为dns服务器IP补充53端口
func DnsServerAddPort(s string) string {
	s = strings.TrimSuffix(strings.ToLower(s), ".")
	if strings.Contains(s, ":") {
		return s
	}
	return s + ":53"
}

// MergeDNSResults 合并去重多个 DNSResult
func MergeDNSResults(results []DNSResult) DNSResult {
	merged := DNSResult{}
	for _, r := range results {
		merged.A = maputils.UniqueMergeSlices(merged.A, r.A)
		merged.AAAA = maputils.UniqueMergeSlices(merged.AAAA, r.AAAA)
		merged.CNAME = maputils.UniqueMergeSlices(merged.CNAME, r.CNAME)
		merged.NS = maputils.UniqueMergeSlices(merged.NS, r.NS)
		merged.MX = maputils.UniqueMergeSlices(merged.MX, r.MX)
		merged.TXT = maputils.UniqueMergeSlices(merged.TXT, r.TXT)
	}
	return merged
}

// isDomain 判断给定字符串是否为域名
func isDomain(s string) bool {
	// 去除首尾空格
	s = strings.TrimSpace(s)

	// 如果是合法的 IP，则不是域名
	if net.ParseIP(s) != nil {
		return false
	}

	// 更宽松的域名匹配：
	// 允许字母、数字、点、连字符、下划线、星号（用于通配符）
	// 放宽对顶级域的要求，不要求一定是 [a-zA-Z]{2,}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\*\-_]+(\.[a-zA-Z0-9\*\-_]+)+\.?$`, s)
	return matched
}

// OptimizeDNSResult 将 A 和 AAAA 记录中的域名移动到 CNAME 中，并确保 CNAME 去重
func OptimizeDNSResult(dnsResult *DNSResult) {
	// 使用 map 实现 CNAME 去重
	cnameSet := make(map[string]struct{})
	for _, cname := range dnsResult.CNAME {
		cnameSet[cname] = struct{}{}
	}

	var newA []string
	for _, record := range dnsResult.A {
		if isDomain(record) {
			if _, exists := cnameSet[record]; !exists {
				dnsResult.CNAME = append(dnsResult.CNAME, record)
				cnameSet[record] = struct{}{}
			}
		} else {
			newA = append(newA, record)
		}
	}
	dnsResult.A = newA

	var newAAAA []string
	for _, record := range dnsResult.AAAA {
		if isDomain(record) {
			if _, exists := cnameSet[record]; !exists {
				dnsResult.CNAME = append(dnsResult.CNAME, record)
				cnameSet[record] = struct{}{}
			}
		} else {
			newAAAA = append(newAAAA, record)
		}
	}
	dnsResult.AAAA = newAAAA
}
