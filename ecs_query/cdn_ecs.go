package ecs_query

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"math/big"
	"net"
	"strings"
)

var defaultServerAddress = GetSystemDefaultAddress()

var cityMap = map[string]string{
	"北京市 移动":      "39.156.128.0",
	"北京市 联通":      "114.240.0.0",
	"北京市 电信":      "1.202.0.0",
	"北京市 教育网":     "58.200.0.0",
	"黑龙江 哈尔滨 移动":  "211.137.243.0",
	"黑龙江 哈尔滨 联通":  "113.0.0.0",
	"黑龙江 哈尔滨 电信":  "42.100.0.0",
	"黑龙江 哈尔滨 教育网": "58.194.0.0",
	"吉林 长春 移动":    "39.134.160.0",
	"吉林 长春 联通":    "119.48.0.0",
	"吉林 长春 电信":    "36.48.0.0",
	"吉林 长春 教育网":   "59.72.0.0",
	"辽宁 沈阳 移动":    "211.137.32.0",
	"辽宁 沈阳 联通":    "113.224.0.0",
	"辽宁 沈阳 电信":    "182.200.0.0",
	"辽宁 沈阳 教育网":   "219.216.0.0",
	"上海 移动":       "211.137.243.0",
	"上海 联通":       "220.248.0.0",
	"上海 电信":       "101.80.0.0",
	"上海 教育网":      "58.194.0.0",
	"天津 天津 移动":    "211.137.160.1",
	"天津 天津 联通":    "61.181.81.1",
	"天津 天津 电信":    "221.238.6.1",
	"天津 天津 教育网":   "58.207.63.253",
	"重庆 重庆 移动":    "218.201.39.204",
	"重庆 重庆 联通":    "221.5.255.1",
	"重庆 重庆 电信":    "218.70.65.254",
	"重庆 重庆 教育网":   "202.202.216.1",
	"河北 石家庄 移动":   "218.207.75.1",
	"河北 石家庄 联通":   "221.192.1.1",
	"河北 石家庄 电信":   "123.180.0.200",
	"河北 石家庄 教育网":  "202.206.232.1",
	"山西 太原 移动":    "211.142.24.17",
	"山西 太原 联通":    "221.204.253.1",
	"山西 太原 电信":    "219.149.144.1",
	"山西 太原 教育网":   "202.207.130.1",
	"广东 广州 移动":    "211.139.145.34",
	"广东 广州 联通":    "211.95.193.69",
	"广东 广州 电信":    "58.61.200.1",
	"广东 广州 教育网":   "202.116.64.8",
	"江苏 南京 移动":    "120.195.118.1",
	"江苏 南京 联通":    "218.104.118.33",
	"江苏 南京 电信":    "58.212.24.1",
	"江苏 南京 教育网":   "202.119.32.7",
	// 美国（硅谷核心区域）
	"美国 加利福尼亚 圣何塞 AWS云":       "54.240.196.0",
	"美国 加利福尼亚 旧金山 Google云":    "8.8.8.8",
	"美国 弗吉尼亚 阿什本 Equinix数据中心": "198.32.118.0",

	// 日本
	"日本 东京 NTT通信":  "203.178.128.0",
	"日本 大阪 KDDI网络": "111.87.128.0",
	"日本 东京 软银集团":   "220.100.0.0",

	// 德国
	"德国 法兰克福 Deutsche Telekom": "62.157.0.0",
	"德国 柏林 1&1 Ionos":          "82.165.0.0",
	"德国 慕尼黑 Hetzner数据中心":       "88.198.0.0",

	// 印度
	"印度 孟买 Reliance Jio": "115.242.0.0",
	"印度 班加罗尔 Airtel宽带":   "125.16.0.0",
	"印度 海得拉巴 Tata通信":     "14.140.0.0",

	// 俄罗斯
	"俄罗斯 莫斯科 Rostelecom": "95.165.0.0",
	"俄罗斯 圣彼得堡 MegaFon":   "31.173.0.0",
	"俄罗斯 新西伯利亚 Beeline":  "31.130.0.0",

	// 韩国
	"韩国 首尔 SK Telecom": "211.174.0.0",
	"韩国 釜山 KT通信":       "175.192.0.0",
	"韩国 仁川 LG Uplus":   "106.240.0.0",

	// 英国
	"英国 伦敦 BT集团":           "82.132.0.0",
	"英国 曼彻斯特 Virgin Media": "62.255.128.0",
	"英国 剑桥 ARM总部网络":        "193.130.104.0",

	// 新加坡
	"新加坡 新加坡-市区 Singtel":    "121.6.0.0",
	"新加坡 裕廊东 StarHub":       "58.185.0.0",
	"新加坡 榜鹅 DigitalOcean节点": "128.199.0.0",

	// 巴西
	"巴西 圣保罗 Vivo电信":    "189.100.0.0",
	"巴西 里约热内卢 Claro网络": "177.128.0.0",
	"巴西 巴西利亚 Oi通信":     "201.95.0.0",

	// 加拿大
	"加拿大 多伦多 Rogers通信":     "24.226.0.0",
	"加拿大 温哥华 Telus网络":      "207.6.0.0",
	"加拿大 蒙特利尔 Bell Canada": "64.230.0.0",
}

