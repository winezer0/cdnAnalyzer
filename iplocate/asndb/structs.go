package asndb

type IpCity struct {
	IpRangeStart string
	IpRangeEnd   string
	CountryCode  string
	State1       string
	State2       string
	City         string
	Postcode     string
	Latitude     float64
	Longitude    float64
	Timezone     string
	IpVersion    int
	DbVersion    int
}

type IpASN struct {
	IpRangeStart   string
	IpRangeEnd     string
	AsNumber       int
	AsOrganisation string
	IpVersion      int
	DbVersion      int
}

type IpCountry struct {
	IpRangeStart string
	IpRangeEnd   string
	CountryCode  string
	IpVersion    int
	DbVersion    int
}

type Ip struct {
	IP                 string  `json:"ip"`
	IPVersion          int     `json:"ip_version"`
	FoundCountry       bool    `json:"found_country"`
	FoundCity          bool    `json:"found_city"`
	FoundASN           bool    `json:"found_asn"`
	CountryCode        string  `json:"country_code"`
	State1             string  `json:"state"`
	State2             string  `json:"state_2"`
	City               string  `json:"city"`
	Postcode           string  `json:"postcode"`
	Latitude           float64 `json:"lat"`
	Longitude          float64 `json:"lon"`
	Timezone           string  `json:"timezone"`
	OrganisationNumber int64   `json:"as_number"`
	OrganisationName   string  `json:"as_organisation"`
	Milliseconds       int64   `json:"ms_taken"`
	Microseconds       int64   `json:"μs_taken"`
}

func NewIp(ipString string, ipVersion int) *Ip {
	return &Ip{ipString, ipVersion, false, false, false, "", "", "", "", "", 0, 0, "", 0, "", 0, 0}
}

type MmdbCountry struct {
	Country struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
}

type MmdbASN struct {
	AsNumber       int64  `maxminddb:"autonomous_system_number"`
	AsOrganisation string `maxminddb:"autonomous_system_organization"`
}

type MmdbCity struct {
	City struct {
		Names struct {
			Value string `maxminddb:"en"`
		} `maxminddb:"names"`
		Postcode string `maxminddb:"postcode"`
		Timezone string `maxminddb:"timezone"`
	} `maxminddb:"city"`
	Country struct {
		ISOCode   string `maxminddb:"iso_code"`
		GeonameID int64  `maxminddb:"geoname_id"`
		Eu        bool   `maxminddb:"is_in_european_union"`
		Names     struct {
			Value string `maxminddb:"en"`
		} `maxminddb:"names"`
	} `maxminddb:"continent"`
	Continent struct {
		Code      string `maxminddb:"code"`
		GeonameID int64  `maxminddb:"geoname_id"`
		Names     struct {
			Value string `maxminddb:"en"`
		} `maxminddb:"names"`
	} `maxminddb:"continent"`
	Location struct {
		Latitude  float64 `maxminddb:"latitude"`
		Longitude float64 `maxminddb:"longitude"`
	} `maxminddb:"location"`
	Subdivisions []struct {
		Names struct {
			Value string `maxminddb:"en"`
		} `maxminddb:"names"`
	} `maxminddb:"subdivisions"`
}
