package cdncheck

import "testing"

func TestIsDomainUsingCDN(t *testing.T) {
	// 初始化一个测试用的 CdnMap
	cdnMap := CdnMap{
		"15cdn.com": {
			Name: "腾正安全加速（原 15CDN）",
			Link: "https://www.15cdn.com",
		},
		"tzcdn.cn": {
			Name: "腾正安全加速（原 15CDN）",
			Link: "https://www.15cdn.com",
		},
	}

	tests := []struct {
		name     string
		domain   string
		expected bool
		cdnMatch string // 预期匹配的 CDN 域名
	}{
		{
			name:     "完全匹配 15cdn.com",
			domain:   "15cdn.com",
			expected: true,
			cdnMatch: "15cdn.com",
		},
		{
			name:     "子域匹配 15cdn.com",
			domain:   "img.15cdn.com",
			expected: true,
			cdnMatch: "15cdn.com",
		},
		{
			name:     "完全匹配 tzcdn.cn",
			domain:   "tzcdn.cn",
			expected: true,
			cdnMatch: "tzcdn.cn",
		},
		{
			name:     "子域匹配 tzcdn.cn",
			domain:   "static.tzcdn.cn",
			expected: true,
			cdnMatch: "tzcdn.cn",
		},
		{
			name:     "不匹配的域名",
			domain:   "example.com",
			expected: false,
			cdnMatch: "",
		},
		{
			name:     "错误后缀不匹配",
			domain:   "bad15cdn.com",
			expected: false,
			cdnMatch: "",
		},
		{
			name:     "多级子域",
			domain:   "a.b.c.15cdn.com",
			expected: true,
			cdnMatch: "15cdn.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, matchedCdn := IsDomainUsingCDN(tt.domain, cdnMap)
			if match != tt.expected {
				t.Errorf("预期匹配结果为 %v，但得到 %v", tt.expected, match)
			}
			if match && matchedCdn != tt.cdnMatch {
				t.Errorf("预期匹配 CDN 域名为 %q，但得到 %q", tt.cdnMatch, matchedCdn)
			}
		})
	}
}
