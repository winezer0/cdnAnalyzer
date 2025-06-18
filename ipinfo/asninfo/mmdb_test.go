package asninfo

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestLookupASNByMMDB(t *testing.T) {
	// 创建配置
	config := &MMDBConfig{
		IPv4Path:             "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\geolite2-asn-ipv4.mmdb",
		IPv6Path:             "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\geolite2-asn-ipv6.mmdb",
		MaxConcurrentQueries: 100,
		QueryTimeout:         5 * time.Second,
	}

	// 创建管理器
	manager := NewMMDBManager(config)

	// 初始化连接
	if err := manager.InitMMDBConn(); err != nil {
		t.Fatalf("初始化数据库连接失败: %v", err)
	}
	defer manager.CloseMMDBConn()

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

	// 测试批量查询
	t.Run("批量查询测试", func(t *testing.T) {
		results := manager.BatchFindASN(testIPs, nil)
		for _, result := range results {
			if result.Error != nil {
				t.Errorf("查询失败 %s: %v", result.IP, result.Error)
				continue
			}
			if result.Result != nil {
				t.Logf("批量查询结果:")
				PrintASNInfo(result.Result)
			}
		}
	})

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

func TestMMDBManagerConcurrency(t *testing.T) {
	config := &MMDBConfig{
		IPv4Path:             "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\geolite2-asn-ipv4.mmdb",
		IPv6Path:             "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\geolite2-asn-ipv6.mmdb",
		MaxConcurrentQueries: 5, // 限制并发数为5
		QueryTimeout:         2 * time.Second,
	}

	manager := NewMMDBManager(config)
	if err := manager.InitMMDBConn(); err != nil {
		t.Fatalf("初始化数据库连接失败: %v", err)
	}
	defer manager.CloseMMDBConn()

	// 创建大量测试IP
	var testIPs []string
	for i := 0; i < 100; i++ {
		testIPs = append(testIPs, fmt.Sprintf("8.8.8.%d", i))
	}

	// 测试并发查询
	start := time.Now()
	results := manager.BatchFindASN(testIPs, nil)
	duration := time.Since(start)

	t.Logf("并发查询 %d 个IP耗时: %v", len(testIPs), duration)

	successCount := 0
	for _, result := range results {
		if result.Error == nil && result.Result != nil {
			successCount++
		}
	}

	t.Logf("成功查询数: %d/%d", successCount, len(testIPs))
}

func TestMMDBErrorHandling(t *testing.T) {
	config := &MMDBConfig{
		IPv4Path:             "不存在的文件.mmdb",
		IPv6Path:             "不存在的文件.mmdb",
		MaxConcurrentQueries: 100,
		QueryTimeout:         5 * time.Second,
	}

	manager := NewMMDBManager(config)

	// 测试初始化错误
	err := manager.InitMMDBConn()
	if err == nil {
		t.Error("预期初始化失败，但未收到错误")
	} else {
		t.Logf("预期的初始化错误: %v", err)
	}

	// 测试无效IP查询
	ipInfo := manager.FindASN("invalid.ip.address")
	if ipInfo == nil {
		t.Error("无效IP查询应返回空结果而不是nil")
	}
	if ipInfo.FoundASN {
		t.Error("无效IP不应找到ASN信息")
	}
}
