package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/winezer0/cdnAnalyzer/internal/analyzer"
	"github.com/winezer0/cdnAnalyzer/internal/generate"
	"github.com/winezer0/cdnAnalyzer/pkg/logging"

	"github.com/jessevdk/go-flags"
	"github.com/winezer0/cdnAnalyzer/pkg/downfile"
	"github.com/winezer0/cdnAnalyzer/pkg/fileutils"
)

// SourcesFilePaths 存储所有依赖文件路径
type SourcesFilePaths struct {
	CdnKeysYaml         string
	CdnDomainsYaml      string
	SourcesChinaJson    string
	SourcesAddedJson    string
	SourcesForeignJson  string
	ProviderForeignYaml string
	UnknownCdnCname     string
}

// Options 命令行参数选项
type Options struct {
	SourcesConfig string `short:"c" description:"资源配置文件路径" default:"sources.yaml"`
	DownloadDir   string `short:"d" description:"资源下载存储目录" default:"sources"`
	SourcesPath   string `short:"o" description:"资源更新后的输出文件" default:"sources/sources.json"`

	// 日志配置参数
	LogFile       string `long:"lf" description:"log file path (default: only stdout)" default:""`
	LogLevel      string `long:"ll" description:"log level: debug/info/warn/error (default debug)" default:"debug" choice:"debug" choice:"info" choice:"warn" choice:"error"`
	ConsoleFormat string `long:"lc" description:"log console format, multiple choice T(time),L(level),C(caller),F(func),M(msg). Empty or off will disable." default:"T L C M"`
}

