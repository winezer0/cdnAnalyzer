package docheck

import "cdnAnalyzer/pkg/downfile"

// AppConfig 表示整个配置文件结构
type AppConfig struct {
	// 与 CmdConfig 结构体相同的字段
	Target     string `yaml:"target"`
	TargetType string `yaml:"target-type"`

	// DNS并发和超时设置
	ResolversNum      int  `yaml:"resolvers-num"`
	CityMapNUm        int  `yaml:"city-map-num"`
	DNSTimeOut        int  `yaml:"dns-timeout"`
	DNSConcurrency    int  `yaml:"dns-concurrency"`
	EDNSConcurrency   int  `yaml:"edns-concurrency"`
	QueryEDNSCNAMES   bool `yaml:"query-edns-cnames"`
	QueryEDNSUseSysNS bool `yaml:"query-edns-use-sys-ns"`

	// 输出配置
	OutputFile  string `yaml:"output-file"`
	OutputType  string `yaml:"output-type"`
	OutputLevel string `yaml:"output-level"`

	// 数据库路径配置
	ResolversFile string `yaml:"resolvers-file"`
	CityMapFile   string `yaml:"city-map-file"`
	AsnIpv4Db     string `yaml:"asn-ipv4-db"`
	AsnIpv6Db     string `yaml:"asn-ipv6-db"`
	Ipv4LocateDb  string `yaml:"ipv4-locate-db"`
	Ipv6LocateDb  string `yaml:"ipv6-locate-db"`
	SourceJson    string `yaml:"source-json"`

	// 下载配置
	DownloadItems []downfile.DownItem `yaml:"download-items"`
}

// DBFilePaths 存储所有数据库文件路径
type DBFilePaths struct {
	ResolversFile string
	CityMapFile   string
	AsnIpv4Db     string
	AsnIpv6Db     string
	Ipv4LocateDb  string
	Ipv6LocateDb  string
	SourceJson    string
}

// 数据库文件名常量
const (
	ModuleDNSResolvers = "dns-resolvers"
	ModuleEDNSCityIP   = "edns-city-ip"

	ModuleIPv4Locate = "qqwry"
	ModuleIPv6Locate = "zxipv6wry"
	ModuleAsnIPv4    = "geolite2-asn-ipv4"
	ModuleAsnIPv6    = "geolite2-asn-ipv6"
	ModuleCDNSource  = "cdn-sources"
)
