package cdncheck

import (
	"github.com/yl2chen/cidranger"
	"net"
	"strconv"
	"strings"
)

var cloudNameList = []string{
	"阿里云", "华为云", "腾讯云",
	"天翼云", "金山云", "UCloud", "青云", "QingCloud", "百度云", "盛大云", "世纪互联蓝云",
	"Azure", "Amazon", "Microsoft", "Google", "vultr", "CloudFlare", "Choopa",
}

// ASNInCdn 检查传入的 uint64 格式的 ASN 号是否在 CDN 厂商的 ASN 列表中
func ASNInCdn(asn uint64, asns []string) bool {
	for _, asnStr := range asns {
		_asnInt, err := strconv.Atoi(asnStr)
		if err != nil {
			continue
		}
		if uint64(_asnInt) == asn {
			return true
		}
	}
	return false
}

// CNAMEInCdn 检查传入的 CNAME 字符串是否包含任一 CDN 厂商的 CNAME
func CNAMEInCdn(cname string, cdnCNAMEs []string) (isCDN bool, CDNName string) {
	// CDN厂商的CNAME Inspired by https://github.com/timwhitez/Frog-checkCDN
	cname = strings.Trim(cname, ".")
	for _, cn := range cdnCNAMEs {
		if strings.Index(cname, cn) >= 0 {
			return true, cn
		}
	}
	return false, ""
}

// IPInCdn 在 CDN 厂商的 IP 段列表中
func IPInCdn(ip string, cdnIPs []string) bool {
	ranger := cidranger.NewPCTrieRanger()
	for _, cdn := range cdnIPs {
		_, network, err := net.ParseCIDR(cdn)
		if err != nil {
			continue
		}
		err = ranger.Insert(cidranger.NewBasicRangerEntry(*network))
		if err != nil {
			continue
		}
	}

	contains, err := ranger.Contains(net.ParseIP(ip))
	if err == nil {
		return contains
	}
	return false
}
