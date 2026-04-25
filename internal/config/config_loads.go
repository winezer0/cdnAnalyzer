package config

import (
	"errors"
	"fmt"
	"github.com/winezer0/cdninfo/internal/embeds"
	"os"

	"github.com/winezer0/cdninfo/pkg/fileutils"
	"github.com/winezer0/cdninfo/pkg/maputils"
	"gopkg.in/yaml.v3"
)

// LoadConfig 加载YAML配置文件
func LoadConfig(configFile string) (*AppConfig, error) {
	if configFile == "" {
		return nil, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, errors.New("读取配置文件失败")
	}

	var appConfig AppConfig
	if err := yaml.Unmarshal(data, &appConfig); err != nil {
		return nil, errors.New("解析YAML配置文件失败")
	}
	return &appConfig, nil
}

// GetDefaultConfig 获取内置默认配置文件内容（从嵌入文件获取）
func GetDefaultConfig() string {
	return embeds.GetConfig()
}

// GenDefaultConfig 生成默认配置文件到指定路径
func GenDefaultConfig(configPath string) error {
	defaultConfig := GetDefaultConfig()
	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}

func LoadResolvers(resolversFile string, resolversNum int) ([]string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(resolversFile); os.IsNotExist(err) {
		return nil, errors.New("DNS解析服务器配置文件不存在")
	}

	resolvers, err := fileutils.ReadTextToList(resolversFile)
	if err != nil {
		return nil, errors.New("加载DNS服务器失败")
	}
	resolvers = maputils.PickRandList(resolvers, resolversNum)
	return resolvers, nil
}

func LoadCityMap(cityMapFile string, randCityNum int) ([]map[string]string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(cityMapFile); os.IsNotExist(err) {
		return nil, errors.New("EDNS城市IP映射文件不存在")
	}

	cityMap, err := fileutils.ReadCSVToMap(cityMapFile)
	if err != nil {
		return nil, fmt.Errorf("加载EDNS城市IP映射文件失败: %v", err)
	}
	selectedCityMap := maputils.PickRandMaps(cityMap, randCityNum)
	return selectedCityMap, nil
}
