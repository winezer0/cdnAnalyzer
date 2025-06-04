package asndb

import (
	"fmt"
	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/oschwald/maxminddb-golang"
	"net"
	"os"
	"strconv"
)

var mmDb = map[string]*maxminddb.Reader{}
var mmDbWriter *mmdbwriter.Tree

func mmdbConnect() {
	mmdbOpenFile("COUNTRY")
	mmdbOpenFile("ASN")
	mmdbOpenFile("CITY")
}

func mmdbClose() {
	for connectionId, conn := range mmDb {
		err := conn.Close()
		if err != nil {
			panic(err)
		}
		delete(mmDb, connectionId)
	}
}

func mmdbInitialised(key string) bool {
	connectionId := key + "ipv4"
	_, ok := mmDb[connectionId]

	return ok
}

func mmdbOpenFile(key string) {
	if len(os.Getenv(key)) > 0 {
		ipVersions := []int{4, 6}
		for _, ipVersion := range ipVersions {
			connectionId := key + "ipv" + strconv.Itoa(ipVersion)
			filePath := "downloads/" + os.Getenv(key) + "-ipv" + strconv.Itoa(ipVersion) + ".mmdb"

			if _, err := os.Stat(filePath); err == nil {
				_, ok := mmDb[connectionId]
				if !ok {
					fmt.Println("Opening MMDB file: " + filePath)
					conn, err := maxminddb.Open(filePath)
					if err != nil {
						panic(err)
					}

					mmDb[connectionId] = conn
				}
			}
		}
	}
}

func mmdbCloseFile(connectionId string, filePath string) {
	conn, ok := mmDb[connectionId]
	if ok {
		fmt.Println("Closing MMDB file: " + filePath)
		err := conn.Close()
		if err != nil {
			panic(err)
		}
		delete(mmDb, connectionId)
	}
}

func mmdbIp(ip net.IP) *Ip {
	ipString := ip.String()
	ipVersion := getIpVersion(ipString)

	ipStruct := NewIp(ipString, ipVersion)

	if hasCityDatabase() {
		connectionId := "CITYipv" + strconv.Itoa(ipVersion)
		_, ok := mmDb[connectionId]
		if ok {
			var mmdbCity MmdbCity
			err := mmDb[connectionId].Lookup(ip, &mmdbCity)
			if err != nil {
				panic(err)
			}

			if len(mmdbCity.City.Names.Value) > 0 {
				ipStruct.City = mmdbCity.City.Names.Value
				ipStruct.Latitude = mmdbCity.Location.Latitude
				ipStruct.Longitude = mmdbCity.Location.Longitude
				ipStruct.FoundCity = true

				for i, subdivision := range mmdbCity.Subdivisions {
					switch i {
					case 0:
						ipStruct.State1 = subdivision.Names.Value
					case 1:
						ipStruct.State2 = subdivision.Names.Value
					}
				}
			}

			if len(mmdbCity.Country.ISOCode) > 0 {
				ipStruct.CountryCode = mmdbCity.Country.ISOCode
				ipStruct.FoundCountry = true
			}
		}
	}

	if !ipStruct.FoundCountry && hasCountryDatabase() {
		connectionId := "COUNTRYipv" + strconv.Itoa(ipVersion)
		_, ok := mmDb[connectionId]
		if ok {
			var mmdbCountry MmdbCountry
			err := mmDb[connectionId].Lookup(ip, &mmdbCountry)
			if err != nil {
				panic(err)
			}

			if len(mmdbCountry.Country.ISOCode) > 0 {
				ipStruct.CountryCode = mmdbCountry.Country.ISOCode
				ipStruct.FoundCountry = true
			}
		}
	}

	if hasASNDatabase() {
		connectionId := "ASNipv" + strconv.Itoa(ipVersion)
		_, ok := mmDb[connectionId]
		if ok {
			var mmdbASN MmdbASN
			err := mmDb[connectionId].Lookup(ip, &mmdbASN)
			if err != nil {
				panic(err)
			}

			if mmdbASN.AsNumber > 0 {
				ipStruct.OrganisationNumber = mmdbASN.AsNumber
				ipStruct.OrganisationName = mmdbASN.AsOrganisation
				ipStruct.FoundASN = true
			}
		}
	}

	return ipStruct
}

