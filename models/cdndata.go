package models

type CDNData struct {
	CDN   Category `json:"cdn"`
	WAF   Category `json:"waf"`
	CLOUD Category `json:"cloud"`
}

type Category struct {
	IP    map[string][]string `json:"ip,omitempty"`
	ASN   map[string][]string `json:"asn,omitempty"`
	CNAME map[string][]string `json:"cname,omitempty"`
	KEYS  map[string][]string `json:"keys,omitempty"`
}

func newEmptyCategory() Category {
	return Category{
		IP:    make(map[string][]string),
		ASN:   make(map[string][]string),
		CNAME: make(map[string][]string),
		KEYS:  make(map[string][]string),
	}
}

func NewEmptyCDNDataPointer() *CDNData {
	return &CDNData{
		CDN:   newEmptyCategory(),
		WAF:   newEmptyCategory(),
		CLOUD: newEmptyCategory(),
	}
}
