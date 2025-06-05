package generate

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
