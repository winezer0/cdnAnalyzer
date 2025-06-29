package main

import (
	"errors"
	"fmt"
	"github.com/winezer0/cdnAnalyzer/pkg/analyzer"
	"github.com/winezer0/cdnAnalyzer/pkg/classify"
	"github.com/winezer0/cdnAnalyzer/pkg/docheck"
	"github.com/winezer0/cdnAnalyzer/pkg/domaininfo/querydomain"
	"github.com/winezer0/cdnAnalyzer/pkg/fileutils"
	"github.com/winezer0/cdnAnalyzer/pkg/ipinfo/queryip"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/winezer0/downtools/downfile"
)

// CmdConfig 存储程序配置，使用结构体标签定义命令行参数
type CmdConfig struct {
	// 配置文件参数
	ConfigFile string `short:"c" long:"config-file" description:"YAML配置文件路径" default:"config.yaml"`
	ConfigDown string `short:"C" long:"config-down" description:"远程加载config文件并保存" default:""`

	// 基本参数 覆盖app Config中的配置
	Target     string `short:"t" long:"target" description:"目标文件路径|目标字符串列表(逗号分隔)"`
	TargetType string `short:"T" long:"target-type" description:"目标数据类型: string/file/sys" default:"string" choice:"string" choice:"file" choice:"sys"`

	// 输出配置参数 覆盖app Config中的配置
	OutputFile  string `long:"of" description:"输出结果文件路径" default:"analyser_output.json"`
	OutputType  string `long:"ot" description:"输出文件类型: csv/json/txt/sys" default:"sys" choice:"csv" choice:"json" choice:"txt" choice:"sys"`
	OutputLevel string `long:"ol" description:"输出详细程度: default/quiet/detail" default:"default" choice:"default" choice:"quiet" choice:"detail"`
	OutputNoCDN bool   `long:"nc" description:"仅保留非CDN&&非WAF的资产"`

	// 数据库更新配置
	StoreDir      string `short:"d" long:"store-dir" description:"数据库文件存储目录" default:"assets"`
	ProxyURL      string `short:"p" long:"proxy" description:"使用代理URL更新 (http|socks5)" default:""`
	UpdateDB      bool   `short:"u" long:"update" description:"自动更新数据库文件 (检查更新间隔)"`
	UpdateDBForce bool   `short:"U" long:"update-force" description:"强制更新数据库并下载未启用配置"`
}

