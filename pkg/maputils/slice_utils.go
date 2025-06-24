package maputils

import "strings"

// UniqueMergeSlices 将任意数量的 []string 字符串切片合并成一个结果切片，并去除重复项。
func UniqueMergeSlices(slices ...[]string) []string {
	unique := make(map[string]struct{})
	for _, slice := range slices {
		for _, v := range slice {
			unique[v] = struct{}{}
		}
	}
	var result []string
	for v := range unique {
		result = append(result, v)
	}
	return result
}

// UniqueMergeSlicesSorted 将任意数量的 []string 字符串切片合并成一个结果切片，并去除重复项，保留首次出现的顺序。
func UniqueMergeSlicesSorted(slices ...[]string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, slice := range slices {
		for _, item := range slice {
			if _, exists := seen[item]; !exists {
				seen[item] = struct{}{}
				result = append(result, item)
			}
		}
	}
	return result
}

// UniqueMergeAnySlices Go1.18+ 泛型合并多个切片并去重（适用于所有 comparable 类型）
func UniqueMergeAnySlices[T comparable](slices ...[]T) []T {
	unique := make(map[T]struct{})
	for _, slice := range slices {
		if len(slice) > 0 {
			for _, v := range slice {
				unique[v] = struct{}{}
			}
		}
	}
	var result []T
	for v := range unique {
		result = append(result, v)
	}
	return result
}

// StripStrings 去除每个字符串两端的空白字符
func StripStrings(targets []string) []string {
	seen := make(map[string]struct{}, len(targets))
	var filteredTargets []string

	for _, target := range targets {
		trimmed := strings.TrimSpace(target)
		if trimmed != "" {
			if _, exists := seen[trimmed]; !exists {
				seen[trimmed] = struct{}{}
				filteredTargets = append(filteredTargets, trimmed)
			}
		}
	}

	return filteredTargets
}
