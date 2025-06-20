package dnsquery

import (
	"cdnCheck/maputils"
	"context"
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"net"
	"regexp"
	"strings"
	"time"
)

// nsServerAddPort 为dns服务器IP补充53端口
func nsServerAddPort(s string) string {
	s = strings.TrimSuffix(strings.ToLower(s), ".")
	if strings.Contains(s, ":") {
		return s
	}
	return s + ":53"
}

func NSServersAddPort(nsServers []string) []string {
	nsList := make([]string, 0, len(nsServers))
	for _, nsServer := range nsServers {
		nsList = append(nsList, nsServerAddPort(nsServer))
	}
	return nsList
}

// mergeResolverResultMap  合并去重多个 ResolverResult
func mergeResolverResultMap(resultMap ResolverDNSResultMap) DNSResult {
	merged := DNSResult{}
	for _, dnsResult := range resultMap {
		merged.A = maputils.UniqueMergeSlices(merged.A, dnsResult.A)
		merged.AAAA = maputils.UniqueMergeSlices(merged.AAAA, dnsResult.AAAA)
		merged.CNAME = maputils.UniqueMergeSlices(merged.CNAME, dnsResult.CNAME)
		merged.NS = maputils.UniqueMergeSlices(merged.NS, dnsResult.NS)
		merged.MX = maputils.UniqueMergeSlices(merged.MX, dnsResult.MX)
		merged.TXT = maputils.UniqueMergeSlices(merged.TXT, dnsResult.TXT)
	}
	return merged
}

// MergeDomainResolverResultMap 将 DomainResolverDNSResultMap 合并为 DomainDNSResultMap（去重所有 resolver 的结果）
func MergeDomainResolverResultMap(resultMap DomainResolverDNSResultMap) DomainDNSResultMap {
	merged := make(DomainDNSResultMap)
	for domain, resolverMap := range resultMap {
		mergedResult := mergeResolverResultMap(resolverMap)
		merged[domain] = &mergedResult
	}
	return merged
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

// LookupNSServers 递归查找最上级可用NS服务器
func LookupNSServers(domain, dnsServer string, timeout time.Duration) ([]string, error) {
	// 规范化域名，去除前后点
	domain = strings.Trim(domain, ".")
	labels := strings.Split(domain, ".")
	for i := 0; i < len(labels); i++ {
		parent := strings.Join(labels[i:], ".")
		nameServers, err := lookupNSServer(parent, dnsServer, timeout)
		if err != nil || len(nameServers) == 0 {
			continue // 没查到就往上一级查
		}
		return nameServers, nil
	}
	return nil, errors.New("no ns record found for any parent domain")
}

// lookupNSServer 查询该域名的权威名称服务器（NS 记录）和 SOA 中的 NS
func lookupNSServer(domain string, dnsServer string, timeout time.Duration) ([]string, error) {
	client := new(dns.Client)
	client.Timeout = timeout

	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeNS)

	resp, _, err := client.Exchange(msg, dnsServer)
	if err != nil {
		return nil, fmt.Errorf("failed to query NS records: %w", err)
	}

	var nameServers []string

	// 检查 Authority Section 是否有 SOA，并提取其中的 NS
	for _, record := range resp.Ns {
		if soa, ok := record.(*dns.SOA); ok {
			nameServers = append(nameServers, soa.Ns)
		}
	}

	// 提取 Answer Section 中的 NS 记录
	for _, answerRecord := range resp.Answer {
		if ns, ok := answerRecord.(*dns.NS); ok {
			nameServers = append(nameServers, ns.Ns)
		}
	}

	if len(nameServers) == 0 {
		return nil, fmt.Errorf("no NS records found for domain: %s", domain)
	}

	return nameServers, nil
}

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

// LookupCNAMEChains 获取域名的cnames链条信息 不包含原域名
func LookupCNAMEChains(domain, dnsServer string, timeout time.Duration) ([]string, string, error) {
	var cnameChains []string
	visited := make(map[string]struct{})
	current := domain

	for {
		if _, ok := visited[current]; ok {
			break
		}
		visited[current] = struct{}{}

		cnames, err := ResolveDNS(current, dnsServer, "CNAME", timeout)
		if err != nil || len(cnames) == 0 {
			break
		}
		//cnameChains = append(cnameChains, cnames...)
		cnameChains = maputils.UniqueMergeSlicesSorted(cnameChains, cnames)
		current = cnames[0]
	}

	return cnameChains, current, nil
}
