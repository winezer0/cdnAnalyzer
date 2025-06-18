package ipv4info

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/zu1k/nali/pkg/wry"
)

type Ipv4Location struct {
	wry.IPDB[uint32]
}

// NewIPv4Location new database from path
func NewIPv4Location(filePath string) (*Ipv4Location, error) {
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

	if !checkIPv4File(fileData) {
		log.Fatalf("IP数据库[%v]内容存在错误\n", filePath)
		return nil, fmt.Errorf("IP数据库[%v]内容存在错误", filePath)
	}

	header := fileData[0:8]
	start := binary.LittleEndian.Uint32(header[:4])
	end := binary.LittleEndian.Uint32(header[4:])

	return &Ipv4Location{
		IPDB: wry.IPDB[uint32]{
			Data: fileData,

			OffLen:   3,
			IPLen:    4,
			IPCnt:    (end-start)/7 + 1,
			IdxStart: start,
			IdxEnd:   end,
		},
	}, nil
}

func (db *Ipv4Location) find(query string) (result *wry.Result, err error) {
	ip := net.ParseIP(query)
	if ip == nil {
		return nil, errors.New("query should be IPv4")
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return nil, errors.New("query should be IPv4")
	}
	ip4uint := binary.BigEndian.Uint32(ip4)

	offset := db.SearchIndexV4(ip4uint)
	if offset <= 0 {
		return nil, errors.New("query not valid")
	}

	reader := wry.NewReader(db.Data)
	reader.Parse(offset + 4)
	return reader.Result.DecodeGBK(), nil
}

func (db *Ipv4Location) Find(query string) string {
	result, err := db.find(query)
	if err != nil || result == nil {
		return ""
	}
	r := strings.ReplaceAll(result.Country, "–", " ")
	return r
}

func checkIPv4File(data []byte) bool {
	if len(data) < 8 {
		return false
	}

	header := data[0:8]
	start := binary.LittleEndian.Uint32(header[:4])
	end := binary.LittleEndian.Uint32(header[4:])

	if start >= end || uint32(len(data)) < end+7 {
		return false
	}

	return true
}
