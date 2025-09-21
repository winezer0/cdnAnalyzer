package querydomain

import (
	"sync"
	"time"

	"github.com/winezer0/cdnAnalyzer/pkg/classify"
	"github.com/winezer0/cdnAnalyzer/pkg/domaininfo/dnsquery"
	"github.com/winezer0/cdnAnalyzer/pkg/domaininfo/ednsquery"
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
	QueryType          string // 新增：查询类型选项 dns, edns, both
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

// Process 根据QueryType执行相应的DNS查询并收集结果
func (pro *DNSProcessor) Process() *dnsquery.DomainDNSResultMap {
	// 1. 收集所有域名
	domains := classify.ExtractFMTsPtr(pro.DomainEntries)
	
	// 根据QueryType决定执行哪种查询
	var dnsResultMap dnsquery.DomainResolverDNSResultMap
	var ednsResultMap ednsquery.DomainCityEDNSResultMap
	
	var wg sync.WaitGroup
	
	// 判断是否需要执行DNS查询
	if pro.DNSQueryConfig.QueryType == "dns" || pro.DNSQueryConfig.QueryType == "both" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dnsResultMap = dnsquery.ResolveDNSWithResolversMulti(
				domains,
				nil, // recordTypes, nil表示默认类型
				pro.DNSQueryConfig.Resolvers,
				pro.DNSQueryConfig.Timeout,
				pro.DNSQueryConfig.MaxDNSConcurrency,
			)
		}()
	}
	
	// 判断是否需要执行EDNS查询
	if pro.DNSQueryConfig.QueryType == "edns" || pro.DNSQueryConfig.QueryType == "both" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ednsResultMap = ednsquery.ResolveEDNSWithCities(
				domains,
				pro.DNSQueryConfig.CityMap,
				pro.DNSQueryConfig.Timeout,
				pro.DNSQueryConfig.MaxEDNSConcurrency,
				pro.DNSQueryConfig.QueryEDNSCNAMES,
				pro.DNSQueryConfig.QueryEDNSUseSysNS,
			)
		}()
	}
	
	// 等待所有查询完成
	wg.Wait()
	
	// 处理DNS查询结果
	var domainDNSResultMap dnsquery.DomainDNSResultMap
	if pro.DNSQueryConfig.QueryType == "dns" || pro.DNSQueryConfig.QueryType == "both" {
		domainDNSResultMap = dnsquery.MergeDomainResolverResultMap(dnsResultMap)
	} else {
		// 如果没有执行DNS查询，初始化空的结果映射
		domainDNSResultMap = make(dnsquery.DomainDNSResultMap)
		for _, domain := range domains {
			domainDNSResultMap[domain] = &dnsquery.DNSResult{
				Error: make(map[string]string),
			}
		}
	}
	
	// 如果执行了EDNS查询，合并结果
	if pro.DNSQueryConfig.QueryType == "edns" || pro.DNSQueryConfig.QueryType == "both" {
		domainEDNSResultMap := ednsquery.MergeDomainCityEDNSResultMap(ednsResultMap)
		MergeEDNSMapToDNSMap(domainDNSResultMap, domainEDNSResultMap)
	}
	
	return &domainDNSResultMap
}