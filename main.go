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

// IPInfo 存储IP解析的中间结果
type IPInfo struct {
	IPv4Locations []map[string]string
	IPv6Locations []map[string]string
	IPv4AsnInfos  []asndb.ASNInfo
	IPv6AsnInfos  []asndb.ASNInfo
}

// DBEngines 存储所有数据库引擎实例
type DBEngines struct {
	IPv6Engine *zxipv6wry.Ipv6Location
}

// DNSQueryConfig 存储DNS查询配置
type DNSQueryConfig struct {
	Resolvers  []string
	RandCities []map[string]string
	Timeout    time.Duration
}

// DNSQueryResult 存储DNS查询结果
type DNSQueryResult struct {
	DomainEntry maputils.TargetEntry
	DNSResult   *dnsquery.DNSResult
}

// DNSProcessor DNS查询处理器
type DNSProcessor struct {
	Config     *DNSQueryConfig
	Classifier *maputils.TargetClassifier
}

// IPProcessor IP信息查询处理器
type IPProcessor struct {
	Engines *DBEngines
	Config  *Config
}

// NewDNSProcessor 创建新的DNS处理器
func NewDNSProcessor(config *DNSQueryConfig, classifier *maputils.TargetClassifier) *DNSProcessor {
	return &DNSProcessor{
		Config:     config,
		Classifier: classifier,
	}
}

// NewIPProcessor 创建新的IP处理器
func NewIPProcessor(engines *DBEngines, config *Config) *IPProcessor {
	return &IPProcessor{
		Engines: engines,
		Config:  config,
	}
}

