package cdncheck

import (
	"cdnCheck/pkg/ipinfo/asninfo"
	"cdnCheck/pkg/maputils"
	"cdnCheck/pkg/models"
	"github.com/yl2chen/cidranger"
	"net"
	"strconv"
	"strings"
	"sync"
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
func keysInMap(cnames []string, cnamesMap map[string][]string) (bool, string) {
	for _, cname := range cnames {
		for companyName, cdnCNames := range cnamesMap {
			if ok := containKeys(cname, cdnCNames); ok {
				return true, companyName
			}
		}
	}
	return false, ""
}

// buildIpRanger 构建一个 CIDR 查找器（ranger）用于快速判断 IP 是否在某组 CIDR 范围内
func buildIpRanger(cidrList []string) cidranger.Ranger {
	ranger := cidranger.NewPCTrieRanger()
	for _, cdn := range cidrList {
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

type CheckResult struct {
	RAW          string `json:"raw"`
	FMT          string `json:"fmt"`
	IsCdn        bool   `json:"is_cdn"`
	CdnCompany   string `json:"cdn_company"`
	IsWaf        bool   `json:"is_waf"`
	WafCompany   string `json:"waf_company"`
	IsCloud      bool   `json:"is_cloud"`
	CloudCompany string `json:"cloud_company"`
	IpSizeIsCdn  bool   `json:"ip_size_is_cdn"`
	IpSize       int    `json:"ip_size"`
}

func checkCDN(cdnData *CDNData, checkInfo *models.CheckInfo) (CheckResult, error) {
	checkResult := CheckResult{
		RAW: checkInfo.RAW,
		FMT: checkInfo.FMT,
	}

	cnameList := checkInfo.CNAME
	ipList := maputils.UniqueMergeSlices(checkInfo.A, checkInfo.AAAA)
	asnList := maputils.UniqueMergeAnySlices(
		asninfo.GetUniqueOrgNumbers(checkInfo.Ipv4Asn),
		asninfo.GetUniqueOrgNumbers(checkInfo.Ipv6Asn),
	)
	ipLocateList := maputils.UniqueMergeSlices(
		maputils.GetMapsValuesUnique(checkInfo.Ipv4Locate),
		maputils.GetMapsValuesUnique(checkInfo.Ipv6Locate),
	)

	// 判断 IP 数量是否符合 CDN 特征
	checkResult.IpSizeIsCdn, checkResult.IpSize = IpsSizeIsCdn(ipList, 3)

	// 封装分类检查逻辑
	type categoryHandler struct {
		db    Category
		setFn func(bool, string)
	}

	categories := []categoryHandler{
		{
			db:    cdnData.CDN,
			setFn: func(b bool, s string) { checkResult.IsCdn, checkResult.CdnCompany = b, s },
		},
		{
			db:    cdnData.WAF,
			setFn: func(b bool, s string) { checkResult.IsWaf, checkResult.WafCompany = b, s },
		},
		{
			db:    cdnData.CLOUD,
			setFn: func(b bool, s string) { checkResult.IsCloud, checkResult.CloudCompany = b, s },
		},
	}

	for _, cat := range categories {
		match, company := CheckCategory(cat.db, ipList, asnList, cnameList, ipLocateList)
		cat.setFn(match, company)
	}

	return checkResult, nil
}

func CheckCDNBatch(cdnData *CDNData, checkInfos []*models.CheckInfo) ([]CheckResult, error) {
	checkResults := make([]CheckResult, len(checkInfos))
	var wg sync.WaitGroup
	var mu sync.Mutex // 如果需要保护共享资源

	for i, checkInfo := range checkInfos {
		wg.Add(1)
		go func(i int, checkInfo *models.CheckInfo) {
			defer wg.Done()
			res, _ := checkCDN(cdnData, checkInfo)
			mu.Lock()
			checkResults[i] = res
			mu.Unlock()
		}(i, checkInfo)
	}

	wg.Wait()
	return checkResults, nil
}
