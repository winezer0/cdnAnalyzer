package ipv6info

import (
	"testing"
)

func TestIpv6Location_Find(t *testing.T) {
	// 集成测试：测试完整的查询流程
	db, err := NewIPv6Location("C:\\Users\\WINDOWS\\Desktop\\cdnAnalyzer\\asset\\zxipv6wry.db")
	if err != nil {
		t.Skipf("跳过测试，因为无法加载数据库: %v", err)
	}
	defer db.Close()

	// 测试一些常见的IPv6地址
	testIPs := []string{
		"2001:db8::1",
		"2405:6f00:c602::1",
		"2409:8c1e:75b0:1120::27",
		"2402:3c00:1000:4::1",
		"2408:8652:200::c101",
		"2409:8900:103f:14f:d7e:cd36:11af:be83",
		"fe80::5c12:27dc:93a4:3426",
		"::1",
	}

	for _, ip := range testIPs {
		t.Run(ip, func(t *testing.T) {
			result := db.Find(ip)

			// 记录结果用于调试
			t.Logf("查询IP: %s -> 结果: %s", ip, result)

			// 验证方法正常工作（不强制要求特定结果）
			// 主要测试方法不会panic或返回异常
		})
	}
}

func TestIpv6Location_BatchFind(t *testing.T) {
	db, err := NewIPv6Location("C:\\Users\\WINDOWS\\Desktop\\cdnAnalyzer\\asset\\zxipv6wry.db")
	if err != nil {
		t.Skipf("跳过测试，因为无法加载数据库: %v", err)
	}
	defer db.Close()

	testIPs := []string{
		"2001:db8::1",
		"2405:6f00:c602::1",
		"2409:8c1e:75b0:1120::27",
		"invalid_ip",
		"192.168.1.1",
	}

	results := db.BatchFind(testIPs)

	if len(results) != len(testIPs) {
		t.Errorf("期望结果数量为 %d，但得到了 %d", len(testIPs), len(results))
	}

	for ip, result := range results {
		t.Logf("批量查询 - IP: %s -> 结果: %s", ip, result)
	}
}

func TestIpv6Location_GetDatabaseInfo(t *testing.T) {
	db, err := NewIPv6Location("C:\\Users\\WINDOWS\\Desktop\\cdnAnalyzer\\asset\\zxipv6wry.db")
	if err != nil {
		t.Skipf("跳过测试，因为无法加载数据库: %v", err)
	}
	defer db.Close()

	info := db.GetDatabaseInfo()

	// 验证返回的信息包含必要的字段
	requiredFields := []string{"ip_count", "index_start", "index_end", "data_size", "off_len", "ip_len"}
	for _, field := range requiredFields {
		if _, exists := info[field]; !exists {
			t.Errorf("数据库信息缺少字段: %s", field)
		}
	}

	t.Logf("数据库信息: %+v", info)
}
