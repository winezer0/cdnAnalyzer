package dns_query

import (
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"net"
	"strings"
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

type ECSResult struct {
	IPs    []string
	CNAMEs []string
	Errors []string // 支持多个错误
}

func EDNSQuery(domain string, EDNSAddr string, dnsServer string) ECSResult {
	domain = dns.Fqdn(domain)

	c := new(dns.Client)
	m := newEDNSMessage(domain, EDNSAddr)

	in, _, err := c.Exchange(m, dnsServer)
	if err != nil {
		return ECSResult{
			Errors: []string{err.Error()},
		}
	}

	var ips []string
	var cnames []string

	for _, answer := range in.Answer {
		switch answer.Header().Rrtype {
		case dns.TypeCNAME:
			cnames = append(cnames, answer.(*dns.CNAME).Target)
		case dns.TypeA:
			ips = append(ips, answer.(*dns.A).A.String())
		}
	}

	return ECSResult{
		IPs:    ips,
		CNAMEs: cnames,
	}
}

func DoEcsQueryWithMultiCities(domain string, timeout time.Duration, cityNum int, queryCNAME bool) map[string]ECSResult {
	finalDomain := domain
	dnsServers := []string{"8.8.8.8:53"}

	if queryCNAME {
		// 获取 CNAME 链并取最后一个
		cnameChain, err := LookupCNAMEChain(domain, defaultServerAddress, timeout)
		if err != nil {
			fmt.Printf("Failed to lookup CNAME chain for %s: %v\n", domain, err)
		} else if len(cnameChain) > 0 {
			fmt.Printf("CNAME chain for %s found\n", cnameChain)
			finalDomain = cnameChain[len(cnameChain)-1]
		}
		// 获取 NS 服务器，并加上默认 DNS // 发现 ns1.a.shifen.com:53 查询原始域名结果是空的, 如果没有查询cname的话 最好为每个域名补充个默认dns服务器
		//nsServers, err := lookupNSServers(finalDomain, defaultServerAddress)
		//nsServers, err := LookupNSServers(finalDomain, defaultServerAddress, timeout)
		nsServers, err := LookupNSServersMutex(finalDomain, defaultServerAddress, timeout)
		if err != nil || len(nsServers) == 0 {
			fmt.Printf("Failed to lookup NS servers for %s, using fallback: 8.8.8.8:53\n", finalDomain)
		} else {
			dnsServers = append([]string{"8.8.8.8:53"}, NSServerAddPort(nsServers[0]))
		}
	}

	// 随机选择 几个 城市
	randCities := getRandomCities(cityMap, cityNum)
	if len(randCities) == 0 {
		fmt.Println("No cities selected, check cityMap or getRandomCities")
		return nil
	}

	// 并发执行查询
	var wg sync.WaitGroup
	resultsChan := make(chan map[string]ECSResult, len(randCities)*len(dnsServers))
	resultMap := make(map[string]ECSResult)
	var mu sync.Mutex // 用于并发安全写入 resultMap

	for _, dnsServer := range dnsServers {
		for city, ecsIP := range randCities {
			wg.Add(1)
			go func(city string, ecsIP string, dnsServer string) {
				defer wg.Done()

				result := EDNSQuery(finalDomain, ecsIP, dnsServer)

				key := fmt.Sprintf("%s@%s", city, dnsServer)

				mu.Lock()
				resultMap[key] = result
				mu.Unlock()

				resultsChan <- map[string]ECSResult{
					key: result,
				}
			}(city, ecsIP, dnsServer)
		}
	}

	// 等待所有协程完成
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// 可选：收集日志或其他用途的结果
	for range resultsChan {
		// 这里只是为了等待所有 goroutine 完成
	}

	return resultMap
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

// lookupNS 查询该域名的权威名称服务器（Name Server, NS）记录
func lookupNS(domain string, dnsServer string) (adders []string, err error) {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	client := &dns.Client{}
	resp, _, err := client.Exchange(msg, dnsServer)
	if err != nil {
		return nil, err
	}
	// 检查 Authority Section 是否有 SOA
	for _, rr := range resp.Ns {
		if soa, ok := rr.(*dns.SOA); ok {
			adders = append(adders, soa.Ns)
		}
	}
	for _, ans := range resp.Answer {
		if ns, ok := ans.(*dns.NS); ok {
			adders = append(adders, ns.Ns)
		}
	}
	if len(adders) == 0 {
		return nil, errors.New("no ns record")
	}
	return adders, nil
}

// lookupNSServers 是递归地查找给定域名的父域名（从最底层开始逐级向上）的名称服务器（NS记录），直到找到一个有效的NS记录或者遍历完所有可能的父域名
func lookupNSServers(domain, serverAddress string) (adders []string, err error) {
	dots := strings.Split(domain, ".")
	for i := 0; i < len(dots)-1; i++ {
		parent := strings.Join(dots[i:], ".")
		adders, err = lookupNS(parent, serverAddress)
		if err != nil {
			var DNSError *net.DNSError
			if errors.As(err, &DNSError) {
				continue
			}
			return nil, err
		}
		return adders, nil
	}
	return nil, errors.New("no ns record")
}

// LookupNSServersMutex 并发地对每个父级域名进行 NS 查询，并返回最长域名的结果
func LookupNSServersMutex(domain, dnsServer string, timeout time.Duration) ([]string, error) {
	// 规范化域名：去除首尾空格和点号
	domain = strings.TrimSpace(strings.Trim(domain, "."))
	domain = dns.Fqdn(domain) // 确保是 FQDN

	// 拆分域名标签
	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return nil, fmt.Errorf("invalid domain: %s", domain)
	}

	var wg sync.WaitGroup
	resultsChan := make(chan struct {
		domain string
		ns     []string
	}, len(labels))

	// 启动并发查询
	for i := 0; i < len(labels); i++ {
		parentDomain := strings.Join(labels[i:], ".")
		parentDomain = dns.Fqdn(parentDomain)

		wg.Add(1)
		go func(pd string) {
			defer wg.Done()

			nsRecords, err := queryDNS(pd, dnsServer, "NS", timeout)
			if err != nil {
				fmt.Printf("[DEBUG] Failed to lookup NS for %s: %v\n", pd, err)
				return
			}
			if len(nsRecords) > 0 {
				resultsChan <- struct {
					domain string
					ns     []string
				}{
					domain: pd,
					ns:     nsRecords,
				}
			}
		}(parentDomain)
	}

	// 等待所有协程完成
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// 收集所有成功的结果
	resultsMap := make(map[string][]string)
	for result := range resultsChan {
		resultsMap[result.domain] = result.ns
	}

	// 如果没有任何结果，返回错误
	if len(resultsMap) == 0 {
		return nil, fmt.Errorf("no ns record found for any parent domain of %s", domain)
	}

	// 查找最长域名（最具体）
	var longestDomain string
	var longestNS []string
	for domain, ns := range resultsMap {
		if len(domain) > len(longestDomain) {
			longestDomain = domain
			longestNS = ns
		}
	}

	fmt.Printf("[INFO] Returning NS from most specific domain: %s\n", longestDomain)
	return longestNS, nil
}
