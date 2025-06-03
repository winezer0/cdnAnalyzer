package main

import (
	"cdnCheck/dnsquery"
	"cdnCheck/filetools"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func main() {
	domain := "www.example.com"
	timeout := time.Second * 5

	resolversFile := "resolvers.txt"
	resolvers, err := filetools.ReadFileToList(resolversFile)
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
	cityMapFile := "city_ip.csv"
	cityMap, err := filetools.ReadCSVToMap(cityMapFile)
	if err != nil {

		fmt.Errorf("Failed to read cityMap CSV: %v\n", err)
	}
	// 随机选择 5 个城市
	randCities := filetools.PickRandMaps(cityMap, 5)
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
	finalInfo, _ := json.MarshalIndent(dnsQueryResult, "", "  ")
	fmt.Println(string(finalInfo))
}
