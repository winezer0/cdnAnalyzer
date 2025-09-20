package docheck

import "github.com/winezer0/cdnAnalyzer/pkg/downfile"

// AppConfig 表示整个配置文件结构
type AppConfig struct {
	//// 与 CmdConfig 结构体相同的字段
	// DNS并发和超时设置
	ResolversNum      int  `yaml:"resolvers-num"`
	CityMapNUm        int  `yaml:"city-map-num"`
	DNSTimeOut        int  `yaml:"dns-timeout"`
	DNSConcurrency    int  `yaml:"dns-concurrency"`
	EDNSConcurrency   int  `yaml:"edns-concurrency"`
	QueryEDNSCNAMES   bool `yaml:"query-edns-cnames"`
	QueryEDNSUseSysNS bool `yaml:"query-edns-use-sys-ns"`

	// 数据库下载配置
	DownloadItems []downfile.DownItem `yaml:"download-items"`
}

// DBFilePaths 存储所有数据库文件路径
type DBFilePaths struct {
	ResolversFile string
	CityMapFile   string
	AsnIpvxDb     string
	Ipv4QQWryDb   string
	Ipv6ZXWryDb   string
	CdnSource     string
}

// 数据库文件名常量
const (
	ModuleDNSResolvers = "dns-resolvers"
	ModuleEDNSCityIP   = "edns-city-ip"
	ModuleIPv4QQWry    = "qqwry"
	ModuleIPv6ZXWry    = "zxipv6wry"
	ModuleAsnIPvx      = "geolite2-asn"
	ModuleCDNSource    = "cdn-sources"
)
