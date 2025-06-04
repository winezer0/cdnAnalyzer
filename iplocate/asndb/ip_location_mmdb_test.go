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
}