func main() {
	// 定义配置并解析命令行参数
	var cmdConfig = &CmdConfig{}
	var appConfig = &docheck.AppConfig{}

	// 使用 PassDoubleDash 选项强制使用 - 前缀
	parser := flags.NewParser(cmdConfig, flags.Default)
	parser.Name = "cdnAnalyzer"
	parser.Usage = "[OPTIONS]"

	// 添加描述信息
	parser.ShortDescription = "CDN信息分析检查工具"
	parser.LongDescription = "CDN信息分析检查工具, 用于检查(URL|Domain|IP)等格式目标所使用的(域名解析|IP分析|CDN|WAF|Cloud)等信息."

	// 解析命令行参数
	if _, err := parser.Parse(); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && errors.Is(flagsErr.Type, flags.ErrHelp) {
			os.Exit(0)
		}
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// 0、进行配置文件远程下载
	configFile := cmdConfig.ConfigFile
	configDown := cmdConfig.ConfigDown
	proxyURL := cmdConfig.ProxyURL
	if configDown != "" {
		fmt.Printf("开始远程下载配置文件: %v\n", configDown)
		configUrl := "https://raw.githubusercontent.com/winezer0/cdnAnalyzer/refs/heads/main/cmd/docheck/config.yaml"
		err := downfile.DownloadFileSimple(configUrl, proxyURL, configDown)
		downfile.CleanupIncompleteDownloads("")
		if err != nil {
			fmt.Printf("远程配置文件下载失败: %v\n", err)
			os.Exit(1)
		} else {
			fmt.Printf("远程配置文件下载成功: %v\n", configDown)
			configFile = configDown
		}
	}

	// 1. 先加载配置文件
	if fileutils.IsNotExists(configFile) {
		fmt.Printf("指定配置文件不存在: %v 请调用远程下载或重新指定配置路径.\n", configFile)
		os.Exit(1)
	}

	appConfig, err := docheck.LoadAppConfigFile(configFile)
	if err != nil || appConfig == nil {
		fmt.Printf("加载配置文件失败: %v\n", err)
		os.Exit(1)
	}

	// 3. 获取并检查数据库文件路径
	dbPaths, err := docheck.CheckAndDownDBFiles(appConfig.DownloadItems, cmdConfig.StoreDir, proxyURL, cmdConfig.UpdateDB, cmdConfig.UpdateDBForce)
	if err != nil {
		fmt.Printf("数据库依赖下载异常: %v\n", err)
		return
	}

	// 检查必要参数
	targets, err := checkTargetInfo(cmdConfig.Target, strings.ToLower(cmdConfig.TargetType))
	if err != nil {
		fmt.Printf("检测目标[%s]加载失败: %v\n", cmdConfig.Target, err)
		os.Exit(1)
	}

	// 分类输入数据为 IP Domain InvalidEntries
	classifier := classify.ClassifyTargets(targets)

	//加载dns解析服务器配置文件，用于dns解析调用
	resolvers, err := docheck.LoadResolvers(dbPaths.ResolversFile, appConfig.ResolversNum)
	if err != nil {
		os.Exit(1)
	}

	//加载本地EDNS城市IP信息
	randCities, err := docheck.LoadCityMap(dbPaths.CityMapFile, appConfig.CityMapNUm)
	if err != nil {
		os.Exit(1)
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
	}

	// 进行DNS解析
	checkInfos := docheck.QueryDomainInfo(dnsConfig, classifier.DomainEntries)

	//将所有IP信息加入到 checkInfos 中
	for _, ipEntries := range classifier.IPEntries {
		checkInfo := analyzer.NewIPCheckInfo(ipEntries.RAW, ipEntries.FMT, ipEntries.IsIPv4, ipEntries.FromUrl)
		checkInfos = append(checkInfos, checkInfo)
	}

	//对 checkInfos 中的IP数据进行分析
	ipDbConfig := &queryip.IpDbConfig{
		AsnIpv4Db:    dbPaths.AsnIpv4Db,
		AsnIpv6Db:    dbPaths.AsnIpv6Db,
		Ipv4LocateDb: dbPaths.Ipv4LocateDb,
		Ipv6LocateDb: dbPaths.Ipv6LocateDb,
	}

	checkInfos = docheck.QueryIPInfo(ipDbConfig, checkInfos)

	// 加载sources.json配置文件
	if _, err := os.Stat(dbPaths.CdnSource); os.IsNotExist(err) {
		fmt.Printf("错误: CDN源数据配置文件不存在: %s\n", dbPaths.CdnSource)
		os.Exit(1)
	}

	cdnData := analyzer.NewEmptyCDNData()
	err = fileutils.ReadJsonToStruct(dbPaths.CdnSource, cdnData)
	if err != nil {
		fmt.Printf("加载CDN源数据失败: %v\n", err)
		os.Exit(1)
	}

	//进行CDN CLOUD WAF 信息分析
	checkResults, err := analyzer.CheckCDNBatch(cdnData, checkInfos)
	if err != nil {
		fmt.Printf("CDN分析异常: %v\n", err)
		os.Exit(1)
	}

	//排除Cdn|WAF部分的结果
	if cmdConfig.OutputNoCDN {
		checkResults = analyzer.FilterNoCdnNoWaf(checkResults)
	}

	// 处理输出详细程度
	var outputData interface{}
	switch strings.ToLower(cmdConfig.OutputLevel) {
	case "quiet":
		// 仅输出fmt部分的内容
		outputData = analyzer.GetFmtList(checkResults)
	case "detail":
		// 合并 checkResults 到 checkInfos
		outputData = analyzer.MergeCheckResultsToCheckInfos(checkInfos, checkResults)
	default:
		// default: 输出 checkResults
		outputData = checkResults
	}

	// 处理输出类型
	fileutils.WriteOutputToFile(outputData, cmdConfig.OutputType, cmdConfig.OutputFile)
}

func checkTargetInfo(target string, targetType string) ([]string, error) {
	var (
		targets []string
		err     error
	)

	if target == "" && (targetType == "string" || targetType == "file") {
		return nil, fmt.Errorf("错误: 必须指定目标(-t, --target)，除非使用 --target-type sys")
	}

	switch targetType {
	case "string":
		targets = strings.Split(target, ",")
	case "file":
		targets, err = fileutils.ReadTextToList(target)
		if err != nil {
			return nil, fmt.Errorf("加载目标文件失败: %v", err)
		}
	case "sys":
		targets, err = fileutils.ReadPipeToList()
		if err != nil {
			return nil, fmt.Errorf("从标准输入读取目标失败: %v", err)
		}
	default:
		return nil, fmt.Errorf("不支持的 target-type: %s", targetType)
	}

	// 过滤空字符串
	filtered := make([]string, 0, len(targets))
	for _, t := range targets {
		t = strings.TrimSpace(t)
		if t != "" {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("没有有效的目标")
	}

	return filtered, nil
}
