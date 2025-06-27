package main

import (
	"cdnAnalyzer/pkg/analyzer"
	"cdnAnalyzer/pkg/classify"
	"cdnAnalyzer/pkg/docheck"
	"cdnAnalyzer/pkg/domaininfo/querydomain"
	"cdnAnalyzer/pkg/downfile"
	"cdnAnalyzer/pkg/fileutils"
	"cdnAnalyzer/pkg/ipinfo/queryip"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
)

// CmdConfig 存储程序配置，使用结构体标签定义命令行参数
type CmdConfig struct {
	// 配置文件参数
	ConfigFile string `short:"c" long:"config" description:"YAML配置文件路径" default:"config.yaml"`

	// 基本参数 覆盖app Config中的配置
	Target     string `short:"t" long:"target" description:"目标文件路径|目标字符串列表(逗号分隔)"`
	TargetType string `short:"T" long:"target-type" description:"目标数据类型: string/file/sys" choice:"string" choice:"file" choice:"sys"`
	// 输出配置参数 覆盖app Config中的配置
	OutputFile  string `short:"o" long:"output-file" description:"输出结果文件路径"`
	OutputType  string `short:"y" long:"output-type" description:"输出文件类型: csv/json/txt/sys" choice:"csv" choice:"json" choice:"txt" choice:"sys"`
	OutputLevel string `short:"l" long:"output-level" description:"输出详细程度: default/quiet/detail" choice:"default" choice:"quiet" choice:"detail"`

	// 数据库更新配置
	ProxyURL      string `short:"p" long:"proxy" description:"代理URL（支持http://和socks5://）" default:""`
	StoreDir      string `short:"d" long:"store-dir" description:"数据库文件保存目录" default:"assets"`
	UpdateDB      bool   `short:"u" long:"update" description:"强制更新数据库文件，忽略缓存"`
	UpdateDBForce bool   `short:"U" long:"update-force" description:"强制更新数据库文件，忽略缓存"`
}

