package queryip

import (
	"testing"
)

func TestQueryIP(t *testing.T) {
	// 测试配置
	config := &IpDbConfig{
		AsnIpv4Db:    "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\geolite2-asn-ipv4.mmdb",
		AsnIpv6Db:    "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\geolite2-asn-ipv6.mmdb",
		Ipv4LocateDb: "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\qqwry.dat",
		Ipv6LocateDb: "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\zxipv6wry.db",
	}

	// 初始化数据库引擎
	engines, err := InitDBEngines(config)
	if err != nil {
		t.Skipf("跳过测试：无法初始化数据库引擎: %v", err)
		return
	}

	// 创建IP处理器
	processor := NewIPProcessor(engines, config)
	defer processor.Close()

	// 测试单个IP查询
	t.Run("TestQuerySingleIP", func(t *testing.T) {
		// 测试IPv4
		location, asnInfo := processor.QuerySingleIP("1.1.1.1")
		t.Logf("IPv4查询结果 - 位置: %s, ASN: %+v", location, asnInfo)

		// 测试IPv6
		location6, asnInfo6 := processor.QuerySingleIP("2001:4860:4860::8888")
		t.Logf("IPv6查询结果 - 位置: %s, ASN: %+v", location6, asnInfo6)
	})

	// 测试批量IP查询
	t.Run("TestQueryIPInfo", func(t *testing.T) {
		ipv4s := []string{"8.8.8.8", "1.1.1.1"}
		ipv6s := []string{"2001:4860:4860::8888", "2606:4700:4700::1111"}

		ipInfo, err := processor.QueryIPInfo(ipv4s, ipv6s)
		if err != nil {
			t.Errorf("批量查询失败: %v", err)
			return
		}

		t.Logf("IPv4位置信息: %+v", ipInfo.IPv4Locations)
		t.Logf("IPv6位置信息: %+v", ipInfo.IPv6Locations)
		t.Logf("IPv4 ASN信息: %+v", ipInfo.IPv4AsnInfos)
		t.Logf("IPv6 ASN信息: %+v", ipInfo.IPv6AsnInfos)

		// 验证结果数量
		if len(ipInfo.IPv4Locations) != len(ipv4s) {
			t.Errorf("IPv4位置信息数量不匹配，期望: %d, 实际: %d", len(ipv4s), len(ipInfo.IPv4Locations))
		}

		if len(ipInfo.IPv6Locations) != len(ipv6s) {
			t.Errorf("IPv6位置信息数量不匹配，期望: %d, 实际: %d", len(ipv6s), len(ipInfo.IPv6Locations))
		}

		if len(ipInfo.IPv4AsnInfos) != len(ipv4s) {
			t.Errorf("IPv4 ASN信息数量不匹配，期望: %d, 实际: %d", len(ipv4s), len(ipInfo.IPv4AsnInfos))
		}

		if len(ipInfo.IPv6AsnInfos) != len(ipv6s) {
			t.Errorf("IPv6 ASN信息数量不匹配，期望: %d, 实际: %d", len(ipv6s), len(ipInfo.IPv6AsnInfos))
		}
	})
}
