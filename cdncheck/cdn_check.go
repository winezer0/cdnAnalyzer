package cdncheck

import (
	"github.com/yl2chen/cidranger"
	"net"
	"strconv"
	"strings"
)

// isContainKeys 检查传入的 字符串是否包含任一子字符串
func isContainKeys(str string, keys []string) bool {
	str = strings.Trim(str, ".")
	for _, cn := range keys {
		if strings.Contains(strings.ToLower(str), strings.ToLower(cn)) {
			return true
		}
	}
	return false
}

// KeysInMap 检查一组 CNAME 是否命中某个 CDN 厂商，并复用 isContainKeys 实现
func KeysInMap(cnames []string, cdnCNAMEsMap map[string][]string) (bool, string) {
	for _, cname := range cnames {
		for companyName, cdnCNames := range cdnCNAMEsMap {
			if ok := isContainKeys(cname, cdnCNames); ok {
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

// IpsInMap 检查多个 IP 是否命中某个 CDN 厂商，并返回厂商名称
func IpsInMap(ips []string, cdnIPsMap map[string][]string) (bool, string) {
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

// ASNsInMap 检查多个 ASN 是否属于某个 CDN 厂商，并返回厂商名称
func ASNsInMap(asns []uint64, cdnASNsMap map[string][]string) (bool, string) {
	for _, asn := range asns {
		for companyName, cdnASNs := range cdnASNsMap {
			if asnInCdn(asn, cdnASNs) {
				return true, companyName
			}
		}
	}
	return false, ""
}

// IpsSizeIsCdn 检查多个 IP 是否命中某个 CDN 厂商，并返回厂商名称
func IpsSizeIsCdn(ips []string, limitSize int) (bool, int) {
	size := len(ips)
	return size > limitSize, size
}
