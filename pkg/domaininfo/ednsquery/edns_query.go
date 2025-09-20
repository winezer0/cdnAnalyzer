package ednsquery

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/winezer0/cdnAnalyzer/pkg/maputils"
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
	NS          []string // NS 记录
	MX          []string // MX 记录
	TXT         []string // TXT 记录
	Errors      []string // 错误信息
	Locations   []string // 所有参与查询的 location（如 Beijing@8.8.8.8）
}

type DomainCityEDNSResultMap = map[string]map[string]*EDNSResult
type CityEDNSResultMap = map[string]*EDNSResult
type DomainEDNSResultMap = map[string]*EDNSResult

// eDNSMessage 创建并返回一个包含 EDNS（扩展 DNS）选项的 DNS 查询消息
func eDNSMessage(domain, EDNSAddr string, qType uint16) *dns.Msg {
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
	m.SetQuestion(domain, qType)
	m.Extra = append(m.Extra, o)
	return m
}

// ResolveEDNS 进行EDNS信息查询
func ResolveEDNS(domain string, EDNSAddr string, dnsServer string, qType uint16, timeout time.Duration) EDNSResult {
	domain = dns.Fqdn(domain)

	Client := new(dns.Client)
	Client.Timeout = timeout
	dnsMsg := eDNSMessage(domain, EDNSAddr, qType)

	in, _, err := Client.Exchange(dnsMsg, dnsServer)
	if err != nil {
		return EDNSResult{
			Errors: []string{err.Error()},
		}
	}

	var ipv4s []string
	var ipv6s []string
	var cnames []string
	var nss []string
	var mxs []string
	var txts []string

	for _, answer := range in.Answer {
		switch answer.Header().Rrtype {
		case dns.TypeA:
			ipv4s = append(ipv4s, answer.(*dns.A).A.String())
		case dns.TypeAAAA:
			ipv6s = append(ipv6s, answer.(*dns.AAAA).AAAA.String())
		case dns.TypeCNAME:
			cnames = append(cnames, answer.(*dns.CNAME).Target)
		case dns.TypeNS:
			nss = append(nss, answer.(*dns.NS).Ns)
		case dns.TypeMX:
			mxs = append(mxs, fmt.Sprintf("%d %s", answer.(*dns.MX).Preference, answer.(*dns.MX).Mx))
		case dns.TypeTXT:
			for _, txt := range answer.(*dns.TXT).Txt {
				txts = append(txts, txt)
			}
		}
	}

	// 处理权威部分的NS记录
	for _, ns := range in.Ns {
		if ns.Header().Rrtype == dns.TypeNS {
			nss = append(nss, ns.(*dns.NS).Ns)
		}
	}

	return EDNSResult{
		Domain: domain,
		A:      ipv4s,
		AAAA:   ipv6s,
		CNAME:  cnames,
		NS:     nss,
		MX:     mxs,
		TXT:    txts,
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
	sem := make(chan struct{}, maxConcurrency)
	resultMap := make(DomainCityEDNSResultMap)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 对每个域名启动任务
	for _, pre := range preResults {
		wg.Add(1)

		go func(pr DomainPreQueryResult) {
			defer wg.Done()

			// 合并权威 DNS 和默认 DNS
			dnsServers := []string{"8.8.8.8:53"}
			if queryCNAMES && len(pr.NameServers) > 0 {
				dnsServers = maputils.UniqueMergeSlices(dnsServers, pr.NameServers)
			}

			// 优化：为该域名创建一个专门的结果通道，以更好地管理并发
			domainResultChan := make(chan struct {
				key string
				res *EDNSResult
			}, len(dnsServers)*len(cities))

			var domainWg sync.WaitGroup

			// 为该域名下的所有 DNS 和 Location 提交子任务
			for _, dnsServer := range dnsServers {
				for _, cityEntry := range cities {
					city := maputils.GetMapValue(cityEntry, "City")
					cityIP := maputils.GetMapValue(cityEntry, "IP")

					domainWg.Add(1)
					sem <- struct{}{} // 获取令牌

					go func(dnsServer, city, cityIP string) {
						defer domainWg.Done()
						defer func() { <-sem }() // 释放令牌

						// 分别执行各种类型的EDNS查询
						aResult := ResolveEDNS(pr.FinalDomain, cityIP, dnsServer, dns.TypeA, timeout)
						aaaaResult := ResolveEDNS(pr.FinalDomain, cityIP, dnsServer, dns.TypeAAAA, timeout)
						cnameResult := ResolveEDNS(pr.FinalDomain, cityIP, dnsServer, dns.TypeCNAME, timeout)
						nsResult := ResolveEDNS(pr.FinalDomain, cityIP, dnsServer, dns.TypeNS, timeout)
						mxResult := ResolveEDNS(pr.FinalDomain, cityIP, dnsServer, dns.TypeMX, timeout)
						txtResult := ResolveEDNS(pr.FinalDomain, cityIP, dnsServer, dns.TypeTXT, timeout)

						// 合并所有结果
						allErrors := []string{}
						allErrors = append(allErrors, aResult.Errors...)
						allErrors = append(allErrors, aaaaResult.Errors...)
						allErrors = append(allErrors, cnameResult.Errors...)
						allErrors = append(allErrors, nsResult.Errors...)
						allErrors = append(allErrors, mxResult.Errors...)
						allErrors = append(allErrors, txtResult.Errors...)

						// 构造 key
						key := fmt.Sprintf("%s@%s", city, dnsServer)

						// 构建完整的 EDNSResult
						ednsRes := &EDNSResult{
							Domain:      pr.Domain,
							FinalDomain: pr.FinalDomain,
							NameServers: dnsServers,
							CNAMEChains: pr.CNAMEChains,
							A:           aResult.A,
							AAAA:        aaaaResult.AAAA,
							CNAME:       cnameResult.CNAME,
							NS:          nsResult.NS,
							MX:          mxResult.MX,
							TXT:         txtResult.TXT,
							Errors:      allErrors,
						}

						domainResultChan <- struct {
							key string
							res *EDNSResult
						}{key: key, res: ednsRes}
					}(dnsServer, city, cityIP)
				}
			}

			// 等待该域名的所有查询完成
			go func() {
				domainWg.Wait()
				close(domainResultChan)
			}()

			// 收集该域名的所有结果
			mu.Lock()
			if _, exists := resultMap[pr.Domain]; !exists {
				resultMap[pr.Domain] = make(CityEDNSResultMap)
			}
			mu.Unlock()

			for domainResult := range domainResultChan {
				mu.Lock()
				resultMap[pr.Domain][domainResult.key] = domainResult.res
				mu.Unlock()
			}
		}(pre)
	}

	// 等待所有任务完成
	wg.Wait()

	return resultMap
}