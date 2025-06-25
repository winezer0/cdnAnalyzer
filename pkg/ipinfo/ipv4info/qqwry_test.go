package ipv4info

import (
	"testing"
)

func TestIpv4Location_Find(t *testing.T) {
	// 集成测试：测试完整的查询流程
	db, err := NewIPv4Location("C:\\Users\\WINDOWS\\Desktop\\cdnAnalyzer\\asset\\qqwry.dat")
	if err != nil {
		t.Skipf("跳过测试，因为无法加载数据库: %v", err)
	}
	defer db.Close()

	// 测试一些常见的IP地址
	testIPs := []string{
		"8.8.8.8",
		"119.29.29.52",
		"114.114.114.114",
		"223.5.5.5",
		"1.1.1.1",
		"208.67.222.222",
		"266.67.222.222",
	}

	for _, ip := range testIPs {
		t.Run(ip, func(t *testing.T) {
			result := db.Find(ip)

			// 记录结果用于调试
			t.Logf("查询IP: %s -> 结果: %s", ip, result)
		})
	}
}

func TestIpv4Location_BatchFind(t *testing.T) {
	db, err := NewIPv4Location("C:\\Users\\WINDOWS\\Desktop\\cdnAnalyzer\\asset\\qqwry.dat")
	if err != nil {
		t.Skipf("跳过测试，因为无法加载数据库: %v", err)
	}
	defer db.Close()

	testIPs := []string{
		"8.8.8.8",
		"119.29.29.52",
		"114.114.114.114",
		"invalid_ip",
		"2001:db8::1",
	}

	results := db.BatchFind(testIPs)

	if len(results) != len(testIPs) {
		t.Errorf("期望结果数量为 %d，但得到了 %d", len(testIPs), len(results))
	}

	for ip, result := range results {
		t.Logf("批量查询 - IP: %s -> 结果: %s", ip, result)
	}
}

func TestIpv4Location_GetDatabaseInfo(t *testing.T) {
	db, err := NewIPv4Location("C:\\Users\\WINDOWS\\Desktop\\cdnAnalyzer\\asset\\qqwry.dat")
	if err != nil {
		t.Skipf("跳过测试，因为无法加载数据库: %v", err)
	}
	defer db.Close()

	info := db.GetDatabaseInfo()

	// 验证返回的信息包含必要的字段
	requiredFields := []string{"ip_count", "index_start", "index_end", "data_size"}
	for _, field := range requiredFields {
		if _, exists := info[field]; !exists {
			t.Errorf("数据库信息缺少字段: %s", field)
		}
	}

	t.Logf("数据库信息: %+v", info)
}
