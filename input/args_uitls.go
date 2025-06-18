package input

import (
	"cdnCheck/fileutils"
	"cdnCheck/maputils"
	"fmt"
)

func LoadTargets(targetFile string) ([]string, error) {
	targets, err := fileutils.ReadTextToList(targetFile)
	if err != nil {
		return nil, fmt.Errorf("加载目标文件失败: %w", err)
	}
	return targets, nil
}

func LoadResolvers(resolversFile string, resolversNum int) ([]string, error) {
	resolvers, err := fileutils.ReadTextToList(resolversFile)
	if err != nil {
		return nil, fmt.Errorf("加载DNS服务器失败: %w", err)
	}
	resolvers = maputils.PickRandList(resolvers, resolversNum)
	return resolvers, nil
}

func LoadCityMap(cityMapFile string, randCityNum int) ([]map[string]string, error) {
	cityMap, err := fileutils.ReadCSVToMap(cityMapFile)
	if err != nil {
		return nil, fmt.Errorf("读取城市IP映射失败: %w", err)
	}
	selectedCityMap := maputils.PickRandMaps(cityMap, randCityNum)
	fmt.Printf("selectedCityMap: %v\n", selectedCityMap)
	return selectedCityMap, nil
}
