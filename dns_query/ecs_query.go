package dns_query

import (
	"github.com/miekg/dns"
	"net"
	"sync"
	"time"
)

var defaultServerAddress = GetSystemDefaultAddress()

var cityMap = map[string]string{
	"北京市 移动":          "39.156.128.0",
	"北京市 联通":          "114.240.0.0",
	"北京市 电信":          "1.202.0.0",
	"北京市 教育网":        "58.200.0.0",
	"黑龙江 哈尔滨 移动":   "211.137.243.0",
	"黑龙江 哈尔滨 联通":   "113.0.0.0",
	"黑龙江 哈尔滨 电信":   "42.100.0.0",
	"黑龙江 哈尔滨 教育网": "58.194.0.0",
	"吉林 长春 移动":       "39.134.160.0",
	"吉林 长春 联通":       "119.48.0.0",
	"吉林 长春 电信":       "36.48.0.0",
	"吉林 长春 教育网":     "59.72.0.0",
	"辽宁 沈阳 移动":       "211.137.32.0",
	"辽宁 沈阳 联通":       "113.224.0.0",
	"辽宁 沈阳 电信":       "182.200.0.0",
	"辽宁 沈阳 教育网":     "219.216.0.0",
	"上海 移动":            "211.137.243.0",
	"上海 联通":            "220.248.0.0",
	"上海 电信":            "101.80.0.0",
	"上海 教育网":          "58.194.0.0",
	"天津 天津 移动":       "211.137.160.1",
	"天津 天津 联通":       "61.181.81.1",
	"天津 天津 电信":       "221.238.6.1",
	"天津 天津 教育网":     "58.207.63.253",
	"重庆 重庆 移动":       "218.201.39.204",
	"重庆 重庆 联通":       "221.5.255.1",
	"重庆 重庆 电信":       "218.70.65.254",
	"重庆 重庆 教育网":     "202.202.216.1",
	"河北 石家庄 移动":     "218.207.75.1",
	"河北 石家庄 联通":     "221.192.1.1",
	"河北 石家庄 电信":     "123.180.0.200",
	"河北 石家庄 教育网":   "202.206.232.1",
	"山西 太原 移动":       "211.142.24.17",
	"山西 太原 联通":       "221.204.253.1",
	"山西 太原 电信":       "219.149.144.1",
	"山西 太原 教育网":     "202.207.130.1",
	"广东 广州 移动":       "211.139.145.34",
	"广东 广州 联通":       "211.95.193.69",
	"广东 广州 电信":       "58.61.200.1",
	"广东 广州 教育网":     "202.116.64.8",
	"江苏 南京 移动":       "120.195.118.1",
	"江苏 南京 联通":       "218.104.118.33",
	"江苏 南京 电信":       "58.212.24.1",
	"江苏 南京 教育网":     "202.119.32.7",
	// 美国（硅谷核心区域）
	"美国 加利福尼亚 圣何塞 AWS云":         "54.240.196.0",
	"美国 加利福尼亚 旧金山 Google云":      "8.8.8.8",
	"美国 弗吉尼亚 阿什本 Equinix数据中心": "198.32.118.0",

	// 日本
	"日本 东京 NTT通信":  "203.178.128.0",
	"日本 大阪 KDDI网络": "111.87.128.0",
	"日本 东京 软银集团": "220.100.0.0",

	// 德国
	"德国 法兰克福 Deutsche Telekom": "62.157.0.0",
	"德国 柏林 1&1 Ionos":            "82.165.0.0",
	"德国 慕尼黑 Hetzner数据中心":    "88.198.0.0",

	// 印度
	"印度 孟买 Reliance Jio":   "115.242.0.0",
	"印度 班加罗尔 Airtel宽带": "125.16.0.0",
	"印度 海得拉巴 Tata通信":   "14.140.0.0",

	// 俄罗斯
	"俄罗斯 莫斯科 Rostelecom":  "95.165.0.0",
	"俄罗斯 圣彼得堡 MegaFon":   "31.173.0.0",
	"俄罗斯 新西伯利亚 Beeline": "31.130.0.0",

	// 韩国
	"韩国 首尔 SK Telecom": "211.174.0.0",
	"韩国 釜山 KT通信":     "175.192.0.0",
	"韩国 仁川 LG Uplus":   "106.240.0.0",

	// 英国
	"英国 伦敦 BT集团":           "82.132.0.0",
	"英国 曼彻斯特 Virgin Media": "62.255.128.0",
	"英国 剑桥 ARM总部网络":      "193.130.104.0",

	// 新加坡
	"新加坡 新加坡-市区 Singtel":   "121.6.0.0",
	"新加坡 裕廊东 StarHub":        "58.185.0.0",
	"新加坡 榜鹅 DigitalOcean节点": "128.199.0.0",

	// 巴西
	"巴西 圣保罗 Vivo电信":      "189.100.0.0",
	"巴西 里约热内卢 Claro网络": "177.128.0.0",
	"巴西 巴西利亚 Oi通信":      "201.95.0.0",

	// 加拿大
	"加拿大 多伦多 Rogers通信":    "24.226.0.0",
	"加拿大 温哥华 Telus网络":     "207.6.0.0",
	"加拿大 蒙特利尔 Bell Canada": "64.230.0.0",
}