// Process 执行DNS查询并收集结果
func (p *DNSProcessor) Process() []*models.CheckInfo {
	var checkResults []*models.CheckInfo
	var wg sync.WaitGroup
	dnsResultChan := make(chan *models.CheckInfo, len(p.Classifier.Domains))

	// 启动并发查询
	for _, domainEntry := range p.Classifier.Domains {
		wg.Add(1)
		go processDomain(domainEntry, p.Config, dnsResultChan, &wg)
	}

	// 启动一个 goroutine 等待所有任务完成，并关闭 channel
	go func() {
		wg.Wait()
		close(dnsResultChan)
	}()

	// 接收DNS查询结果，实时输出，并保存到列表中
	for dnsResult := range dnsResultChan {
		checkResults = append(checkResults, dnsResult)
		finalInfo, _ := json.MarshalIndent(dnsResult, "", "  ")
		fmt.Println(string(finalInfo))
	}

	return checkResults
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

// queryDNS 执行DNS和EDNS查询
func queryDNS(domainEntry maputils.TargetEntry, config *DNSQueryConfig) (*DNSQueryResult, error) {
	domain := domainEntry.Fmt

	// 常规 DNS 查询
	dnsResults := dnsquery.QueryAllDNSWithMultiResolvers(domain, config.Resolvers, config.Timeout)
	dnsQueryResult := dnsquery.MergeDNSResults(dnsResults)

	// EDNS 查询
	eDNSQueryResults := dnsquery.EDNSQueryWithMultiCities(domain, config.Timeout, config.RandCities, false)
	if len(eDNSQueryResults) == 0 {
		fmt.Fprintf(os.Stderr, "EDNS 查询结果为空: %s\n", domain)
	} else {
		eDNSQueryResult := dnsquery.MergeEDNSResults(eDNSQueryResults)
		dnsQueryResult = dnsquery.MergeEDNSToDNS(eDNSQueryResult, dnsQueryResult)
	}

	// 优化DNS查询结果
	dnsquery.OptimizeDNSResult(&dnsQueryResult)

	return &DNSQueryResult{
		DomainEntry: domainEntry,
		DNSResult:   &dnsQueryResult,
	}, nil
}

// processDomain 处理单个域名查询任务
func processDomain(
	domainEntry maputils.TargetEntry,
	config *DNSQueryConfig,
	resultChan chan<- *models.CheckInfo,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	// 执行DNS查询
	queryResult, err := queryDNS(domainEntry, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DNS查询失败 [%s]: %v\n", domainEntry.Fmt, err)
		return
	}

	// 填充结果
	dnsResult := populateDNSResult(queryResult.DomainEntry, queryResult.DNSResult)

	// 发送结果到 channel
	resultChan <- dnsResult
}

// initDBEngines 初始化所有数据库引擎
func initDBEngines(config *Config) (*DBEngines, error) {
	// 初始化ASN数据库
	if err := asndb.InitMMDBConn(config.AsnIpv4Db, config.AsnIpv6Db); err != nil {
		return nil, fmt.Errorf("初始化ASN数据库失败: %w", err)
	}

	// 初始化IPv4数据库
	if err := qqwry.LoadDBFile(config.Ipv4LocateDb); err != nil {
		return nil, fmt.Errorf("初始化IPv4数据库失败: %w", err)
	}

	// 初始化IPv6数据库
	ipv6Engine, err := zxipv6wry.NewIPv6Location(config.Ipv6LocateDb)
	if err != nil {
		return nil, fmt.Errorf("初始化IPv6数据库失败: %w", err)
	}

	return &DBEngines{
		IPv6Engine: ipv6Engine,
	}, nil
}

// parseIPInfo 解析IP信息并返回中间结果
func parseIPInfo(ipv4s, ipv6s []string, engines *DBEngines) (*IPInfo, error) {
	info := &IPInfo{}

	// 使用通道来并发处理IP信息
	type ipv4Result struct {
		location map[string]string
		asn      asndb.ASNInfo
	}

	type ipv6Result struct {
		location map[string]string
		asn      asndb.ASNInfo
	}

	ipv4Chan := make(chan ipv4Result, len(ipv4s))
	ipv6Chan := make(chan ipv6Result, len(ipv6s))

	// 并发处理IPv4信息
	for _, ipv4 := range ipv4s {
		go func(ip string) {
			// 查询位置信息
			location, _ := qqwry.QueryIP(ip)
			locationToStr := qqwry.LocationToStr(*location)
			locationMap := map[string]string{ip: locationToStr}

			// 查询ASN信息
			asnInfo := asndb.FindASN(ip)

			ipv4Chan <- ipv4Result{
				location: locationMap,
				asn:      *asnInfo,
			}
		}(ipv4)
	}

	// 并发处理IPv6信息
	for _, ipv6 := range ipv6s {
		go func(ip string) {
			// 查询位置信息
			location := engines.IPv6Engine.Find(ip)
			locationMap := map[string]string{ip: location}

			// 查询ASN信息
			asnInfo := asndb.FindASN(ip)

			ipv6Chan <- ipv6Result{
				location: locationMap,
				asn:      *asnInfo,
			}
		}(ipv6)
	}

	// 收集IPv4结果
	for i := 0; i < len(ipv4s); i++ {
		result := <-ipv4Chan
		info.IPv4Locations = append(info.IPv4Locations, result.location)
		info.IPv4AsnInfos = append(info.IPv4AsnInfos, result.asn)
	}

	// 收集IPv6结果
	for i := 0; i < len(ipv6s); i++ {
		result := <-ipv6Chan
		info.IPv6Locations = append(info.IPv6Locations, result.location)
		info.IPv6AsnInfos = append(info.IPv6AsnInfos, result.asn)
	}

	return info, nil
}

// ProcessIPInfo 处理单个IP信息
func (p *IPProcessor) ProcessIPInfo(ipv4s, ipv6s []string) (*IPInfo, error) {
	return parseIPInfo(ipv4s, ipv6s, p.Engines)
}

// ProcessCheckInfo 处理单个CheckInfo的IP信息
func (p *IPProcessor) ProcessCheckInfo(checkInfo *models.CheckInfo) error {
	ipInfo, err := p.ProcessIPInfo(checkInfo.A, checkInfo.AAAA)
	if err != nil {
		return fmt.Errorf("处理IP信息失败: %w", err)
	}

	// 更新CheckInfo的IP信息
	checkInfo.Ipv6Asn = ipInfo.IPv6AsnInfos
	checkInfo.Ipv4Asn = ipInfo.IPv4AsnInfos
	checkInfo.Ipv4Locate = ipInfo.IPv4Locations
	checkInfo.Ipv6Locate = ipInfo.IPv6Locations

	return nil
}

// ProcessCheckInfos 批量处理CheckInfo列表
func (p *IPProcessor) ProcessCheckInfos(checkInfos []*models.CheckInfo) error {
	for _, checkInfo := range checkInfos {
		if err := p.ProcessCheckInfo(checkInfo); err != nil {
			fmt.Printf("处理CheckInfo失败: %v\n", err)
			continue
		}
	}
	return nil
}

// ProcessCDNInfo 处理CDN相关信息
func (p *IPProcessor) ProcessCDNInfo(checkInfo *models.CheckInfo, sourceData *models.CDNData) error {
	cnameList := checkInfo.CNAME
	// 合并 A 和 AAAA 记录
	ipList := maputils.UniqueMergeSlices(checkInfo.A, checkInfo.AAAA)
	// 合并 asn记录列表
	asnList := maputils.UniqueMergeAnySlices(
		asndb.GetUniqueOrgNumbers(checkInfo.Ipv4Asn),
		asndb.GetUniqueOrgNumbers(checkInfo.Ipv6Asn),
	)
	// 合并 IP定位列表
	ipLocateList := maputils.UniqueMergeSlices(
		maputils.GetMapsUniqueValues(checkInfo.Ipv4Locate),
		maputils.GetMapsUniqueValues(checkInfo.Ipv6Locate),
	)

	// 判断IP解析结果数量是否在CDN内
	checkInfo.IpSizeIsCdn, checkInfo.IpSize = cdncheck.IpsSizeIsCdn(ipList, 3)

	// 检查结果中的 CDN | WAF | CLOUD 情况
	checkInfo.IsCdn, checkInfo.CdnCompany = cdncheck.CheckCategory(sourceData.CDN, ipList, asnList, cnameList, ipLocateList)
	checkInfo.IsWaf, checkInfo.WafCompany = cdncheck.CheckCategory(sourceData.WAF, ipList, asnList, cnameList, ipLocateList)
	checkInfo.IsCloud, checkInfo.CloudCompany = cdncheck.CheckCategory(sourceData.CLOUD, ipList, asnList, cnameList, ipLocateList)

	return nil
}

// ProcessCDNInfos 批量处理CDN信息
func (p *IPProcessor) ProcessCDNInfos(checkInfos []*models.CheckInfo, sourceData *models.CDNData) error {
	for _, checkInfo := range checkInfos {
		if err := p.ProcessCDNInfo(checkInfo, sourceData); err != nil {
			fmt.Printf("处理CDN信息失败: %v\n", err)
			continue
		}
	}
	return nil
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

	// 配置DNS查询参数
	dnsConfig := &DNSQueryConfig{
		Resolvers:  resolvers,
		RandCities: randCities,
		Timeout:    5 * time.Second,
	}

	// 创建DNS处理器并执行查询
	dnsProcessor := NewDNSProcessor(dnsConfig, classifier)
	checkResults = dnsProcessor.Process()

	// 初始化数据库引擎
	engines, err := initDBEngines(config)
	if err != nil {
		fmt.Printf("初始化数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer asndb.CloseMMDBConn()

	// 创建IP处理器
	ipProcessor := NewIPProcessor(engines, config)

	// 处理IP信息
	if err := ipProcessor.ProcessCheckInfos(checkResults); err != nil {
		fmt.Printf("处理IP信息失败: %v\n", err)
	}

	// 加载source.json配置文件
	sourceData := models.NewEmptyCDNDataPointer()
	if err := fileutils.ReadJsonToStruct(config.SourceJson, sourceData); err != nil {
		fmt.Printf("加载CDN源数据失败: %v\n", err)
		os.Exit(1)
	}

	// 处理CDN信息
	if err := ipProcessor.ProcessCDNInfos(checkResults, sourceData); err != nil {
		fmt.Printf("处理CDN信息失败: %v\n", err)
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
