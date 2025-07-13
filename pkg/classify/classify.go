package classify

import (
	"fmt"
)

// TargetEntry 表示单个目标条目
type TargetEntry struct {
	RAW     string // 原始输入（如带端口或路径的字符串）
	FMT     string // 格式化后的内容（纯 IP 或 domain）
	IsIPv4  bool   // 是否IPV4
	FromUrl bool   // 是否来源于URL
}

// TargetClassifier 分类器结构体
type TargetClassifier struct {
	IPEntries      []TargetEntry
	DomainEntries  []TargetEntry
	InvalidEntries []string
}

// NewTargetClassifier 创建新的分类器实例
func NewTargetClassifier() *TargetClassifier {
	return &TargetClassifier{
		IPEntries:      make([]TargetEntry, 0),
		DomainEntries:  make([]TargetEntry, 0),
		InvalidEntries: make([]string, 0),
	}
}

// Classify 对输入字符串切片进行分类
func (tc *TargetClassifier) Classify(targets []string) {
	for _, target := range targets {
		category, raw, fmtVal, isURL, isIPv4 := classifyTarget(target)

		switch category {
		case "IP":
			tc.IPEntries = append(tc.IPEntries, TargetEntry{
				RAW:     raw,
				FMT:     fmtVal,
				FromUrl: isURL,
				IsIPv4:  isIPv4,
			})
		case "Domain":
			tc.DomainEntries = append(tc.DomainEntries, TargetEntry{
				RAW:     raw,
				FMT:     fmtVal,
				FromUrl: isURL,
				IsIPv4:  isIPv4,
			})
		case "InvalidEntries":
			tc.InvalidEntries = append(tc.InvalidEntries, raw)
		}
	}
}

// Total 获取所有分类的数量总和
func (tc *TargetClassifier) Total() int {
	return len(tc.IPEntries) + len(tc.DomainEntries) + len(tc.InvalidEntries)
}

// ShowSummary 打印摘要信息
func (tc *TargetClassifier) ShowSummary() {
	total := tc.Total()
	fmt.Printf("Total targets: %d", total)
	fmt.Printf("IPEntries: %d (from URL: %d)", len(tc.IPEntries), countURLs(tc.IPEntries))
	fmt.Printf("DomainEntry: %d (from URL: %d)", len(tc.DomainEntries), countURLs(tc.DomainEntries))
	fmt.Printf("InvalidEntries: %d", len(tc.InvalidEntries))
}

func ClassifyTargets(targets []string) *TargetClassifier {
	classifier := NewTargetClassifier()
	classifier.Classify(targets)
	return classifier
}
