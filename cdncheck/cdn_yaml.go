package cdncheck

import (
	"encoding/json"
	"strings"
)

// CdnProvider 表示每个 CDN 提供商的结构体
type CdnProvider struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

// CdnMap 是域名到 CdnProvider 的映射
type CdnMap map[string]CdnProvider

// ParseCdnJSON 将 JSON 字符串解析为 CdnMap
func ParseCdnJSON(jsonData string) (CdnMap, error) {
	var cdnMap CdnMap = make(map[string]CdnProvider)
	err := json.Unmarshal([]byte(jsonData), &cdnMap)
	if err != nil {
		return nil, err
	}
	return cdnMap, nil
}

// IsDomainUsingCDN 判断给定域名是否使用了 CDN 提供商
func IsDomainUsingCDN(domain string, cdnMap CdnMap) (bool, string) {
	for cdnDomain := range cdnMap {
		if containsDomain(domain, cdnDomain) {
			return true, cdnDomain
		}
	}
	return false, ""
}

// containsDomain 判断 domain 是否是 cdnDomain 的子域或等于它本身
func containsDomain(domain, cdnDomain string) bool {
	return domain == cdnDomain || strings.HasSuffix(domain, "."+cdnDomain)
}
