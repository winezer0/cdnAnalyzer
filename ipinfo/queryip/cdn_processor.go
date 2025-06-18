package queryip

import (
	"cdnCheck/cdncheck"
	"cdnCheck/ipinfo/asninfo"
	"cdnCheck/maputils"
	"cdnCheck/models"
	"fmt"
)

// CDNProcessor CDN信息处理器
type CDNProcessor struct{}

// NewCDNProcessor 创建新的CDN处理器
func NewCDNProcessor() *CDNProcessor {
	return &CDNProcessor{}
}

// ProcessCDNInfo 处理CDN相关信息
func (p *CDNProcessor) ProcessCDNInfo(checkInfo *models.CheckInfo, sourceData *models.CDNData) error {
	cnameList := checkInfo.CNAME
	// 合并 A 和 AAAA 记录
	ipList := maputils.UniqueMergeSlices(checkInfo.A, checkInfo.AAAA)
	// 合并 asn记录列表
	asnList := maputils.UniqueMergeAnySlices(
		asninfo.GetUniqueOrgNumbers(checkInfo.Ipv4Asn),
		asninfo.GetUniqueOrgNumbers(checkInfo.Ipv6Asn),
	)
	// 合并 IP定位列表
	ipLocateList := maputils.UniqueMergeSlices(
		maputils.GetMapsUniqueValues(checkInfo.Ipv4Locate),
		maputils.GetMapsUniqueValues(checkInfo.Ipv6Locate),
	)

	// 判断IP解析结果数量是否在CDN内
	checkInfo.IpSizeIsCdn, checkInfo.IpSize = cdncheck.IpsSizeIsCdn(ipList, 3)

	// 检查结果中的 CDN | WAF | CLOUD 情况
	checkInfo.IsCdn, checkInfo.CdnCompany = cdncheck.CheckCategory(sourceData.CDN, ipList, asnList, cnameList, ipLocateList)
	checkInfo.IsWaf, checkInfo.WafCompany = cdncheck.CheckCategory(sourceData.WAF, ipList, asnList, cnameList, ipLocateList)
	checkInfo.IsCloud, checkInfo.CloudCompany = cdncheck.CheckCategory(sourceData.CLOUD, ipList, asnList, cnameList, ipLocateList)

	return nil
}

// ProcessCDNInfos 批量处理CDN信息
func (p *CDNProcessor) ProcessCDNInfos(checkInfos []*models.CheckInfo, sourceData *models.CDNData) error {
	for _, checkInfo := range checkInfos {
		if err := p.ProcessCDNInfo(checkInfo, sourceData); err != nil {
			fmt.Printf("处理CDN信息失败: %v\n", err)
			continue
		}
	}
	return nil
}
