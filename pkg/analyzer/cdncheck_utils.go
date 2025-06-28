package analyzer

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
