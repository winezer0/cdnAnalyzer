package docheck

import (
	"cdnAnalyzer/pkg/fileutils"
	"cdnAnalyzer/pkg/maputils"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// LoadAppConfigFile 加载YAML配置文件
func LoadAppConfigFile(configFile string) (*AppConfig, error) {
	if configFile == "" {
		return nil, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var appConfig AppConfig
	if err := yaml.Unmarshal(data, &appConfig); err != nil {
		return nil, fmt.Errorf("解析YAML配置文件失败: %w", err)
	}
	return &appConfig, nil
}

// GetDBPathsFromConfig 从YAMLConfig和下载配置中获取数据库文件路径
func GetDBPathsFromConfig(appConfig *AppConfig) DBFilePaths {
	// 初始化空路径
	paths := DBFilePaths{}
	if appConfig == nil || appConfig.DownloadItems == nil {
		return paths
	}

	// 遍历下载配置，根据m odule 名称查找对应的文件路径
	for _, item := range appConfig.DownloadItems {
		module := item.Module
		filepath := item.FileName

		// 根据module名称匹配数据库类型
		switch module {
		case ModuleDNSResolvers:
			paths.ResolversFile = filepath
		case ModuleEDNSCityIP:
			paths.CityMapFile = filepath
		case ModuleAsnIPv4:
			paths.AsnIpv4Db = filepath
		case ModuleAsnIPv6:
			paths.AsnIpv6Db = filepath
		case ModuleIPv4Locate:
			paths.Ipv4LocateDb = filepath
		case ModuleIPv6Locate:
			paths.Ipv6LocateDb = filepath
		case ModuleCDNSource:
			paths.SourceJson = filepath
		}
	}
	return paths
}

// DBFilesIsExists 检查paths列表中是否每个元素都存在
func DBFilesIsExists(paths DBFilePaths) []string {
	// 检查必要的数据库文件路径是否存在
	var missingPaths []string

	if !fileutils.IsFileExists(paths.ResolversFile) {
		missingPaths = append(missingPaths, fmt.Sprintf("DNS解析服务器配置文件(resolvers-file) -> Not File [%s]", paths.ResolversFile))
	}
	if !fileutils.IsFileExists(paths.CityMapFile) {
		missingPaths = append(missingPaths, fmt.Sprintf("EDNS城市IP映射文件(city-map-file) -> Not File [%s]", paths.CityMapFile))
	}
	if !fileutils.IsFileExists(paths.AsnIpv4Db) {
		missingPaths = append(missingPaths, fmt.Sprintf("IPv4 ASN数据库(asn-ipv4-db) -> Not File [%s]", paths.AsnIpv4Db))
	}
	if !fileutils.IsFileExists(paths.AsnIpv6Db) {
		missingPaths = append(missingPaths, fmt.Sprintf("IPv6 ASN数据库(asn-ipv6-db) -> Not File [%s]", paths.AsnIpv6Db))
	}
	if !fileutils.IsFileExists(paths.Ipv4LocateDb) {
		missingPaths = append(missingPaths, fmt.Sprintf("IPv4地理位置数据库(ipv4-locate-db) -> Not File [%s]", paths.Ipv4LocateDb))
	}
	if !fileutils.IsFileExists(paths.Ipv6LocateDb) {
		missingPaths = append(missingPaths, fmt.Sprintf("IPv6地理位置数据库(ipv6-locate-db) -> Not File [%s]", paths.Ipv6LocateDb))
	}
	if fileutils.IsFileExists(paths.SourceJson) {
		missingPaths = append(missingPaths, fmt.Sprintf("CDN源数据配置文件(source-json) -> Not File [%s]", paths.SourceJson))
	}

	return missingPaths
}

func LoadResolvers(resolversFile string, resolversNum int) ([]string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(resolversFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("DNS解析服务器配置文件不存在: %s", resolversFile)
	}

	resolvers, err := fileutils.ReadTextToList(resolversFile)
	if err != nil {
		return nil, fmt.Errorf("加载DNS服务器失败: %w", err)
	}
	resolvers = maputils.PickRandList(resolvers, resolversNum)
	return resolvers, nil
}

func LoadCityMap(cityMapFile string, randCityNum int) ([]map[string]string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(cityMapFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("EDNS城市IP映射文件不存在: %s", cityMapFile)
	}

	cityMap, err := fileutils.ReadCSVToMap(cityMapFile)
	if err != nil {
		return nil, fmt.Errorf("读取城市IP映射失败: %w", err)
	}
	selectedCityMap := maputils.PickRandMaps(cityMap, randCityNum)
	//fmt.Printf("selectedCityMap: %v\n", selectedCityMap)
	return selectedCityMap, nil
}
