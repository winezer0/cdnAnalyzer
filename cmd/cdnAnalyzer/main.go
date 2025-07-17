package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/winezer0/cdnAnalyzer/internal/analyzer"
	"github.com/winezer0/cdnAnalyzer/internal/docheck"
	"github.com/winezer0/cdnAnalyzer/internal/embeds"
	"github.com/winezer0/cdnAnalyzer/pkg/classify"
	"github.com/winezer0/cdnAnalyzer/pkg/domaininfo/querydomain"
	"github.com/winezer0/cdnAnalyzer/pkg/fileutils"
	"github.com/winezer0/cdnAnalyzer/pkg/ipinfo/queryip"

	"github.com/jessevdk/go-flags"
	"github.com/winezer0/cdnAnalyzer/pkg/downfile"
	"github.com/winezer0/cdnAnalyzer/pkg/logging"
)

// CmdConfig 存储程序配置，使用结构体标签定义命令行参数
type CmdConfig struct {
	// 配置文件参数
	ConfigFile   string `short:"c" long:"config-file" description:"config yaml file path (default ~/cdnAnalyzer/config.yaml)" default:""`
	UpdateConfig bool   `short:"C" long:"update-config" description:"updated config content by remote url (default false)"`

	// 基本参数 覆盖app Config中的配置
	Input     string `short:"i" long:"input" description:"input file or str list (separated by commas)"`
	InputType string `short:"I" long:"input-type" description:"input data type: str/file/sys (default str)" default:"str" choice:"file" choice:"str" choice:"sys"`

	// 输出配置参数 覆盖app Config中的配置
	Output      string `short:"o" long:"output" description:"output file path (default result.json)" default:"result.json"`
	OutputType  string `short:"O" long:"output-type" description:"output file type: csv/json/txt/sys (default sys)" default:"sys" choice:"csv" choice:"json" choice:"txt" choice:"sys"`
	OutputLevel int    `short:"l" long:"output-level" description:"Output verbosity level: 1=quiet, 2=default, 3=detail (default 2)" default:"2" choice:"1" choice:"2" choice:"3"`
	OutputNoCDN bool   `short:"n" long:"output-no-cdn" description:"only output Info where not CDN and not WAF."`

	// 数据库更新配置
	Proxy    string `short:"p" long:"proxy" description:"use the proxy URL down files (support http|socks5)" default:""`
	Folder   string `short:"d" long:"folder" description:"db files storage dir (default ~/cdnAnalyzer)" default:""`
	UpdateDB bool   `short:"u" long:"update-db" description:"Auto update db files by interval (default: false)"`

	// DNS 相关参数（新增）
	DNSTimeout        int    `short:"t" long:"dns-timeout" description:"Cover Config, Set DNS query timeout in seconds" default:"0"`
	ResolversNum      int    `short:"r" long:"resolvers-num" description:"Cover Config, Set number of resolvers to use" default:"0"`
	CityMapNum        int    `short:"m" long:"city-map-num" description:"Cover Config, Set number of city map workers" default:"0"`
	DNSConcurrency    int    `short:"w" long:"dns-concurrency" description:"Cover Config, Set concurrent DNS queries" default:"0"`
	EDNSConcurrency   int    `short:"W" long:"edns-concurrency" description:"Cover Config, Set concurrent EDNS queries" default:"0"`
	QueryEDNSCNAMES   string `short:"q" long:"query-ednscnames" description:"Cover Config, Set enable CNAME resolution via EDNS (allow:|false|true)" default:"" choice:"" choice:"false" choice:"true"`
	QueryEDNSUseSysNS string `short:"s" long:"query-edns-use-sys-ns" description:"Cover Config, Set use system nameservers for EDNS (allow:|false|true)" default:"" choice:"" choice:"false" choice:"true"`

	// 日志配置参数
	LogFile       string `long:"lf" description:"log file path (default: only stdout)" default:""`
	LogLevel      string `long:"ll" description:"log level: debug/info/warn/error (default error)" default:"error" choice:"debug" choice:"info" choice:"warn" choice:"error"`
	ConsoleFormat string `long:"lc" description:"log console format, multiple choice T(time),L(level),C(caller),F(func),M(msg). Empty or off will disable." default:"T L C M"`
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
	parser.ShortDescription = "CDN Information Analysis Tool"
	parser.LongDescription = "CDN Information Analysis Tool, Analysis Such as (Domain resolution|IP analysis|CDN|WAF|Cloud)."

	// 解析命令行参数
	if _, err := parser.Parse(); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && errors.Is(flagsErr.Type, flags.ErrHelp) {
			os.Exit(0)
		}
		fmt.Printf("Error:%v\n", err)
		os.Exit(1)
	}

	// 初始化日志记录器
	logCfg := logging.NewLogConfig(cmdConfig.LogLevel, cmdConfig.LogFile, cmdConfig.ConsoleFormat)
	if err := logging.InitLogger(logCfg); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logging.Sync()

	// 初始化数据库文件默认存储目录
	if cmdConfig.Folder == "" {
		cmdConfig.Folder = fileutils.GetUserSubDir("cdnAnalyzer")
	}

	//初始化配置文件默认路径 存储到 cmdConfig.Folder 下
	if cmdConfig.ConfigFile == "" {
		cmdConfig.ConfigFile = fileutils.JoinPath(cmdConfig.Folder, "config.yaml")
		logging.Debugf("Use default current config path: %v", cmdConfig.ConfigFile)
	}

	// 进行配置文件远程更新
	if cmdConfig.UpdateConfig {
		url := "https://raw.githubusercontent.com/winezer0/cdnAnalyzer/refs/heads/main/cmd/cdnAnalyzer/config.yaml"
		err := downfile.DownloadFileSimple(url, cmdConfig.Proxy, cmdConfig.ConfigFile)
		downfile.CleanupIncompleteDownloads(cmdConfig.Folder) //清理下载失败的数据
		if err != nil {
			logging.Fatalf("Failed to load <%s> -> error: %v\n", url, err)
		} else {
			logging.Debugf("Success update config from url: %v", url)
		}
	}

	//生成默认配置文件
	if fileutils.IsEmptyFile(cmdConfig.ConfigFile) {
		fileutils.MakeDirs(cmdConfig.ConfigFile, true)
		fileutils.WriteAny(cmdConfig.ConfigFile, embeds.GetConfig())
		logging.Debugf("Success creat config from embed: %v", cmdConfig.ConfigFile)
	}

	// 加载配置文件远程更新
	appConfig, err := docheck.LoadAppConfigFile(cmdConfig.ConfigFile)
	if err != nil || appConfig == nil {
		logging.Fatalf("Failed to load the config: %v\n", err)
	} else {
		logging.Debugf("Success loaded config: %v", cmdConfig.ConfigFile)
	}

	// 获取并检查数据库文件路径
	dbPaths, err := docheck.DownloadDbFiles(appConfig.DownloadItems, cmdConfig.Folder, cmdConfig.Proxy, cmdConfig.UpdateDB)
	if err != nil {
		logging.Fatalf("Check and down db files failed: %v\n", err)
	} else {
		logging.Debugf("Success Check db files: %v", cmdConfig.Folder)
	}

	// 检查必要参数
	targets, err := checkTargetInfo(cmdConfig.Input, strings.ToLower(cmdConfig.InputType))
	if err != nil {
		logging.Fatalf("target [%s] loading failed: %v\n", cmdConfig.Input, err)
	}

	//使用cmd配置更新配置文件中的某些配置
	appConfig = updateAppConfig(appConfig, cmdConfig)

	// 分类输入数据为 IP Domain InvalidEntries
	classifier := classify.ClassifyTargets(targets)

	//加载dns解析服务器配置文件，用于dns解析调用
	resolvers, err := docheck.LoadResolvers(dbPaths.ResolversFile, appConfig.ResolversNum)
	if err != nil {
		logging.Fatalf("Failed to load resolvers: %v", err)
	} else {
		logging.Debugf("Success get rand resolvers: %v", resolvers)
	}

	//加载本地EDNS城市IP信息
	randCities, err := docheck.LoadCityMap(dbPaths.CityMapFile, appConfig.CityMapNUm)
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
		logging.Fatalf("Error: The CDN source data not exist.: %s\n", dbPaths.CdnSource)
	}

	cdnData := analyzer.NewEmptyCDNData()
	err = fileutils.ReadJsonToStruct(dbPaths.CdnSource, cdnData)
	if err != nil {
		logging.Fatalf("Failed to load CDN source data: %v\n", err)
	} else {
		logging.Debugf("Success load CDN source data: %v", dbPaths.CdnSource)
	}

	//进行CDN CLOUD WAF 信息分析
	checkResults, err := analyzer.CheckCDNBatch(cdnData, checkInfos)
	if err != nil {
		logging.Fatalf("Failed to analysis CDN info: %v\n", err)
	} else {
		logging.Debugf("Success analysis CDN info: %v", dbPaths.CdnSource)
	}

	//排除Cdn|WAF部分的结果
	if cmdConfig.OutputNoCDN {
		checkResults = analyzer.FilterNoCdnNoWaf(checkResults)
	}

	// 处理输出详细程度
	var outputData interface{}
	switch cmdConfig.OutputLevel {
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
	if err = fileutils.WriteOutputToFile(outputData, cmdConfig.OutputType, cmdConfig.Output); err != nil {
		logging.Debugf("Write analysis results to [%v] occur error: %v", cmdConfig.Output, err)
	}
}

