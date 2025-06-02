package dns_query

import (
	"fmt"
	"testing"
	"time"
)

func TestDoEcsQueryWithMultiCities(t *testing.T) {
	domain := "www.baidu.com" // 替换为你要测试的域名

	ecsQueryResults := DoEcsQueryWithMultiCities(domain, 3*time.Second)

	if len(ecsQueryResults) == 0 {
		t.Fatal("Expected non-empty ecsQueryResults")
	}

	t.Logf("Query ecsQueryResults for %s:", domain)
	for location, records := range ecsQueryResults {
		t.Logf("[%s] => %v", location, records)
	}

	uniqueAddrs := DedupEcsResults(ecsQueryResults)
	fmt.Println("Unique addresses:")
	for _, addr := range uniqueAddrs {
		fmt.Println(addr)
	}
	if len(uniqueAddrs) == 0 {
		t.Fatal("Expected non-empty uniqueAddrs")
	}
}
