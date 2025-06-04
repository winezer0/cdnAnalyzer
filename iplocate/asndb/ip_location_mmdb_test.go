package asndb

import (
	"fmt"
	"github.com/oschwald/maxminddb-golang"
	"net"
	"os"
	"testing"
)

func TestMMDB_ASN_Lookup(t *testing.T) {
	// 打开数据库连接
	ipv4_filePath := "C:\\Users\\WINDOWS\\Downloads\\geolite2-asn-ipv4.mmdb"
	ipv6_filePath := "C:\\Users\\WINDOWS\\Downloads\\geolite2-asn-ipv6.mmdb"

	if _, err := os.Stat(ipv4_filePath); err == nil {
		connectionId := "ipv" + "4"
		_, ok := mmDb[connectionId]
		if !ok {
			fmt.Println("Opening MMDB file: " + ipv4_filePath)
			conn, err := maxminddb.Open(ipv4_filePath)
			if err != nil {
				panic(err)
			}

			mmDb[connectionId] = conn
		}
	}

	if _, err := os.Stat(ipv6_filePath); err == nil {
		connectionId := "ipv" + "6"
		_, ok := mmDb[connectionId]
		if !ok {
			fmt.Println("Opening MMDB file: " + ipv6_filePath)
			conn, err := maxminddb.Open(ipv6_filePath)
			if err != nil {
				panic(err)
			}

			mmDb[connectionId] = conn
		}
	}
	defer mmdbClose()

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

		ipInfo := mmdbIp(ip)
		if ipInfo == nil {
			t.Errorf("无法解析IP信息: %s", ipStr)
			continue
		}

		fmt.Printf("IP: %15s | ASN: %6d | 组织: %s\n",
			ipStr,
			ipInfo.OrganisationNumber,
			ipInfo.OrganisationName,
		)
	}
}
