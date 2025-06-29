package generate

import (
	"github.com/winezer0/cdnAnalyzer/pkg/analyzer"
	"github.com/winezer0/cdnAnalyzer/pkg/fileutils"
	"github.com/winezer0/cdnAnalyzer/pkg/maputils"
)

// TransferCdnDomainsYaml  实现NaLi cdn.yml到json的转换
func TransferCdnDomainsYaml(path string) *analyzer.CDNData {
	// 数据来源 https://github.com/4ft35t/cdn/blob/master/src/cdn.yml
	// CloudKeysData 是整个 YAML 文件的结构
	type naliCdnData map[string]struct {
		Name string `yaml:"name"`
		Link string `yaml:"link"`
	}
	// 1. 读取 YAML 到 CloudKeysData
	var yamlData naliCdnData
	err := fileutils.ReadYamlToStruct(path, &yamlData)
	if err != nil {
		panic(err)
	}

	// 2. 构建 cname map[string][]string 并赋值给 cdnData.CDN.CNAME
	// 初始化 CDNData 结构
	cdnData := analyzer.NewEmptyCDNData()
	for domain, info := range yamlData {
		cdnData.CDN.CNAME[info.Name] = append(cdnData.CDN.CNAME[info.Name], domain)
	}

	return cdnData
}

// TransferPDCdnCheckJson 实现cdn check json 数据源的转换
func TransferPDCdnCheckJson(path string) *analyzer.CDNData {
	// PDCheckData 表示整个配置结构
	type PDCheckData struct {
		CDN    map[string][]string `json:"cdn"`
		WAF    map[string][]string `json:"waf"`
		Cloud  map[string][]string `json:"cloud"`
		Common map[string][]string `json:"common"`
	}

	// 加载cdn check json数据源
	var pdCheckData PDCheckData
	err := fileutils.ReadJsonToStruct(path, &pdCheckData)
	if err != nil {
		panic(err)
	}

	// 将 cdn/waf/cloud 的值作为 IP 数据填充到对应字段
	cdnData := analyzer.NewEmptyCDNData()
	cdnData.CDN.IP = maputils.CopyMap(pdCheckData.CDN)
	cdnData.WAF.IP = maputils.CopyMap(pdCheckData.WAF)
	cdnData.CLOUD.IP = maputils.CopyMap(pdCheckData.Cloud)

	// 合并 common 到 cdn.cname
	for provider, cnames := range pdCheckData.Common {
		cdnData.CDN.CNAME[provider] = append([]string{}, cnames...)
	}
	return cdnData
}

// TransferCloudKeysYaml  实现 cloud keys yml到json的转换
func TransferCloudKeysYaml(path string) *analyzer.CDNData {
	// 数据来源 用户自己数据到cloud_keys.yml中
	// 是整个 YAML 文件的结构
	var cloudKeysYaml map[string]struct {
		Keys []string `yaml:"keys"`
	}

	// 1. 读取 YAML 到 CloudKeysData
	err := fileutils.ReadYamlToStruct(path, &cloudKeysYaml)
	if err != nil {
		panic(err)
	}
	//
	// 2. 构建 cname map[string][]string 并赋值给 cdnData.CDN.CNAME
	// 初始化 CDNData 结构
	cdnData := analyzer.NewEmptyCDNData()
	for cloudName, yamEntry := range cloudKeysYaml {
		cdnData.CLOUD.KEYS[cloudName] = yamEntry.Keys
	}

	return cdnData
}

// TransferProviderYAML 读取并解析 provider.yaml 文件
func TransferProviderYAML(filePath string) *analyzer.CDNData {
	// 定义结构体匹配 provider.yaml 格式
	type Component struct {
		FQDN map[string][]string `yaml:"fqdn"`
		ASN  map[string][]string `yaml:"asn"`
		URLs map[string][]string `yaml:"urls"`
		CIDR map[string][]string `yaml:"cidr"`
	}

	// 主配置结构体，使用嵌套的 Component
	type ProviderConfig struct {
		CDN    Component `yaml:"cdn"`
		WAF    Component `yaml:"waf"`
		CLOUD  Component `yaml:"CLOUD"`
		COMMON Component `yaml:"common"`
	}

	// 初始化配置结构体
	var providerConfig ProviderConfig
	err := fileutils.ReadYamlToStruct(filePath, &providerConfig)
	if err != nil {
		panic(err)
	}

	// 初始化 CDNData 结构
	cdnData := analyzer.NewEmptyCDNData()
	for name, asns := range providerConfig.CDN.ASN {
		cdnData.CDN.ASN[name] = analyzer.NormalizeASNList(asns)
	}

	for name, cidr := range providerConfig.CDN.CIDR {
		cdnData.CDN.IP[name] = cidr
	}

	for name, asns := range providerConfig.WAF.ASN {
		cdnData.WAF.ASN[name] = analyzer.NormalizeASNList(asns)
	}

	for name, cidr := range providerConfig.WAF.CIDR {
		cdnData.WAF.IP[name] = cidr
	}

	for name, asns := range providerConfig.CLOUD.ASN {
		cdnData.CLOUD.ASN[name] = analyzer.NormalizeASNList(asns)
	}

	for name, cidr := range providerConfig.CLOUD.CIDR {
		cdnData.CLOUD.IP[name] = cidr
	}

	// 添加 common 的 子域名到 CDN 的CNAMES子域名
	for name, fqdn := range providerConfig.COMMON.FQDN {
		cdnData.CDN.CNAME[name] = fqdn
	}

	return cdnData
}
