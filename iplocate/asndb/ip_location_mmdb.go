package asndb

import (
	"fmt"
	"github.com/oschwald/maxminddb-golang"
	"github.com/praserx/ipconv"
	"github.com/seancfoley/ipaddress-go/ipaddr"
	"net"
	"os"
	"strconv"
	"strings"
)

var mmDb = map[string]*maxminddb.Reader{}

type Ip struct {
	IP                 string `json:"ip"`
	IPVersion          int    `json:"ip_version"`
	FoundASN           bool   `json:"found_asn"`
	OrganisationNumber uint64 `json:"as_number"`
	OrganisationName   string `json:"as_organisation"`
}

type ASNRecord struct {
	AsNumber       uint64 `maxminddb:"autonomous_system_number"`
	AsOrganisation string `maxminddb:"autonomous_system_organization"`
}

func NewIp(ipString string, ipVersion int) *Ip {
	return &Ip{ipString, ipVersion, false, 0, ""}
}

// initMMDBConn 初始化 MaxMind ASN 数据库连接，接受 IPv4 和 IPv6 数据库的完整路径
func initMMDBConn(ipv4Path, ipv6Path string) error {
	type dbInfo struct {
		filePath     string
		connectionId string
	}

	dbFiles := []dbInfo{
		{ipv4Path, "ipv4"},
		{ipv6Path, "ipv6"},
	}

	for _, db := range dbFiles {
		if _, err := os.Stat(db.filePath); os.IsNotExist(err) {
			fmt.Printf("文件不存在: %s\n", db.filePath)
			continue
		}

		if _, ok := mmDb[db.connectionId]; ok {
			fmt.Printf("数据库已加载: %s\n", db.connectionId)
			continue
		}

		conn, err := maxminddb.Open(db.filePath)
		if err != nil {
			return fmt.Errorf("打开数据库失败 [%s]: %v", db.filePath, err)
		}

		mmDb[db.connectionId] = conn
		fmt.Printf("数据库已加载: %s -> %s\n", db.connectionId, db.filePath)
	}

	return nil
}

func closeMMDBConn() {
	for connectionId, conn := range mmDb {
		fmt.Printf("正在关闭数据库连接: %s\n", connectionId)
		err := conn.Close()
		if err != nil {
			fmt.Printf("关闭数据库失败 [%s]: %v\n", connectionId, err)
			continue
		}
		fmt.Printf("数据库已成功关闭: %s\n", connectionId)
		delete(mmDb, connectionId)
	}
}

func findASN(ip net.IP) *Ip {
	ipString := ip.String()
	ipVersion := getIpVersion(ipString)
	ipStruct := NewIp(ipString, ipVersion)
	connectionId := "ipv" + strconv.Itoa(ipVersion)
	_, ok := mmDb[connectionId]
	if ok {
		var asnRecord ASNRecord
		err := mmDb[connectionId].Lookup(ip, &asnRecord)
		if err != nil {
			panic(err)
		}

		if asnRecord.AsNumber > 0 {
			ipStruct.OrganisationNumber = asnRecord.AsNumber
			ipStruct.OrganisationName = asnRecord.AsOrganisation
			ipStruct.FoundASN = true
		}
	}

	return ipStruct
}

func getIpVersion(ipString string) int {
	ipVersion := 4
	if strings.Contains(ipString, ":") {
		ipVersion = 6
	}

	return ipVersion
}

func ipv4ToNumber(ipString string) int64 {
	ip, ipVersion, err := ipconv.ParseIP(ipString)
	if err == nil && ipVersion == 4 {
		number, err := ipconv.IPv4ToInt(ip)
		if err == nil {
			return int64(number)
		}
		panic(err)
	}

	return 0
}

func findIPRanges(ipRangeStart string, ipRangeEnd string) []*net.IPNet {
	ipStart := ipaddr.NewIPAddressString(ipRangeStart)
	ipEnd := ipaddr.NewIPAddressString(ipRangeEnd)

	addressStart := ipStart.GetAddress()
	addressEnd := ipEnd.GetAddress()

	ipRange := addressStart.SpanWithRange(addressEnd)
	rangeSlice := ipRange.SpanWithPrefixBlocks()

	var ipNets []*net.IPNet
	for _, val := range rangeSlice {
		_, network, err := net.ParseCIDR(val.String())
		if err != nil {
			panic(err)
		}

		ipNets = append(ipNets, network)
	}

	return ipNets
}