// DoEcsQueryWithMultiCities 并发进行多个地理位置的EDNS查询，返回每个位置的原始解析结果
func DoEcsQueryWithMultiCities(domain string, timeout time.Duration) map[string][]string {
	domain = dns.Fqdn(domain)

	// 获取 CNAME 链并取最后一个
	lastCNAME := domain
	cnameChain, _ := LookupCNAMEChain(domain, defaultServerAddress, timeout)
	if len(cnameChain) > 0 {
		lastCNAME = cnameChain[len(cnameChain)-1]
	}

	// 获取 NS 服务器
	nsServer := "8.8.8.8:53"
	nsServers, _ := LookupNSServers(lastCNAME, defaultServerAddress, timeout)
	if len(nsServers) > 0 {
		nsServer = nsServers[0]
	}

	// 随机选择 5 个城市
	randCities := getRandomCities(cityMap, 5)

	// 并发执行查询
	var wg sync.WaitGroup
	resultsChan := make(chan map[string][]string, len(randCities))
	errorChan := make(chan error, len(randCities))

	for city, ecsIP := range randCities {
		wg.Add(1)
		go func(city string, ecsIP string) {
			defer wg.Done()
			addresses, err := EDNSQuery(domain, ecsIP, NSServerAddPort(nsServer))
			if err != nil {
				errorChan <- err
				return
			}
			resultsChan <- map[string][]string{
				city: addresses,
			}
		}(city, ecsIP)
	}

	// 等待所有协程完成
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorChan)
	}()

	// 收集结果
	queryResults := make(map[string][]string)
	for res := range resultsChan {
		for city, records := range res {
			queryResults[city] = append(queryResults[city], records...)
		}
	}

	return queryResults
}

// EDNSQuery 使用 EDNS 模拟地域，来查询解析的IP地址，
// 需要 DNSServer 支持 EDNS0SUBNET，目前国内主流 DNS 均不支持该功能
// 通过 NS 记录查询得到的权威 DNS 服务器，通常会支持该功能
func EDNSQuery(domain string, EDNSAddr string, dnsServer string) (adders []string, err error) {
	c := new(dns.Client)
	m := newEDNSMessage(domain, EDNSAddr)
	in, _, err := c.Exchange(m, dnsServer) //注意:要选择支持自定义 EDNS 的DNS 或者是 目标NS服务器  国内DNS大部分不支持自定义 EDNS 数据
	if err != nil {
		return nil, err
	}
	for _, answer := range in.Answer {
		if answer.Header().Rrtype == dns.TypeCNAME {
			adders = append(adders, answer.(*dns.CNAME).Target)
		}
		if answer.Header().Rrtype == dns.TypeA {
			adders = append(adders, answer.(*dns.A).A.String())
		}
	}
	return adders, nil
}

// newEDNSMessage 创建并返回一个包含 EDNS（扩展 DNS）选项的 DNS 查询消息
// 参数： domain：要查询的域名。 EDNSAddr：用于设置 EDNS0_SUBNET 选项的 IP 地址。
// 返回值： 返回一个指向 dns_query.Msg 结构体的指针，表示生成的 DNS 查询消息。
func newEDNSMessage(domain, EDNSAddr string) *dns.Msg {
	e := new(dns.EDNS0_SUBNET) //EDNS
	e.Code = dns.EDNS0SUBNET
	e.Family = 1
	e.SourceNetmask = 24
	e.Address = net.ParseIP(EDNSAddr).To4()
	e.SourceScope = 0

	o := new(dns.OPT)
	o.Hdr.Name = "."
	o.Hdr.Rrtype = dns.TypeOPT
	o.Option = append(o.Option, e)
	o.SetUDPSize(dns.DefaultMsgSize)

	m := new(dns.Msg)
	m.SetQuestion(domain, dns.TypeA)
	m.Extra = append(m.Extra, o)
	return m
}