func main() {
	// 定义配置并解析命令行参数
	var cmdConfig = &CmdConfig{}
	var appConfig = &docheck.AppConfig{}
	var downloadItems []downfile.DownItem

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

	// 1. 先加载配置文件
	configFile := cmdConfig.ConfigFile
	if !fileutils.IsFileExists(configFile) {
		fmt.Printf("指定配置文件不存在: %v\n", configFile)
		os.Exit(1)
	}

	appConfig, err := docheck.LoadAppConfigFile(configFile)
	if err != nil || appConfig == nil {
		fmt.Printf("加载配置文件失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 用命令行中的其他参数覆盖配置文件中的值
	downloadItems = appConfig.DownloadItems
	mergeConfigFromCmd(appConfig, cmdConfig)

	// 3. 获取并检查数据库文件路径
	dbPaths := docheck.GetDBPathsFromConfig(appConfig)
	missedDB := docheck.DBFilesIsExists(dbPaths)
	if len(missedDB) > 0 || appConfig.UpdateDB || appConfig.UpdateDBForce {
		// 更新数据库配置
		downfile.CleanupExpiredCache()
		// 创建HTTP客户端配置
		clientConfig := &downfile.ClientConfig{
			ConnectTimeout: 30,
			IdleTimeout:    30,
			ProxyURL:       appConfig.ProxyURL,
		}
		// 创建HTTP客户端
		httpClient, err := downfile.CreateHTTPClient(clientConfig)
		if err != nil {
			fmt.Printf("创建HTTP客户端失败: %v\n", err)
			return
		}
		// 处理所有配置组
		fmt.Printf("开始下载数据库文件: %s\n")
		totalItems := len(downloadItems)
		successItems := downfile.ProcessDownItems(httpClient, downloadItems, appConfig.StoreDir, appConfig.UpdateDBForce, true, 3)
		fmt.Printf("\n数据库依赖下载任务完成: 成功 %d/%d\n", successItems, totalItems)
	}

	// 检查必要参数
	targetType := strings.ToLower(appConfig.TargetType)
	if appConfig.Target == "" && targetType != "sys" {
		fmt.Println("错误: 必须指定目标(-t, --target)，除非使用 --target-type sys")
		os.Exit(1)
	}

	var targets []string

	// 根据 target-type 参数决定如何处理输入
	switch targetType {
	case "string":
		// 直接输入的目标字符串，按逗号分隔
		if appConfig.Target == "" {
			fmt.Println("错误: 使用 --target-type string 时必须指定 --target")
			os.Exit(1)
		}
		targets = strings.Split(appConfig.Target, ",")

	case "file":
		// 从文件读取目标
		if appConfig.Target == "" {
			fmt.Println("错误: 使用 --target-type file 时必须指定 --target")
			os.Exit(1)
		}
		targets, err = fileutils.ReadTextToList(appConfig.Target)
		if err != nil {
			fmt.Printf("加载目标文件失败: %v\n", err)
			os.Exit(1)
		}

	case "sys":
		// 从标准输入读取目标
		targets, err = fileutils.ReadPipeToList()
		if err != nil {
			fmt.Printf("从标准输入读取目标失败: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("错误: 不支持的 target-type: %s\n", targetType)
		os.Exit(1)
	}

	// 过滤空目标
	var filteredTargets []string
	for _, target := range targets {
		if strings.TrimSpace(target) != "" {
			filteredTargets = append(filteredTargets, strings.TrimSpace(target))
		}
	}
	targets = filteredTargets

	if len(targets) == 0 {
		fmt.Println("错误: 没有有效的目标")
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

	// 检查数据库文件是否存在
	dbFiles := map[string]string{
		"IPv4 ASN数据库":     dbPaths.AsnIpv4Db,
		"IPv6 ASN数据库":     dbPaths.AsnIpv6Db,
		"IPv4地理位置数据库": dbPaths.Ipv4LocateDb,
		"IPv6地理位置数据库": dbPaths.Ipv6LocateDb,
	}

	for name, path := range dbFiles {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("警告: %s文件不存在: %s\n", name, path)
		}
	}

	checkInfos = docheck.QueryIPInfo(ipDbConfig, checkInfos)

	// 加载source.json配置文件
	if _, err := os.Stat(dbPaths.SourceJson); os.IsNotExist(err) {
		fmt.Printf("错误: CDN源数据配置文件不存在: %s\n", dbPaths.SourceJson)
		os.Exit(1)
	}

	cdnData := analyzer.NewEmptyCDNData()
	err = fileutils.ReadJsonToStruct(dbPaths.SourceJson, cdnData)
	if err != nil {
		fmt.Printf("加载CDN源数据失败: %v\n", err)
		os.Exit(1)
	}

	//进行CDN CLOUD WAF 信息分析
	checkResults, err := analyzer.CheckCDNBatch(cdnData, checkInfos)
	if err != nil {
		fmt.Printf("CDN分析失败: %v\n", err)
		os.Exit(1)
	}

	// 处理输出详细程度
	var outputData interface{}
	switch strings.ToLower(appConfig.OutputLevel) {
	case "quiet":
		// 仅输出不是CDN的fmt内容
		outputData = analyzer.GetNoCDNs(checkResults)
	case "detail":
		// 合并 checkResults 到 checkInfos
		outputData = analyzer.MergeCheckResultsToCheckInfos(checkInfos, checkResults)
	default:
		// default: 输出 checkResults
		outputData = checkResults
	}

	// 处理输出类型
	fileutils.WriteOutputToFile(outputData, appConfig.OutputType, appConfig.OutputFile)
}

// mergeConfigFromCmd 将命令行参数合并到YAML配置中
func mergeConfigFromCmd(yamlConfig *docheck.AppConfig, cmdConfig *CmdConfig) {
	if cmdConfig.Target != "" {
		yamlConfig.Target = cmdConfig.Target
	}

	if cmdConfig.TargetType != "" {
		yamlConfig.TargetType = cmdConfig.TargetType
	}

	if cmdConfig.OutputFile != "" {
		yamlConfig.OutputFile = cmdConfig.OutputFile
	}

	if cmdConfig.OutputType != "" {
		yamlConfig.OutputType = cmdConfig.OutputType
	}

	if cmdConfig.OutputLevel != "" {
		yamlConfig.OutputLevel = cmdConfig.OutputLevel
	}

	if cmdConfig.ProxyURL != "" {
		yamlConfig.ProxyURL = cmdConfig.ProxyURL
	}

	if cmdConfig.StoreDir != "" {
		yamlConfig.StoreDir = cmdConfig.StoreDir
	}

	if cmdConfig.UpdateDB || !yamlConfig.UpdateDB {
		yamlConfig.UpdateDB = cmdConfig.UpdateDB
	}

	if cmdConfig.UpdateDBForce || !yamlConfig.UpdateDB {
		yamlConfig.UpdateDB = cmdConfig.UpdateDBForce
	}
}
