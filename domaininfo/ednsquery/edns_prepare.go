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

// preQueryDomains 辅助函数：异步并发预查 CNAME / NS
func preQueryDomains(ctx context.Context, domains []string, timeout time.Duration, maxConcurrency int, useSysNS bool) []DomainPreQueryResult {

	defaultNS := "8.8.8.8:53"
	if useSysNS {
		defaultNS = dnsquery.GetSystemDefaultAddress()
	}

	pool, err := ants.NewPool(maxConcurrency)
	if err != nil {
		panic(err)
	}
	defer pool.Release()

	resultChan := make(chan DomainPreQueryResult, len(domains))
	var wg sync.WaitGroup

	for _, domain := range domains {
		domain := domain // 避免闭包捕获问题
		wg.Add(1)

		_ = pool.Submit(func() {
			defer wg.Done()

			var (
				cnameChains []string
				finalDomain string
				cnameErr    error
				nsServers   []string
				nsErr       error
			)

			var stepWg sync.WaitGroup
			stepWg.Add(2)

			// Step 1: 并发执行 CNAME 查询
			go func() {
				defer stepWg.Done()
				cnameChains, finalDomain, cnameErr = dnsquery.LookupCNAMEChains(domain, defaultNS, timeout)
				if cnameErr != nil {
					fmt.Printf("failed to lookup [%v] CNAME chains: %v\n", domain, cnameErr)
				}
			}()

			// Step 2: 并发执行 NS 查询
			go func() {
				defer stepWg.Done()
				nsServers, nsErr = dnsquery.LookupNSServers(domain, defaultNS, timeout)
				if nsErr != nil {
					fmt.Printf("failed to lookup [%v] NS servers: %v\n", domain, nsErr)
				}
			}()

			// 等待两个步骤完成
			stepWg.Wait()

			// 填充结果
			res := NewEmptyDomainPreQueryResult(domain)
			if cnameErr == nil && len(cnameChains) > 0 {
				res.CNAMEChains = cnameChains
				res.FinalDomain = finalDomain
				//fmt.Printf("success to lookup [%v] Final Domain [%v] CNAME Chains: %v\n", domain, finalDomain, cnameChains)
			}

			if nsErr == nil && len(nsServers) > 0 {
				res.NameServers = dnsquery.NSServersAddPort(nsServers)
				//fmt.Printf("success to lookup [%v] NS servers: %v\n", domain, res.NameServers)
			}

			// 防止在 context 被取消后继续发送结果
			select {
			case <-ctx.Done():
				return
			case resultChan <- res:
			}
		})
	}

	// 启动 goroutine 等待所有任务完成并关闭 channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	var results []DomainPreQueryResult
	for r := range resultChan {
		results = append(results, r)
	}

	return results
}
