package ednsquery

import (
	"cdnCheck/domaininfo/dnsquery"
	"context"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"sync"
	"time"
)

type DomainPreQueryResult struct {
	Domain      string
	FinalDomain string
	NameServers []string
	CNAMEs      []string
	Err         error
}

// getAuthoritativeNameservers   辅助函数：获取权威 DNS 服务器列表
func getAuthoritativeNameservers(domain, defaultNS string, timeout time.Duration) []string {
	if len(defaultNS) == 0 {
		defaultNS = defaultServerAddress
	}

	nsServers, err := lookupNSServers(domain, defaultNS, timeout)
	if err != nil || len(nsServers) == 0 {
		return []string{defaultNS} // fallback
	}

	var servers []string
	for _, ns := range nsServers {
		servers = append(servers, dnsquery.DnsServerAddPort(ns))
	}

	return servers
}

// getFinalDomain 辅助函数：CNAME 跟踪
func getFinalDomain(domain string, defaultDNS string, timeout time.Duration) (string, error) {
	cnameChain, err := lookupCNAMEChains(domain, defaultDNS, timeout)
	if err != nil {
		return domain, fmt.Errorf("failed to lookup CNAME chain: %w", err)
	}
	if len(cnameChain) > 0 {
		return cnameChain[len(cnameChain)-1], nil
	}

	return domain, nil
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
			cnameChain, err := lookupCNAMEChains(domain, defaultDNS, timeout)
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
