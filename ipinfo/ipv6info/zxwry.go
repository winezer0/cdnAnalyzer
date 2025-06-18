package ipv6info

// forked from https://github.com/zu1k/nali
// ipv6db数据使用http://ip.zxinc.org的免费离线数据（更新到2021年）

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/zu1k/nali/pkg/wry"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

type Ipv6Location struct {
	wry.IPDB[uint64]
}

func NewIPv6Location(filePath string) (*Ipv6Location, error) {
	var fileData []byte

	_, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		log.Fatalf("IP数据库文件[%v]不存在\n", filePath)
		return nil, fmt.Errorf("IP数据库文件[%v]不存在", filePath)
	}

	fileBase, err := os.OpenFile(filePath, os.O_RDONLY, 0400)
	if err != nil {
		return nil, err
	}
	defer fileBase.Close()

	fileData, err = io.ReadAll(fileBase)
	if err != nil {
		return nil, err
	}

	if !checkIPv6File(fileData) {
		log.Fatalf("IP数据库[%v]内容存在错误\n", filePath)
		return nil, fmt.Errorf("IP数据库[%v]内容存在错误", filePath)
	}

	header := fileData[:24]
	offLen := header[6]
	ipLen := header[7]

	start := binary.LittleEndian.Uint64(header[16:24])
	counts := binary.LittleEndian.Uint64(header[8:16])
	end := start + counts*11

	return &Ipv6Location{
		IPDB: wry.IPDB[uint64]{
			Data: fileData,

			OffLen:   offLen,
			IPLen:    ipLen,
			IPCnt:    counts,
			IdxStart: start,
			IdxEnd:   end,
		},
	}, nil
}

func (db *Ipv6Location) find(query string) (result *wry.Result, err error) {
	ip := net.ParseIP(query)
	if ip == nil {
		return nil, errors.New("query should be IPv6")
	}
	ip6 := ip.To16()
	if ip6 == nil {
		return nil, errors.New("query should be IPv6")
	}
	ip6 = ip6[:8]
	ipu64 := binary.BigEndian.Uint64(ip6)

	offset := db.SearchIndexV6(ipu64)
	reader := wry.NewReader(db.Data)
	reader.Parse(offset)

	return &reader.Result, nil
}

func (db *Ipv6Location) Find(query string) string {
	result, err := db.find(query)
	if err != nil || result == nil {
		return ""
	}
	r := strings.ReplaceAll(result.Country, "\t", " ")
	return r
}

func checkIPv6File(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	if string(data[:4]) != "IPDB" {
		return false
	}

	if len(data) < 24 {
		return false
	}
	header := data[:24]
	start := binary.LittleEndian.Uint64(header[16:24])
	counts := binary.LittleEndian.Uint64(header[8:16])
	end := start + counts*11
	if start >= end || uint64(len(data)) < end {
		return false
	}

	return true
}
