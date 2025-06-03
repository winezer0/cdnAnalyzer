package dnsquery

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
	"sync"
	"time"
)

var defaultServerAddress = GetSystemDefaultAddress()

type ECSResult struct {
	IPs    []string
	CNAMEs []string
	Errors []string // 支持多个错误
}

func EDNSQuery(domain string, EDNSAddr string, dnsServer string, timeout time.Duration) ECSResult {
	domain = dns.Fqdn(domain)

	Client := new(dns.Client)
	Client.Timeout = timeout
	dnsMsg := newEDNSMessage(domain, EDNSAddr)

	in, _, err := Client.Exchange(dnsMsg, dnsServer)
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

func EDNSQueryWithMultiCities(domain string, timeout time.Duration, Cities []map[string]string, queryCNAME bool) map[string]ECSResult {

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
		nsServers, err := LookupNSServers(finalDomain, defaultServerAddress, timeout)
		if err != nil || len(nsServers) == 0 {
			fmt.Printf("Failed to lookup NS servers for %s, using fallback: 8.8.8.8:53\n", finalDomain)
		} else {
			dnsServers = append(dnsServers, DnsServerAddPort(nsServers[0]))
			fmt.Printf("Success lookup NS servers: %v, using fallback: %v\n", nsServers, nsServers[0])
		}
	}

	// 并发执行查询
	var wg sync.WaitGroup
	resultsChan := make(chan map[string]ECSResult, len(Cities)*len(dnsServers))
	resultMap := make(map[string]ECSResult)
	var mu sync.Mutex // 用于并发安全写入 resultMap

	for _, dnsServer := range dnsServers {
		for _, entry := range Cities {
			city := entry["City"]
			ecsIP := entry["IP"]
			wg.Add(1)
			go func(city string, ecsIP string, dnsServer string) {
				defer wg.Done()

				result := EDNSQuery(finalDomain, ecsIP, dnsServer, timeout)

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
// 返回值： 返回一个指向 dnsquery.Msg 结构体的指针，表示生成的 DNS 查询消息。
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
