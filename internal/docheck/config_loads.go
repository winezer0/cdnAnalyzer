package docheck

import (
	"errors"
	"os"

	"github.com/winezer0/cdnAnalyzer/pkg/downfile"
	"github.com/winezer0/cdnAnalyzer/pkg/fileutils"
	"github.com/winezer0/cdnAnalyzer/pkg/logging"
	"github.com/winezer0/cdnAnalyzer/pkg/maputils"
	"gopkg.in/yaml.v3"
)

// LoadAppConfigFile 加载YAML配置文件
func LoadAppConfigFile(configFile string) (*AppConfig, error) {
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
			paths.CdnSource = storePath
		}
	}
	return paths
}

// DBFilesIsExists 检查paths列表中是否每个元素都存在
func DBFilesIsExists(paths DBFilePaths) []string {
	// 检查必要的数据库文件路径是否存在
	var missingPaths []string

	if fileutils.IsNotExists(paths.ResolversFile) {
		missingPaths = append(missingPaths, "DNS解析服务器配置文件(resolvers-file) -> Not File")
	}
	if fileutils.IsNotExists(paths.CityMapFile) {
		missingPaths = append(missingPaths, "EDNS城市IP映射文件(city-map-file) -> Not File")
	}
	if fileutils.IsNotExists(paths.AsnIpv4Db) {
		missingPaths = append(missingPaths, "IPv4 ASN数据库(asn-ipv4-db) -> Not File")
	}
	if fileutils.IsNotExists(paths.AsnIpv6Db) {
		missingPaths = append(missingPaths, "IPv6 ASN数据库(asn-ipv6-db) -> Not File")
	}
	if fileutils.IsNotExists(paths.Ipv4LocateDb) {
		missingPaths = append(missingPaths, "IPv4地理位置数据库(ipv4-locate-db) -> Not File")
	}
	if fileutils.IsNotExists(paths.Ipv6LocateDb) {
		missingPaths = append(missingPaths, "IPv6地理位置数据库(ipv6-locate-db) -> Not File")
	}
	if fileutils.IsNotExists(paths.CdnSource) {
		missingPaths = append(missingPaths, "CDN源数据配置文件(sources-json) -> Not File")
	}

	return missingPaths
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
		return nil, errors.New("读取城市IP映射失败")
	}
	selectedCityMap := maputils.PickRandMaps(cityMap, randCityNum)
	return selectedCityMap, nil
}

// DownloadDbFiles 进行数据库文件检查和下载
func DownloadDbFiles(downItems []downfile.DownItem, storeDir string, proxyUrl string, updateDB bool) (DBFilePaths, error) {
	var errs error
	dbPaths := GetDBPathsFromConfig(downItems, storeDir)
	missedDB := DBFilesIsExists(dbPaths)
	if len(missedDB) > 0 || updateDB {
		logging.Warnf("Missing DB files, starting download: %v", missedDB)
		// 更新数据库缓存记录文件
		downfile.CleanupExpiredCache()

		// 创建HTTP客户端配置
		clientConfig := &downfile.ClientConfig{
			ConnectTimeout: 30,
			IdleTimeout:    30,
			ProxyURL:       proxyUrl,
		}
		httpClient, err := downfile.CreateHTTPClient(clientConfig)
		if err != nil {
			errs = errors.New("Failed to create HTTP client")
		}

		// 仅保留已启用的项
		downItems = downfile.FilterEnableItems(downItems)

		// 处理所有配置组
		totalItems := len(downItems)
		if totalItems > 0 {
			logging.Debugf("Start downloading the dependent files for the IP database, totalItems:[%v]...", totalItems)
			downedItems := downfile.ProcessDownItems(httpClient, downItems, storeDir, false, false, 3)
			downfile.CleanupIncompleteDownloads(storeDir)
			if downedItems != totalItems {
				errs = errors.New("file download progress is missing")
			}
		}
	}
	return dbPaths, errs
}
