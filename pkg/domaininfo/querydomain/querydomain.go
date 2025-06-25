package querydomain

import (
	classify2 "cdnAnalyzer/pkg/classify"
	dnsquery2 "cdnAnalyzer/pkg/domaininfo/dnsquery"
	ednsquery2 "cdnAnalyzer/pkg/domaininfo/ednsquery"
	"fmt"
	"sync"
	"time"
)

// DNSQueryConfig 存储DNS查询配置
type DNSQueryConfig struct {
	Resolvers          []string
	CityMap            []map[string]string
	Timeout            time.Duration
	MaxDNSConcurrency  int
	MaxEDNSConcurrency int
	QueryEDNSCNAMES    bool
	QueryEDNSUseSysNS  bool
}

// DNSProcessor DNS查询处理器
type DNSProcessor struct {
	DNSQueryConfig *DNSQueryConfig
	DomainEntries  *[]classify2.TargetEntry
}

// NewDNSProcessor 创建新的DNS处理器
func NewDNSProcessor(dnsQueryConfig *DNSQueryConfig, domainEntry *[]classify2.TargetEntry) *DNSProcessor {
	return &DNSProcessor{
		DNSQueryConfig: dnsQueryConfig,
		DomainEntries:  domainEntry,
	}
}

// DNSQueryResult 存储DNS查询结果
type DNSQueryResult struct {
	DomainEntry *classify2.TargetEntry
	DNSResult   *dnsquery2.DNSResult
}

// Process 执行DNS查询并收集结果
func (pro *DNSProcessor) Process() *dnsquery2.DomainDNSResultMap {
	// 1. 收集所有域名
	domains := classify2.ExtractFMTsPtr(pro.DomainEntries)

	// 2. 批量常规DNS查询
	dnsResultMap := dnsquery2.ResolveDNSWithResolversMulti(
		domains,
		nil, // recordTypes, nil表示默认类型
		pro.DNSQueryConfig.Resolvers,
		pro.DNSQueryConfig.Timeout,
		pro.DNSQueryConfig.MaxDNSConcurrency,
	)
	// 合并为每域名一个DNSResult DomainDNSResultMap
	domainDNSResultMap := dnsquery2.MergeDomainResolverResultMap(dnsResultMap)

	// 3. 批量EDNS查询
	ednsResultMap := ednsquery2.ResolveEDNSWithCities(
		domains,
		pro.DNSQueryConfig.CityMap,
		pro.DNSQueryConfig.Timeout,
		pro.DNSQueryConfig.MaxEDNSConcurrency,
		pro.DNSQueryConfig.QueryEDNSCNAMES,
		pro.DNSQueryConfig.QueryEDNSUseSysNS,
	)
	// 合并为每域名一个EDNSResult
	domainEDNSResultMap := ednsquery2.MergeDomainCityEDNSResultMap(ednsResultMap)

	//合并 domainDNSResultMap 和 domainEDNSResultMap
	domainDNSResultMap = MergeEDNSMapToDNSMap(domainDNSResultMap, domainEDNSResultMap)
	return &domainDNSResultMap
}

// FastProcess 并行执行DNS和EDNS查询，并允许分别指定并发线程数
func (pro *DNSProcessor) FastProcess() *dnsquery2.DomainDNSResultMap {
	domains := classify2.ExtractFMTsPtr(pro.DomainEntries)

	var finalDNSMap dnsquery2.DomainDNSResultMap
	var finalEDNSMap ednsquery2.DomainEDNSResultMap

	var wg sync.WaitGroup
	wg.Add(2)

	// DNS 查询 channel
	dnsChan := make(chan dnsquery2.DomainDNSResultMap, 1)
	// EDNS 查询 channel
	ednsChan := make(chan ednsquery2.DomainEDNSResultMap, 1)

	// 启动 DNS 查询 goroutine
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("DNS goroutine panic:", r)
				dnsChan <- dnsquery2.DomainDNSResultMap{}
			}
		}()
		resultMap := dnsquery2.ResolveDNSWithResolversMulti(
			domains,
			nil,
			pro.DNSQueryConfig.Resolvers,
			pro.DNSQueryConfig.Timeout,
			pro.DNSQueryConfig.MaxDNSConcurrency,
		)
		merged := dnsquery2.MergeDomainResolverResultMap(resultMap)
		dnsChan <- merged
	}()

	// 启动 EDNS 查询 goroutine
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("EDNS goroutine panic:", r)
				ednsChan <- ednsquery2.DomainEDNSResultMap{}
			}
		}()
		resultMap := ednsquery2.ResolveEDNSWithCities(
			domains,
			pro.DNSQueryConfig.CityMap,
			pro.DNSQueryConfig.Timeout,
			pro.DNSQueryConfig.MaxEDNSConcurrency,
			pro.DNSQueryConfig.QueryEDNSCNAMES,
			pro.DNSQueryConfig.QueryEDNSUseSysNS,
		)
		merged := ednsquery2.MergeDomainCityEDNSResultMap(resultMap)
		ednsChan <- merged
	}()

	// 等待两个任务完成并关闭channel
	go func() {
		wg.Wait()
		close(dnsChan)
		close(ednsChan)
	}()

	// 阻塞等待结果
	finalDNSMap = <-dnsChan
	finalEDNSMap = <-ednsChan

	// 合并结果（如果需要）
	if len(finalEDNSMap) > 0 {
		MergeEDNSMapToDNSMap(finalDNSMap, finalEDNSMap)
	}
	return &finalDNSMap
}
