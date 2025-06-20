package querydomain

import (
	"cdnCheck/classify"
	"cdnCheck/domaininfo/dnsquery"
	"cdnCheck/domaininfo/ednsquery"
	"fmt"
	"time"
)

// DNSQueryConfig 存储DNS查询配置
type DNSQueryConfig struct {
	Resolvers          []string
	CityMap            []map[string]string
	Timeout            time.Duration
	MaxDNSConcurrency  int
	MaxEDNSConcurrency int
}

// DNSProcessor DNS查询处理器
type DNSProcessor struct {
	DNSQueryConfig *DNSQueryConfig
	DomainEntries  *[]classify.TargetEntry
}

// NewDNSProcessor 创建新的DNS处理器
func NewDNSProcessor(dnsQueryConfig *DNSQueryConfig, domainEntry *[]classify.TargetEntry) *DNSProcessor {
	return &DNSProcessor{
		DNSQueryConfig: dnsQueryConfig,
		DomainEntries:  domainEntry,
	}
}

// DNSQueryResult 存储DNS查询结果
type DNSQueryResult struct {
	DomainEntry *classify.TargetEntry
	DNSResult   *dnsquery.DNSResult
}

// Process 执行DNS查询并收集结果
func (pro *DNSProcessor) Process() *dnsquery.DomainDNSResultMap {
	// 1. 收集所有域名
	domains := classify.ExtractFMTsPtr(pro.DomainEntries)

	// 2. 批量常规DNS查询
	dnsResultMap := dnsquery.ResolveDNSWithResolversMulti(
		domains,
		nil, // recordTypes, nil表示默认类型
		pro.DNSQueryConfig.Resolvers,
		pro.DNSQueryConfig.Timeout,
		pro.DNSQueryConfig.MaxDNSConcurrency,
	)
	// 合并为每域名一个DNSResult DomainDNSResultMap
	domainDNSResultMap := dnsquery.MergeDomainResolverResultMap(dnsResultMap)

	// 3. 批量EDNS查询
	ednsResultMap := ednsquery.ResolveEDNSWithCities(
		domains,
		pro.DNSQueryConfig.CityMap,
		pro.DNSQueryConfig.Timeout,
		pro.DNSQueryConfig.MaxEDNSConcurrency,
		true, // queryCNAMES
		true, // useSysNSQueryCNAMES
	)
	// 合并为每域名一个EDNSResult
	domainEDNSResultMap := ednsquery.MergeDomainCityEDNSResultMap(ednsResultMap)

	//合并 domainDNSResultMap 和 domainEDNSResultMap
	domainDNSResultMap = MergeEDNSMapToDNSMap(domainDNSResultMap, domainEDNSResultMap)
	return &domainDNSResultMap
}

// FastProcess 并行执行DNS和EDNS查询，并允许分别指定并发线程数
func (pro *DNSProcessor) FastProcess() *dnsquery.DomainDNSResultMap {
	// 1. 获取所有域名（不重复）
	domains := classify.ExtractFMTsPtr(pro.DomainEntries)

	fmt.Printf("domains: %v\n", domains)
	fmt.Printf("resolvers: %v\n", pro.DNSQueryConfig.Resolvers)
	fmt.Printf("cityMap: %v\n", pro.DNSQueryConfig.CityMap)

	type dnsResult struct {
		domainDNSResultMap dnsquery.DomainDNSResultMap
	}
	type ednsResult struct {
		domainEDNSResultMap ednsquery.DomainEDNSResultMap
	}

	dnsChan := make(chan dnsResult, 1)
	ednsChan := make(chan ednsResult, 1)

	go func() {
		defer close(dnsChan)
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("DNS goroutine panic:", r)
				dnsChan <- dnsResult{domainDNSResultMap: make(dnsquery.DomainDNSResultMap)}
			}
		}()
		dnsResultMap := dnsquery.ResolveDNSWithResolversMulti(
			domains,
			nil,
			pro.DNSQueryConfig.Resolvers,
			pro.DNSQueryConfig.Timeout,
			pro.DNSQueryConfig.MaxDNSConcurrency,
		)
		fmt.Printf("dnsResultMap len: %d  -> %v\n", len(dnsResultMap), dnsResultMap)
		merged := dnsquery.MergeDomainResolverResultMap(dnsResultMap)
		fmt.Printf("merged DNSResultMap len: %d  -> %v\n", len(merged), merged)
		dnsChan <- dnsResult{domainDNSResultMap: merged}
	}()

	go func() {
		defer close(ednsChan)
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("EDNS goroutine panic:", r)
				ednsChan <- ednsResult{domainEDNSResultMap: make(ednsquery.DomainEDNSResultMap)}
			}
		}()
		ednsResultMap := ednsquery.ResolveEDNSWithCities(
			domains,
			pro.DNSQueryConfig.CityMap,
			pro.DNSQueryConfig.Timeout,
			pro.DNSQueryConfig.MaxEDNSConcurrency,
			true,
			true,
		)
		fmt.Printf("ednsResultMap len: %d -> %v\n", len(ednsResultMap), ednsResultMap)
		merged := ednsquery.MergeDomainCityEDNSResultMap(ednsResultMap)
		fmt.Printf("merged EDNSResultMap len: %d -> %v\n", len(merged), merged)
		ednsChan <- ednsResult{domainEDNSResultMap: merged}
	}()

	var finalDNSMap dnsquery.DomainDNSResultMap
	var finalEDNSMap ednsquery.DomainEDNSResultMap

	gotDNS := false
	gotEDNS := false
	for !(gotDNS && gotEDNS) {
		select {
		case dnsRes, ok := <-dnsChan:
			if ok {
				finalDNSMap = dnsRes.domainDNSResultMap
				fmt.Printf("main: finalDNSMap after assign: %d -> %v\n", len(finalDNSMap), finalDNSMap)
			}
			gotDNS = true
		case ednsRes, ok := <-ednsChan:
			if ok {
				finalEDNSMap = ednsRes.domainEDNSResultMap
				fmt.Printf("main: finalEDNSMap after assign: %d -> %v\n", len(finalEDNSMap), finalEDNSMap)
			}
			gotEDNS = true
		}
	}

	if len(finalDNSMap) == 0 {
		fmt.Println("No DNS results found")
	}
	if len(finalEDNSMap) == 0 {
		fmt.Println("No EDNS results found")
	}

	MergeEDNSMapToDNSMap(finalDNSMap, finalEDNSMap)
	return &finalDNSMap
}
