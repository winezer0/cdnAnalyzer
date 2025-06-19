package dnsquery

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/panjf2000/ants/v2"
)

// DNSResult 统一返回结构
type DNSResult struct {
	A     []string          `json:"A"`
	AAAA  []string          `json:"AAAA"`
	CNAME []string          `json:"CNAME"`
	NS    []string          `json:"NS"`
	MX    []string          `json:"MX"`
	TXT   []string          `json:"TXT"`
	Error map[string]string `json:"error,omitempty"`
}

// ResolverResult 表示单个解析器的完整结果
type ResolverResult struct {
	Resolver string    `json:"resolver"`
	Result   DNSResult `json:"result,omitempty"`
}

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
func setRecord(res *DNSResult, qType string, records []string) {
	switch qType {
	case "A":
		res.A = records
	case "AAAA":
		res.AAAA = records
	case "CNAME":
		res.CNAME = records
	case "NS":
		res.NS = records
	case "MX":
		res.MX = records
	case "TXT":
		res.TXT = records
	}
}

var DefaultRecordTypesSlice = []string{"A", "AAAA", "CNAME", "NS", "MX", "TXT"}

// ResolveDNSWithResolvers 并发查询多个 DNS 解析器的指定记录类型（默认全部）
func ResolveDNSWithResolvers(
	ctx context.Context,
	domain string,
	recordTypes []string,
	resolvers []string,
	timeout time.Duration,
	maxConcurrency int,
) []ResolverResult {

	if recordTypes == nil || len(recordTypes) == 0 {
		recordTypes = DefaultRecordTypesSlice
	}

	results := make([]ResolverResult, len(resolvers))
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 创建 ants 线程池
	pool, err := ants.NewPool(maxConcurrency, ants.WithOptions(ants.Options{
		Nonblocking: true, // 提交失败不阻塞主线程
	}))
	if err != nil {
		panic(err)
	}
	defer pool.Release()

	// 为每个 resolver 提交一个任务
	for i, resolver := range resolvers {
		i, resolver := i, resolver
		wg.Add(1)

		_ = pool.Submit(func() {
			defer wg.Done()

			result := DNSResult{Error: make(map[string]string)}

			for _, qType := range recordTypes {
				select {
				case <-ctx.Done():
					mu.Lock()
					result.Error[qType] = ctx.Err().Error()
					mu.Unlock()
					return
				default:
				}

				recs, err := ResolveDNS(domain, resolver, qType, timeout)
				mu.Lock()
				if err != nil {
					result.Error[qType] = err.Error()
				} else {
					setRecord(&result, qType, recs)
				}
				mu.Unlock()
			}

			mu.Lock()
			results[i] = ResolverResult{
				Resolver: resolver,
				Result:   result,
			}
			mu.Unlock()
		})
	}

	// 等待所有任务完成
	wg.Wait()
	return results
}

type DNSResultForTask struct {
	Resolver string
	QType    string
	Records  []string
	Error    error
}

// ResolveDNSWithResolversAtom 更加原子性的查询操作  没有实现上下文处理
func ResolveDNSWithResolversAtom(
	domain string,
	recordTypes []string,
	resolvers []string,
	timeout time.Duration,
	maxConcurrency int,
) []ResolverResult {

	if len(recordTypes) == 0 {
		recordTypes = DefaultRecordTypesSlice
	}

	// 最终结果数组
	results := make([]ResolverResult, len(resolvers))
	resolverMap := make(map[string]int) // resolver -> index
	for i, r := range resolvers {
		resolverMap[r] = i
		results[i] = ResolverResult{
			Resolver: r,
			Result:   DNSResult{Error: make(map[string]string)},
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// 创建线程池
	pool, err := ants.NewPool(maxConcurrency, ants.WithOptions(ants.Options{
		Nonblocking: true,
	}))
	if err != nil {
		panic(err)
	}
	defer pool.Release()

	// 为每个 resolver 的每个 recordType 提交任务
	for _, resolver := range resolvers {
		for _, qType := range recordTypes {
			wg.Add(1)

			// 捕获变量
			resolver := resolver
			qType := qType

			err := pool.Submit(func() {
				defer wg.Done()

				recs, err := ResolveDNS(domain, resolver, qType, timeout)

				mu.Lock()
				idx := resolverMap[resolver]
				if err != nil {
					results[idx].Result.Error[qType] = err.Error()
				} else {
					setRecord(&results[idx].Result, qType, recs)
				}
				mu.Unlock()
			})

			if err != nil {
				// 回退：同步执行
				go func() {
					defer wg.Done()

					recs, err := ResolveDNS(domain, resolver, qType, timeout)

					mu.Lock()
					idx := resolverMap[resolver]
					if err != nil {
						results[idx].Result.Error[qType] = err.Error()
					} else {
						setRecord(&results[idx].Result, qType, recs)
					}
					mu.Unlock()
				}()
			}
		}
	}

	// 等待所有任务完成
	wg.Wait()
	return results
}
