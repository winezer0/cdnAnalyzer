package dns_query

import (
	"fmt"
	"github.com/miekg/dns"
	"time"
)

// DNSResult 统一返回结构
type DNSResult struct {
	A     []string          `json:"A"`
	AAAA  []string          `json:"AAAA"`
	CNAME []string          `json:"CNAME"`
	NS    []string          `json:"NS"`
	MX    []string          `json:"MX"`
	TXT   []string          `json:"TXT"`
	Error map[string]string `json:"error,omitempty"`
}

// queryDNS 查询指定类型的DNS记录，支持超时
func queryDNS(domain, dnsServer, qtype string, timeout time.Duration) ([]string, error) {
	c := new(dns.Client)
	c.Timeout = timeout
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.StringToType[qtype])
	resp, _, err := c.Exchange(m, dnsServer)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, ans := range resp.Answer {
		switch rr := ans.(type) {
		case *dns.A:
			result = append(result, rr.A.String())
		case *dns.AAAA:
			result = append(result, rr.AAAA.String())
		case *dns.CNAME:
			result = append(result, rr.Target)
		case *dns.NS:
			result = append(result, rr.Ns)
		case *dns.MX:
			result = append(result, fmt.Sprintf("%d %s", rr.Preference, rr.Mx))
		case *dns.TXT:
			result = append(result, rr.Txt...)
		}
	}
	return result, nil
}

// QueryAllDNS 查询所有常用DNS记录，支持超时
func QueryAllDNS(domain, dnsServer string, timeout time.Duration) DNSResult {
	result := DNSResult{Error: make(map[string]string)}
	// A
	a, err := queryDNS(domain, dnsServer, "A", timeout)
	if err != nil {
		result.Error["A"] = err.Error()
	}
	result.A = a
	// AAAA
	aaaa, err := queryDNS(domain, dnsServer, "AAAA", timeout)
	if err != nil {
		result.Error["AAAA"] = err.Error()
	}
	result.AAAA = aaaa
	// CNAME
	cname, err := queryDNS(domain, dnsServer, "CNAME", timeout)
	if err != nil {
		result.Error["CNAME"] = err.Error()
	}
	result.CNAME = cname // 列表
	// NS
	ns, err := queryDNS(domain, dnsServer, "NS", timeout)
	if err != nil {
		result.Error["NS"] = err.Error()
	}
	result.NS = ns // 列表
	// MX
	mx, err := queryDNS(domain, dnsServer, "MX", timeout)
	if err != nil {
		result.Error["MX"] = err.Error()
	}
	result.MX = mx
	// TXT
	txt, err := queryDNS(domain, dnsServer, "TXT", timeout)
	if err != nil {
		result.Error["TXT"] = err.Error()
	}
	result.TXT = txt

	if len(result.Error) == 0 {
		result.Error = nil
	}
	return result
}

//// 仅用于单文件测试
//func main() {
//	domain := "example.com"
//	dnsServer := "8.8.8.8:53"
//	timeout := 3 * time.Second
//
//	result := QueryAllDNS(domain, dnsServer, timeout)
//	b, _ := json.MarshalIndent(result, "", "  ")
//	fmt.Println(string(b))
//}
