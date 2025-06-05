package main

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

//TODO 实现 任意数据格式到本数据Json的转换
//TODO 实现 本数据Json到结构体的转换
//TODO 基于结构体进行数据分析
