package main

import (
	"cdnCheck/classify"
	"cdnCheck/domaininfo/querydomain"
	"cdnCheck/fileutils"
	"cdnCheck/input"
	"cdnCheck/ipinfo/queryip"
	"cdnCheck/maputils"
	"cdnCheck/models"
	"flag"
	"fmt"
	"os"
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
}

// parseFlags 解析命令行参数并返回配置
func parseFlags() *Config {
	config := &Config{}

	// 定义命令行参数
	flag.StringVar(&config.TargetFile, "target", "targets.txt", "目标文件路径")
	flag.StringVar(&config.OutputFile, "output", "", "输出结果文件路径（默认为目标文件名.results.json）")
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

	// 解析命令行参数
	flag.Parse()

	// 设置默认输出文件路径
	if config.OutputFile == "" {
		config.OutputFile = config.TargetFile + ".results.json"
	}

	return config
}

func main() {
	// 解析命令行参数
	config := parseFlags()

	// 检查目标文件是否存在
	if !fileutils.IsFileExists(config.TargetFile) {
		fmt.Printf("目标文件 [%v] 不存在\n", config.TargetFile)
		os.Exit(1)
	}

	//加载输入目标
	targets, err := input.LoadTargets(config.TargetFile)
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
	// 创建DNS处理器并执行查询
	dnsProcessor := querydomain.NewDNSProcessor(dnsConfig, &classifier.DomainEntries)
	dnsResult := dnsProcessor.Process()
	fmt.Printf(maputils.AnyToJsonStr(dnsResult))

	//将dns查询结果合并到 CheckInfo 中去
	var checkInfos []*models.CheckInfo
	for _, domainEntry := range classifier.DomainEntries {
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

	fmt.Printf(maputils.AnyToJsonStr(checkInfos))

	//将所有IP信息加入到[]checkInfo中
	ipDbConfig := &queryip.IpDbConfig{
		AsnIpv4Db:    config.AsnIpv4Db,
		AsnIpv6Db:    config.AsnIpv6Db,
		Ipv4LocateDb: config.Ipv4LocateDb,
		Ipv6LocateDb: config.Ipv6LocateDb,
	}

	// 初始化IP数据库引擎
	ipEngines, err := queryip.InitDBEngines(ipDbConfig)
	if err != nil {
		fmt.Printf("初始化数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer ipEngines.Close()

	//// 处理IP信息
	//if err := ipProcessor.ProcessCheckInfos(checkResults); err != nil {
	//	fmt.Printf("处理IP信息失败: %v\n", err)
	//}
	//
	//// 加载source.json配置文件
	//sourceData := models.NewEmptyCDNDataPointer()
	//if err := fileutils.ReadJsonToStruct(config.SourceJson, sourceData); err != nil {
	//	fmt.Printf("加载CDN源数据失败: %v\n", err)
	//	os.Exit(1)
	//}
	//
	//// 处理CDN信息
	//if err := ipProcessor.ProcessCDNInfos(checkResults, sourceData); err != nil {
	//	fmt.Printf("处理CDN信息失败: %v\n", err)
	//}
	//
	//// Step 8: 将结果写入文件
	//err = os.WriteFile(config.OutputFile, maputils.AnyToJsonBytes(checkResults), 0644)
	////将结果写入到CSV
	//sliceToMaps, err := maputils.ConvertStructSliceToMaps(checkResults)
	//if err == nil {
	//	fmt.Printf("%v", maputils.AnyToJsonStr(checkResults))
	//	csvOutputFile := config.OutputFile + ".csv"
	//	fileutils.WriteCSV(csvOutputFile, sliceToMaps, true)
	//} else {
	//	fmt.Errorf("Convert Struct Slice To Maps error: %v\n", err)
	//}
}
