package dnsquery

import (
	"cdnCheck/filetools"
	"fmt"
	"testing"
	"time"
)

func TestEDNSQuery(t *testing.T) {
	domain := "www.baidu.com" // 替换为你要测试的域名
	eDNSQueryResults := EDNSQuery(domain, "175.1.238.1", "8.8.8.8:53", 5*time.Second)

	t.Logf("Query results for %s:", domain)
	t.Logf("  IPs:    %v", eDNSQueryResults.IPs)
	t.Logf("  CNAMEs: %v", eDNSQueryResults.CNAMEs)
	t.Logf("  Errors: %v", eDNSQueryResults.Errors)
}

func TestEDNSQueryWithMultiCities(t *testing.T) {
	domain := "www.example.com" // 替换为你要测试的域名

	var filename = "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\city_ip.csv"
	cityMap, err := filetools.ReadCSVToMap(filename)
	if err != nil {
		t.Errorf("Failed to read CSV: %v", err)
	}

	// 随机选择 5 个城市
	randCities := filetools.PickRandMaps(cityMap, 5)
	// 随机选择 几个 城市
	if len(randCities) == 0 {
		fmt.Println("No cities selected, check cityMap or getRandMaps")
	} else {
		t.Logf("randCities: %v", randCities)
	}

	eDNSQueryResults := EDNSQueryWithMultiCities(domain, 5*time.Second, randCities, true)

	if len(eDNSQueryResults) == 0 {
		t.Fatal("Expected non-empty ENS QueryResults")
	}

	t.Logf("Query EDNS QueryResults for %s:", domain)
	for location, records := range eDNSQueryResults {
		t.Logf("[%s] => %v", location, records)
	}

	t.Logf("Query results for %s:", domain)
	for location, result := range eDNSQueryResults {
		t.Logf("[%s]:", location)
		t.Logf("  IPs:    %v", result.IPs)
		t.Logf("  CNAMEs: %v", result.CNAMEs)
		t.Logf("  Errors: %v", result.Errors)
	}
}
