package asndb

import (
	"fmt"
	"github.com/maxmind/mmdbwriter"
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

	connectionId := "ipv" + strconv.Itoa(ipVersion)
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

	return ipStruct
}
