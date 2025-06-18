package ednsquery

import (
	"cdnCheck/maputils"
	"context"
	"fmt"
	"github.com/miekg/dns"
	"github.com/panjf2000/ants/v2"
	"net"
	"sync"
	"time"
)

var defaultServerAddress = GetSystemDefaultAddress()

type EDNSResult struct {
	Domain      string   // 原始域名
	FinalDomain string   // 最终解析域名（CNAME 链尾部）
	NameServers []string // 权威 DNS 列表
	CNAMEs      []string // CNAME 链条
	A           []string
	AAAA        []string
	CNAME       []string
	Errors      []string
}

// eDNSMessage 创建并返回一个包含 EDNS（扩展 DNS）选项的 DNS 查询消息
func eDNSMessage(domain, EDNSAddr string) *dns.Msg {
	e := new(dns.EDNS0_SUBNET) // EDNS
	e.Code = dns.EDNS0SUBNET
	e.Family = 1
	e.SourceNetmask = 24
	e.Address = net.ParseIP(EDNSAddr).To4()
	e.SourceScope = 0

	o := new(dns.OPT)
	o.Hdr.Name = "."
	o.Hdr.Rrtype = dns.TypeOPT
	o.Option = append(o.Option, e)
	o.SetUDPSize(dns.DefaultMsgSize)

	m := new(dns.Msg)
	m.SetQuestion(domain, dns.TypeA)
	m.Extra = append(m.Extra, o)
	return m
}

// ResolveEDNS 进行EDNS信息查询
func ResolveEDNS(domain string, EDNSAddr string, dnsServer string, timeout time.Duration) EDNSResult {
	domain = dns.Fqdn(domain)

	Client := new(dns.Client)
	Client.Timeout = timeout
	dnsMsg := eDNSMessage(domain, EDNSAddr)

	in, _, err := Client.Exchange(dnsMsg, dnsServer)
	if err != nil {
		return EDNSResult{
			Errors: []string{err.Error()},
		}
	}

	var ipv4s []string
	var ipv6s []string
	var cnames []string

	for _, answer := range in.Answer {
		switch answer.Header().Rrtype {
		case dns.TypeA:
			ipv4s = append(ipv4s, answer.(*dns.A).A.String())
		case dns.TypeAAAA:
			ipv6s = append(ipv6s, answer.(*dns.AAAA).AAAA.String())
		case dns.TypeCNAME:
			cnames = append(cnames, answer.(*dns.CNAME).Target)
		}
	}

	return EDNSResult{
		A:     ipv4s,
		AAAA:  ipv6s,
		CNAME: cnames,
	}
}

// ResolveEDNSWithCities 不再负责 CNAME / NS 查询，只执行 EDNS 查询任务
func ResolveEDNSWithCities(
	domain string,
	finalDomain string,
	authoritativeNameservers []string,
	cnames []string,
	Cities []map[string]string,
	timeout time.Duration,
	maxConcurrency int,
) map[string]EDNSResult {

	// 合并权威 DNS + 默认 DNS
	dnsServers := append([]string{"8.8.8.8:53"}, authoritativeNameservers...)

	// 创建线程池
	pool, err := ants.NewPool(maxConcurrency, ants.WithOptions(ants.Options{
		Nonblocking: true,
	}))
	if err != nil {
		panic(err)
	}
	defer pool.Release()

	var wg sync.WaitGroup
	resultMap := make(map[string]EDNSResult)
	var mu sync.Mutex

	results := make(chan struct{}, len(Cities)*len(dnsServers)) // 仅用于等待完成

	// 提交任务
	for _, dnsServer := range dnsServers {
		for _, entry := range Cities {
			city := maputils.GetMapValue(entry, "City")
			cityIP := maputils.GetMapValue(entry, "IP")

			wg.Add(1)

			task := func(city, cityIP, dnsServer string) {
				defer wg.Done()

				res := ResolveEDNS(finalDomain, cityIP, dnsServer, timeout)

				key := fmt.Sprintf("%s@%s", city, dnsServer)

				mu.Lock()
				resultMap[key] = EDNSResult{
					Domain:      domain,
					FinalDomain: finalDomain,
					NameServers: dnsServers,
					CNAMEs:      cnames,
					A:           res.A,
					AAAA:        res.AAAA,
					CNAME:       res.CNAME,
					Errors:      res.Errors,
				}
				mu.Unlock()

				results <- struct{}{}
			}

			err := pool.Submit(func() {
				task(city, cityIP, dnsServer)
			})

			if err != nil {
				go task(city, cityIP, dnsServer)
			}
		}
	}

	// 等待所有任务完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 等待所有结果
	for range results {
		// 仅用于接收完所有信号
	}

	return resultMap
}

