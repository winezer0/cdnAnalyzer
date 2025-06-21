package main

import (
	"cdnCheck/cdncheck"
	"cdnCheck/classify"
	"cdnCheck/domaininfo/querydomain"
	"cdnCheck/fileutils"
	"cdnCheck/input"
	"cdnCheck/ipinfo/queryip"
	"cdnCheck/models"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config 存储程序配置
type Config struct {
	TargetFile string
	OutputFile string

	ResolversFile     string
	ResolversNum      int
	CityMapFile       string
	CityMapNum        int
	QueryEDNSCNAMES   bool
	QueryEDNSUseSysNS bool
	TimeOut           int

	AsnIpv4Db       string
	AsnIpv6Db       string
	Ipv4LocateDb    string
	Ipv6LocateDb    string
	SourceJson      string
	DNSConcurrency  int
	EDNSConcurrency int

	OutputType  string // "CSV" "JSON" "TXT" "SYSOUT"
	OutputLevel string // "default" "quiet" "detail"
}

// parseFlags 解析命令行参数并返回配置
func parseFlags() *Config {
	config := &Config{}

	// 定义命令行参数
	flag.StringVar(&config.TargetFile, "target", "targets.txt", "目标文件路径")
	flag.StringVar(&config.OutputFile, "WriteOutputToFile", "", "输出结果文件路径（默认为目标文件名.results.json）")
	flag.StringVar(&config.ResolversFile, "resolvers", "asset/resolvers.txt", "DNS解析服务器配置文件路径")
	flag.IntVar(&config.ResolversNum, "resolvers-num", 5, "选择用于解析的最大DNS服务器数量")
	flag.StringVar(&config.CityMapFile, "city-map", "asset/city_ip.csv", "EDNS城市IP映射文件路径")
	flag.IntVar(&config.CityMapNum, "city-num", 5, "随机选择的城市数量")

	flag.IntVar(&config.DNSConcurrency, "max-dns-concurrency", 5, "DNS Concurrency")
	flag.IntVar(&config.EDNSConcurrency, "max-edns-concurrency", 5, "EDNS Concurrency")
	flag.IntVar(&config.TimeOut, "max-timeout-second", 5, "TimeOut Second")

	flag.BoolVar(&config.QueryEDNSCNAMES, "query-edns-cnames", false, "enable query edns cnames")
	flag.BoolVar(&config.QueryEDNSUseSysNS, "query-edns-use-sys-ns", false, "enable query edns use sys ns")

	flag.StringVar(&config.AsnIpv4Db, "asn-ipv4", "asset/geolite2-asn-ipv4.mmdb", "IPv4 ASN数据库路径")
	flag.StringVar(&config.AsnIpv6Db, "asn-ipv6", "asset/geolite2-asn-ipv6.mmdb", "IPv6 ASN数据库路径")
	flag.StringVar(&config.Ipv4LocateDb, "ipv4-db", "asset/qqwry.dat", "IPv4地理位置数据库路径")
	flag.StringVar(&config.Ipv6LocateDb, "ipv6-db", "asset/zxipv6wry.db", "IPv6地理位置数据库路径")
	flag.StringVar(&config.SourceJson, "source", "asset/source.json", "CDN源数据配置文件路径")

	flag.StringVar(&config.OutputLevel, "WriteOutputToFile-level", "quiet", "输出详细程度: default/quiet/detail")
	flag.StringVar(&config.OutputType, "WriteOutputToFile-type", "csv", "输出文件类型: csv/json/txt/sys")

	// 解析命令行参数
	flag.Parse()

	return config
}

// queryDomainInfo 进行域名信息解析
func queryDomainInfo(dnsConfig *querydomain.DNSQueryConfig, domainEntries []classify.TargetEntry) []*models.CheckInfo {
	// 创建DNS处理器并执行查询
	dnsProcessor := querydomain.NewDNSProcessor(dnsConfig, &domainEntries)
	dnsResult := dnsProcessor.Process()

	//将dns查询结果合并到 CheckInfo 中去
	var checkInfos []*models.CheckInfo
	for _, domainEntry := range domainEntries {
		var checkInfo *models.CheckInfo
		//当存在dns查询结果时,补充
		if result, ok := (*dnsResult)[domainEntry.FMT]; ok && result != nil {
			checkInfo = querydomain.PopulateDNSResult(domainEntry, result)
		} else {
			fmt.Printf("No DNS result for domain: %s\n", domainEntry.FMT)
			checkInfo = models.NewDomainCheckInfo(domainEntry.RAW, domainEntry.FMT, domainEntry.FromUrl)
		}
		checkInfos = append(checkInfos, checkInfo)
	}
	return checkInfos
}

// queryIPInfo 进行IP信息查询
func queryIPInfo(ipDbConfig *queryip.IpDbConfig, checkInfos []*models.CheckInfo) []*models.CheckInfo {
	// 初始化IP数据库引擎
	ipEngines, err := queryip.InitDBEngines(ipDbConfig)
	if err != nil {
		fmt.Printf("初始化数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer ipEngines.Close()

	//对 checkInfos 中的A/AAAA记录进行IP信息查询，并赋值回去
	for _, checkInfo := range checkInfos {
		if len(checkInfo.A) > 0 || len(checkInfo.AAAA) > 0 {
			ipInfo, err := ipEngines.QueryIPInfo(checkInfo.A, checkInfo.AAAA)
			if err != nil {
				fmt.Printf("查询IP信息失败: %v\n", err)
			} else {
				checkInfo.Ipv4Locate = ipInfo.IPv4Locations
				checkInfo.Ipv4Asn = ipInfo.IPv4AsnInfos
				checkInfo.Ipv6Locate = ipInfo.IPv6Locations
				checkInfo.Ipv6Asn = ipInfo.IPv6AsnInfos
			}
		}
	}

	return checkInfos
}

func main() {
	// 解析命令行参数
	config := parseFlags()

	// 检查目标文件是否存在
	targetFile := config.TargetFile
	if !fileutils.IsFileExists(targetFile) {
		fmt.Printf("目标文件 [%v] 不存在\n", targetFile)
		os.Exit(1)
	}

	//加载输入目标
	targets, err := input.LoadTargets(targetFile)
	if err != nil {
		os.Exit(1)
	}
	//分类输入数据为 IP Domain InvalidEntries
	classifier := classify.ClassifyTargets(targets)

	//加载dns解析服务器配置文件，用于dns解析调用
	resolvers, err := input.LoadResolvers(config.ResolversFile, config.ResolversNum)
	if err != nil {
		os.Exit(1)
	}

	//加载本地EDNS城市IP信息
	randCities, err := input.LoadCityMap(config.CityMapFile, config.CityMapNum)
	if err != nil {
		os.Exit(1)
	}

	// 配置DNS查询参数
	dnsConfig := &querydomain.DNSQueryConfig{
		Resolvers:          resolvers,
		CityMap:            randCities,
		Timeout:            time.Second * time.Duration(config.TimeOut),
		MaxDNSConcurrency:  config.DNSConcurrency,
		MaxEDNSConcurrency: config.EDNSConcurrency,
		QueryEDNSCNAMES:    config.QueryEDNSCNAMES,
		QueryEDNSUseSysNS:  config.QueryEDNSUseSysNS,
	}
	// 进行DNS解析
	checkInfos := queryDomainInfo(dnsConfig, classifier.DomainEntries)

	//将所有IP信息加入到 checkInfos 中
	for _, ipEntries := range classifier.IPEntries {
		checkInfo := models.NewIPCheckInfo(ipEntries.RAW, ipEntries.FMT, ipEntries.IsIPv4, ipEntries.FromUrl)
		checkInfos = append(checkInfos, checkInfo)
	}

	//对 checkInfos 中的IP数据进行分析
	ipDbConfig := &queryip.IpDbConfig{
		AsnIpv4Db:    config.AsnIpv4Db,
		AsnIpv6Db:    config.AsnIpv6Db,
		Ipv4LocateDb: config.Ipv4LocateDb,
		Ipv6LocateDb: config.Ipv6LocateDb,
	}
	checkInfos = queryIPInfo(ipDbConfig, checkInfos)

	// 加载source.json配置文件
	cdnData := cdncheck.NewEmptyCDNData()
	err = fileutils.ReadJsonToStruct(config.SourceJson, cdnData)
	if err != nil {
		fmt.Printf("加载CDN源数据失败: %v\n", err)
		os.Exit(1)
	}

	//进行CDN CLOUD WAF 信息分析
	checkResults, err := cdncheck.CheckCDNBatch(cdnData, checkInfos)
	if err != nil {
		fmt.Printf("CDN分析失败: %v\n", err)
		os.Exit(1)
	}

	// 处理输出详细程度
	var outputData interface{}
	switch strings.ToLower(config.OutputLevel) {
	case "quiet":
		// 仅输出不是CDN的fmt内容
		outputData = cdncheck.GetNoCDNs(checkResults)
	case "detail":
		// 合并 checkResults 到 checkInfos
		outputData = cdncheck.MergeCheckResultsToCheckInfos(checkInfos, checkResults)
	default:
		// default: 输出 checkResults
		outputData = checkResults
	}

	// 处理输出类型
	fileutils.WriteOutputToFile(outputData, config.OutputType, config.OutputFile, targetFile)
}
