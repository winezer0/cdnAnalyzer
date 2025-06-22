package main

import (
	cdncheck2 "cdnCheck/pkg/cdncheck"
	"cdnCheck/pkg/classify"
	querydomain2 "cdnCheck/pkg/domaininfo/querydomain"
	fileutils2 "cdnCheck/pkg/fileutils"
	"cdnCheck/pkg/ipinfo/queryip"
	"cdnCheck/pkg/maputils"
	"cdnCheck/pkg/models"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
)

// Config 存储程序配置，使用结构体标签定义命令行参数
type Config struct {
	// 基本参数
	Target string `short:"t" long:"target" description:"目标文件路径或直接输入的目标(多个目标用逗号分隔)"`

	// DNS配置参数
	ResolversFile     string `short:"r" long:"resolvers" description:"DNS解析服务器配置文件路径" default:"asset/resolvers.txt"`
	ResolversNum      int    `short:"n" long:"resolvers-num" description:"选择用于解析的最大DNS服务器数量" default:"5"`
	CityMapFile       string `short:"c" long:"city-map" description:"EDNS城市IP映射文件路径" default:"asset/city_ip.csv"`
	CityMapNum        int    `short:"m" long:"city-num" description:"随机选择的城市数量" default:"5"`
	DNSConcurrency    int    `short:"d" long:"dns-concurrency" description:"DNS并发数" default:"5"`
	EDNSConcurrency   int    `short:"e" long:"edns-concurrency" description:"EDNS并发数" default:"5"`
	TimeOut           int    `short:"w" long:"timeout" description:"超时时间(秒)" default:"5"`
	QueryEDNSCNAMES   bool   `short:"C" long:"query-edns-cnames" description:"启用EDNS CNAME查询"`
	QueryEDNSUseSysNS bool   `short:"S" long:"query-edns-use-sys-ns" description:"启用EDNS系统NS查询"`

	// 数据库配置参数
	AsnIpv4Db    string `short:"a" long:"asn-ipv4" description:"IPv4 ASN数据库路径" default:"asset/geolite2-asn-ipv4.mmdb"`
	AsnIpv6Db    string `short:"A" long:"asn-ipv6" description:"IPv6 ASN数据库路径" default:"asset/geolite2-asn-ipv6.mmdb"`
	Ipv4LocateDb string `short:"4" long:"ipv4-db" description:"IPv4地理位置数据库路径" default:"asset/qqwry.dat"`
	Ipv6LocateDb string `short:"6" long:"ipv6-db" description:"IPv6地理位置数据库路径" default:"asset/zxipv6wry.db"`
	SourceJson   string `short:"s" long:"source" description:"CDN源数据配置文件路径" default:"asset/source.json"`

	// 输出配置参数
	OutputFile  string `short:"o" long:"output-file" description:"输出结果文件路径"  default:""`
	OutputType  string `short:"y" long:"output-type" description:"输出文件类型: csv/json/txt/sys" default:"sys" choice:"csv" choice:"json" choice:"txt" choice:"sys"`
	OutputLevel string `short:"l" long:"output-level" description:"输出详细程度: default/quiet/detail" default:"default" choice:"default" choice:"quiet" choice:"detail"`
}

func LoadResolvers(resolversFile string, resolversNum int) ([]string, error) {
	resolvers, err := fileutils2.ReadTextToList(resolversFile)
	if err != nil {
		return nil, fmt.Errorf("加载DNS服务器失败: %w", err)
	}
	resolvers = maputils.PickRandList(resolvers, resolversNum)
	return resolvers, nil
}

func LoadCityMap(cityMapFile string, randCityNum int) ([]map[string]string, error) {
	cityMap, err := fileutils2.ReadCSVToMap(cityMapFile)
	if err != nil {
		return nil, fmt.Errorf("读取城市IP映射失败: %w", err)
	}
	selectedCityMap := maputils.PickRandMaps(cityMap, randCityNum)
	//fmt.Printf("selectedCityMap: %v\n", selectedCityMap)
	return selectedCityMap, nil
}

// queryDomainInfo 进行域名信息解析
func queryDomainInfo(dnsConfig *querydomain2.DNSQueryConfig, domainEntries []classify.TargetEntry) []*models.CheckInfo {
	// 创建DNS处理器并执行查询
	dnsProcessor := querydomain2.NewDNSProcessor(dnsConfig, &domainEntries)
	dnsResult := dnsProcessor.Process()

	//将dns查询结果合并到 CheckInfo 中去
	var checkInfos []*models.CheckInfo
	for _, domainEntry := range domainEntries {
		var checkInfo *models.CheckInfo
		//当存在dns查询结果时,补充
		if result, ok := (*dnsResult)[domainEntry.FMT]; ok && result != nil {
			checkInfo = querydomain2.PopulateDNSResult(domainEntry, result)
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
	// 定义配置并解析命令行参数
	var config Config

	// 使用 PassDoubleDash 选项强制使用 - 前缀
	parser := flags.NewParser(&config, flags.Default)
	parser.Name = "cdncheck"
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

	// 检查必要参数
	if config.Target == "" {
		fmt.Println("错误: 必须指定目标(-t, --target)")
		os.Exit(1)
	}

	var targets []string
	var err error

	// 判断 target 是文件路径还是直接输入的目标
	if fileutils2.IsFileExists(config.Target) {
		// 是文件路径，从文件加载目标
		targets, err = fileutils2.ReadTextToList(config.Target)
		if err != nil {
			fmt.Printf("加载目标文件失败: %v\n", err)
			os.Exit(1)
		}
	} else {
		// 不是文件路径，视为直接输入的目标 支持 按逗号分隔
		targets = strings.Split(config.Target, ",")
		for i, target := range targets {
			targets[i] = strings.TrimSpace(target)
		}
	}

	// 分类输入数据为 IP Domain InvalidEntries
	classifier := classify.ClassifyTargets(targets)

	//加载dns解析服务器配置文件，用于dns解析调用
	resolvers, err := LoadResolvers(config.ResolversFile, config.ResolversNum)
	if err != nil {
		os.Exit(1)
	}

	//加载本地EDNS城市IP信息
	randCities, err := LoadCityMap(config.CityMapFile, config.CityMapNum)
	if err != nil {
		os.Exit(1)
	}

	// 配置DNS查询参数
	dnsConfig := &querydomain2.DNSQueryConfig{
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
	cdnData := cdncheck2.NewEmptyCDNData()
	err = fileutils2.ReadJsonToStruct(config.SourceJson, cdnData)
	if err != nil {
		fmt.Printf("加载CDN源数据失败: %v\n", err)
		os.Exit(1)
	}

	//进行CDN CLOUD WAF 信息分析
	checkResults, err := cdncheck2.CheckCDNBatch(cdnData, checkInfos)
	if err != nil {
		fmt.Printf("CDN分析失败: %v\n", err)
		os.Exit(1)
	}

	// 处理输出详细程度
	var outputData interface{}
	switch strings.ToLower(config.OutputLevel) {
	case "quiet":
		// 仅输出不是CDN的fmt内容
		outputData = cdncheck2.GetNoCDNs(checkResults)
	case "detail":
		// 合并 checkResults 到 checkInfos
		outputData = cdncheck2.MergeCheckResultsToCheckInfos(checkInfos, checkResults)
	default:
		// default: 输出 checkResults
		outputData = checkResults
	}

	// 处理输出类型
	fileutils2.WriteOutputToFile(outputData, config.OutputType, config.OutputFile)
}