// 使用命令行非默认值参数更新配置文件中的参数
func updateAppConfig(appConfig *docheck.AppConfig, cmdConfig *CmdConfig) *docheck.AppConfig {
	if cmdConfig.DNSTimeout > 0 {
		appConfig.DNSTimeOut = cmdConfig.DNSTimeout
	}

	if cmdConfig.ResolversNum > 0 {
		appConfig.ResolversNum = cmdConfig.ResolversNum
	}

	if cmdConfig.CityMapNum > 0 {
		appConfig.CityMapNUm = cmdConfig.CityMapNum
	}

	if cmdConfig.DNSConcurrency > 0 {
		appConfig.DNSConcurrency = cmdConfig.DNSConcurrency
	}

	if cmdConfig.EDNSConcurrency > 0 {
		appConfig.EDNSConcurrency = cmdConfig.EDNSConcurrency
	}

	if cmdConfig.QueryEDNSCNAMES != "" {
		appConfig.QueryEDNSCNAMES = cmdConfig.QueryEDNSCNAMES == "true"
	}

	if cmdConfig.QueryEDNSUseSysNS != "" {
		appConfig.QueryEDNSUseSysNS = cmdConfig.QueryEDNSUseSysNS == "true"
	}

	// 确保并发数有一个合理的默认值，防止死锁
	if appConfig.DNSConcurrency <= 0 {
		appConfig.DNSConcurrency = 10
	}
	if appConfig.EDNSConcurrency <= 0 {
		appConfig.EDNSConcurrency = 10
	}
	return appConfig
}

func checkTargetInfo(target string, targetType string) ([]string, error) {
	var (
		targets []string
		err     error
	)

	if target == "" && (targetType == "str" || targetType == "file") {
		return nil, fmt.Errorf("the target must be specified")
	}

	switch targetType {
	case "str":
		targets = strings.Split(target, ",")
	case "file":
		targets, err = fileutils.ReadTextToList(target)
		if err != nil {
			return nil, fmt.Errorf("failed to load the target file: %v", err)
		}
	case "sys":
		targets, err = fileutils.ReadPipeToList()
		if err != nil {
			return nil, fmt.Errorf("failed to load the system pipe: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported target type: %s", targetType)
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
		return nil, fmt.Errorf("no valid target has been entered")
	}

	return filtered, nil
}
