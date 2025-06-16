package maputils

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

// UniqueMergeAnySlices Go1.18+ 泛型合并多个切片并去重（适用于所有 comparable 类型）
func UniqueMergeAnySlices[T comparable](slices ...[]T) []T {
	unique := make(map[T]struct{})
	for _, slice := range slices {
		for _, v := range slice {
			unique[v] = struct{}{}
		}
	}
	var result []T
	for v := range unique {
		result = append(result, v)
	}
	return result
}