// 获取随机城市
func getRandomCities(cityMap map[string]string, count int) map[string]string {
	if len(cityMap) <= count {
		return cityMap
	}

	//获取键切片
	keys := make([]string, 0, len(cityMap))
	for key := range cityMap {
		keys = append(keys, key)
	}

	//存储随机城市数据
	selectedCities := make(map[string]string, count)
	for i := 0; i < count; i++ {
		index, _ := rand.Int(rand.Reader, big.NewInt(int64(len(keys))))
		randIndex := index.Int64()
		selectedCities[keys[randIndex]] = cityMap[keys[randIndex]]
		// 通过切片发移除指定索引位置元素,防止被多次选择
		keys = append(keys[:randIndex], keys[randIndex+1:]...)
	}
	return selectedCities
}

// DoCDNCheck 进行CDN检查
func DoCDNCheck(domain string) (isCDN bool, err error) {
	domain = fixDomain(domain)
	cname, err := LookupCNAME(domain)
	if err != nil {
		return true, err
	}

	nsServers, err := LookupNSServer(cname)
	if err != nil {
		return true, err
	}
	nsServer := nsServers[0]
	return doCDNCheck(cname, nsServer)
}

// doCDNCheck 检查给定的域名是否使用CDN 通过模拟多个地理位置发送DNS查询来获取该域名解析出的IP地址，根据结果数量判断
func doCDNCheck(domain, serverAddr string) (isCDN bool, err error) {
	addrsMap := map[string]struct{}{}
	times := 0
	maxTimeout := 0
	for _, cityAddr := range cityMap {
		query, err := EDNSQuery(dns.Fqdn(domain), cityAddr, ToServerAddr(serverAddr))
		if err != nil {
			var DNSError *net.DNSError
			if errors.Is(err, DNSError) {
				return true, err
			}
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				maxTimeout++
				if maxTimeout > 3 {
					return true, err
				}
			}
			continue
		}
		for _, addr := range query {
			addrsMap[addr] = struct{}{}
		}
		// 如果已经找到大于5个不同的IP地址解析，则认为该IP地址使用了CDN
		if len(addrsMap) > 5 {
			return true, nil
		}
		// 如果查询次数大于5次，查询到的IP地址仍然小于3个，则提前认为该IP地址没有使用CDN
		if times >= 5 && len(addrsMap) <= 3 {
			return false, nil
		}
		times++
	}
	return false, nil
}

// DoEcsQuery 进行CDN查询IP
func DoEcsQuery(domain string) (adders []string, err error) {
	domain = fixDomain(domain)
	cname, err := LookupCNAME(domain)
	if err != nil {
		return nil, err
	}
	nsServers, err := LookupNSServer(cname)
	if err != nil {
		return nil, err
	}
	nsServer := nsServers[0]
	return doEcsQuery(cname, nsServer)
}

// doEcsQuery 进行CDN查询IP
func doEcsQuery(domain, serverAddr string) (adders []string, err error) {
	addersMap := map[string]struct{}{}
	times := 0
	maxTimeout := 0

	// 如果 cityMap 数量大于5，随机获取5个城市
	//var selectedCities map[string]string
	//selectedCities = getRandomCities(cityMap, 5)
	var selectedCities = getRandomCities(cityMap, 5)
	for _, cityAddr := range selectedCities {
		queryResult, err := EDNSQuery(dns.Fqdn(domain), cityAddr, ToServerAddr(serverAddr))
		if err != nil {
			var DNSError *net.DNSError
			if errors.Is(err, DNSError) {
				return nil, err
			}
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				maxTimeout++
				if maxTimeout > 3 {
					return nil, err
				}
			}
			continue
		}

		for _, addr := range queryResult {
			addersMap[addr] = struct{}{}
		}

		// 如果已经找到大于5个不同的IP地址解析，则认为该域名使用了CDN
		if len(addersMap) > 5 {
			break
		}

		times++
	}

	// 将 addersMap 中的 IP 地址转换为切片
	for addr := range addersMap {
		adders = append(adders, addr)
	}

	return adders, nil
}

