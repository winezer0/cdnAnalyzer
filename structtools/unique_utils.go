package structutils

// UniqueMergeSlices 合并去重字符串切片
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
