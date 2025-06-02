package dns_query

import (
	"testing"
	"time"
)

func TestDoEcsQueryWithMultiCities(t *testing.T) {
	domain := "www.baidu.com"

	ecsQueryResults := DoEcsQueryWithMultiCities(domain, 30*time.Second, 3, true)

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
}
