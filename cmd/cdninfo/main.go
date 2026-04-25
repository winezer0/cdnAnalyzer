package main

import (
	"github.com/winezer0/cdninfo/internal/config"
	"os"
	"strings"
	"time"

	"github.com/winezer0/ipinfo/pkg/queryip"

	"github.com/winezer0/cdninfo/internal/analyzer"
	"github.com/winezer0/cdninfo/internal/docheck"
	"github.com/winezer0/cdninfo/pkg/classify"
	"github.com/winezer0/cdninfo/pkg/domaininfo/querydomain"
	"github.com/winezer0/cdninfo/pkg/fileutils"
	"github.com/winezer0/xutils/logging"
)

func main() {
	// 定义命令行参数
	// 打印命令行输入配置
	opts, _ := InitOptionsArgs(0)
	defer logging.Sync()

	// 加载配置文件远程更新
	var appConfig = &config.AppConfig{}
	appConfig, err := config.LoadConfig(opts.ConfigFile)
	if err != nil || appConfig == nil {
		logging.Fatalf("Failed to load the config: %v\n", err)
	} else {
		logging.Debugf("Success loaded config: %v", opts.ConfigFile)
	}

	// 获取并检查数据库文件路径
	dbPathsInfo, err := downloadDbFiles(appConfig.DownloadItems, opts.StoreDB, opts.UpdateDB, opts.ProxyUrl)
	if err != nil {
		logging.Fatalf("Check and down db files failed: %v\n", err)
	} else {
		logging.Debugf("Success Check db files: %v", opts.StoreDB)
	}

	// 检查必要参数
	targets, err := parseTargetFomat(opts.Input, strings.ToLower(opts.InputType))
	if err != nil {
		logging.Fatalf("target [%s] loading failed: %v\n", opts.Input, err)
	}

	//使用cmd配置更新配置文件中的某些配置
	appConfig = updateAppConfigByOpts(appConfig, opts)

	// 分类输入数据为 IP Domain InvalidEntries
	classifier := classify.ClassifyTargets(targets)

	//加载dns解析服务器配置文件，用于dns解析调用
	resolvers, err := config.LoadResolvers(dbPathsInfo.ResolversFile, appConfig.ResolversNum)
	if err != nil {
		logging.Fatalf("Failed to load resolvers: %v", err)
	} else {
		logging.Debugf("Success get rand resolvers: %v", resolvers)
	}

	//加载本地EDNS城市IP信息
	randCities, err := config.LoadCityMap(dbPathsInfo.CityMapFile, appConfig.CityMapNUm)
	if err != nil {
		logging.Fatalf("Failed to load city map: %v", err)
	} else {
		logging.Debugf("Success get rand cities: %v", randCities)
	}

	// 配置DNS查询参数
	dnsConfig := &querydomain.DNSQueryConfig{
		Resolvers:          resolvers,
		CityMap:            randCities,
		Timeout:            time.Second * time.Duration(appConfig.DNSTimeOut),
		MaxDNSConcurrency:  appConfig.DNSConcurrency,
		MaxEDNSConcurrency: appConfig.EDNSConcurrency,
		QueryEDNSCNAMES:    appConfig.QueryEDNSCNAMES,
		QueryEDNSUseSysNS:  appConfig.QueryEDNSUseSysNS,
		QueryType:          appConfig.QueryMethod, // 添加查询类型配置
	}

	// 进行DNS解析
	checkInfos := docheck.QueryDomainInfo(dnsConfig, classifier.DomainEntries)

	//将所有IP信息加入到 checkInfos 中
	for _, ipEntries := range classifier.IPEntries {
		checkInfo := analyzer.NewIPCheckInfo(ipEntries.RAW, ipEntries.FMT, ipEntries.IsIPv4, ipEntries.FromUrl)
		checkInfos = append(checkInfos, checkInfo)
	}

	//对 checkInfos 中的IP数据进行分析
	ipDbConfig := &queryip.IPDbConfig{
		AsnIpvxDb:    dbPathsInfo.AsnIpvxDb,
		Ipv4LocateDb: dbPathsInfo.Ipv4LocateDb,
		Ipv6LocateDb: dbPathsInfo.Ipv6LocateDb,
	}

	checkInfos = docheck.QueryIPInfo(ipDbConfig, checkInfos)

	// 加载sources.json配置文件
	if _, err := os.Stat(dbPathsInfo.CdnSource); os.IsNotExist(err) {
		logging.Fatalf("Error: The CDN source data not exist.: %s\n", dbPathsInfo.CdnSource)
	}

	cdnData := analyzer.NewEmptyCDNData()
	err = fileutils.ReadJsonToStruct(dbPathsInfo.CdnSource, cdnData)
	if err != nil {
		logging.Fatalf("Failed to load CDN source data: %v\n", err)
	} else {
		logging.Debugf("Success load CDN source data: %v", dbPathsInfo.CdnSource)
	}

	//进行CDN CLOUD WAF 信息分析
	checkResults, err := analyzer.CheckCDNBatch(cdnData, checkInfos)
	if err != nil {
		logging.Fatalf("Failed to analysis CDN info: %v\n", err)
	} else {
		logging.Debugf("Success analysis CDN info: %v", dbPathsInfo.CdnSource)
	}

	//排除Cdn|WAF部分的结果
	if opts.OutputNoCDN {
		checkResults = analyzer.FilterNoCdnNoWaf(checkResults)
	}

	// 处理输出详细程度
	var outputData interface{}
	switch opts.OutputLevel {
	case 1:
		// 仅输出fmt部分的内容
		outputData = analyzer.GetFmtList(checkResults)
	case 3:
		// 合并 checkResults 到 checkInfos
		outputData = analyzer.MergeCheckResultsToCheckInfos(checkInfos, checkResults)
	default:
		// 输出 checkResults
		outputData = checkResults
	}

	// 处理输出类型
	if err = fileutils.WriteOutputToFile(outputData, opts.OutputType, opts.Output); err != nil {
		logging.Debugf("Write analysis results to [%v] occur error: %v", opts.Output, err)
	}
}
