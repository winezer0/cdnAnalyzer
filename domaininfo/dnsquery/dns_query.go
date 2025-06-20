package dnsquery

import (
	"cdnCheck/maputils"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/panjf2000/ants/v2"
)

// DNSResult 表示一个解析器对某个域名的查询结果（按 recordType 分类）
type DNSResult struct {
	A     []string          `json:"A,omitempty"`
	AAAA  []string          `json:"AAAA,omitempty"`
	CNAME []string          `json:"CNAME,omitempty"`
	MX    []string          `json:"MX,omitempty"`
	NS    []string          `json:"NS,omitempty"`
	TXT   []string          `json:"TXT,omitempty"`
	Error map[string]string `json:"Error,omitempty"` // key: record type, value: error message
}

// NewEmptyDNSQueryResult 返回一个空的 DNS 查询结果对象
func NewEmptyDNSQueryResult() *DNSResult {
	return &DNSResult{
		A:     nil,
		AAAA:  nil,
		CNAME: nil,
		MX:    nil,
		NS:    nil,
		TXT:   nil,
		Error: make(map[string]string),
	}
}

var DefaultRecordTypesSlice = []string{"A", "AAAA", "CNAME", "NS", "MX", "TXT"}

// DomainResolverDNSResultMap 是最终返回的结构：domain -> resolver -> 查询结果（使用指针避免拷贝）
type DomainResolverDNSResultMap = map[string]map[string]*DNSResult
type ResolverDNSResultMap = map[string]*DNSResult
type DomainDNSResultMap = map[string]*DNSResult

// ResolveDNS 查询指定类型的DNS记录，支持超时
func ResolveDNS(domain, dnsServer, queryType string, timeout time.Duration) ([]string, error) {
	client := &dns.Client{Timeout: timeout}
	m := &dns.Msg{}
	m.SetQuestion(dns.Fqdn(domain), dns.StringToType[queryType])
	dnsServer = nsServerAddPort(dnsServer)
	resp, _, err := client.Exchange(m, dnsServer)
	if err != nil {
		return nil, err
	}
	return parseRecord(resp), nil
}

func parseRecord(resp *dns.Msg) []string {
	var result []string
	for _, ans := range resp.Answer {
		switch rr := ans.(type) {
		case *dns.A:
			result = append(result, rr.A.String())
		case *dns.AAAA:
			result = append(result, rr.AAAA.String())
		case *dns.CNAME:
			result = append(result, strings.TrimSuffix(rr.Target, "."))
		case *dns.NS:
			result = append(result, strings.TrimSuffix(rr.Ns, "."))
		case *dns.MX:
			result = append(result, fmt.Sprintf("%d %s", rr.Preference, strings.TrimSuffix(rr.Mx, ".")))
		case *dns.TXT:
			result = append(result, rr.Txt...)
		}
	}
	return result
}

// setRecord 将结果写入对应的字段
func setRecord(result *DNSResult, qType string, recs []string) {
	switch qType {
	case "A":
		result.A = maputils.UniqueMergeSlices(result.A, recs)
	case "AAAA":
		result.AAAA = maputils.UniqueMergeSlices(result.AAAA, recs)
	case "CNAME":
		result.CNAME = maputils.UniqueMergeSlices(result.CNAME, recs)
	case "MX":
		result.MX = maputils.UniqueMergeSlices(result.MX, recs)
	case "NS":
		result.NS = maputils.UniqueMergeSlices(result.NS, recs)
	case "TXT":
		result.TXT = maputils.UniqueMergeSlices(result.TXT, recs)
	}
}

// ResolveDNSWithResolversMulti 支持多个 domain，并发控制，返回结构化结果（使用指针优化 map 操作）
func ResolveDNSWithResolversMulti(
	domains []string,
	recordTypes []string,
	resolvers []string,
	timeout time.Duration,
	maxConcurrency int,
) DomainResolverDNSResultMap {

	if len(recordTypes) == 0 {
		recordTypes = DefaultRecordTypesSlice
	}

	results := make(map[string]map[string]*DNSResult)
	var mu sync.Mutex
	var wgAll sync.WaitGroup

	// 创建 goroutine 池
	pool, err := ants.NewPool(maxConcurrency, ants.WithOptions(ants.Options{
		Nonblocking: true,
	}))
	if err != nil {
		panic(err)
	}
	defer pool.Release()

	// 初始化每个 domain 的结果容器
	for _, domain := range domains {
		domain := domain
		results[domain] = make(map[string]*DNSResult)
		for _, resolver := range resolvers {
			resolver := resolver
			results[domain][resolver] = NewEmptyDNSQueryResult()
		}
	}

	// 提交并发任务
	for _, domain := range domains {
		domain := domain
		for _, resolver := range resolvers {
			resolver := resolver
			for _, qType := range recordTypes {
				qType := qType
				wgAll.Add(1)

				err := pool.Submit(func() {
					defer wgAll.Done()

					recs, err := ResolveDNS(domain, resolver, qType, timeout)

					mu.Lock()
					if err != nil {
						results[domain][resolver].Error[qType] = err.Error()
					} else {
						setRecord(results[domain][resolver], qType, recs)
					}
					mu.Unlock()
				})

				if err != nil {
					// 回退到同步执行
					go func() {
						defer wgAll.Done()

						recs, err := ResolveDNS(domain, resolver, qType, timeout)

						mu.Lock()
						if err != nil {
							results[domain][resolver].Error[qType] = err.Error()
						} else {
							setRecord(results[domain][resolver], qType, recs)
						}
						mu.Unlock()
					}()
				}
			}
		}
	}

	// 等待所有任务完成
	wgAll.Wait()

	return results
}
