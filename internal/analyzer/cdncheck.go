package analyzer

import (
	"github.com/winezer0/cdnAnalyzer/pkg/ipinfo/asninfo"
	"github.com/winezer0/cdnAnalyzer/pkg/maputils"
	"sync"
)

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

func checkCDN(cdnData *CDNData, checkInfo *CheckInfo) (CheckResult, error) {
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

func CheckCDNBatch(cdnData *CDNData, checkInfos []*CheckInfo) ([]CheckResult, error) {
	checkResults := make([]CheckResult, len(checkInfos))
	var wg sync.WaitGroup
	var mu sync.Mutex // 如果需要保护共享资源

	for i, checkInfo := range checkInfos {
		wg.Add(1)
		go func(i int, checkInfo *CheckInfo) {
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
