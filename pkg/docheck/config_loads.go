package docheck

import (
	"cdnAnalyzer/pkg/fileutils"
	"cdnAnalyzer/pkg/maputils"
	"fmt"
	"github.com/winezer0/downtools/downfile"
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
func GetDBPathsFromConfig(downloadItems []downfile.DownItem, storeDir string) DBFilePaths {
	// 初始化空路径
	paths := DBFilePaths{}
	if downloadItems == nil || len(downloadItems) == 0 {
		return paths
	}

	// 遍历下载配置，根据 module 名称查找对应的文件路径
	for _, item := range downloadItems {
		module := item.Module
		storePath := downfile.GetItemFilePath(item.FileName, storeDir)
		// 根据module名称匹配数据库类型
		switch module {
		case ModuleDNSResolvers:
			paths.ResolversFile = storePath
		case ModuleEDNSCityIP:
			paths.CityMapFile = storePath
		case ModuleAsnIPv4:
			paths.AsnIpv4Db = storePath
		case ModuleAsnIPv6:
			paths.AsnIpv6Db = storePath
		case ModuleIPv4Locate:
			paths.Ipv4LocateDb = storePath
		case ModuleIPv6Locate:
			paths.Ipv6LocateDb = storePath
		case ModuleCDNSource:
			paths.SourceJson = storePath
		}
	}
	return paths
}

// DBFilesIsExists 检查paths列表中是否每个元素都存在
func DBFilesIsExists(paths DBFilePaths) []string {
	// 检查必要的数据库文件路径是否存在
	var missingPaths []string

	if fileutils.IsNotExists(paths.ResolversFile) {
		missingPaths = append(missingPaths, fmt.Sprintf("DNS解析服务器配置文件(resolvers-file) -> Not File [%s]", paths.ResolversFile))
	}
	if fileutils.IsNotExists(paths.CityMapFile) {
		missingPaths = append(missingPaths, fmt.Sprintf("EDNS城市IP映射文件(city-map-file) -> Not File [%s]", paths.CityMapFile))
	}
	if fileutils.IsNotExists(paths.AsnIpv4Db) {
		missingPaths = append(missingPaths, fmt.Sprintf("IPv4 ASN数据库(asn-ipv4-db) -> Not File [%s]", paths.AsnIpv4Db))
	}
	if fileutils.IsNotExists(paths.AsnIpv6Db) {
		missingPaths = append(missingPaths, fmt.Sprintf("IPv6 ASN数据库(asn-ipv6-db) -> Not File [%s]", paths.AsnIpv6Db))
	}
	if fileutils.IsNotExists(paths.Ipv4LocateDb) {
		missingPaths = append(missingPaths, fmt.Sprintf("IPv4地理位置数据库(ipv4-locate-db) -> Not File [%s]", paths.Ipv4LocateDb))
	}
	if fileutils.IsNotExists(paths.Ipv6LocateDb) {
		missingPaths = append(missingPaths, fmt.Sprintf("IPv6地理位置数据库(ipv6-locate-db) -> Not File [%s]", paths.Ipv6LocateDb))
	}
	if fileutils.IsNotExists(paths.SourceJson) {
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

// CheckAndDownDBFiles 进行数据库文件检查和下载
func CheckAndDownDBFiles(downItems []downfile.DownItem, storeDir string, proxyURL string, updateDB bool, updateDBForce bool) (DBFilePaths, error) {
	var errs error = nil
	dbPaths := GetDBPathsFromConfig(downItems, storeDir)
	missedDB := DBFilesIsExists(dbPaths)
	if len(missedDB) > 0 || updateDB || updateDBForce {
		// 更新数据库缓存记录文件
		downfile.CleanupExpiredCache()

		// 创建HTTP客户端配置
		clientConfig := &downfile.ClientConfig{
			ConnectTimeout: 30,
			IdleTimeout:    30,
			ProxyURL:       proxyURL,
		}
		httpClient, err := downfile.CreateHTTPClient(clientConfig)
		if err != nil {
			errs = fmt.Errorf("创建HTTP客户端失败: %v\n", err)
		}

		// 仅保留已启用的项
		if !updateDBForce {
			downItems = downfile.FilterEnableItems(downItems)
			fmt.Printf("过滤下载数据库文件数量: %v\n", len(downItems))
		}

		// 处理所有配置组
		totalItems := len(downItems)
		if totalItems > 0 {
			fmt.Printf("开始下载IP数据库依赖文件, totalItems:[%v]...\n", totalItems)
			downedItems := downfile.ProcessDownItems(httpClient, downItems, storeDir, updateDBForce, true, 3)
			downfile.CleanupIncompleteDownloads(storeDir)
			if downedItems != totalItems {
				errs = fmt.Errorf("文件下载进度缺失[%v/%v]", downedItems, totalItems)
			}
		}
	}
	return dbPaths, errs
}
