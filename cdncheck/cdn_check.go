package cdncheck

import (
	"cdnCheck/models"
	"github.com/yl2chen/cidranger"
	"net"
	"strconv"
	"strings"
)

// containKeys 检查传入的 字符串是否包含任一子字符串
func containKeys(str string, keys []string) bool {
	str = strings.Trim(str, ".")
	for _, cn := range keys {
		if strings.Contains(strings.ToLower(str), strings.ToLower(cn)) {
			return true
		}
	}
	return false
}

// keysInMap 检查一组 CNAME 是否命中某个 CDN 厂商，并复用 containKeys 实现
func keysInMap(cnames []string, cdnCNAMEsMap map[string][]string) (bool, string) {
	for _, cname := range cnames {
		for companyName, cdnCNames := range cdnCNAMEsMap {
			if ok := containKeys(cname, cdnCNames); ok {
				return true, companyName
			}
		}
	}
	return false, ""
}

// buildIpRanger 构建一个 CIDR 查找器（ranger）用于快速判断 IP 是否在某组 CIDR 范围内
func buildIpRanger(cdnIPs []string) cidranger.Ranger {
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

// ipInRanger 使用预构建的 CIDR 查找器来判断 IP 是否属于该 CDN 厂商的网段
func ipInRanger(ip string, ranger cidranger.Ranger) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	contains, err := ranger.Contains(parsedIP)
	return err == nil && contains
}

// ipsInMap 检查多个 IP 是否命中某个 CDN 厂商，并返回厂商名称
func ipsInMap(ips []string, cdnIPsMap map[string][]string) (bool, string) {
	for companyName, cdnIPs := range cdnIPsMap {
		ranger := buildIpRanger(cdnIPs) // 为每个厂商只构建一次 CIDR 查找器
		for _, ip := range ips {
			if ipInRanger(ip, ranger) {
				return true, companyName
			}
		}
	}
	return false, ""
}

// asnInList 检查传入的 uint64 格式的 ASN 号是否在 CDN 厂商的 ASN 列表中
func asnInList(asn uint64, cndASNs []string) bool {
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

// asnsInMap 检查多个 ASN 是否属于某个 CDN 厂商，并返回厂商名称
func asnsInMap(asns []uint64, cdnASNsMap map[string][]string) (bool, string) {
	for _, asn := range asns {
		for companyName, cdnASNs := range cdnASNsMap {
			if asnInList(asn, cdnASNs) {
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

func CheckCategory(
	categoryMap models.Category,
	cnames []string,
	ips []string,
	asns []uint64,
	ipLocates []string,
) (bool, string) {

	// 定义检查项：每个检查项是一个函数调用
	checks := []func() (bool, string){
		func() (bool, string) { return keysInMap(cnames, categoryMap.CNAME) },
		func() (bool, string) { return keysInMap(ipLocates, categoryMap.KEYS) },
		func() (bool, string) { return asnsInMap(asns, categoryMap.ASN) },
		func() (bool, string) { return ipsInMap(ips, categoryMap.IP) },
	}

	// 依次执行检查
	for _, check := range checks {
		if ok, company := check(); ok {
			return ok, company
		}
	}

	return false, ""
}
