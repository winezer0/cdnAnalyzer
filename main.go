package main

import (
	"cdnCheck/dnsquery"
	"cdnCheck/fileutils"
	"cdnCheck/maputils"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func main() {
	//TODO 实现命令行调用
	//TODO 实现功能整合
	//1. 域名DNS解析
	//2. 域名EDNS解析
	//3. 对解析结果进行IP locate调用 查询IPv4 、IPv6 、ASN号
	//4. 加载新的cdn检查文件 对 域名、Cnames、IPv4、IPv6、ASN进行 判断

	domain := "www.example.com"
	timeout := time.Second * 5

	resolversFile := "asset/resolvers.txt"
	resolvers, err := fileutils.ReadTextToList(resolversFile)
	fmt.Printf("load resolvers: %v\n", len(resolvers))
	if err != nil {
		fmt.Println("加载DNS服务器失败:", err)
		os.Exit(1)
	}

	//进行常规 DNS 信息解析查询
	dnsResults := dnsquery.QueryAllDNSWithMultiResolvers(domain, resolvers, 5, timeout)
	fmt.Printf("DNS dnsResults: %v\n", dnsResults)
	dnsQueryResult := dnsquery.MergeDNSResults(dnsResults)
	b, _ := json.MarshalIndent(dnsQueryResult, "", "  ")
	fmt.Println(string(b))

	//进行 EDNS 信息查询
	cityMapFile := "asset/city_ip.csv"
	cityMap, err := fileutils.ReadCSVToMap(cityMapFile)
	if err != nil {

		fmt.Errorf("Failed to read cityMap CSV: %v\n", err)
	}
	// 随机选择 5 个城市
	randCities := maputils.PickRandMaps(cityMap, 5)
	// 随机选择 几个 城市
	if len(randCities) == 0 {
		fmt.Println("No cities selected, check cityMap or getRandMaps")
	} else {
		fmt.Printf("randCities: %v\n", randCities)
	}

	eDNSQueryResults := dnsquery.EDNSQueryWithMultiCities(domain, timeout, randCities, true)
	if len(eDNSQueryResults) == 0 {
		fmt.Errorf("Expected non-empty EDNS QueryResults\n")
	}

	for location, records := range eDNSQueryResults {
		fmt.Printf("[%s] => %v\n", location, records)
	}

	for location, result := range eDNSQueryResults {
		fmt.Printf("[%s]:\n", location)
		fmt.Printf("  A:    %v\n", result.A)
		fmt.Printf("  CNAME: %v\n", result.CNAME)
		fmt.Printf("  Errors: %v\n", result.Errors)
	}

	eDNSQueryResult := dnsquery.MergeEDNSResults(eDNSQueryResults)
	fmt.Printf("  A:    %v\n", eDNSQueryResult.A)
	fmt.Printf("  CNAME: %v\n", eDNSQueryResult.CNAME)
	fmt.Printf("  Errors: %v\n", eDNSQueryResult.Errors)

	dnsQueryResult = dnsquery.MergeEDNSToDNS(eDNSQueryResult, dnsQueryResult)
	dnsquery.OptimizeDNSResult(&dnsQueryResult)
	finalInfo, _ := json.MarshalIndent(dnsQueryResult, "", "  ")
	fmt.Println(string(finalInfo))
}
