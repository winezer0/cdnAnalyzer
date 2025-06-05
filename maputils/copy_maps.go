package maputils

// 辅助函数：深拷贝 map[string][]string
func CopyMap(src map[string][]string) map[string][]string {
	dst := make(map[string][]string)
	for k, v := range src {
		dst[k] = append([]string{}, v...)
	}
	return dst
}
