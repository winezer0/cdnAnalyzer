package dns_query

import (
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"strings"
	"sync"
	"time"
)

// DNSResult 统一返回结构
type DNSResult struct {
	A     []string          `json:"A"`
	AAAA  []string          `json:"AAAA"`
	CNAME []string          `json:"CNAME"`
	NS    []string          `json:"NS"`
	MX    []string          `json:"MX"`
	TXT   []string          `json:"TXT"`
	Error map[string]string `json:"error,omitempty"`
}

// queryDNS 查询指定类型的DNS记录，支持超时
func queryDNS(domain, dnsServer, queryType string, timeout time.Duration) ([]string, error) {
	client := new(dns.Client)
	client.Timeout = timeout
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.StringToType[queryType])
	resp, _, err := client.Exchange(m, dnsServer)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, ans := range resp.Answer {
		switch rr := ans.(type) {
		case *dns.A:
			result = append(result, rr.A.String())
		case *dns.AAAA:
			result = append(result, rr.AAAA.String())
		case *dns.CNAME:
			result = append(result, rr.Target)
		case *dns.NS:
			result = append(result, rr.Ns)
		case *dns.MX:
			result = append(result, fmt.Sprintf("%d %s", rr.Preference, rr.Mx))
		case *dns.TXT:
			result = append(result, rr.Txt...)
		}
	}
	return result, nil
}

// QueryAllDNS 查询所有常用DNS记录，支持超时
func QueryAllDNS(domain, dnsServer string, timeout time.Duration) DNSResult {
	result := DNSResult{Error: make(map[string]string)}

	// 定义要查询的记录类型和对应的结果字段
	records := []struct {
		qType  string
		record *[]string
	}{
		{"A", &result.A},
		{"AAAA", &result.AAAA},
		{"CNAME", &result.CNAME},
		{"NS", &result.NS},
		{"MX", &result.MX},
		{"TXT", &result.TXT},
	}

	// 依次查询
	for _, r := range records {
		res, err := queryDNS(domain, dnsServer, r.qType, timeout)
		if err != nil {
			result.Error[r.qType] = err.Error()
		} else {
			*r.record = res
		}
	}

	// 如果没有错误，则返回 nil 更语义化
	if len(result.Error) == 0 {
		result.Error = nil
	}

	return result
}

// LookupNSServers 递归查找最上级可用NS服务器
func LookupNSServers(domain, dnsServer string, timeout time.Duration) ([]string, error) {
	// 规范化域名，去除前后点
	domain = strings.Trim(domain, ".")
	labels := strings.Split(domain, ".")
	for i := 0; i < len(labels); i++ {
		parent := strings.Join(labels[i:], ".")
		ns, err := queryDNS(parent, dnsServer, "NS", timeout)
		if err != nil || len(ns) == 0 {
			continue // 没查到就往上一级查
		}
		return ns, nil
	}
	return nil, errors.New("no ns record found for any parent domain")
}

// LookupCNAMEChain 查询域名的CNAME链路，直到没有CNAME为止 //只取第一个CNAME（通常只会有一个）
func LookupCNAMEChain(domain, dnsServer string, timeout time.Duration) ([]string, error) {
	var chain []string
	visited := make(map[string]struct{})
	current := domain

	for {
		// 防止死循环
		if _, ok := visited[current]; ok {
			break
		}
		visited[current] = struct{}{}
		chain = append(chain, current)

		cnames, err := queryDNS(current, dnsServer, "CNAME", timeout)
		if err != nil || len(cnames) == 0 {
			break // 没有CNAME或查询出错，终止
		}
		// 只取第一个CNAME（通常只会有一个）
		current = cnames[0]
	}

	return chain, nil
}

// QueryAllDNSWithMultiResolvers 随机选5个DNS服务器进行并发查询
func QueryAllDNSWithMultiResolvers(domain string, resolvers []string, timeout time.Duration, pick int) DNSResult {
	picked := PickRandomElements(resolvers, pick)
	var wg sync.WaitGroup
	results := make([]DNSResult, pick)
	wg.Add(pick)
	for i, resolver := range picked {
		go func(i int, resolver string) {
			defer wg.Done()
			results[i] = QueryAllDNS(domain, resolver, timeout)
		}(i, resolver)
	}
	wg.Wait()
	return MergeDNSResults(results)
}