func main() {
	// 定义命令行参数
	var options Options
	parser := flags.NewParser(&options, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && errors.Is(flagsErr.Type, flags.ErrHelp) {
			os.Exit(0)
		}
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志记录器
	logCfg := logging.NewLogConfig(options.LogLevel, options.LogFile, options.ConsoleFormat)
	if err := logging.InitLogger(logCfg); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logging.Sync()

	sourcesConfig := options.SourcesConfig
	downloadDir := options.DownloadDir
	sourcesPath := options.SourcesPath

	//1、实现相关配置文件自动下载
	downloadConfig, err := downfile.LoadConfig(sourcesConfig)
	if err != nil {
		logging.Errorf("加载配置文件失败: %v", err)
		return
	}
	logging.Debugf("加载配置文件成功: %v", downloadConfig)

	// 清理过期缓存记录
	downfile.CacheFileName = ".sources_cache.json"
	downfile.CleanupExpiredCache()

	// 创建HTTP客户端配置
	clientConfig := &downfile.ClientConfig{
		ConnectTimeout: 10,
		IdleTimeout:    10,
		ProxyURL:       "",
	}

	// 创建HTTP客户端
	httpClient, err := downfile.CreateHTTPClient(clientConfig)
	if err != nil {
		logging.Errorf("创建HTTP客户端失败: %v", err)
		return
	}

	// 提取每个配置组中需要下载的内容
	var downItems []downfile.DownItem
	for _, items := range downloadConfig {
		downItems = append(downItems, items...)
	}

	// 下载所有配置
	if len(downItems) > 0 {
		logging.Debugf("总计需要下载的文件数量:[%v]", len(downItems))
		downfile.ProcessDownItems(httpClient, downItems, downloadDir, false, false, 3)
	} else {
		logging.Debug("未发现需要下载的文件")
		return
	}

	// 清理未完成的下载文件
	downfile.CleanupIncompleteDownloads(downloadDir)

	// 格式化 downItems 中的物理路径
	absDir, err := fileutils.GetAbsolutePath(downloadDir)
	if err != nil {
		logging.Errorf("获取目录[%v]的绝对路径出现错误: %v", absDir, err)
		return
	} else {
		logging.Debugf("DownDir:[%v] -> AbsPath:[%v]", downloadDir, absDir)
	}

	for index, downedItem := range downItems {
		absPath := downfile.GetItemFilePath(downedItem.FileName, absDir)
		logging.Debugf("FileName:[%v] -> AbsPath:[%v]", downedItem.FileName, absPath)
		downedItem.FileName = absPath
		downItems[index] = downedItem
	}

	//检查是否存在未下载的依赖文件
	var missingPaths []string
	for _, downedItem := range downItems {
		if fileutils.IsNotExists(downedItem.FileName) {
			missingPaths = append(missingPaths, fmt.Sprintf("Not File [%s]", downedItem.FileName))
		}
	}

	if len(missingPaths) > 0 {
		logging.Errorf("存在未下载完成的必须依赖文件: %v...", missingPaths)
		return
	} else {
		logging.Debug("所有必须依赖文件已经下载完成.")
	}

	sourcesPaths := getSourcesPaths(downItems)
	logging.Debugf("sourcesPaths: %v.", sourcesPaths)

	// 合并已下载的数据库文件到source
	// 加载并转换 cloud_keys.yml
	cdnKeysTransData := generate.TransferCloudKeysYaml(sourcesPaths.CdnKeysYaml)

	// 加载并转换 cdn.yml
	cdnYamlTransData := generate.TransferCdnDomainsYaml(sourcesPaths.CdnDomainsYaml)

	// 加载sources_foreign.json 数据的合并
	sourceData := generate.TransferPDCdnCheckJson(sourcesPaths.SourcesForeignJson)

	// 加载 sources_provider.json
	sourceProvider := generate.TransferProviderYAML(sourcesPaths.ProviderForeignYaml)

	// 合并以上几种旧的数据源
	mergedCdnData, _ := analyzer.MergeCdnDataList(*sourceData, *cdnYamlTransData, *cdnKeysTransData, *sourceProvider)
	if err != nil {
		logging.Fatalf("Merge Cdn Data List (sourceData, cdnYamlTransData, cdnKeysTransData, sourceProvider) error: %+v.", err)
	}

	// 加载 sources_china.json
	chinaCdnData := analyzer.NewEmptyCDNData()
	if fileutils.ReadJsonToStruct(sourcesPaths.SourcesChinaJson, chinaCdnData); err != nil {
		logging.Errorf("Read Json [%s] To Struct error: %+v.", sourcesPaths.SourcesChinaJson, err)
	}

	//// 加载 sources_added.json
	addedCdnData := analyzer.NewEmptyCDNData()
	if fileutils.ReadJsonToStruct(sourcesPaths.SourcesAddedJson, addedCdnData); err != nil {
		logging.Errorf("Read Json [%s] To Struct error: %+v.", sourcesPaths.SourcesAddedJson, err)
	}

	// 使用合并三种类型的数据
	finalCdnData, err := analyzer.MergeCdnDataList(*mergedCdnData, *chinaCdnData, *addedCdnData)
	if err != nil {
		logging.Fatalf("Merge Cdn Data List (mergedCdnData, chinaCdnData, addedCdnData) error: %+v.", err)
	}

	// 添加未知供应商的CDN CNAME 最后添加
	cnameList, err := fileutils.ReadTextToList(sourcesPaths.UnknownCdnCname)
	cleanedCnameList := cleanCnameList(cnameList)
	analyzer.AddDataToCdnDataCategory(finalCdnData, analyzer.CategoryCDN, analyzer.FieldCNAME, "UNKNOWN", cleanedCnameList)

	// 写入结构体到文件中去
	err = fileutils.WriteSortedJson(sourcesPath, finalCdnData)
	if err != nil {
		logging.Errorf("数据文件[%v]生成失败 -> [%+v]!!!", sourcesPath, err)
	} else {
		logging.Debugf("数据文件[%v]生成成功...", sourcesPath)
	}
}

// getSourcesPaths 从downloadItems中获取数据库文件路径
func getSourcesPaths(downloadItems []downfile.DownItem) SourcesFilePaths {
	// 数据库文件名常量
	const (
		CdnKeysYaml         = "cdn-keys"
		CdnDomainsYaml      = "cdn-domains"
		SourcesChinaJson    = "cdn-sources-china"
		SourcesAddedJson    = "cdn-sources-added"
		SourcesForeignJson  = "cdn-sources-foreign"
		ProviderForeignYaml = "cdn-provider-foreign"
		UnknownCdnCname     = "unknown-cdn-cname"
	)

	// 初始化空路径
	paths := SourcesFilePaths{}
	if downloadItems == nil || len(downloadItems) == 0 {
		return paths
	}

	// 遍历下载配置，根据 module 名称查找对应的文件路径
	for _, item := range downloadItems {
		module := item.Module
		storePath := item.FileName
		// 根据module名称匹配数据库类型
		switch module {
		case CdnKeysYaml:
			paths.CdnKeysYaml = storePath
		case CdnDomainsYaml:
			paths.CdnDomainsYaml = storePath
		case SourcesForeignJson:
			paths.SourcesForeignJson = storePath
		case SourcesChinaJson:
			paths.SourcesChinaJson = storePath
		case SourcesAddedJson:
			paths.SourcesAddedJson = storePath
		case ProviderForeignYaml:
			paths.ProviderForeignYaml = storePath
		case UnknownCdnCname:
			paths.UnknownCdnCname = storePath
		}
	}
	return paths
}

// 遍历列表，清理每个元素
func cleanCnameList(cnameList []string) []string {
	var result []string
	for _, cname := range cnameList {
		cname = strings.Trim(strings.TrimSpace(cname), ".")
		result = append(result, cname)
	}
	return result
}