func mmdbSaveRestart(table string, ipVersion int) {
	if mmDbWriter != nil {
		var key string
		switch table {
		case "ip_country":
			key = "COUNTRY"
		case "ip_asn":
			key = "ASN"
		case "ip_city":
			key = "CITY"
		}

		connectionId := key + "ipv" + strconv.Itoa(ipVersion)
		filePath := "downloads/" + os.Getenv(key) + "-ipv" + strconv.Itoa(ipVersion) + ".mmdb"

		mmdbCloseFile(connectionId, filePath)

		fileHandle, err := os.Create(filePath)
		if err != nil {
			panic(err)
		}

		fmt.Println("Writing MMDB file: " + filePath)
		_, err = mmDbWriter.WriteTo(fileHandle)
		if err != nil {
			panic(err)
		}

		err = fileHandle.Close()
		if err != nil {
			panic(err)
		}

		mmDbWriter = nil
		mmdbOpenFile(key)
	}
}

func mmdbSaveCountries(countries []IpCountry) {
	mmdbInitWriter("COUNTRY", countries[0].IpVersion, 24)

	for _, country := range countries {
		record := mmdbtype.Map{
			"country": mmdbtype.Map{
				"iso_code": mmdbtype.String(country.CountryCode),
			},
		}

		ipRanges := findIPRanges(country.IpRangeStart, country.IpRangeEnd)
		for _, ipRange := range ipRanges {
			err := mmDbWriter.Insert(ipRange, record)
			if err != nil {
				panic(err)
			}
		}
	}
}

func mmdbSaveASNs(ASNs []IpASN) {
	mmdbInitWriter("ASN", ASNs[0].IpVersion, 24)

	for _, ASN := range ASNs {
		record := mmdbtype.Map{
			"autonomous_system_number":       mmdbtype.Uint32(ASN.AsNumber),
			"autonomous_system_organization": mmdbtype.String(ASN.AsOrganisation),
		}

		ipRanges := findIPRanges(ASN.IpRangeStart, ASN.IpRangeEnd)
		for _, ipRange := range ipRanges {
			err := mmDbWriter.Insert(ipRange, record)
			if err != nil {
				panic(err)
			}
		}
	}
}

func mmdbSaveCities(cities []IpCity) {
	mmdbInitWriter("CITY", cities[0].IpVersion, 28)

	for _, city := range cities {
		record := mmdbtype.Map{
			"city": mmdbtype.Map{
				"names": mmdbtype.Map{
					"en": mmdbtype.String(city.City),
				},
				"postcode": mmdbtype.String(city.Postcode),
				"timezone": mmdbtype.String(city.Timezone),
			},
			"country": mmdbtype.Map{
				"iso_code": mmdbtype.String(city.CountryCode),
			},
			"location": mmdbtype.Map{
				"latitude":  mmdbtype.Float64(city.Latitude),
				"longitude": mmdbtype.Float64(city.Longitude),
			},
			"subdivisions": mmdbtype.Slice{
				mmdbtype.Map{
					"names": mmdbtype.Map{
						"en": mmdbtype.String(city.State1),
					},
				},
				mmdbtype.Map{
					"names": mmdbtype.Map{
						"en": mmdbtype.String(city.State2),
					},
				},
			},
		}

		ipRanges := findIPRanges(city.IpRangeStart, city.IpRangeEnd)
		for _, ipRange := range ipRanges {
			err := mmDbWriter.Insert(ipRange, record)
			if err != nil {
				panic(err)
			}
		}
	}
}

func mmdbInitWriter(dbType string, ipVersion int, recordSize int) {
	if mmDbWriter == nil {
		var err error
		mmDbWriter, err = mmdbwriter.New(
			mmdbwriter.Options{
				DatabaseType:            os.Getenv(dbType) + "-ipv" + strconv.Itoa(ipVersion),
				RecordSize:              recordSize,
				IPVersion:               ipVersion,
				IncludeReservedNetworks: true,
				DisableIPv4Aliasing:     true,
			},
		)
		if err != nil {
			panic(err)
		}
	}
}
