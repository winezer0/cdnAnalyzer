package main

import (
	"cdnCheck/cdncheck"
	"cdnCheck/dnsquery"
	"cdnCheck/fileutils"
	"cdnCheck/iplocate/asndb"
	"cdnCheck/iplocate/qqwry"
	"cdnCheck/iplocate/zxipv6wry"
	"cdnCheck/maputils"
	"cdnCheck/models"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"
)

// Config 存储程序配置
type Config struct {
	TargetFile    string
	OutputFile    string
	ResolversFile string
	ResolversNum  int
	CityMapFile   string
	RandCityNum   int
	AsnIpv4Db     string
	AsnIpv6Db     string
	Ipv4LocateDb  string
	Ipv6LocateDb  string
	SourceJson    string
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
	flag.IntVar(&config.RandCityNum, "city-num", 5, "随机选择的城市数量")
	flag.StringVar(&config.AsnIpv4Db, "asn-ipv4", "asset/geolite2-asn-ipv4.mmdb", "IPv4 ASN数据库路径")
	flag.StringVar(&config.AsnIpv6Db, "asn-ipv6", "asset/geolite2-asn-ipv6.mmdb", "IPv6 ASN数据库路径")
	flag.StringVar(&config.Ipv4LocateDb, "ipv4-db", "asset/qqwry.ipdb", "IPv4地理位置数据库路径")
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

func loadTargets(targetFile string) ([]string, error) {
	targets, err := fileutils.ReadTextToList(targetFile)
	if err != nil {
		return nil, fmt.Errorf("加载目标文件失败: %w", err)
	} else {
		fmt.Printf("load target from file: %v\n", targets)
	}
	return targets, nil
}

func classifyTargets(targets []string) *maputils.TargetClassifier {
	classifier := maputils.NewTargetClassifier()
	classifier.Classify(targets)
	classifier.Summary()
	return classifier
}

func loadResolvers(resolversFile string, resolversNum int) ([]string, error) {
	resolvers, err := fileutils.ReadTextToList(resolversFile)
	if err != nil {
		return nil, fmt.Errorf("加载DNS服务器失败: %w", err)
	}
	resolvers = maputils.PickRandList(resolvers, resolversNum)
	fmt.Printf("choise resolvers: %v\n", resolvers)
	return resolvers, nil
}

func loadCityMap(cityMapFile string, randCityNum int) ([]map[string]string, error) {
	cityMap, err := fileutils.ReadCSVToMap(cityMapFile)
	if err != nil {
		return nil, fmt.Errorf("读取城市IP映射失败: %w", err)
	}
	randCities := maputils.PickRandMaps(cityMap, randCityNum)
	fmt.Printf("randCities: %v\n", randCities)
	return randCities, nil
}

// populateDNSResult 将 DNS 查询结果填充到 CheckInfo 中
func populateDNSResult(domainEntry maputils.TargetEntry, query *dnsquery.DNSResult) *models.CheckInfo {
	result := models.NewDomainCheckInfo(domainEntry.Raw, domainEntry.Fmt, domainEntry.FromUrl)

	// 逐个复制 DNS 记录
	result.A = append(result.A, query.A...)
	result.AAAA = append(result.AAAA, query.AAAA...)
	result.CNAME = append(result.CNAME, query.CNAME...)
	result.NS = append(result.NS, query.NS...)
	result.MX = append(result.MX, query.MX...)
	result.TXT = append(result.TXT, query.TXT...)

	return result
}

// 处理单个 domain 查询任务
func processDomain(
	domainEntry maputils.TargetEntry,
	resolvers []string,
	randCities []map[string]string,
	timeout time.Duration,
	resultChan chan<- *models.CheckInfo,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	domain := domainEntry.Fmt

	// 常规 DNS 查询
	dnsResults := dnsquery.QueryAllDNSWithMultiResolvers(domain, resolvers, timeout)
	dnsQueryResult := dnsquery.MergeDNSResults(dnsResults)

	// EDNS 查询
	eDNSQueryResults := dnsquery.EDNSQueryWithMultiCities(domain, timeout, randCities, false)
	if len(eDNSQueryResults) == 0 {
		fmt.Fprintf(os.Stderr, "EDNS 查询结果为空: %s\n", domain)
	} else {
		eDNSQueryResult := dnsquery.MergeEDNSResults(eDNSQueryResults)
		dnsQueryResult = dnsquery.MergeEDNSToDNS(eDNSQueryResult, dnsQueryResult)
	}
	//优化dns查询结果
	dnsquery.OptimizeDNSResult(&dnsQueryResult)
	// 填充结果
	dnsResult := populateDNSResult(domainEntry, &dnsQueryResult)
	// 发送结果到 channel
	resultChan <- dnsResult
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
	targets, err := loadTargets(config.TargetFile)
	if err != nil {
		os.Exit(1)
	}
	//分类输入数据为 IP Domain Invalid
	classifier := classifyTargets(targets)
	//存储所有结果
	var checkResults []*models.CheckInfo

	//加载dns解析服务器配置文件，用于dns解析调用
	resolvers, err := loadResolvers(config.ResolversFile, config.ResolversNum)
	if err != nil {
		os.Exit(1)
	}

	//加载本地EDNS城市IP信息
	randCities, err := loadCityMap(config.CityMapFile, config.RandCityNum)
	if err != nil {
		os.Exit(1)
	}

	// Step 5: 并发查询并收集DNS解析结果
	timeout := 5 * time.Second // DNS 查询超时时间
	var wg sync.WaitGroup
	dnsResultChan := make(chan *models.CheckInfo, len(classifier.Domains))

	for _, domainEntry := range classifier.Domains {
		wg.Add(1)
		go processDomain(domainEntry, resolvers, randCities, timeout, dnsResultChan, &wg)
	}

	// Step 6: 启动一个 goroutine 等待所有任务完成，并关闭 channel
	go func() {
		wg.Wait()
		close(dnsResultChan)
	}()

	// Step 7: 接收DNS查询结果，实时输出，并保存到列表中
	for dnsResult := range dnsResultChan {
		checkResults = append(checkResults, dnsResult)
		finalInfo, _ := json.MarshalIndent(dnsResult, "", "  ")
		fmt.Println(string(finalInfo))
	}

	//初始化IP数据库信息
	//加载ASN db数据库
	asndb.InitMMDBConn(config.AsnIpv4Db, config.AsnIpv6Db)
	defer asndb.CloseMMDBConn()

	//加载IPv4数据库
	if err := qqwry.LoadDBFile(config.Ipv4LocateDb); err != nil {
		panic(err)
	}

	//加载IPv6数据库
	ipv6Engine, _ := zxipv6wry.NewIPv6Location(config.Ipv6LocateDb)

	//将 IP初始化到 checkResults
	//将所有IP转换到
	for _, ipEntry := range classifier.IPS {
		ipResult := models.NewIPCheckInfo(ipEntry.Raw, ipEntry.Fmt, ipEntry.IsIPv4, ipEntry.FromUrl)
		checkResults = append(checkResults, ipResult)
	}

	//循环 checkResults
	for _, checkResult := range checkResults {
		//对 每个 checkResult.AipResult.AAAA 进行IP定位查询
		ipv4s := checkResult.A
		ipv6s := checkResult.AAAA

		//查询Ipv4的IP定位信息
		var ipv4Locations []map[string]string
		for _, ipv4 := range ipv4s {
			ipv4Location, _ := qqwry.QueryIP(ipv4)
			locationToStr := qqwry.LocationToStr(*ipv4Location)
			fmt.Printf("ipv4Locations: %v\n", locationToStr)
			ipv4Locations = append(ipv4Locations, map[string]string{ipv4: locationToStr})
		}

		//查询IPV6的IP定位信息
		var ipv6Locations []map[string]string
		for _, ipv6 := range ipv6s {
			//查询Ipv6的Locate信息
			ipv6Location := ipv6Engine.Find(ipv6)
			fmt.Printf("ipv6Location: %v\n", ipv6Location)
			ipv6Locations = append(ipv6Locations, map[string]string{ipv6: ipv6Location})
		}

		//查询Ipv4的ASN信息
		var ipv4AsnInfos []asndb.ASNInfo
		for _, ipv4 := range ipv4s {
			fmt.Printf("ipv4: %v\n", ipv4)
			ipv4AsnInfo := asndb.FindASN(ipv4)
			asndb.PrintASNInfo(ipv4AsnInfo)
			ipv4AsnInfos = append(ipv4AsnInfos, *ipv4AsnInfo)
		}

		//查询Ipv6的ASN信息
		var ipv6AsnInfos []asndb.ASNInfo
		for _, ipv6 := range ipv6s {
			//查询Ipv6的ASN信息
			fmt.Printf("ipv6: %v\n", ipv6)
			ipv6AsnInfo := asndb.FindASN(ipv6)
			asndb.PrintASNInfo(ipv6AsnInfo)
			ipv6AsnInfos = append(ipv6AsnInfos, *ipv6AsnInfo)
		}

		checkResult.Ipv6Asn = ipv6AsnInfos
		checkResult.Ipv4Asn = ipv4AsnInfos
		checkResult.Ipv4Locate = ipv4Locations
		checkResult.Ipv6Locate = ipv6Locations
	}

	//加载source.json配置文件 检查当前结果是否存在CDN
	sourceData := models.NewEmptyCDNDataPointer()
	if err := fileutils.ReadJsonToStruct(config.SourceJson, sourceData); err != nil {
		panic(err)
	}
	fmt.Printf("%v", maputils.AnyToJsonStr(sourceData))

	for _, checkResult := range checkResults {
		cnameList := checkResult.CNAME
		// 合并 A 和 AAAA 记录
		ipList := maputils.UniqueMergeSlices(checkResult.A, checkResult.AAAA)
		//合并 asn记录列表 需要处理后合并
		asnList := maputils.UniqueMergeAnySlices(asndb.GetUniqueOrgNumbers(checkResult.Ipv4Asn), asndb.GetUniqueOrgNumbers(checkResult.Ipv6Asn))
		//合并 IP定位列表 需要处理后合并
		ipLocateList := maputils.UniqueMergeSlices(maputils.GetMapsUniqueValues(checkResult.Ipv4Locate), maputils.GetMapsUniqueValues(checkResult.Ipv6Locate))
		//判断IP解析结果数量是否在CDN内
		checkResult.IpSizeIsCdn, checkResult.IpSize = cdncheck.IpsSizeIsCdn(ipList, 3)
		//检查结果中的 CDN | WAF | CLOUD 情况
		checkResult.IsCdn, checkResult.CdnCompany = cdncheck.CheckCategory(sourceData.CDN, ipList, asnList, cnameList, ipLocateList)
		checkResult.IsWaf, checkResult.WafCompany = cdncheck.CheckCategory(sourceData.WAF, ipList, asnList, cnameList, ipLocateList)
		checkResult.IsCloud, checkResult.CloudCompany = cdncheck.CheckCategory(sourceData.CLOUD, ipList, asnList, cnameList, ipLocateList)
	}

	// Step 8: 将结果写入文件
	err = os.WriteFile(config.OutputFile, maputils.AnyToJsonBytes(checkResults), 0644)
	//将结果写入到CSV
	sliceToMaps, err := maputils.ConvertStructSliceToMaps(checkResults)
	if err == nil {
		fmt.Printf("%v", maputils.AnyToJsonStr(checkResults))
		csvOutputFile := config.OutputFile + ".csv"
		fileutils.WriteCSV(csvOutputFile, sliceToMaps, true)
	} else {
		fmt.Errorf("Convert Struct Slice To Maps error: %v\n", err)
	}
}
