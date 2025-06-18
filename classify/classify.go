package classify

import (
	"fmt"
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

// Total 获取所有分类的数量总和
func (tc *TargetClassifier) Total() int {
	return len(tc.IPS) + len(tc.Domains) + len(tc.Invalid)
}

// ShowSummary 打印摘要信息
func (tc *TargetClassifier) ShowSummary() {
	total := tc.Total()
	fmt.Printf("Total targets: %d\n", total)
	fmt.Printf("IPS: %d (from URL: %d)\n", len(tc.IPS), countURLs(tc.IPS))
	fmt.Printf("Domains: %d (from URL: %d)\n", len(tc.Domains), countURLs(tc.Domains))
	fmt.Printf("Invalid: %d\n", len(tc.Invalid))
}

func ClassifyTargets(targets []string) *TargetClassifier {
	classifier := NewTargetClassifier()
	classifier.Classify(targets)
	classifier.ShowSummary()
	return classifier
}
