package dnsquery

import (
	"cdnCheck/filetools"
	"fmt"
	"testing"
	"time"
)

func TestEDNSQuery(t *testing.T) {
	domain := "www.baidu.com" // 替换为你要测试的域名
	ecsQueryResults := EDNSQuery(domain, "175.1.238.1", "8.8.8.8:53", 5*time.Second)

	t.Logf("Query results for %s:", domain)
	t.Logf("  IPs:    %v", ecsQueryResults.IPs)
	t.Logf("  CNAMEs: %v", ecsQueryResults.CNAMEs)
	t.Logf("  Errors: %v", ecsQueryResults.Errors)
}

func TestEDNSQueryWithMultiCities(t *testing.T) {
	domain := "www.example.com" // 替换为你要测试的域名

	var filename = "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\city_ip.csv"
	cityMap, err := filetools.ReadCSVToMap(filename)
	if err != nil {
		t.Errorf("Failed to read CSV: %v", err)
	}

	// 随机选择 5 个城市
	randCities := filetools.GetRandMaps(cityMap, 5)
	// 随机选择 几个 城市
	if len(randCities) == 0 {
		fmt.Println("No cities selected, check cityMap or getRandMaps")
	} else {
		t.Logf("randCities: %v", randCities)
	}

	ecsQueryResults := EDNSQueryWithMultiCities(domain, 5*time.Second, randCities, true)

	if len(ecsQueryResults) == 0 {
		t.Fatal("Expected non-empty ecsQueryResults")
	}

	t.Logf("Query ecsQueryResults for %s:", domain)
	for location, records := range ecsQueryResults {
		t.Logf("[%s] => %v", location, records)
	}

	if len(ecsQueryResults) == 0 {
		t.Fatal("Expected non-empty ecsQueryResults")
	}

	t.Logf("Query results for %s:", domain)
	for location, result := range ecsQueryResults {
		t.Logf("[%s]:", location)
		t.Logf("  IPs:    %v", result.IPs)
		t.Logf("  CNAMEs: %v", result.CNAMEs)
		t.Logf("  Errors: %v", result.Errors)
	}

	//uniqueAddrs := DedupEcsResults(ecsQueryResults)
	//fmt.Println("Unique addresses:")
	//for _, addr := range uniqueAddrs {
	//	fmt.Println(addr)
	//}
	//if len(uniqueAddrs) == 0 {
	//	t.Fatal("Expected non-empty uniqueAddrs")
	//}
}
