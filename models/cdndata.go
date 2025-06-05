package models

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

func NewCDNData() *CDNData {
	return &CDNData{
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
}
