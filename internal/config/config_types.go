package config

import (
	"github.com/winezer0/downutils/downutils"
)

// AppConfig 表示整个配置文件结构
type AppConfig struct {
	//// 与 CmdConfig 结构体相同的字段
	// DNS并发和超时设置
	ResolversNum      int    `yaml:"resolvers-num"`
	CityMapNUm        int    `yaml:"city-map-num"`
	DNSTimeOut        int    `yaml:"dns-timeout"`
	DNSConcurrency    int    `yaml:"dns-concurrency"`
	EDNSConcurrency   int    `yaml:"edns-concurrency"`
	QueryEDNSCNAMES   bool   `yaml:"query-edns-cnames"`
	QueryEDNSUseSysNS bool   `yaml:"query-edns-use-sys-ns"`
	QueryMethod       string `yaml:"query-method"`

	// 数据库下载配置
	DownloadItems []downutils.DownItem `yaml:"download-items"`
}
