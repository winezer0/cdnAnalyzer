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
	CNAMEChains []string
	Err         error
}

// NewEmptyDomainPreQueryResult 创建一个默认空值的 DomainPreQueryResult，只填充域名信息
func NewEmptyDomainPreQueryResult(domain string) DomainPreQueryResult {
	return DomainPreQueryResult{
		Domain:      domain,
		FinalDomain: domain,     // 默认空字符串
		NameServers: []string{}, // 默认空切片
		CNAMEChains: []string{}, // 默认空切片
		Err:         nil,        // 默认无错误
	}
}

// NewEmptyDomainPreQueryResults 创建一批默认空值的 DomainPreQueryResult，每个对应一个域名
func NewEmptyDomainPreQueryResults(domains []string) []DomainPreQueryResult {
	results := make([]DomainPreQueryResult, 0, len(domains))
	for _, domain := range domains {
		results = append(results, NewEmptyDomainPreQueryResult(domain))
	}
	return results
}

// 辅助函数：异步并发预查 CNAME / NS
func preQueryDomains(
	ctx context.Context,
	domains []string,
	defaultNS string,
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
			res.FinalDomain = domain
			res.NameServers = []string{}
			res.CNAMEChains = []string{}

			// Step 1: 获取最终域名（基于 CNAMES 跟踪）
			cnameChain, err := dnsquery.LookupCNAMEChains(domain, defaultNS, timeout)
			if err != nil {
				fmt.Errorf("failed to lookup [%v] CNAME chains: %v\n", domain, err)
			} else if len(cnameChain) > 0 {
				fmt.Printf("success to lookup [%v] CNAME chains: %v\n", domain, cnameChain)
				res.CNAMEChains = cnameChain
				res.FinalDomain = cnameChain[len(cnameChain)-1]
			}

			// Step 2: 获取权威 DNS 服务器
			nsServers, err := dnsquery.LookupNSServers(domain, defaultNS, timeout)
			if err != nil {
				fmt.Errorf("failed to lookup [%v] NS servers: %v\n", domain, err)
			} else if len(nsServers) > 0 {
				nsList := make([]string, 0, len(nsServers))
				for _, nsServer := range nsServers {
					nsList = append(nsList, dnsquery.DnsServerAddPort(nsServer))
				}
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
