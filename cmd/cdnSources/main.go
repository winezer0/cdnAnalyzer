package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/winezer0/cdninfo/internal/analyzer"
	"github.com/winezer0/cdninfo/internal/generate"
	"github.com/winezer0/downutils/downutils"
	"github.com/winezer0/xutils/logging"

	"github.com/jessevdk/go-flags"
	"github.com/winezer0/cdninfo/pkg/fileutils"
)

// SourcesFilePathInfo 存储所有依赖文件路径
type SourcesFilePathInfo struct {
	CdnKeysYaml         string
	CdnDomainsYaml      string
	SourcesChinaJson    string
	SourcesAddedJson    string
	SourcesForeignJson  string
	ProviderForeignYaml string
	UnknownCdnCname     string
}

// 版本信息常量(根据实际情况修改)
const (
	AppName      = "cdnsources"
	AppShortDesc = "cdn auto sources update"
	AppLongDesc  = "cdn auto sources update"
	AppVersion   = "0.0.3"
	BuildDate    = "2026-04-27"
)

// Options 命令行参数选项
type Options struct {
	SourcesConfig string `short:"c" description:"资源配置文件路径" default:"sources.yaml"`
	SourcesOut    string `short:"o" description:"资源更新后的输出文件" default:"sources/sources.json"`

	// 下载配置参数
	DownloadDir string `short:"d" description:"资源下载存储目录" default:"sources"`
	NoProgress  bool   `long:"no-progress" description:"不显示下载进度条"`
	UpdateForce bool   `long:"update-force" description:"强制更新，忽略更新策略"`
	EnableForce bool   `long:"enable-force" description:"强制启用，忽略Enable字段"`
	ProxyURL    string `long:"proxy-url" description:"代理URL" default:""`

	Version bool `short:"v" long:"version" description:"显示版本信息"`
	// 日志配置参数
	LogFile       string `long:"lf" description:"log file path (default: only stdout)" default:""`
	LogLevel      string `long:"ll" description:"log level: debug/info/warn/error (default debug)" default:"debug" choice:"debug" choice:"info" choice:"warn" choice:"error"`
	ConsoleFormat string `long:"lc" description:"log console format, multiple choice T(time),L(level),C(caller),F(func),M(msg). Empty or off will disable." default:"T L C M"`
}

// InitOptionsArgs 常用的工具函数，解析parser和logging配置
func InitOptionsArgs(minimumParams int) (*Options, *flags.Parser) {
	opts := &Options{}
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = AppName
	parser.Usage = "[OPTIONS]"
	parser.ShortDescription = AppShortDesc
	parser.LongDescription = AppLongDesc

	// 命令行参数数量检查 指不包含程序名本身的参数数量
	if minimumParams > 0 && len(os.Args)-1 < minimumParams {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	// 命令行参数解析检查
	if _, err := parser.Parse(); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && errors.Is(flagsErr.Type, flags.ErrHelp) {
			os.Exit(0)
		}
		fmt.Printf("Error:%v\n", err)
		os.Exit(1)
	}

	// 版本号输出
	if opts.Version {
		fmt.Printf("%s version %s\n", AppName, AppVersion)
		fmt.Printf("Build Date: %s\n", BuildDate)
		os.Exit(0)
	}

	// 初始化日志器
	logCfg := logging.NewLogConfig(opts.LogLevel, opts.LogFile, opts.ConsoleFormat)
	if err := logging.InitDefaultLogger(logCfg); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	if opts.SourcesConfig == "" {
		logging.Errorf("config file path is empty")
		os.Exit(1)
	}

	return opts, parser
}

func main() {
	// 定义命令行参数
	// 打印命令行输入配置
	opts, _ := InitOptionsArgs(0)
	defer logging.Sync()

	// 加载并验证配置文件
	downConfig, err := downutils.LoadConfig(opts.SourcesConfig)
	if err != nil {
		logging.Errorf("failed to load config: %v", err)
		return
	}
	logging.Debugf("加载配置文件成功: %v", downConfig)

	// 显示配置信息
	displayOptions(opts)

	// 构建下载选项
	downOptions := &downutils.DownOptions{
		OutputForce:    opts.DownloadDir,
		UpdateForce:    opts.UpdateForce,
		EnableForce:    opts.EnableForce,
		ShowProgress:   !opts.NoProgress,
		ProxyURL:       opts.ProxyURL,
		MaxRetries:     3,
		MaxConcurrent:  1,
		MaxSpeed:       0,
		ConnectTimeout: 30,
		IdleTimeout:    30,
	}

	// 执行下载流程
	result, err := downutils.ExecuteDownloads(downOptions, downConfig)
	if err != nil {
		logging.Errorf("download failed: %v", err)
		return
	}

	// 显示下载结果
	downutils.DisplayDownloadResult(result)

	// 从下载配置中提取文件路径信息
	var downItems []downutils.DownItem
	for _, items := range downConfig {
		downItems = append(downItems, items...)
	}

	//检查是否存在未下载的依赖文件
	var missingPaths []string
	for _, item := range downItems {
		finalPath := downutils.GetDownItemFinalPath(item.FileName, item.StorageDir, downOptions.OutputForce)
		if fileutils.IsNotExists(finalPath) {
			missingPaths = append(missingPaths, fmt.Sprintf("Not File [%s]", finalPath))
		}
	}

	if len(missingPaths) > 0 {
		logging.Errorf("存在未下载完成的必须依赖文件: %v...", missingPaths)
		return
	} else {
		logging.Debug("所有必须依赖文件已经下载完成.")
	}

	sourcesPaths := getSourcesPaths(downItems, opts.DownloadDir)
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
	err = fileutils.WriteSortedJson(opts.SourcesOut, finalCdnData)
	if err != nil {
		logging.Errorf("数据文件[%v]生成失败 -> [%+v]!!!", opts.SourcesOut, err)
	} else {
		logging.Debugf("数据文件[%v]生成成功...", opts.SourcesOut)
	}
}

// getSourcesPaths 从downloadItems中获取数据库文件路径
func getSourcesPaths(downloadItems []downutils.DownItem, outputForceDir string) SourcesFilePathInfo {
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
	paths := SourcesFilePathInfo{}
	if downloadItems == nil || len(downloadItems) == 0 {
		return paths
	}

	// 遍历下载配置，根据 module 名称查找对应的文件路径
	for _, item := range downloadItems {
		module := item.Module
		storePath := downutils.GetModuleFinalPath(module, downloadItems, outputForceDir)
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

// displayOptions 显示配置信息
func displayOptions(opts *Options) {
	logging.Infof("Configuration loaded:")
	logging.Infof("  SourcesConfig: %s", opts.SourcesConfig)
	logging.Infof("  DownloadDir: %s", opts.DownloadDir)
	logging.Infof("  SourcesPath: %s", opts.SourcesOut)
	logging.Infof("  NoProgress: %v", opts.NoProgress)
	logging.Infof("  OutputForce: %s", opts.DownloadDir)
	logging.Infof("  UpdateForce: %v", opts.UpdateForce)
	logging.Infof("  EnableForce: %v", opts.EnableForce)
	logging.Infof("  ProxyURL: %s", opts.ProxyURL)
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
