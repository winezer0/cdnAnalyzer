package analyzer

import (
	"github.com/yl2chen/cidranger"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// containsAny 检查字符串是否包含任一子字符串（忽略大小写） 注意：keys 必须已经是小写形式
func containsAny(str string, keys []string) bool {
	for _, key := range keys {
		if strings.Contains(str, key) {
			return true
		}
	}
	return false
}

var regexCache sync.Map
var regexMark = []string{"]", ")", "}", "*", "+", "^", "$", "?", "|", "\\"}

func containKeysSupportRegex(str string, keys []string) bool {
	str = strings.Trim(str, ".")
	strLower := strings.ToLower(str)

	for _, key := range keys {
		keyLower := strings.ToLower(key)

		if containsAny(keyLower, regexMark) {
			val, ok := regexCache.Load(key)
			if !ok {
				regex, err := regexp.Compile("(?i)" + key)
				if err != nil {
					if strings.Contains(strLower, keyLower) {
						return true
					}
					continue
				}
				val, _ = regexCache.LoadOrStore(key, regex)
			}

			if re, ok := val.(*regexp.Regexp); ok && re.MatchString(str) {
				return true
			}
		} else {
			if strings.Contains(strLower, keyLower) {
				return true
			}
		}
	}
	return false
}

// keysInMap 检查一组 CNAME 是否命中某个 CDN 厂商，并复用 containsAny 实现
func keysInMap(cnames []string, cnamesMap map[string][]string) (bool, string) {
	for _, cname := range cnames {
		for companyName, cdnCNames := range cnamesMap {
			if ok := containKeysSupportRegex(cname, cdnCNames); ok {
				return true, companyName
			}
		}
	}
	return false, ""
}

// buildIpRanger 构建一个 CIDR 查找器（ranger）用于快速判断 IP 是否在某组 CIDR 范围内
func buildIpRanger(cidrList []string) cidranger.Ranger {
	ranger := cidranger.NewPCTrieRanger()
	for _, cidr := range cidrList {
		_, network, err := net.ParseCIDR(cidr)
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
func ipsInMap(ips []string, ipsMap map[string][]string) (bool, string) {
	for companyName, cdnIPs := range ipsMap {
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
func asnInList(asn uint64, asnList []string) bool {
	for _, asnStr := range asnList {
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

// asnInMap 检查多个 ASN 是否属于某个 CDN 厂商，并返回厂商名称
func asnInMap(asns []uint64, cdnASNsMap map[string][]string) (bool, string) {
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

func CheckCategory(categoryMap Category, ipList []string, asnList []uint64, cnameList []string, ipLocateList []string) (bool, string) {
	// 定义检查项：每个检查项是一个函数调用
	checks := []func() (bool, string){
		func() (bool, string) { return keysInMap(cnameList, categoryMap.CNAME) },
		func() (bool, string) { return keysInMap(ipLocateList, categoryMap.KEYS) },
		func() (bool, string) { return asnInMap(asnList, categoryMap.ASN) },
		func() (bool, string) { return ipsInMap(ipList, categoryMap.IP) },
	}

	// 依次执行检查
	for _, check := range checks {
		if ok, company := check(); ok {
			return ok, company
		}
	}

	return false, ""
}

// GetFmtList 获取NoCDN && NoWAF的fmt数据
func GetFmtList(checkResults []CheckResult) []string {
	var fmtList []string
	for _, r := range checkResults {
		fmtList = append(fmtList, r.FMT)
	}
	return fmtList
}

// FilterNoCdnNoWaf 获取NoCDN && NoWAF的数据
func FilterNoCdnNoWaf(checkResults []CheckResult) []CheckResult {
	var nonCDNResult []CheckResult
	for _, r := range checkResults {
		if !r.IsCdn && !r.IsWaf && !r.IpSizeIsCdn {
			nonCDNResult = append(nonCDNResult, r)
		}
	}
	return nonCDNResult
}

// MergeCheckResultsToCheckInfos  通过 FMT 字段匹配对应条目将 checkResults 合并到 checkInfos
func MergeCheckResultsToCheckInfos(checkInfos []*CheckInfo, checkResults []CheckResult) []*CheckInfo {
	// 创建 FMT 到 CheckResult 的映射，便于快速查找
	resultMap := make(map[string]CheckResult)
	for _, result := range checkResults {
		resultMap[result.FMT] = result
	}

	// 遍历 checkInfos，将对应的 CheckResult 信息合并进去
	for _, checkInfo := range checkInfos {
		// 查找对应的 CheckResult
		if result, ok := resultMap[checkInfo.FMT]; ok {
			// 合并 CDN 相关信息
			checkInfo.IsCdn = result.IsCdn
			checkInfo.CdnCompany = result.CdnCompany

			// 合并 WAF 相关信息
			checkInfo.IsWaf = result.IsWaf
			checkInfo.WafCompany = result.WafCompany

			// 合并 Cloud 相关信息
			checkInfo.IsCloud = result.IsCloud
			checkInfo.CloudCompany = result.CloudCompany

			// 合并 IP 大小相关信息
			checkInfo.IpSizeIsCdn = result.IpSizeIsCdn
			checkInfo.IpSize = result.IpSize
		}
	}
	return checkInfos
}
