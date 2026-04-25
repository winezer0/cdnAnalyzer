package main

import (
	"errors"
	"fmt"
	"github.com/winezer0/downutils/downutils"
)

// downloadDbFiles 进行数据库文件检查和下载
func downloadDbFiles(downItems []downutils.DownItem, outputForce string, updateForce bool, proxyUrl string) (DBFilePathInfo, error) {
	var errs error
	dbPathsInfo := InitDBPathsInfo(downItems, outputForce)
	// 构建下载选项
	downOptions := &downutils.DownOptions{
		OutputForce:    outputForce,
		UpdateForce:    updateForce,
		EnableForce:    false,
		ShowProgress:   true,
		MaxRetries:     3,
		MaxConcurrent:  1,
		MaxSpeed:       0,
		ConnectTimeout: 30,
		IdleTimeout:    30,
		ProxyURL:       proxyUrl,
	}

	// 加载配置
	downConfig := downutils.DownConfig{
		"default": downItems,
	}

	// 执行下载
	result, err := downutils.ExecuteDownloads(downOptions, downConfig)
	if err != nil {
		errs = fmt.Errorf("download failed: %v", err)
	}

	// 检查下载结果
	if result != nil && result.FailedItems > 0 {
		errs = errors.New("some files failed to download")
	}
	return dbPathsInfo, errs
}

// DBFilePathInfo 存储所有数据库文件路径
type DBFilePathInfo struct {
	ResolversFile string
	CityMapFile   string
	AsnIpvxDb     string
	Ipv4LocateDb  string
	Ipv6LocateDb  string
	CdnSource     string
}

// 数据库文件名常量
const (
	ModuleDNSResolvers = "dns-resolvers"
	ModuleEDNSCityIP   = "edns-city-ip"
	ModuleIPv4Locate   = "ipv4locate"
	ModuleIPv6Locate   = "ipv6locate"
	ModuleAsnIPvx      = "geolite2-asn"
	ModuleCDNSource    = "cdn-sources"
)

// InitDBPathsInfo 从YAMLConfig和下载配置中获取数据库文件路径
func InitDBPathsInfo(downItems []downutils.DownItem, outputForce string) DBFilePathInfo {
	// 初始化空路径
	paths := DBFilePathInfo{}
	if downItems == nil || len(downItems) == 0 {
		return paths
	}

	// 遍历下载配置，根据 module 名称查找对应的文件路径
	for _, item := range downItems {
		module := item.Module
		storePath := downutils.GetModuleFinalPath(module, downItems, outputForce)
		// 根据module名称匹配数据库类型
		switch module {
		case ModuleDNSResolvers:
			paths.ResolversFile = storePath
		case ModuleEDNSCityIP:
			paths.CityMapFile = storePath
		case ModuleAsnIPvx:
			paths.AsnIpvxDb = storePath
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
