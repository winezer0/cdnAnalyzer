package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/winezer0/cdnAnalyzer/pkg/analyzer"
	"github.com/winezer0/cdnAnalyzer/pkg/fileutils"
	"github.com/winezer0/cdnAnalyzer/pkg/generate"
	"github.com/winezer0/downtools/downfile"
)

// SourcesFilePaths 存储所有依赖文件路径
type SourcesFilePaths struct {
	CdnKeysYaml         string
	CdnDomainsYaml      string
	SourcesChinaJson    string
	SourcesChinaJson2   string
	SourcesForeignJson  string
	ProviderForeignYaml string
	UnknownCdnCname     string
}

// Options 命令行参数选项
type Options struct {
	SourcesConfig string `short:"c" description:"资源配置文件路径" default:"sources.yaml"`
	DownloadDir   string `short:"d" description:"资源下载存储目录" default:"sources"`
	SourcesPath   string `short:"o" description:"资源更新后的输出文件" default:"sources/sources.json"`
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
	}

	sourcesConfig := options.SourcesConfig
	downloadDir := options.DownloadDir
	sourcesPath := options.SourcesPath

	//1、实现相关配置文件自动下载
	downloadConfig, err := downfile.LoadConfig(sourcesConfig)
	if err != nil {
		fmt.Printf("加载配置文件失败: %v\n", err)
		return
	}
	fmt.Printf("加载配置文件成功: %v\n", downloadConfig)

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
		fmt.Printf("创建HTTP客户端失败: %v\n", err)
		return
	}

	// 提取每个配置组中需要下载的内容
	var downItems []downfile.DownItem
	for _, items := range downloadConfig {
		downItems = append(downItems, items...)
	}

	// 下载所有配置
	if len(downItems) > 0 {
		fmt.Printf("总计需要下载的文件数量:[%v]\n", len(downItems))
		downfile.ProcessDownItems(httpClient, downItems, downloadDir, true, false, 3)
	} else {
		fmt.Printf("未发现需要下载的文件\n")
		return
	}

	// 清理未完成的下载文件
	downfile.CleanupIncompleteDownloads(downloadDir)

	// 格式化 downItems 中的物理路径
	absDir, err := fileutils.GetAbsolutePath(downloadDir)
	if err != nil {
		fmt.Printf("获取目录[%v]的绝对路径出现错误: %v\n", absDir, err)
		return
	} else {
		fmt.Printf("DownDir:[%v] -> AbsPath:[%v]\n", downloadDir, absDir)
	}

	for index, downedItem := range downItems {
		absPath := downfile.GetItemFilePath(downedItem.FileName, absDir)
		fmt.Printf("FileName:[%v] -> AbsPath:[%v]\n", downedItem.FileName, absPath)
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
		fmt.Printf("存在未下载完成的必须依赖文件: %v...\n", missingPaths)
		return
	} else {
		fmt.Printf("所有必须依赖文件已经下载完成.\n")
	}

	sourcesPaths := getSourcesPaths(downItems)
	fmt.Printf("sourcesPaths: %v.\n", sourcesPaths)

	// 合并已下载的数据库文件到source
	// 加载并转换 cloud_keys.yml
	cdnKeysTransData := generate.TransferCloudKeysYaml(sourcesPaths.CdnKeysYaml)

	// 加载并转换 cdn.yml
	cdnYamlTransData := generate.TransferCdnDomainsYaml(sourcesPaths.CdnDomainsYaml)

	// 加载sources_foreign.json 数据的合并
	sourceData := generate.TransferPDCdnCheckJson(sourcesPaths.SourcesForeignJson)

	// 加载 sources_china.json
	sourceChina := generate.TransferPDCdnCheckJson(sourcesPaths.SourcesChinaJson)

	// 加载 sources_china2.json
	sourceChina2 := generate.TransferPDCdnCheckJson(sourcesPaths.SourcesChinaJson2)

	// 加载 sources_provider.json
	sourceProvider := generate.TransferProviderYAML(sourcesPaths.ProviderForeignYaml)

	// 合并以上几种数据源
	sourceMerge, _ := analyzer.CdnDataMergeSafe(*sourceData, *sourceChina, *sourceChina2, *cdnYamlTransData, *cdnKeysTransData, *sourceProvider)

	// 合并后的json重新转为CDNData结构体
	sourceCdnData := analyzer.NewEmptyCDNData()
	sourceMergeBytes, err := json.Marshal(sourceMerge)
	json.Unmarshal(sourceMergeBytes, sourceCdnData)

	// 添加未知供应商的CDN CNAME
	cnameList, err := fileutils.ReadTextToList(sourcesPaths.UnknownCdnCname)
	cleanedCnameList := cleanCnameList(cnameList)
	analyzer.AddDataToCdnDataCategory(sourceCdnData, analyzer.CategoryCDN, analyzer.FieldCNAME, "UNKNOWN", cleanedCnameList)

	// 写入结构体到文件中去
	fileutils.WriteSortedJson(sourcesPath, sourceCdnData)
	if fileutils.IsEmptyFile(sourcesPath) {
		fmt.Printf("数据文件[%v]生成失败!!!\n", sourcesPath)
	} else {
		fmt.Printf("数据文件[%v]生成成功...\n", sourcesPath)
	}
}

// getSourcesPaths 从downloadItems中获取数据库文件路径
func getSourcesPaths(downloadItems []downfile.DownItem) SourcesFilePaths {
	// 数据库文件名常量
	const (
		CdnKeysYaml         = "cdn-keys"
		CdnDomainsYaml      = "cdn-domains"
		SourcesChinaJson    = "cdn-sources-china"
		SourcesChinaJson2   = "cdn-sources-china2"
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
		case SourcesChinaJson2:
			paths.SourcesChinaJson2 = storePath
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
