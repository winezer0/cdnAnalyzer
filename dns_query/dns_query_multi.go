package dns_query

import (
	"bufio"
	"math/rand"
	"os"
	"sync"
	"time"
)

// 读取本地 resolvers.txt，返回所有DNS服务器（加上:53端口）
func LoadResolvers(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var resolvers []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			if !hasPort(line) {
				line += ":53"
			}
			resolvers = append(resolvers, line)
		}
	}
	return resolvers, scanner.Err()
}

func hasPort(addr string) bool {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return true
		}
		if addr[i] < '0' || addr[i] > '9' {
			break
		}
	}
	return false
}

// 随机选取n个不重复的DNS服务器
func PickRandomResolvers(resolvers []string, n int) []string {
	if len(resolvers) <= n {
		return resolvers
	}
	rand.Seed(time.Now().UnixNano())
	perm := rand.Perm(len(resolvers))
	var picked []string
	for i := 0; i < n; i++ {
		picked = append(picked, resolvers[perm[i]])
	}
	return picked
}

// 合并去重字符串切片
func mergeUniqueStringSlices(slices ...[]string) []string {
	unique := make(map[string]struct{})
	for _, slice := range slices {
		for _, v := range slice {
			unique[v] = struct{}{}
		}
	}
	var result []string
	for v := range unique {
		result = append(result, v)
	}
	return result
}

// 合并多个DNSResult，去重
func MergeDNSResults(results []DNSResult) DNSResult {
	merged := DNSResult{}
	for _, r := range results {
		merged.A = mergeUniqueStringSlices(merged.A, r.A)
		merged.AAAA = mergeUniqueStringSlices(merged.AAAA, r.AAAA)
		merged.CNAME = mergeUniqueStringSlices(merged.CNAME, r.CNAME)
		merged.NS = mergeUniqueStringSlices(merged.NS, r.NS)
		merged.MX = mergeUniqueStringSlices(merged.MX, r.MX)
		merged.TXT = mergeUniqueStringSlices(merged.TXT, r.TXT)
	}
	return merged
}

// 并发查询，随机选5个DNS服务器
func QueryAllDNSWithMultiResolvers(domain string, resolvers []string, timeout time.Duration) DNSResult {
	const n = 5
	picked := PickRandomResolvers(resolvers, n)
	var wg sync.WaitGroup
	results := make([]DNSResult, n)
	wg.Add(n)
	for i, resolver := range picked {
		go func(i int, resolver string) {
			defer wg.Done()
			results[i] = QueryAllDNS(domain, resolver, timeout)
		}(i, resolver)
	}
	wg.Wait()
	return MergeDNSResults(results)
}

//// 示例main函数
//func main() {
//	domain := "example.com"
//	timeout := 3 * time.Second
//
//	resolvers, err := LoadResolvers("resolvers.txt")
//	if err != nil {
//		panic(err)
//	}
//
//	result := QueryAllDNSWithMultiResolvers(domain, resolvers, timeout)
//	b, _ := json.MarshalIndent(result, "", "  ")
//	os.Stdout.Write(b)
//}
