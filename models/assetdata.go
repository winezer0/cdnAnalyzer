package models

import (
	"cdnCheck/ipinfo/asninfo"
)

// CheckInfo 用于保存资产结果时间的类型
type CheckInfo struct {
	Raw     string `json:"raw"`               // 存储原始输入信息
	Fmt     string `json:"fmt,omitempty"`     // 存储格式化后的输入信息（可选）
	IsIpv4  bool   `json:"isIpv4,omitempty"`  // 存储格式化后的输入信息（可选）
	FromUrl bool   `json:"fromUrl,omitempty"` // 存储格式化后的输入信息（可选）

	A     []string `json:"A,omitempty"`     // A记录
	AAAA  []string `json:"AAAA,omitempty"`  // AAAA记录
	CNAME []string `json:"CNAME,omitempty"` // CNAME记录
	NS    []string `json:"NS,omitempty"`    // NS记录
	MX    []string `json:"MX,omitempty"`    // MX记录
	TXT   []string `json:"TXT,omitempty"`   // TXT记录

	Ipv4Locate []map[string]string `json:"Ipv4Locate,omitempty"` // A记录的IP解析信息
	Ipv6Locate []map[string]string `json:"Ipv6Locate,omitempty"` // AAAA记录的IP解析信息

	Ipv4Asn []asninfo.ASNInfo `json:"Ipv4Asn,omitempty"` // A记录的ASN查询信息
	Ipv6Asn []asninfo.ASNInfo `json:"Ipv6Asn,omitempty"` // AAAA记录的ASN查询信息

	IsCdn      bool   `json:"IsCdn,omitempty"`
	CdnCompany string `json:"CdnCompany,omitempty"`

	IsWaf      bool   `json:"IsWaf,omitempty"`
	WafCompany string `json:"WafCompany,omitempty"`

	IsCloud      bool   `json:"IsCloud,omitempty"`
	CloudCompany string `json:"CloudCompany,omitempty"`

	IpSize      int  `json:"IpSize,omitempty"`
	IpSizeIsCdn bool `json:"IpSizeIsCdn,omitempty"`
}

// NewDomainCheckInfo 初始化一个新的 CheckInfo 实例
func NewDomainCheckInfo(raw, fmt string, fromUrl bool) *CheckInfo {
	return &CheckInfo{
		Raw:     raw,
		Fmt:     fmt,
		FromUrl: fromUrl,
	}
}

// NewIPCheckInfo 初始化一个新的 CheckInfo 实例
func NewIPCheckInfo(raw, fmt string, isIpv4 bool, fromUrl bool) *CheckInfo {
	result := &CheckInfo{
		Raw:     raw,
		Fmt:     fmt,
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
