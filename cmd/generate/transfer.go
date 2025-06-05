package generate

import (
	"cdnCheck/fileutils"
	"cdnCheck/maputils"
	"cdnCheck/models"
)

// NaliCdnData 是整个 YAML 文件的结构
type NaliCdnData map[string]struct {
	Name string `yaml:"name"`
	Link string `yaml:"link"`
}

// CdnCheckData 表示整个配置结构
type CdnCheckData struct {
	CDN    map[string][]string `json:"cdn"`
	WAF    map[string][]string `json:"waf"`
	Cloud  map[string][]string `json:"cloud"`
	Common map[string][]string `json:"common"`
}

// TransferNaliCdnYaml  实现nali cdn.yml到json的转换
func TransferNaliCdnYaml(path string) *models.CDNData {
	// 数据来源 https://github.com/4ft35t/cdn/blob/master/src/cdn.yml
	// 初始化 CDNData 结构
	cdnData := models.NewEmptyCDNDataPointer()

	// 1. 读取 YAML 到 NaliCdnData
	var yamlData NaliCdnData
	err := fileutils.ReadYamlToStruct(path, &yamlData)
	if err != nil {
		panic(err)
	}

	// 2. 构建 cname map[string][]string 并赋值给 cdnData.CDN.CNAME
	for domain, info := range yamlData {
		cdnData.CDN.CNAME[info.Name] = append(cdnData.CDN.CNAME[info.Name], domain)
	}

	return cdnData
}

// TransferCdnCheckJson 实现cdn check json 数据源的转换
func TransferCdnCheckJson(path string) *models.CDNData {
	// 加载cdn check json数据源
	var cdnCheckData CdnCheckData
	err := fileutils.ReadJsonToStruct(path, &cdnCheckData)
	if err != nil {
		panic(err)
	}

	// 将 cdn/waf/cloud 的值作为 IP 数据填充到对应字段
	cdnData := models.NewEmptyCDNDataPointer()
	cdnData.CDN.IP = maputils.CopyMap(cdnCheckData.CDN)
	cdnData.WAF.IP = maputils.CopyMap(cdnCheckData.WAF)
	cdnData.Cloud.IP = maputils.CopyMap(cdnCheckData.Cloud)

	// 合并 common 到 cdn.cname
	for provider, cnames := range cdnCheckData.Common {
		cdnData.CDN.CNAME[provider] = append([]string{}, cnames...)
	}
	return cdnData
}
