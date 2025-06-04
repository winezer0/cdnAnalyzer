package cdncheck

import (
	"net"
	"strings"
)

// Inspired by https://github.com/timwhitez/Frog-checkCDN
// CDN厂商的CNAME
var cnames = []string{
	"cdn.net", "fwdns.net", "bitgravity.com", "21okglb.cn", "kxcdn", "fastwebcdn.com", "cachefly.net",
	"simplecdn.net", "tbcache.com", "footprint.net", "cloudflare.net", "51cdn.com", "google.", "bluehatnetwork.com",
	"hadns.net", "incapdns", "skyparkcdn", "akamai", "hwcdn", "cdn77.org", "aicdn.com", "akamaitechnologies.com",
	"fastly", "fpbns", "cdn77.net", "zenedge.net", "akadns.net", "customcdn.com", "fastly.net", "lswcdn",
	"googleusercontent.com", "mncdn.com", "21speedcdn.com", "hiberniacdn.com", "mirror-image.net", "anankecdn.com.br",
	"cncssr.chinacache.net", "hichina.net", "insnw.net", "jiashule.com", "llnwd", "cdn.dnsv1.com", "bitgravity",
	"mwcloudcdn.com", "amazonaws.com", "systemcdn.net", "wscdns.com", "cdnvideo", "ccgslb", "fpbns.net", "dnsv1",
	"360wzb.com", "inscname.net", "ytcdn.net", "21vokglb.cn", "aliyuncs.com", "cdntip", "netdna-ssl.com", "att-dsa.net",
	"tcdn.qq.com", "netdna", "ccgslb.com.cn", "netdna.com", "l.doubleclick.net", "chinaidns.net", "turbobytes-cdn.com",
	"instacontent.net", "speedcdns", "clients.turbobytes.net", "akamai-staging.net", "fastcdn.cn", "wscloudcdn",
	"gslb.taobao.com", "hichina.com", "fastcache.com", "cachecn.com", "verygslb.com", "cdnzz.net", "fwcdn.com",
	"kunlunca.com", "cdn.cloudflare.net", "customcdn.cn", "vo.llnwd.net", "swiftserve.com", "lldns.net", "afxcdn.net",
	"ourwebpic.com", "edgekey", "ucloud.cn", "cdn20.com", "swiftcdn1.com", "cdn77", "azioncdn.net", "akamaized.net",
	"cdnvideo.ru", "incapdns.net", "tlgslb.com", "kunlun.com", "cloudflare.com", "anankecdn", "cdnudns.com",
	"footprint", "txnetworks.cn", "akamai.com", "cdnsun.net", "wpc.", "qiniudns.com", "okglb.com", "cloudflare",
	"ngenix", "cloudfront", "belugacdn.com", "edgecast", "cdnsun.net.", "alicdn.com", "cdn.telefonica.com", "lxdns.com",
	"internapcdn.net", "ewcache.com", "llnwd.net", "c3cdn.net", "chinacache.net", "21vianet.com.cn", "qingcdn.com",
	"yunjiasu-cdn", "cdn.ngenix.net", "skyparkcdn.net", "ccgslb.com", "adn.", "presscdn", "panthercdn.com",
	"edgecastcdn.net", "ay1.b.yahoo.com", "alicloudsec.com", "cachefly", "kunlunar.com", "bdydns.com", "cloudfront.net",
	"acadn.com", "cap-mii.net", "gslb.tbcache.com", "awsdns", "cdn.bitgravity.com", "cdnify.io", "kxcdn.com",
	"00cdn.com", "cdnetworks.net", "fastweb.com", "googlesyndication.", "akamaitech.net", "presscdn.com", "cdnetworks",
	"cdntip.com", "cdnify", "hacdn.net", "azureedge.net", "alicloudlayer.com", "internapcdn", "speedcdns.com", "cdnsun",
	"cdngc.net", "gccdn.net", "fastlylb.net", "cdnnetworks.com", "mwcloudcdn", "21cvcdn.com", "ccgslb.net", "azioncdn",
	"wac.", "unicache.com", "vo.msecnd.net", "stackpathdns.com", "lswcdn.net", "dnspao.com", "akamai.net", "azureedge",
	"aodianyun.com", "dnion.com", "wscloudcdn.com", "ourwebcdn.net", "netdna-cdn.com", "chinacache", "c3cache.net",
	"aliyun-inc.com", "sprycdn.com", "hwcdn.net", "yimg.", "telefonica", "aqb.so", "alikunlun.com",
	"chinanetcenter.com", "cloudcdn.net", "xgslb.net", "gccdn.cn", "globalcdn.cn", "lxcdn.com", "rncdn1.com",
	"youtube.", "txcdn.cn", "edgesuite.net", "okcdn.com", "akamaiedge.net"}

// CheckCName 检查域名的CNAME，判断是否是CDN
func (c *CDNCheck) CheckCName(domain string) (isCDN bool, CDNName string, CName string) {
	cname, err := net.LookupCNAME(domain)
	if err != nil || len(cname) == 0 {
		return
	}
	cname = strings.Trim(cname, ".")
	if cname == domain {
		return
	}
	for _, cn := range cnames {
		if strings.Index(cname, cn) >= 0 {
			return true, cn, cname
		}
	}
	return false, "", cname
}
