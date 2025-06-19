package ednsquery

import (
	"cdnCheck/domaininfo/dnsquery"
	"context"
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"net"
	"strings"
	"time"
)

// lookupNSServers 递归查找最上级可用NS服务器
func lookupNSServers(domain, dnsServer string, timeout time.Duration) ([]string, error) {
	// 规范化域名，去除前后点
	domain = strings.Trim(domain, ".")
	labels := strings.Split(domain, ".")
	for i := 0; i < len(labels); i++ {
		parent := strings.Join(labels[i:], ".")
		//ns, err := queryDNS(parent, dnsServer, "NS", timeout)
		ns, err := lookupNSSOA(parent, dnsServer, timeout)
		if err != nil || len(ns) == 0 {
			continue // 没查到就往上一级查
		}
		return ns, nil
	}
	return nil, errors.New("no ns record found for any parent domain")
}

// lookupNSSOA 查询该域名的权威名称服务器（Name Server, NS）记录
func lookupNSSOA(domain string, dnsServer string, timeout time.Duration) (adders []string, err error) {
	client := new(dns.Client)
	client.Timeout = timeout
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	resp, _, err := client.Exchange(msg, dnsServer)
	if err != nil {
		return nil, err
	}
	// 检查 Authority Section 是否有 SOA
	for _, rr := range resp.Ns {
		if soa, ok := rr.(*dns.SOA); ok {
			adders = append(adders, soa.Ns)
		}
	}
	for _, ans := range resp.Answer {
		if ns, ok := ans.(*dns.NS); ok {
			adders = append(adders, ns.Ns)
		}
	}
	if len(adders) == 0 {
		return nil, errors.New("no ns record")
	}
	return adders, nil
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

// lookupCNAMEChains 查询域名的CNAME链路，直到没有CNAME为止 //只取第一个CNAME（通常只会有一个）
func lookupCNAMEChains(domain, dnsServer string, timeout time.Duration) ([]string, error) {
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

		cnames, err := dnsquery.ResolveDNS(current, dnsServer, "CNAME", timeout)
		if err != nil || len(cnames) == 0 {
			break // 没有CNAME或查询出错，终止
		}
		// 只取第一个CNAME（通常只会有一个）
		current = cnames[0]
	}

	return chain, nil
}
