package maputils

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

// TargetClassifier 定义结构体用于存储分类结果
type TargetClassifier struct {
	IPs     []string `json:"ips"`
	Domains []string `json:"domains"`
	URLs    []string `json:"urls"`
	Invalid []string `json:"invalid"`
}

// 创建新的 TargetClassifier 实例
func NewTargetClassifier() *TargetClassifier {
	return &TargetClassifier{
		IPs:     make([]string, 0),
		Domains: make([]string, 0),
		URLs:    make([]string, 0),
		Invalid: make([]string, 0),
	}
}

// 使用正则表达式判断是否为合法域名
func isValidDomain(domain string) bool {
	domainRegex := `^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(domainRegex, domain)
	return matched
}

// 判断字符串类型
func classifyTarget(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "Invalid", fmt.Errorf("empty input")
	}

	// 判断 URL
	if u, err := url.ParseRequestURI(target); err == nil && u.Scheme != "" && u.Host != "" {
		return "URL", nil
	}

	// 判断 IP 地址
	if ip := net.ParseIP(target); ip != nil {
		return "IP", nil
	}

	// 判断域名
	if isValidDomain(target) {
		return "Domain", nil
	}

	return "Invalid", nil
}

// 对输入的字符串切片进行分类
func (tc *TargetClassifier) Classify(targets []string) {
	for _, target := range targets {
		category, _ := classifyTarget(target)
		switch category {
		case "IP":
			tc.IPs = append(tc.IPs, target)
		case "Domain":
			tc.Domains = append(tc.Domains, target)
		case "URL":
			tc.URLs = append(tc.URLs, target)
		default:
			tc.Invalid = append(tc.Invalid, target)
		}
	}
}

// 获取所有分类的数量总和
func (tc *TargetClassifier) Total() int {
	return len(tc.IPs) + len(tc.Domains) + len(tc.URLs) + len(tc.Invalid)
}

// 打印分类结果摘要
func (tc *TargetClassifier) Summary() {
	fmt.Printf("Total targets: %d\n", tc.Total())
	fmt.Printf("IPs: %d\n", len(tc.IPs))
	fmt.Printf("Domains: %d\n", len(tc.Domains))
	fmt.Printf("URLs: %d\n", len(tc.URLs))
	fmt.Printf("Invalid: %d\n", len(tc.Invalid))
}

// 获取特定类型的列表
func (tc *TargetClassifier) Get(category string) ([]string, bool) {
	switch strings.ToLower(category) {
	case "ip", "ips":
		return tc.IPs, true
	case "domain", "domains":
		return tc.Domains, true
	case "url", "urls":
		return tc.URLs, true
	case "invalid":
		return tc.Invalid, true
	default:
		return nil, false
	}
}
