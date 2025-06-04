package asndb

import (
	"fmt"
	"net"
	"testing"
)

func TestMMDB_ASN_Lookup(t *testing.T) {
	// 打开数据库连接
	ipv4Filepath := "C:\\Users\\WINDOWS\\Downloads\\geolite2-asn-ipv4.mmdb"
	ipv6Filepath := "C:\\Users\\WINDOWS\\Downloads\\geolite2-asn-ipv6.mmdb"
	initMMDBConn(ipv4Filepath, ipv6Filepath)
	defer closeMMDBConn()

	ipv4dbsize, _ := CountMMDBSize(mmDb["ipv4"])
	t.Logf("CountMMDBSize ipv4: %d\n", ipv4dbsize)
	ipv6dbsize, _ := CountMMDBSize(mmDb["ipv6"])
	t.Logf("CountMMDBSize ipv6: %d\n", ipv6dbsize)

	// 定义要测试的 IPs
	testIPs := []string{
		"8.8.8.8",         // Google DNS (IPv4)
		"2606:4700::6813", // Cloudflare (IPv6)
		"192.168.1.1",     // 内网地址
		"1.1.1.1",         // Cloudflare
	}

	for _, ipStr := range testIPs {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			t.Errorf("无效的IP地址: %s", ipStr)
			continue
		}

		ipInfo := findASN(ip)
		if ipInfo == nil {
			t.Errorf("无法解析IP信息: %s", ipStr)
			continue
		}
		fmt.Printf("ipInfo: %v\n", ipInfo)
		fmt.Printf("IP: %15s | ASN: %6d | 组织: %s\n",
			ipStr,
			ipInfo.OrganisationNumber,
			ipInfo.OrganisationName,
		)
	}

	results, err := ASNToIPRanges(13335)
	t.Logf("ASNToIPRanges results: %v Error:%v", len(results), err)
}

func TestMMDB_To_CSV(t *testing.T) {
	// 打开数据库连接
	ipv4Filepath := "C:\\Users\\WINDOWS\\Downloads\\geolite2-asn-ipv4.mmdb"
	ipv6Filepath := "C:\\Users\\WINDOWS\\Downloads\\geolite2-asn-ipv6.mmdb"
	initMMDBConn(ipv4Filepath, ipv6Filepath)
	defer closeMMDBConn()

	outputPath := "C:\\Users\\WINDOWS\\Downloads\\geolite2-asn-all.csv"
	ExportASNToCSV(outputPath)
}

func TestFirstASNToIPRanges(t *testing.T) {
	// 打开数据库连接
	ipv4Filepath := "C:\\Users\\WINDOWS\\Downloads\\geolite2-asn-ipv4.mmdb"
	ipv6Filepath := "C:\\Users\\WINDOWS\\Downloads\\geolite2-asn-ipv6.mmdb"
	initMMDBConn(ipv4Filepath, ipv6Filepath)
	defer closeMMDBConn()

	// 初始化缓存（只需调用一次）
	err := PreloadASNCache()
	if err != nil {
		fmt.Println("预加载缓存失败:", err)
		return
	}

	// 查询某个 ASN（例如 Google 的 ASN 15169）
	asn := uint64(13335)
	ipNets, found := FirstASNToIPRanges(asn)
	if !found {
		fmt.Printf("未找到 ASN %d 对应的 IP 段\n", asn)
		return
	}

	fmt.Printf("找到 ASN %d 的 %d 个 IP 段:\n", asn, len(ipNets))
	for _, ipNet := range ipNets {
		fmt.Println("  ", ipNet.String())
	}
}
