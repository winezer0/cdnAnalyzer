package models

// 用于保存资产结果时间的类型
type CheckResult struct {
	Raw   string   `json:"raw"`   // 存储原始输入信息
	Fmt   string   `json:"fmt"`   // 存储格式化后的输入信息
	A     []string `json:"A"`     // A记录
	AAAA  []string `json:"AAAA"`  // AAAA记录
	CNAME []string `json:"CNAME"` // CNAME记录
	NS    []string `json:"NS"`    // NS记录
	MX    []string `json:"MX"`    // MX记录
	TXT   []string `json:"TXT"`   // TXT记录

	ALocate     map[string]string `json:"ALocate"`     // A记录的IP解析信息
	AAAAALocate map[string]string `json:"AAAAALocate"` // AAAA记录的IP解析信息

	AAsn    map[string]string `json:"AAsn"`    // A记录的ASN查询信息
	AAAAAsn map[string]string `json:"AAAAAsn"` // AAAA记录的ASN查询信息
}

// NewCheckResult 初始化一个新的 CheckResult 实例
func NewCheckResult(raw, fmt string) *CheckResult {
	return &CheckResult{
		Raw: raw,
		Fmt: fmt,

		A:     make([]string, 0),
		AAAA:  make([]string, 0),
		CNAME: make([]string, 0),
		NS:    make([]string, 0),
		MX:    make([]string, 0),
		TXT:   make([]string, 0),

		ALocate:     make(map[string]string),
		AAAAALocate: make(map[string]string),

		AAsn:    make(map[string]string),
		AAAAAsn: make(map[string]string),
	}
}
