package dnsquery

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

// GetSystemDefaultAddress 获取系统默认的本地dns服务器
func GetSystemDefaultAddress() (addr string) {
	resolver := net.Resolver{}
	resolver.PreferGo = true
	resolver.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		addr = address
		return nil, fmt.Errorf(address)
	}
	_, _ = resolver.LookupHost(context.Background(), "Addr")
	return addr
}

// NSServerAddPort 为dns服务器IP补充53端口
func NSServerAddPort(s string) string {
	s = strings.TrimSuffix(strings.ToLower(s), ".")
	if strings.Contains(s, ":") {
		return s
	}
	return s + ":53"
}

func HasPort(addr string) bool {
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

// PickRandomElements 随机选取n个不重复的DNS服务器
func PickRandomElements(resolvers []string, n int) []string {
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

// 获取随机城市
func getRandomCities(cityList []map[string]string, count int) []map[string]string {
	if len(cityList) <= count {
		return cityList
	}

	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 存储随机城市数据
	selectedCities := make([]map[string]string, 0, count)
	for i := 0; i < count; i++ {
		randIndex := rand.Intn(len(cityList))
		selectedCities = append(selectedCities, cityList[randIndex])
		// 通过切片操作移除指定索引位置元素,防止被多次选择
		cityList = append(cityList[:randIndex], cityList[randIndex+1:]...)
	}
	return selectedCities
}

// 合并去重字符串切片
func MergeUniqueStringSlices(slices ...[]string) []string {
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
		merged.A = MergeUniqueStringSlices(merged.A, r.A)
		merged.AAAA = MergeUniqueStringSlices(merged.AAAA, r.AAAA)
		merged.CNAME = MergeUniqueStringSlices(merged.CNAME, r.CNAME)
		merged.NS = MergeUniqueStringSlices(merged.NS, r.NS)
		merged.MX = MergeUniqueStringSlices(merged.MX, r.MX)
		merged.TXT = MergeUniqueStringSlices(merged.TXT, r.TXT)
	}
	return merged
}

// LoadFilesToList 读取本地文件
func LoadFilesToList(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			if !HasPort(line) {
				line += ":53"
			}
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

// DedupEcsResults 将多城市查询结果中的地址进行去重整合
func DedupEcsResults(results map[string][]string) []string {
	unique := make(map[string]struct{})
	var list []string

	for _, addresses := range results {
		for _, addr := range addresses {
			if _, exists := unique[addr]; !exists {
				unique[addr] = struct{}{}
				list = append(list, addr)
			}
		}
	}
	return list
}