// ToServerAddr 为dns服务器IP补充53端口
func ToServerAddr(s string) string {
	s = strings.TrimSuffix(strings.ToLower(s), ".")
	if strings.Contains(s, ":") {
		return s
	}
	return s + ":53"

}

// LookupNSServer 是递归地查找给定域名的父域名（从最底层开始逐级向上）的名称服务器（NS记录），直到找到一个有效的NS记录或者遍历完所有可能的父域名
func LookupNSServer(domain string) (adders []string, err error) {
	dots := strings.Split(domain, ".")
	for i := 0; i < len(dots)-1; i++ {
		parent := strings.Join(dots[i:], ".")
		adders, err = LookupNS(parent, defaultServerAddress)
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

// LookupCNAME 查询CNAME记录(别名记录)
func LookupCNAME(domain string) (cname string, err error) {
	return lookupCNAME(domain, defaultServerAddress)
}

// lookupCNAME 查询CNAME记录(别名记录)
func lookupCNAME(domain, dnsServer string) (cname string, err error) {
	// 创建 DNS 客户端
	domain = dns.Fqdn(domain)
	c := new(dns.Client)

	// 构造 DNS 请求消息
	m := new(dns.Msg)
	m.SetQuestion(domain, dns.TypeA)
	m.RecursionDesired = true

	// 发送 DNS 请求
	r, _, err := c.Exchange(m, dnsServer)
	if err != nil {
		return "", err
	}

	// 遍历响应记录
	for _, ans := range r.Answer {
		if v, ok := ans.(*dns.CNAME); ok {
			cname = v.Target
		}
	}

	if len(cname) == 0 {
		return domain, nil
	}

	return cname, nil
}

// LookupNS 查询该域名的权威名称服务器（Name Server, NS）记录
func LookupNS(domain string, dnsServer string) (adders []string, err error) {
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

// fixDomain 加.用于格式化域名
func fixDomain(domain string) string {
	if strings.HasSuffix(domain, ".") {
		return domain
	}
	return domain + "."

}

// LookupIPAAdders 查询该域名的A记录（即IPv4地址记录）
func LookupIPAAdders(domain string) (ipAdders []string, err error) {
	return lookupIPAAdders(domain, defaultServerAddress)
}

// lookupIPAAdders 查询该域名的A记录（即IPv4地址记录）
func lookupIPAAdders(domain, dnsServer string) (ipAdders []string, err error) {
	// 创建 DNS 客户端
	domain = dns.Fqdn(domain)
	c := new(dns.Client)

	// 构造 DNS 请求消息
	m := new(dns.Msg)
	m.SetQuestion(domain, dns.TypeA)
	m.RecursionDesired = true

	// 发送 DNS 请求
	r, _, err := c.Exchange(m, dnsServer)
	if err != nil {
		return nil, err
	}

	// 遍历响应记录
	for _, ans := range r.Answer {
		if v, ok := ans.(*dns.A); ok {
			ipAdders = append(ipAdders, v.A.String())
		}
	}
	return ipAdders, nil
}

// GetSystemDefaultAddress 获取系统默认的本地dns服务器
func GetSystemDefaultAddress() (addr string) {
	resolver := net.Resolver{}
	resolver.PreferGo = true
	resolver.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		addr = address
		return nil, fmt.Errorf(address)
	}
	_, _ = resolver.LookupHost(context.Background(), "Addr")
	return addr
}

// EDNSQuery 使用 EDNS 模拟地域，来查询解析的IP地址，
// 需要 DNSServer 支持 EDNS0SUBNET，目前国内主流 DNS 均不支持该功能
// 通过 NS 记录查询得到的权威 DNS 服务器，通常会支持该功能
func EDNSQuery(domain string, EDNSAddr string, dnsServer string) (adders []string, err error) {
	c := new(dns.Client)
	m := newEDNSMessage(domain, EDNSAddr)
	in, _, err := c.Exchange(m, dnsServer) //注意:要选择支持自定义 EDNS 的DNS 或者是 目标NS服务器  国内DNS大部分不支持自定义EDNS数据
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
// 返回值： 返回一个指向 dns.Msg 结构体的指针，表示生成的 DNS 查询消息。
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
