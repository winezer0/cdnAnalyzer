package filetools

import (
	"math/rand"
	"time"
)

// PickRandList 随机选取n个不重复的DNS服务器
func PickRandList(resolvers []string, n int) []string {
	if len(resolvers) <= n {
		return resolvers
	}
	//rand.Seed(time.Now().UnixNano())
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	perm := r.Perm(len(resolvers))
	var picked []string
	for i := 0; i < n; i++ {
		picked = append(picked, resolvers[perm[i]])
	}
	return picked
}

// PickRandMaps 获取随机城市
func PickRandMaps(cityList []map[string]string, count int) []map[string]string {
	if len(cityList) <= count {
		return cityList
	}
	// 设置随机种子
	//rand.Seed(time.Now().UnixNano())
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 存储随机城市数据
	selectedCities := make([]map[string]string, 0, count)
	for i := 0; i < count; i++ {
		randIndex := r.Intn(len(cityList))
		selectedCities = append(selectedCities, cityList[randIndex])
		// 通过切片操作移除指定索引位置元素,防止被多次选择
		cityList = append(cityList[:randIndex], cityList[randIndex+1:]...)
	}
	return selectedCities
}

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
