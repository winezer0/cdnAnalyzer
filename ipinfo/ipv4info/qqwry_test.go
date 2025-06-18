package ipv4info

import (
	"testing"
)

func TestIpv4LocationFind(t *testing.T) {
	// 集成测试：测试完整的查询流程
	db, err := NewIPv4Location("C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\qqwry.dat")
	if err != nil {
		t.Skipf("跳过集成测试，因为无法加载数据库: %v", err)
	}

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

			// 验证方法正常工作（不强制要求特定结果）
			// 主要测试方法不会panic或返回异常
		})
	}
}
