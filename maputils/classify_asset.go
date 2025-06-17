package maputils

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

// TargetEntry 表示单个目标条目
type TargetEntry struct {
	Raw     string // 原始输入（如带端口或路径的字符串）
	Fmt     string // 格式化后的内容（纯 IP 或 domain）
	IsIPv4  bool   // 是否IPV4
	FromUrl bool   // 是否来源于URL
}

// TargetClassifier 分类器结构体
type TargetClassifier struct {
	IPS     []TargetEntry
	Domains []TargetEntry
	Invalid []string
}

// NewTargetClassifier 创建新的分类器实例
func NewTargetClassifier() *TargetClassifier {
	return &TargetClassifier{
		IPS:     make([]TargetEntry, 0),
		Domains: make([]TargetEntry, 0),
		Invalid: make([]string, 0),
	}
}

// Classify 对输入字符串切片进行分类
func (tc *TargetClassifier) Classify(targets []string) {
	for _, target := range targets {
		category, raw, fmtVal, isURL, isIPv4 := classifyTarget(target)

		switch category {
		case "IP":
			tc.IPS = append(tc.IPS, TargetEntry{
				Raw:     raw,
				Fmt:     fmtVal,
				FromUrl: isURL,
				IsIPv4:  isIPv4,
			})
		case "Domain":
			tc.Domains = append(tc.Domains, TargetEntry{
				Raw:     raw,
				Fmt:     fmtVal,
				FromUrl: isURL,
				IsIPv4:  isIPv4,
			})
		case "Invalid":
			tc.Invalid = append(tc.Invalid, raw)
		}
	}
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
		return "Invalid", "", "", false, false
	}

	// 先尝试作为 URL 解析
	u, err := url.ParseRequestURI(target)
	if err == nil && u.Scheme != "" && u.Host != "" {
		host, _, err := net.SplitHostPort(u.Host)
		if err == nil {
			ipStr := strings.Trim(host, "[]") // 去掉 IPv6 方括号
			if isIP, pureIP := IsValidIP(ipStr); isIP {
				return "IP", target, pureIP, true, IsIpv4(pureIP)
			}
			if isValidDomain(host) {
				return "Domain", target, host, true, false
			}
		} else {
			ipStr := strings.Trim(u.Host, "[]")
			if isIP, pureIP := IsValidIP(ipStr); isIP {
				return "IP", target, pureIP, true, IsIpv4(pureIP)
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
			return "IP", target, pureIP, false, IsIpv4(pureIP)
		}
		if isValidDomain(host) {
			return "Domain", target, host, false, false
		}
	}

	// 判断纯 IP 地址
	if isIP, pureIP := IsValidIP(target); isIP {
		return "IP", target, pureIP, false, IsIpv4(pureIP)
	}

	// 判断纯域名
	if isValidDomain(target) {
		return "Domain", target, target, false, false
	}

	return "Invalid", target, "", false, false
}

// Total 获取所有分类的数量总和
func (tc *TargetClassifier) Total() int {
	return len(tc.IPS) + len(tc.Domains) + len(tc.Invalid)
}

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

// Summary 打印摘要信息
func (tc *TargetClassifier) Summary() {
	total := tc.Total()

	fmt.Printf("Total targets: %d\n", total)
	fmt.Printf("IPS: %d (from URL: %d)\n", len(tc.IPS), countURLs(tc.IPS))
	fmt.Printf("Domains: %d (from URL: %d)\n", len(tc.Domains), countURLs(tc.Domains))
	fmt.Printf("Invalid: %d\n", len(tc.Invalid))
}

// IsIpv4 判断输入字符串是否是 IPv4
func IsIpv4(ipStr string) bool {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	return ip != nil && ip.To4() != nil
}
