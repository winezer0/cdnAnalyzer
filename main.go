package main

import (
	"cdnCheck/dnsquery"
	"cdnCheck/fileutils"
	"cdnCheck/iplocate/asndb"
	"cdnCheck/iplocate/qqwry"
	"cdnCheck/iplocate/zxipv6wry"
	"cdnCheck/maputils"
	"cdnCheck/models"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	targetFile := "C:\\Users\\WINDOWS\\Desktop\\demo.txt" //需要进行查询的目标文件
	timeout := time.Second * 5                            //dns查询的超时时间配置
	resolversFile := "asset/resolvers.txt"                //dns解析服务器
	resolversNum := 5                                     //选择用于解析的最大DNS服务器数量 每个服务器将触发至少5次DNS解析
	cityMapFile := "asset/city_ip.csv"                    //用于 EDNS 查询时模拟城市的IP
	randCityNum := 5                                      //用于 EDNS 查询时模拟城市的IP数量，每个IP将触发至少一次EDNS查询
	asnIpv4Db := "asset/geolite2-asn-ipv4.mmdb"           //IPv4的IP ASN数据库地址
	asnIpv6Db := "asset/geolite2-asn-ipv6.mmdb"           //IPv6的IP ASN数据库地址

	//加载并分类 需要进行查询的目
	targets, err := fileutils.ReadTextToList(targetFile)
	if err != nil {
		fmt.Println("加载目标文件失败:", err)
		os.Exit(1)
	}
	fmt.Printf("load target from file: %v\n", targets)
	classifier := maputils.NewTargetClassifier()
	classifier.Classify(targets)
	classifier.Summary()

	//加载dns解析服务器配置文件，用于dns解析调用
	resolvers, err := fileutils.ReadTextToList(resolversFile)
	fmt.Printf("load resolvers: %v\n", len(resolvers))
	if err != nil {
		fmt.Println("加载DNS服务器失败:", err)
		os.Exit(1)
	}

	resolvers = maputils.PickRandList(resolvers, resolversNum)
	fmt.Printf("choise resolvers: %v\n", resolvers)

	//加载本地EDNS城市IP信息
	cityMap, err := fileutils.ReadCSVToMap(cityMapFile)
	if err != nil {
		fmt.Errorf("Failed to read cityMap CSV: %v\n", err)
	}
	// 随机选择 5 个城市
	randCities := maputils.PickRandMaps(cityMap, randCityNum)
	fmt.Printf("randCities: %v\n", randCities)

	//加载ASN db数据库 //TODO 在函数内实现自动初始化调用,避免多次加载
	asndb.InitMMDBConn(asnIpv4Db, asnIpv6Db)
	defer asndb.CloseMMDBConn()

	//加载IPv4数据库
	ipv4LocateDb := "asset/qqwry.ipdb"
	if err := qqwry.LoadFile(ipv4LocateDb); err != nil {
		panic(err)
	}

	//加载IPv6数据库 //TODO 在函数内实现自动初始化调用,避免多次加载
	ipv6LocateDb := "asset/zxipv6wry.db"
	ipv6Engine, _ := zxipv6wry.NewIPv6Location(ipv6LocateDb)

	//循环进行查询操作
	domains := classifier.Domains
	for _, domain := range domains {
		findResult := models.NewCheckResult(domain, domain)

		//进行常规 DNS 信息解析查询
		dnsResults := dnsquery.QueryAllDNSWithMultiResolvers(domain, resolvers, timeout)
		//合并多次DNS查询结果
		dnsQueryResult := dnsquery.MergeDNSResults(dnsResults)

		//进行 EDNS 信息查询
		eDNSQueryResults := dnsquery.EDNSQueryWithMultiCities(domain, timeout, randCities, false)
		if len(eDNSQueryResults) == 0 {
			fmt.Errorf("Expected non-empty EDNS QueryResults\n")
		}
		//合并多次EDNS查询结果
		eDNSQueryResult := dnsquery.MergeEDNSResults(eDNSQueryResults)

		//合并EDNS结果到DNS结果中去
		dnsQueryResult = dnsquery.MergeEDNSToDNS(eDNSQueryResult, dnsQueryResult)
		dnsquery.OptimizeDNSResult(&dnsQueryResult)

		//合并DNS EDNS结果到最终结果数据中
		findResult.A = append(findResult.A, dnsQueryResult.A...)
		findResult.AAAA = append(findResult.AAAA, dnsQueryResult.AAAA...)
		findResult.CNAME = append(findResult.CNAME, dnsQueryResult.CNAME...)
		findResult.NS = append(findResult.NS, dnsQueryResult.NS...)
		findResult.MX = append(findResult.MX, dnsQueryResult.MX...)
		findResult.TXT = append(findResult.TXT, dnsQueryResult.TXT...)

		//对DNS查询结果中的IPS数据进行IP定位信息查询
		ipv4s := findResult.A

		var ipv4AsnInfos []asndb.ASNInfo
		var ipv4Locations []qqwry.Location

		for _, ipv4 := range ipv4s {
			//查询Ipv4的ASN信息
			ipv4AsnInfo := asndb.FindASN(net.IP(ipv4))
			ipv4AsnInfos = append(ipv4AsnInfos, *ipv4AsnInfo)
			//查询Ipv4的Locate信息
			ipv4Location, err := qqwry.QueryIP(ipv4)
			if err != nil {
				print("IPv4 %s 定位查询失败：%v", ipv4, err)
			}
			ipv4Locations = append(ipv4Locations, *ipv4Location)
		}

		fmt.Printf("ipv4AsnInfos: %v\n", ipv4AsnInfos)
		fmt.Printf("ipv4Locations: %v\n", ipv4Locations)

		//进行IPV6数据查询
		ipv6s := findResult.AAAA
		var ipv6AsnInfos []asndb.ASNInfo
		var ipv6Locations []string //TODO 需要统一IPv6和IPv4的定位查询结果
		for _, ipv6 := range ipv6s {
			//查询Ipv6的ASN信息
			ipv6AsnInfo := asndb.FindASN(net.IP(ipv6))
			ipv6AsnInfos = append(ipv6AsnInfos, *ipv6AsnInfo)
			//查询Ipv6的Locate信息
			ipv6Location := ipv6Engine.Find(ipv6)
			ipv6Locations = append(ipv6Locations, ipv6Location)
		}
		fmt.Printf("ipv6AsnInfos: %v\n", ipv6AsnInfos)
		fmt.Printf("ipv6Location: %v\n", ipv6Locations)

		//输出DNS记录
		finalInfo, _ := json.MarshalIndent(findResult, "", "  ")
		fmt.Println(string(finalInfo))

		os.Exit(1)

	}

}
