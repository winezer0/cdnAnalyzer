package ednsquery

import (
	"cdnCheck/maputils"
	"fmt"
	"github.com/miekg/dns"
	"github.com/panjf2000/ants/v2"
	"net"
	"sync"
	"time"
)

var defaultServerAddress = GetSystemDefaultAddress()

type EDNSResult struct {
	A      []string
	AAAA   []string
	CNAME  []string
	Errors []string
}

// eDNSMessage 创建并返回一个包含 EDNS（扩展 DNS）选项的 DNS 查询消息
// 参数： domain：要查询的域名。 EDNSAddr：用于设置 EDNS0_SUBNET 选项的 A 地址。
// 返回值： 返回一个指向 dnsquery.Msg 结构体的指针，表示生成的 DNS 查询消息。
func eDNSMessage(domain, EDNSAddr string) *dns.Msg {
	e := new(dns.EDNS0_SUBNET) // EDNS
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

// ResolveEDNS 进行EDNS信息查询
func ResolveEDNS(domain string, EDNSAddr string, dnsServer string, timeout time.Duration) EDNSResult {
	domain = dns.Fqdn(domain)

	Client := new(dns.Client)
	Client.Timeout = timeout
	dnsMsg := eDNSMessage(domain, EDNSAddr)

	in, _, err := Client.Exchange(dnsMsg, dnsServer)
	if err != nil {
		return EDNSResult{
			Errors: []string{err.Error()},
		}
	}

	var ipv4s []string
	var ipv6s []string
	var cnames []string

	for _, answer := range in.Answer {
		switch answer.Header().Rrtype {
		case dns.TypeA:
			ipv4s = append(ipv4s, answer.(*dns.A).A.String())
		case dns.TypeAAAA:
			ipv6s = append(ipv6s, answer.(*dns.AAAA).AAAA.String())
		case dns.TypeCNAME:
			cnames = append(cnames, answer.(*dns.CNAME).Target)
		}
	}

	return EDNSResult{
		A:     ipv4s,
		AAAA:  ipv6s,
		CNAME: cnames,
	}
}

// ResolveEDNSWithCities  批量EDNS查询
func ResolveEDNSWithCities(
	domain string,
	timeout time.Duration,
	Cities []map[string]string,
	queryCNAME bool,
	maxConcurrency int,
) map[string]EDNSResult {

	finalDomain := domain
	dnsServers := []string{"8.8.8.8:53"} // 默认 DNS

	if queryCNAME {
		var err error
		finalDomain, err = getFinalDomain(domain, defaultServerAddress, timeout)
		if err != nil {
			fmt.Printf("Failed to resolve final domain for %s: %v\n", domain, err)
		}

		nsServers := getAuthoritativeNameservers(finalDomain, defaultServerAddress, timeout)
		dnsServers = append(dnsServers, nsServers...)
	}

	// 创建线程池（ants v2）
	pool, err := ants.NewPool(maxConcurrency, ants.WithOptions(ants.Options{
		Nonblocking: true,
	}))
	if err != nil {
		panic(err)
	}
	defer pool.Release()

	var wg sync.WaitGroup
	resultMap := make(map[string]EDNSResult)
	var mu sync.Mutex // 保护 resultMap 并发写入

	results := make(chan struct{}, len(Cities)*len(dnsServers)) // 只用于等待完成

	// 提交任务
	for _, dnsServer := range dnsServers {
		for _, entry := range Cities {
			city := maputils.GetMapValue(entry, "City")
			cityIP := maputils.GetMapValue(entry, "IP")

			wg.Add(1)

			task := func(city, cityIP, dnsServer string) {
				defer wg.Done()

				result := ResolveEDNS(finalDomain, cityIP, dnsServer, timeout)
				key := fmt.Sprintf("%s@%s", city, dnsServer)

				mu.Lock()
				resultMap[key] = result
				mu.Unlock()

				results <- struct{}{}
			}

			err := pool.Submit(func() {
				task(city, cityIP, dnsServer)
			})

			if err != nil {
				// 回退：同步执行
				go task(city, cityIP, dnsServer)
			}
		}
	}

	// 等待所有任务完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 等待所有结果（只为了阻塞主线程）
	for range results {
		// 仅用于接收完所有信号
	}

	return resultMap
}
