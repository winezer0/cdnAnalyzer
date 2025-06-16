package cdncheck

import (
	"github.com/yl2chen/cidranger"
	"net"
	"strconv"
	"strings"
)

// cnameInCdn 检查传入的 CNAME 字符串是否包含任一 CDN 厂商的 CNAME
func cnameInCdn(cname string, cdnCNAMEs []string) bool {
	cname = strings.Trim(cname, ".")
	for _, cn := range cdnCNAMEs {
		if strings.Contains(strings.ToLower(cname), strings.ToLower(cn)) {
			return true
		}
	}
	return false
}

// CnamesInCdnMap 检查一组 CNAME 是否命中某个 CDN 厂商，并复用 cnameInCdn 实现
func CnamesInCdnMap(cnames []string, cdnCNAMEsMap map[string][]string) (bool, string) {
	for _, cname := range cnames {
		for companyName, cdnCNames := range cdnCNAMEsMap {
			if ok := cnameInCdn(cname, cdnCNames); ok {
				return true, companyName
			}
		}
	}
	return false, ""
}

// buildRanger 构建一个 CIDR 查找器（ranger）用于快速判断 IP 是否在某组 CIDR 范围内
func buildRanger(cdnIPs []string) cidranger.Ranger {
	ranger := cidranger.NewPCTrieRanger()
	for _, cdn := range cdnIPs {
		_, network, err := net.ParseCIDR(cdn)
		if err != nil {
			continue
		}
		_ = ranger.Insert(cidranger.NewBasicRangerEntry(*network))
	}
	return ranger
}

// ipInCdn 使用预构建的 CIDR 查找器来判断 IP 是否属于该 CDN 厂商的网段
func ipInCdn(ip string, ranger cidranger.Ranger) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	contains, err := ranger.Contains(parsedIP)
	return err == nil && contains
}

// IpsInCdnMap 检查多个 IP 是否命中某个 CDN 厂商，并返回厂商名称
func IpsInCdnMap(ips []string, cdnIPsMap map[string][]string) (bool, string) {
	for companyName, cdnIPs := range cdnIPsMap {
		ranger := buildRanger(cdnIPs) // 为每个厂商只构建一次 CIDR 查找器
		for _, ip := range ips {
			if ipInCdn(ip, ranger) {
				return true, companyName
			}
		}
	}
	return false, ""
}

// asnInCdn 检查传入的 uint64 格式的 ASN 号是否在 CDN 厂商的 ASN 列表中
func asnInCdn(asn uint64, cndASNs []string) bool {
	for _, asnStr := range cndASNs {
		_asnInt, err := strconv.Atoi(asnStr)
		if err != nil {
			continue
		}
		if uint64(_asnInt) == asn {
			return true
		}
	}
	return false
}

// ASNsInCdnMap 检查多个 ASN 是否属于某个 CDN 厂商，并返回厂商名称
func ASNsInCdnMap(asns []uint64, cdnASNsMap map[string][]string) (bool, string) {
	for _, asn := range asns {
		for companyName, cdnASNs := range cdnASNsMap {
			if asnInCdn(asn, cdnASNs) {
				return true, companyName
			}
		}
	}
	return false, ""
}

// ipLocateInCdn 检查传入的 IP定位 字符串是否包含任一 CDN 厂商的 IP归属地范围关键字中
func ipLocateInCdn(ipLocate string, cdnIpLocates []string) bool {
	ipLocate = strings.Trim(ipLocate, ".")
	for _, cn := range cdnIpLocates {
		if strings.Contains(strings.ToLower(ipLocate), strings.ToLower(cn)) {
			return true
		}
	}
	return false
}

// IpLocatesInCdn 检查一组 IP定位 是否命中某个 CDN 厂商
func IpLocatesInCdn(ipLocates []string, cdnIpLocatesMap map[string][]string) (bool, string) {
	for _, ipLocate := range ipLocates {
		for companyName, cdnIpLocates := range cdnIpLocatesMap {
			if ok := ipLocateInCdn(ipLocate, cdnIpLocates); ok {
				return true, companyName
			}
		}
	}
	return false, ""
}
