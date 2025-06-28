package ednsquery

import (
	"context"
	"fmt"
	"github.com/miekg/dns"
	"github.com/panjf2000/ants/v2"
	maputils2 "github.com/winezer0/cdnAnalyzer/pkg/maputils"
	"net"
	"sync"
	"time"
)

// EDNSResult 存放最后格式化的结果
type EDNSResult struct {
	Domain      string   // 原始域名
	FinalDomain string   // 最终解析出的域名（CNAME 链尾部）
	NameServers []string // 权威 DNS 列表
	CNAMEChains []string // CNAME 链条
	A           []string // A 记录
	AAAA        []string // AAAA 记录
	CNAME       []string // CNAME 记录
	Errors      []string // 错误信息
	Locations   []string // 所有参与查询的 location（如 Beijing@8.8.8.8）
}

type DomainCityEDNSResultMap = map[string]map[string]*EDNSResult
type CityEDNSResultMap = map[string]*EDNSResult
type DomainEDNSResultMap = map[string]*EDNSResult

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
		Domain: domain,
		A:      ipv4s,
		AAAA:   ipv6s,
		CNAME:  cnames,
	}
}

// ResolveEDNSWithCities 批量解析多个域名在多个 DNS 和 location 下的 EDNS 响应
func ResolveEDNSWithCities(
	domains []string,
	cities []map[string]string,
	timeout time.Duration,
	maxConcurrency int,
	queryCNAMES bool,
	useSysNSQueryCNAMES bool,
) DomainCityEDNSResultMap {

	// Step 1: 异步并发预查所有域名的 CNAME / NS
	ctx := context.Background()
	var preResults []DomainPreQueryResult

	if queryCNAMES {
		preResults = preQueryDomains(ctx, domains, timeout, maxConcurrency, useSysNSQueryCNAMES)
	} else {
		preResults = NewEmptyDomainPreQueryResults(domains)
	}

	// Step 2: 创建协程池进行并发查询
	pool, err := ants.NewPool(maxConcurrency)
	if err != nil {
		panic(err)
	}
	defer pool.Release()

	resultMap := make(DomainCityEDNSResultMap)
	var mu sync.Mutex
	var wg sync.WaitGroup

	resultsChan := make(chan struct{}, len(preResults)) // 控制等待完成

	// 对每个域名启动任务
	for _, pre := range preResults {
		wg.Add(1)

		go func(pr DomainPreQueryResult) {
			defer wg.Done()

			// 合并权威 DNS 和默认 DNS
			dnsServers := []string{"8.8.8.8:53"}
			if queryCNAMES && len(pr.NameServers) > 0 {
				dnsServers = maputils2.UniqueMergeSlices(dnsServers, pr.NameServers)
			}

			// 为该域名下的所有 DNS 和 Location 提交子任务
			for _, dnsServer := range dnsServers {
				for _, cityEntry := range cities {
					city := maputils2.GetMapValue(cityEntry, "City")
					cityIP := maputils2.GetMapValue(cityEntry, "IP")

					wg.Add(1)

					pool.Submit(func() {
						defer wg.Done()

						// 执行 EDNS 查询
						res := ResolveEDNS(pr.FinalDomain, cityIP, dnsServer, timeout)

						// 构造 key
						key := fmt.Sprintf("%s@%s", city, dnsServer)

						// 构建完整的 EDNSResult
						ednsRes := EDNSResult{
							Domain:      pr.Domain,
							FinalDomain: pr.FinalDomain,
							NameServers: dnsServers,
							CNAMEChains: pr.CNAMEChains,
							A:           res.A,
							AAAA:        res.AAAA,
							CNAME:       res.CNAME,
							Errors:      res.Errors,
						}

						// 写入结果
						mu.Lock()
						if _, exists := resultMap[pr.Domain]; !exists {
							resultMap[pr.Domain] = make(CityEDNSResultMap)
						}
						resultMap[pr.Domain][key] = &ednsRes
						mu.Unlock()

						resultsChan <- struct{}{}
					})
				}
			}
		}(pre)
	}

	// 等待所有任务完成
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// 等待所有结果
	for range resultsChan {
		// 只用于阻塞主 goroutine
	}

	return resultMap
}
