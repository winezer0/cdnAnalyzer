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
	"fmt"
	"os"
	"sync"
	"time"
)

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

	dnsquery.OptimizeDNSResult(&dnsQueryResult)

	// 填充结果
	findResult := populateDNSResult(domainEntry, &dnsQueryResult)

	// 发送结果到 channel
	resultChan <- findResult
}

func main() {

	targetFile := "C:\\Users\\WINDOWS\\Desktop\\demo.txt" //需要进行查询的目标文件
	if !fileutils.IsFileExists(targetFile) {
		fmt.Printf("file [%v] Is Not File Exists .", targetFile)
		os.Exit(1)
	}
	//加载输入目标
	targets, err := loadTargets(targetFile)
	if err != nil {
		os.Exit(1)
	}
	//分类输入数据为 IP Domain Invalid
	classifier := classifyTargets(targets)
	//存储所有结果
	var allResults []*models.CheckInfo

	//加载dns解析服务器配置文件，用于dns解析调用
	resolversFile := "asset/resolvers.txt" //dns解析服务器
	resolversNum := 5                      //选择用于解析的最大DNS服务器数量 每个服务器将触发至少5次DNS解析
	resolvers, err := loadResolvers(resolversFile, resolversNum)
	if err != nil {
		os.Exit(1)
	}

	//加载本地EDNS城市IP信息
	cityMapFile := "asset/city_ip.csv" //用于 EDNS 查询时模拟城市的IP
	randCityNum := 5
	randCities, err := loadCityMap(cityMapFile, randCityNum)
	if err != nil {
		os.Exit(1)
	}

	// Step 5: 并发查询并收集DNS解析结果
	timeout := 5 * time.Second // DNS 查询超时时间
	var wg sync.WaitGroup
	resultChan := make(chan *models.CheckInfo, len(classifier.Domains))

	for _, domainEntry := range classifier.Domains {
		wg.Add(1)
		go processDomain(domainEntry, resolvers, randCities, timeout, resultChan, &wg)
	}

	// Step 6: 启动一个 goroutine 等待所有任务完成，并关闭 channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Step 7: 接收DNS查询结果，实时输出，并保存到列表中
	for result := range resultChan {
		allResults = append(allResults, result)

		// 实时输出结果
		finalInfo, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(finalInfo))
	}

	//初始化IP数据库信息
	asnIpv4Db := "asset/geolite2-asn-ipv4.mmdb" //IPv4的IP ASN数据库地址
	asnIpv6Db := "asset/geolite2-asn-ipv6.mmdb" //IPv6的IP ASN数据库地址
	//加载ASN db数据库
	asndb.InitMMDBConn(asnIpv4Db, asnIpv6Db)
	defer asndb.CloseMMDBConn()

	//加载IPv4数据库
	ipv4LocateDb := "asset/qqwry.ipdb"
	if err := qqwry.LoadDBFile(ipv4LocateDb); err != nil {
		panic(err)
	}

	//加载IPv6数据库
	ipv6LocateDb := "asset/zxipv6wry.db"
	ipv6Engine, _ := zxipv6wry.NewIPv6Location(ipv6LocateDb)

	//将 IP初始化到 allResults
	//将所有IP转换到
	for _, ipEntry := range classifier.IPs {
		ipResult := models.NewIPCheckInfo(ipEntry.Raw, ipEntry.Fmt, ipEntry.IsIPv4, ipEntry.FromUrl)
		allResults = append(allResults, ipResult)
	}

	//循环 allResults
	for _, findResult := range allResults {
		//对 每个 findResult.AipResult.AAAA 进行IP定位查询
		ipv4s := findResult.A
		ipv6s := findResult.AAAA

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

		findResult.Ipv6Asn = ipv6AsnInfos
		findResult.Ipv4Asn = ipv4AsnInfos
		findResult.Ipv4Locate = ipv4Locations
		findResult.Ipv6Locate = ipv6Locations
	}

	//加载source.json配置文件 检查当前结果是否存在CDN
	sourceJson := "asset/source.json"
	sourceData := models.NewEmptyCDNDataPointer()
	if err := fileutils.ReadJsonToStruct(sourceJson, sourceData); err != nil {
		panic(err)
	}
	fmt.Printf("%v", maputils.AnyToJsonStr(sourceData))

	for _, result := range allResults {
		cnames := result.CNAME
		// 合并 A 和 AAAA 记录
		allIPs := maputils.UniqueMergeSlices(result.A, result.AAAA)
		//合并 asn记录列表 需要处理后合并
		allAsns := maputils.UniqueMergeAnySlices(asndb.GetUniqueOrgNumbers(result.Ipv4Asn), asndb.GetUniqueOrgNumbers(result.Ipv6Asn))
		//合并 IP定位列表 需要处理后合并
		ipLocates := maputils.UniqueMergeSlices(maputils.GetMapsUniqueValues(result.Ipv4Locate), maputils.GetMapsUniqueValues(result.Ipv6Locate))
		//判断IP解析结果数量是否在CDN内
		ipSizeIsCdn, ipSize := cdncheck.IpsSizeIsCdn(allIPs, 3)
		fmt.Printf("ipSizeIsCdn: %v ipSize:%v\n", ipSizeIsCdn, ipSize)
		//检查结果中的CDN情况
		//判断cname是否在cdn内
		cnameIsCDN, cnameFindCdnCompany := cdncheck.KeysInMap(cnames, sourceData.CDN.CNAME)
		fmt.Printf("cnameIsCDN: %v cnameFindCdnCompany:%v\n", cnameIsCDN, cnameFindCdnCompany)
		// 判断IP是否在cdn内
		ipIsCDN, ipFindCdnCompany := cdncheck.IpsInMap(allIPs, sourceData.CDN.IP)
		fmt.Printf("ipIsCDN: %v ipFindCdnCompany:%v\n", ipIsCDN, ipFindCdnCompany)

		// 判断asn是否在cdn内
		asnIsCDN, asnFindCdnCompany := cdncheck.ASNsInMap(allAsns, sourceData.CDN.ASN)
		fmt.Printf("asnIsCDN: %v asnFindCdnCompany:%v\n", asnIsCDN, asnFindCdnCompany)

		//判断IP定位是否在CDN内
		ipLocateIsCDN, ipLocateFindCdnCompany := cdncheck.KeysInMap(ipLocates, sourceData.CDN.KEYS)
		fmt.Printf("ipLocateIsCDN: %v ipLocateFindCdnCompany:%v\n", ipLocateIsCDN, ipLocateFindCdnCompany)
		result.IpLocateIsCDN = ipLocateIsCDN

		result.IpSizeIsCdn = ipSizeIsCdn
		result.AsnIsCDN = asnIsCDN
		result.CnameIsCDN = cnameIsCDN
		result.IpIsCDN = ipIsCDN

		//判断cname是否在waf内
		result.CnameIsWAF, result.CnameFindCdnCompany = cdncheck.KeysInMap(cnames, sourceData.WAF.CNAME)
		// 判断IP是否在waf内
		result.IpIsWAF, result.IpFindWafCompany = cdncheck.IpsInMap(allIPs, sourceData.WAF.IP)
		// 判断asn是否在waf内
		result.AsnIsWAF, result.AsnFindWafCompany = cdncheck.ASNsInMap(allAsns, sourceData.WAF.ASN)
		//判断IP定位是否在WAF内
		result.IpLocateIsWAF, result.IpLocateFindWafCompany = cdncheck.KeysInMap(ipLocates, sourceData.WAF.KEYS)

		//判断cname是否在cloud内
		result.CnameIsCLOUD, result.CnameFindCloudCompany = cdncheck.KeysInMap(cnames, sourceData.CLOUD.CNAME)
		// 判断IP是否在cloud内
		result.IpIsCLOUD, result.IpFindCloudCompany = cdncheck.IpsInMap(allIPs, sourceData.CLOUD.IP)
		// 判断asn是否在cloud内
		result.AsnIsCLOUD, result.AsnFindCloudCompany = cdncheck.ASNsInMap(allAsns, sourceData.CLOUD.ASN)
		//判断IP定位是否在CLOUD内
		result.IpLocateIsCLOUD, result.IpLocateFindCloudCompany = cdncheck.KeysInMap(ipLocates, sourceData.CLOUD.KEYS)

		result.FinalIsCdn = ipSizeIsCdn || ipIsCDN || cnameIsCDN || asnIsCDN
	}

	// Step 8: 可选：将结果写入文件
	err = os.WriteFile(targetFile+".results.json", maputils.AnyToJsonBytes(allResults), 0644)
	//将结果写入到CSV
	sliceToMaps, err := maputils.ConvertStructSliceToMaps(allResults)
	if err == nil {
		fmt.Printf("%v", maputils.AnyToJsonStr(allResults))
		fileutils.WriteCSV(targetFile+".results.csv", sliceToMaps, true)
		//fileutils.WriteCSV2(targetFile+".results.csv", sliceToMaps, true, "a+", nil)
	} else {
		fmt.Errorf("Convert Struct Slice To Maps error: %v\n", err)
	}

}
