package classify

import (
	"net"
	"net/url"
	"regexp"
	"strings"
)

// countURLs 统计列表中来自 URL 的条目数量
func countURLs(entries []TargetEntry) int {
	count := 0
	for _, e := range entries {
		if e.FromUrl {
			count++
		}
	}
	return count
}

// isIpv4 判断输入字符串是否是 IPv4
func isIpv4(ipStr string) bool {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	return ip != nil && ip.To4() != nil
}

// isValidDomain 使用正则判断是否是合法域名
func isValidDomain(domain string) bool {
	domainRegex := `^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(domainRegex, domain)
	return matched
}

// IsValidIP 检查字符串是否是一个纯 IP 地址（不包含端口、路径等）
func IsValidIP(target string) (bool, string) {
	ip := net.ParseIP(strings.TrimSpace(target))
	if ip == nil {
		return false, ""
	}
	return true, ip.String()
}

// classifyTarget 返回目标类型、原始输入、提取后的内容 和 是否是 URL
func classifyTarget(target string) (string, string, string, bool, bool) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "InvalidEntries", "", "", false, false
	}

	// 先尝试作为 URL 解析
	u, err := url.ParseRequestURI(target)
	if err == nil && u.Scheme != "" && u.Host != "" {
		host, _, err := net.SplitHostPort(u.Host)
		if err == nil {
			ipStr := strings.Trim(host, "[]") // 去掉 IPv6 方括号
			if isIP, pureIP := IsValidIP(ipStr); isIP {
				return "IP", target, pureIP, true, isIpv4(pureIP)
			}
			if isValidDomain(host) {
				return "Domain", target, host, true, false
			}
		} else {
			ipStr := strings.Trim(u.Host, "[]")
			if isIP, pureIP := IsValidIP(ipStr); isIP {
				return "IP", target, pureIP, true, isIpv4(pureIP)
			}
			if isValidDomain(u.Host) {
				return "Domain", target, u.Host, true, false
			}
		}
	}

	// 再尝试作为 host:port 解析
	host, _, err := net.SplitHostPort(target)
	if err == nil {
		ipStr := strings.Trim(host, "[]") // 去掉 IPv6 方括号
		if isIP, pureIP := IsValidIP(ipStr); isIP {
			return "IP", target, pureIP, false, isIpv4(pureIP)
		}
		if isValidDomain(host) {
			return "Domain", target, host, false, false
		}
	}

	// 判断纯 IP 地址
	if isIP, pureIP := IsValidIP(target); isIP {
		return "IP", target, pureIP, false, isIpv4(pureIP)
	}

	// 判断纯域名
	if isValidDomain(target) {
		return "Domain", target, target, false, false
	}

	return "InvalidEntries", target, "", false, false
}

// ExtractFMTsPtr 从 TargetEntry 切片中提取所有 FMT 字段，返回字符串切片
func ExtractFMTsPtr(entries *[]TargetEntry) []string {
	fmtList := make([]string, 0, len(*entries))
	for _, entry := range *entries {
		fmtList = append(fmtList, entry.FMT)
	}
	return fmtList
}
