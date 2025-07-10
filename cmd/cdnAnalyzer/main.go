package main

import (
	"errors"
	"fmt"
	"github.com/winezer0/cdnAnalyzer/internal/analyzer"
	"github.com/winezer0/cdnAnalyzer/internal/docheck"
	"github.com/winezer0/cdnAnalyzer/pkg/classify"
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
	ConfigFile   string `short:"c" long:"config-file" description:"YAML配置文件路径" default:""`
	UpdateConfig bool   `short:"C" long:"update-config" description:"从Github URL 远程下载更新配置文件内容"`

	// 基本参数 覆盖app Config中的配置
	Input     string `short:"i" long:"input" description:"目标文件路径|目标字符串列表(逗号分隔)"`
	InputType string `short:"I" long:"input-type" description:"目标数据类型: string/file/sys" default:"string" choice:"string" choice:"string" choice:"string"`

	// 输出配置参数 覆盖app Config中的配置
	Output        string `short:"o" long:"output" description:"输出结果文件路径" default:"analyser_output.json"`
	OutputType    string `short:"O" long:"output-type" description:"输出文件类型: csv/json/txt/sys" default:"sys" choice:"csv" choice:"json" choice:"txt" choice:"sys"`
	OutputVerbose string `short:"V" long:"ov" description:"输出详细程度: default/quiet/detail" default:"default" choice:"default" choice:"quiet" choice:"detail"`
	OutputNoCDN   bool   `short:"N" long:"nc" description:"仅保留非CDN&&非WAF的资产"`

	// 数据库更新配置
	Proxy    string `short:"p" long:"proxy" description:"使用代理URL进行数据库下载 (http|socks5)" default:""`
	Folder   string `short:"d" long:"folder" description:"依赖文件存储目录 (默认用户目录)" default:""`
	UpdateDB bool   `short:"u" long:"update-db" description:"自动更新数据库文件 (检查更新间隔)"`
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

	// 初始化数据库文件默认存储目录
	if cmdConfig.Folder == "" {
		cmdConfig.Folder = fileutils.GetUserSubDir("cdnAnalyzer")
	}

	//初始化配置文件默认路径 存储到 cmdConfig.Folder 下
	if cmdConfig.ConfigFile == "" {
		cmdConfig.ConfigFile = fileutils.JoinPath(cmdConfig.Folder, "config.yaml")
	}

	// 进行配置文件远程更新
	if cmdConfig.UpdateConfig {
		url := "https://raw.githubusercontent.com/winezer0/cdnAnalyzer/refs/heads/main/cmd/cdnAnalyzer/config.yaml"
		err := downfile.DownloadFileSimple(url, cmdConfig.Proxy, cmdConfig.ConfigFile)
		downfile.CleanupIncompleteDownloads(cmdConfig.Folder) //清理下载失败的数据
		if err != nil {
			fmt.Printf("Failed to load <%s> -> error: %v\n", url, err)
			os.Exit(1)
		}
	}

	// 加载配置文件远程更新
	appConfig, err := docheck.LoadAppConfigFile(cmdConfig.ConfigFile)
	if err != nil || appConfig == nil {
		fmt.Printf("Failed to load the config: %v\n", err)
		os.Exit(1)
	}

	// 获取并检查数据库文件路径
	dbPaths, err := docheck.CheckAndDownDbFiles(appConfig.DownItems, cmdConfig.Folder, cmdConfig.Proxy, cmdConfig.UpdateDB)
	if err != nil {
		fmt.Printf("Check And Down DbF files failed: %v\n", err)
		os.Exit(1)
	}

	// 检查必要参数
	targets, err := checkTargetInfo(cmdConfig.Input, strings.ToLower(cmdConfig.InputType))
	if err != nil {
		fmt.Printf("target [%s] loading failed: %v\n", cmdConfig.Input, err)
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
	switch strings.ToLower(cmdConfig.OutputVerbose) {
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
	fileutils.WriteOutputToFile(outputData, cmdConfig.OutputType, cmdConfig.Output)
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
