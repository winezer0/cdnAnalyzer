package analyzer

import (
	"cdnAnalyzer/pkg/ipinfo/asninfo"
)

// CheckInfo 用于保存资产结果时间的类型
type CheckInfo struct {
	RAW     string `json:"raw"`     // 存储原始输入信息
	FMT     string `json:"fmt"`     // 存储格式化后的输入信息（可选）
	IsIpv4  bool   `json:"isIpv4"`  // 存储格式化后的输入信息（可选）
	FromUrl bool   `json:"fromUrl"` // 存储格式化后的输入信息（可选）

	A     []string `json:"A"`     // A记录
	AAAA  []string `json:"AAAA"`  // AAAA记录
	CNAME []string `json:"CNAME"` // CNAME记录
	NS    []string `json:"NS"`    // NS记录
	MX    []string `json:"MX"`    // MX记录
	TXT   []string `json:"TXT"`   // TXT记录

	Ipv4Locate []map[string]string `json:"Ipv4Locate"` // A记录的IP解析信息
	Ipv6Locate []map[string]string `json:"Ipv6Locate"` // AAAA记录的IP解析信息

	Ipv4Asn []asninfo.ASNInfo `json:"Ipv4Asn"` // A记录的ASN查询信息
	Ipv6Asn []asninfo.ASNInfo `json:"Ipv6Asn"` // AAAA记录的ASN查询信息

	IsCdn      bool   `json:"IsCdn"`
	CdnCompany string `json:"CdnCompany"`

	IsWaf      bool   `json:"IsWaf"`
	WafCompany string `json:"WafCompany"`

	IsCloud      bool   `json:"IsCloud"`
	CloudCompany string `json:"CloudCompany"`

	IpSize      int  `json:"IpSize"`
	IpSizeIsCdn bool `json:"IpSizeIsCdn"`
}

// NewDomainCheckInfo 初始化一个新的 CheckInfo 实例
func NewDomainCheckInfo(raw, fmt string, fromUrl bool) *CheckInfo {
	return &CheckInfo{
		RAW:     raw,
		FMT:     fmt,
		FromUrl: fromUrl,
	}
}

// NewIPCheckInfo 初始化一个新的 CheckInfo 实例
func NewIPCheckInfo(raw, fmt string, isIpv4 bool, fromUrl bool) *CheckInfo {
	result := &CheckInfo{
		RAW:     raw,
		FMT:     fmt,
		IsIpv4:  isIpv4,
		FromUrl: fromUrl,
	}

	if isIpv4 {
		result.A = []string{fmt}
	} else {
		result.AAAA = []string{fmt}
	}

	return result
}
