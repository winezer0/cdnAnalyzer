package querydomain

import (
	"cdnCheck/classify"
	"cdnCheck/domaininfo/dnsquery"
	"cdnCheck/domaininfo/ednsquery"
	"cdnCheck/models"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// QueryConfig 存储DNS查询配置
type QueryConfig struct {
	Resolvers  []string
	RandCities []map[string]string
	Timeout    time.Duration
}

// QueryResult 存储DNS查询结果
type QueryResult struct {
	DomainEntry classify.TargetEntry
	DNSResult   *dnsquery.DNSResult
}

// DNSProcessor DNS查询处理器
type DNSProcessor struct {
	Config     *QueryConfig
	Classifier *classify.TargetClassifier
}

// NewDNSProcessor 创建新的DNS处理器
func NewDNSProcessor(config *QueryConfig, classifier *classify.TargetClassifier) *DNSProcessor {
	return &DNSProcessor{
		Config:     config,
		Classifier: classifier,
	}
}

// Process 执行DNS查询并收集结果
func (p *DNSProcessor) Process() []*models.CheckInfo {
	var checkResults []*models.CheckInfo
	var wg sync.WaitGroup
	dnsResultChan := make(chan *models.CheckInfo, len(p.Classifier.Domains))

	// 启动并发查询
	for _, domainEntry := range p.Classifier.Domains {
		wg.Add(1)
		go processDomain(domainEntry, p.Config, dnsResultChan, &wg)
	}

	// 启动一个 goroutine 等待所有任务完成，并关闭 channel
	go func() {
		wg.Wait()
		close(dnsResultChan)
	}()

	// 接收DNS查询结果，实时输出，并保存到列表中
	for dnsResult := range dnsResultChan {
		checkResults = append(checkResults, dnsResult)
		finalInfo, _ := json.MarshalIndent(dnsResult, "", "  ")
		fmt.Println(string(finalInfo))
	}

	return checkResults
}

// queryDNS 执行DNS和EDNS查询
func queryDNS(domainEntry classify.TargetEntry, config *QueryConfig) (*QueryResult, error) {
	domain := domainEntry.Fmt

	// 常规 DNS 查询
	dnsResults := dnsquery.QueryAllDNSWithMultiResolvers(domain, config.Resolvers, config.Timeout)
	dnsQueryResult := dnsquery.MergeDNSResults(dnsResults)

	// EDNS 查询
	eDNSQueryResults := ednsquery.EDNSQueryWithMultiCities(domain, config.Timeout, config.RandCities, false)
	if len(eDNSQueryResults) == 0 {
		fmt.Fprintf(os.Stderr, "EDNS 查询结果为空: %s\n", domain)
	} else {
		eDNSQueryResult := MergeEDNSResults(eDNSQueryResults)
		dnsQueryResult = MergeEDNSToDNS(eDNSQueryResult, dnsQueryResult)
	}

	// 优化DNS查询结果
	dnsquery.OptimizeDNSResult(&dnsQueryResult)

	return &QueryResult{
		DomainEntry: domainEntry,
		DNSResult:   &dnsQueryResult,
	}, nil
}

// processDomain 处理单个域名查询任务
func processDomain(
	domainEntry classify.TargetEntry,
	config *QueryConfig,
	resultChan chan<- *models.CheckInfo,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	// 执行DNS查询
	queryResult, err := queryDNS(domainEntry, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DNS查询失败 [%s]: %v\n", domainEntry.Fmt, err)
		return
	}

	// 填充结果
	dnsResult := populateDNSResult(queryResult.DomainEntry, queryResult.DNSResult)

	// 发送结果到 channel
	resultChan <- dnsResult
}

// populateDNSResult 将 DNS 查询结果填充到 CheckInfo 中
func populateDNSResult(domainEntry classify.TargetEntry, query *dnsquery.DNSResult) *models.CheckInfo {
	result := models.NewDomainCheckInfo(domainEntry.Raw, domainEntry.Fmt, domainEntry.FromUrl)

	// 逐个复制 DNS 记录
	result.A = append(result.A, query.A...)
	result.AAAA = append(result.AAAA, query.AAAA...)
	result.CNAME = append(result.CNAME, query.CNAME...)
	result.NS = append(result.NS, query.NS...)
	result.MX = append(result.MX, query.MX...)
	result.TXT = append(result.TXT, query.TXT...)

	return result
}
