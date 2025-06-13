package models

import (
	"cdnCheck/iplocate/asndb"
)

// 用于保存资产结果时间的类型
type CheckResult struct {
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

	Ipv4Asn []asndb.ASNInfo `json:"Ipv4Asn,omitempty"` // A记录的ASN查询信息
	Ipv6Asn []asndb.ASNInfo `json:"Ipv6Asn,omitempty"` // AAAA记录的ASN查询信息
}

// NewDomainCheckResult 初始化一个新的 CheckResult 实例
func NewDomainCheckResult(raw, fmt string, fromUrl bool) *CheckResult {
	return &CheckResult{
		Raw:     raw,
		Fmt:     fmt,
		FromUrl: fromUrl,
	}
}

// NewIPCheckResult 初始化一个新的 CheckResult 实例
func NewIPCheckResult(raw, fmt string, isIpv4 bool, fromUrl bool) *CheckResult {
	result := &CheckResult{
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
