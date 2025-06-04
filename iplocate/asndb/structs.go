package asndb

type Ip struct {
	IP                 string `json:"ip"`
	IPVersion          int    `json:"ip_version"`
	FoundASN           bool   `json:"found_asn"`
	OrganisationNumber uint64 `json:"as_number"`
	OrganisationName   string `json:"as_organisation"`
}

func NewIp(ipString string, ipVersion int) *Ip {
	return &Ip{ipString, ipVersion, false, 0, ""}
}

type ASNRecord struct {
	AsNumber       uint64 `maxminddb:"autonomous_system_number"`
	AsOrganisation string `maxminddb:"autonomous_system_organization"`
}
