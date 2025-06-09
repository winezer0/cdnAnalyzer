package asndb

import (
	"fmt"
	"github.com/oschwald/maxminddb-golang"
	"net"
	"os"
	"strconv"
	"strings"
)

var mmDb = map[string]*maxminddb.Reader{}

type ASNInfo struct {
	IP                 string `json:"ip"`
	IPVersion          int    `json:"ip_version"`
	FoundASN           bool   `json:"found_asn"`
	OrganisationNumber uint64 `json:"as_number"`
	OrganisationName   string `json:"as_organisation"`
}

type ASNRecord struct {
	AutonomousSystemNumber uint64 `maxminddb:"autonomous_system_number"`
	AutonomousSystemOrg    string `maxminddb:"autonomous_system_organization"`
}

func NewIp(ipString string, ipVersion int) *ASNInfo {
	return &ASNInfo{ipString, ipVersion, false, 0, ""}
}

// InitMMDBConn 初始化 MaxMind ASN 数据库连接，接受 IPv4 和 IPv6 数据库的完整路径
func InitMMDBConn(ipv4Path, ipv6Path string) error {
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

func CloseMMDBConn() {
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

func FindASN(ip net.IP) *ASNInfo {
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

		if asnRecord.AutonomousSystemNumber > 0 {
			ipStruct.OrganisationNumber = asnRecord.AutonomousSystemNumber
			ipStruct.OrganisationName = asnRecord.AutonomousSystemOrg
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

// ASNToIPRanges 通过ASN号反查所有IP段
func ASNToIPRanges(targetASN uint64) ([]*net.IPNet, error) {
	var findIPs []*net.IPNet
	connectionIds := []string{"ipv4", "ipv6"}
	for _, connectionId := range connectionIds {
		reader, ok := mmDb[connectionId]
		if !ok {
			continue // 跳过未加载的数据库
		}
		networks := reader.Networks()
		for networks.Next() {
			var record ASNRecord
			ipNet, err := networks.Network(&record)
			if err != nil {
				return nil, err
			}
			if record.AutonomousSystemNumber == targetASN {
				findIPs = append(findIPs, ipNet)
			}
		}
		if err := networks.Err(); err != nil {
			return nil, err
		}
	}
	return findIPs, nil
}

func CountMMDBSize(reader *maxminddb.Reader) (int, error) {
	count := 0
	networks := reader.Networks()
	for networks.Next() {
		count++
	}
	if err := networks.Err(); err != nil {
		return 0, err
	}
	return count, nil
}