type DomainPreQueryResult struct {
	Domain      string
	FinalDomain string
	NameServers []string
	CNAMEs      []string
	Err         error
}

// ResolveEDNSDomainsWithCities 接收多个域名，先批量查询 CNAME / NS，再并发 EDNS 查询
// ResolveEDNSDomainsWithCities 接收多个域名，先批量查询 CNAME / NS，再并发 EDNS 查询
func ResolveEDNSDomainsWithCities(
	domains []string,
	Cities []map[string]string,
	timeout time.Duration,
	maxConcurrency int,
) map[string]map[string]EDNSResult {

	// Step 1: 异步并发预查所有域名的 CNAME / NS
	ctx := context.Background()
	preResults := preQueryDomains(ctx, domains, defaultServerAddress, timeout, maxConcurrency)

	// Step 2: 并发执行每个域名的 EDNS 查询
	resultMap := make(map[string]map[string]EDNSResult)
	var mu sync.Mutex
	var wg sync.WaitGroup

	pool, _ := ants.NewPool(maxConcurrency)
	defer pool.Release()

	for _, pr := range preResults {
		wg.Add(1)

		pr := pr // 捕获变量
		pool.Submit(func() {
			defer wg.Done()

			ednsResults := ResolveEDNSWithCities(
				pr.Domain,
				pr.FinalDomain,
				pr.NameServers,
				pr.CNAMEs,
				Cities,
				timeout,
				maxConcurrency,
			)

			mu.Lock()
			resultMap[pr.Domain] = ednsResults
			mu.Unlock()
		})
	}

	wg.Wait()
	return resultMap
}

// 辅助函数：异步并发预查 CNAME / NS
func preQueryDomains(
	ctx context.Context,
	domains []string,
	defaultDNS string,
	timeout time.Duration,
	maxConcurrency int,
) []DomainPreQueryResult {

	pool, _ := ants.NewPool(maxConcurrency)
	defer pool.Release()

	resultChan := make(chan DomainPreQueryResult, len(domains))
	var wg sync.WaitGroup

	for _, domain := range domains {
		wg.Add(1)

		go func(domain string) {
			defer wg.Done()

			var res DomainPreQueryResult
			res.Domain = domain

			// Step 1: 获取最终域名（CNAME 跟踪）
			finalDomain, err := getFinalDomain(domain, defaultDNS, timeout)
			if err != nil {
				res.Err = err
				res.FinalDomain = domain
			} else {
				res.FinalDomain = finalDomain
			}

			// Step 2: 获取 CNAME 链
			cnameChain, err := LookupCNAMEChain(domain, defaultDNS, timeout)
			if err == nil && len(cnameChain) > 0 {
				res.CNAMEs = cnameChain
			}

			// Step 3: 获取权威 DNS 服务器
			nsList := getAuthoritativeNameservers(res.FinalDomain, defaultDNS, timeout)
			if err == nil && len(nsList) > 0 {
				res.NameServers = nsList
			}

			select {
			case <-ctx.Done():
			case resultChan <- res:
			}
		}(domain)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var results []DomainPreQueryResult
	for r := range resultChan {
		results = append(results, r)
	}

	return results
}
