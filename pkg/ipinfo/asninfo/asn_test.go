package asninfo

import (
	"net"
	"testing"
)

func TestMMDBManager_FindASN(t *testing.T) {
	// 创建配置
	config := &MMDBConfig{
		IPv4Path:             "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\geolite2-asn-ipv4.mmdb",
		IPv6Path:             "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\geolite2-asn-ipv6.mmdb",
		MaxConcurrentQueries: 100,
	}

	// 创建管理器
	manager := NewMMDBManager(config)

	// 初始化连接
	if err := manager.InitMMDBConn(); err != nil {
		t.Fatalf("初始化数据库连接失败: %v", err)
	}
	defer manager.Close()

	// 测试数据库大小统计
	ipv4DbSize, err := manager.CountMMDBSize("ipv4")
	if err != nil {
		t.Errorf("统计IPv4数据库大小失败: %v", err)
	}
	t.Logf("IPv4数据库大小: %d", ipv4DbSize)

	ipv6DbSize, err := manager.CountMMDBSize("ipv6")
	if err != nil {
		t.Errorf("统计IPv6数据库大小失败: %v", err)
	}
	t.Logf("IPv6数据库大小: %d", ipv6DbSize)

	// 定义测试IP列表
	testIPs := []string{
		"8.8.8.8",         // Google DNS (IPv4)
		"2606:4700::6813", // Cloudflare (IPv6)
		"192.168.1.1",     // 内网地址
		"1.1.1.1",         // Cloudflare
		"116.162.1.1",     // Cloudflare
	}

	// 测试单个IP查询
	t.Run("单个IP查询测试", func(t *testing.T) {
		for _, ipStr := range testIPs {
			ip := net.ParseIP(ipStr)
			if ip == nil {
				t.Errorf("无效的IP地址: %s", ipStr)
				continue
			}

			ipInfo := manager.FindASN(ipStr)
			if ipInfo == nil {
				t.Errorf("无法解析IP信息: %s", ipStr)
				continue
			}

			t.Logf("单个IP查询结果:")
			PrintASNInfo(ipInfo)
		}
	})
}

func TestMMDBManager_ASNToIPRanges(t *testing.T) {
	// 创建配置
	config := &MMDBConfig{
		IPv4Path:             "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\geolite2-asn-ipv4.mmdb",
		IPv6Path:             "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\geolite2-asn-ipv6.mmdb",
		MaxConcurrentQueries: 100,
	}

	// 创建管理器
	manager := NewMMDBManager(config)

	// 初始化连接
	if err := manager.InitMMDBConn(); err != nil {
		t.Fatalf("初始化数据库连接失败: %v", err)
	}
	defer manager.Close()

	// 测试ASN到IP范围查询
	t.Run("ASN到IP范围查询测试", func(t *testing.T) {
		results, err := manager.ASNToIPRanges(13335)
		if err != nil {
			t.Errorf("ASN到IP范围查询失败: %v", err)
			return
		}
		t.Logf("找到 %d 个IP范围", len(results))
		for _, ipNet := range results {
			t.Logf("IP范围: %s", ipNet.String())
		}
	})
}

func TestMMDBManager_BatchFindASN(t *testing.T) {
	config := &MMDBConfig{
		IPv4Path:             "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\geolite2-asn-ipv4.mmdb",
		IPv6Path:             "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\geolite2-asn-ipv6.mmdb",
		MaxConcurrentQueries: 100,
	}

	manager := NewMMDBManager(config)

	// 初始化数据库连接
	if err := manager.InitMMDBConn(); err != nil {
		t.Skipf("跳过测试，因为无法加载数据库: %v", err)
	}
	defer manager.Close()

	// 验证数据库是否已初始化
	if !manager.IsInitialized() {
		t.Fatal("数据库未正确初始化")
	}

	testIPs := []string{
		"8.8.8.8",
		"1.1.1.1",
		"114.114.114.114",
		"2001:db8::1",
		"invalid_ip",
		"",
	}

	results := manager.BatchFindASN(testIPs)

	if len(results) != len(testIPs) {
		t.Errorf("期望结果数量为 %d，但得到了 %d", len(testIPs), len(results))
	}

	for i, result := range results {
		if result == nil {
			t.Errorf("索引 %d 的结果为 nil", i)
			continue
		}

		t.Logf("IP: %s, 版本: %d, 找到ASN: %v, ASN: %d, 组织: %s",
			result.IP, result.IPVersion, result.FoundASN, result.OrganisationNumber, result.OrganisationName)
	}
}
