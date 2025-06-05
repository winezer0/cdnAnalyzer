package generate

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// NaliCDNInfo 表示每个 CDN 的信息
type NaliCDNInfo struct {
	Name string `yaml:"name"`
	Link string `yaml:"link"`
}

// NaliCdnYamlData 是整个 YAML 文件的结构
type NaliCdnYamlData map[string]NaliCDNInfo

// TransferCdnCnameYamlToJson  实现cdn.yml到json的转换 https://github.com/4ft35t/cdn/blob/master/src/cdn.yml
func TransferCdnCnameYamlToJson(path string) {
	// 1. 读取 YAML 文件内容
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	// 2. 解析 YAML 到 map
	var yamlData NaliCdnYamlData
	err = yaml.Unmarshal(data, &yamlData)
	if err != nil {
		panic(err)
	}

	// 3. 构建 cname map[string][]string
	cnameMap := make(map[string][]string)
	for domain, info := range yamlData {
		cnameMap[info.Name] = append(cnameMap[info.Name], domain)
	}

	// 4. 打印结果
	fmt.Printf("CNAME Map: %+v\n", cnameMap)

	// 如果你需要嵌套进一个更大的结构，比如 Category：
	type Category struct {
		CNAME map[string][]string `json:"cname"`
	}
	category := Category{
		CNAME: cnameMap,
	}

	// 可选：输出为 JSON 格式
	jsonBytes, _ := json.MarshalIndent(category, "", "  ")
	fmt.Println("\nJSON 格式输出:")
	fmt.Println(string(jsonBytes))
}

// CdnCheckData 表示整个配置结构
type CdnCheckData struct {
	CDN    map[string][]string `json:"cdn"`
	WAF    map[string][]string `json:"waf"`
	Cloud  map[string][]string `json:"cloud"`
	Common map[string][]string `json:"common"`
}

// LoadCdnCheckSourceDataJson 实现转换 projectdiscovery cdncheck 的数据源到新的格式
func LoadCdnCheckSourceDataJson(path string) (*CdnCheckData, error) {
	// 1. 读取文件内容
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	// 2. 解析 JSON 到结构体
	var config CdnCheckData
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %v", err)
	}

	return &config, nil

}

type CDNData struct {
	CDN   Category `json:"cdn"`
	WAF   Category `json:"waf"`
	Cloud Category `json:"cloud"`
}

type Category struct {
	IP    map[string][]string `json:"ip"`
	ASN   map[string][]string `json:"asn"`
	CNAME map[string][]string `json:"cname"`
}

func ConvertConfigToCDNData(config *CdnCheckData) *CDNData {
	// 初始化结果结构
	cdnData := &CDNData{
		CDN: Category{
			IP:    make(map[string][]string),
			ASN:   make(map[string][]string),
			CNAME: make(map[string][]string),
		},
		WAF: Category{
			IP:    make(map[string][]string),
			ASN:   make(map[string][]string),
			CNAME: make(map[string][]string),
		},
		Cloud: Category{
			IP:    make(map[string][]string),
			ASN:   make(map[string][]string),
			CNAME: make(map[string][]string),
		},
	}

	// 将 cdn/waf/cloud 的值作为 IP 数据填充到对应字段
	cdnData.CDN.IP = copyMap(config.CDN)
	cdnData.WAF.IP = copyMap(config.WAF)
	cdnData.Cloud.IP = copyMap(config.Cloud)

	// 合并 common 到 cdn.cname
	for provider, cnames := range config.Common {
		cdnData.CDN.CNAME[provider] = append([]string{}, cnames...)
	}
	return cdnData
}

// 辅助函数：深拷贝 map[string][]string
func copyMap(src map[string][]string) map[string][]string {
	dst := make(map[string][]string)
	for k, v := range src {
		dst[k] = append([]string{}, v...)
	}
	return dst
}
