package qqwry

import (
	"encoding/binary"
	"errors"
	"github.com/ipipdotnet/ipdb-go"
	"net"
	"os"
	"strings"
	"sync"
)

var (
	data          []byte
	dataLen       uint32
	ipdbCity      *ipdb.City
	dataType      = dataTypeDat
	locationCache = &sync.Map{}
)

const (
	dataTypeDat  = 0
	dataTypeIpdb = 1
)

const (
	indexLen      = 7
	redirectMode1 = 0x01
	redirectMode2 = 0x02
)

type Location struct {
	Country  string // 国家
	Province string // 省份
	City     string // 城市
	District string // 区县
	ISP      string // 运营商
	IP       string // IP地址
}

// IsLoaded 检查数据库是否已加载
func IsLoaded() bool {
	switch dataType {
	case dataTypeIpdb:
		return ipdbCity != nil
	case dataTypeDat:
		return data != nil && dataLen > 0
	default:
		return false // 不支持的类型或未初始化
	}
}

// LoadDBFile 从文件加载IP数据库
func LoadDBFile(filepath string) error {
	info, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return errors.New("file does not exist: " + filepath)
	}
	if !info.Mode().IsRegular() {
		return errors.New("not a regular file: " + filepath)
	}
	if info.Size() == 0 {
		return errors.New("file is empty: " + filepath)
	}

	body, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	LoadDBData(body)
	return nil
}

// LoadDBData 从内存加载IP数据库
func LoadDBData(database []byte) {
	if string(database[6:11]) == "build" {
		dataType = dataTypeIpdb
		loadCity, err := ipdb.NewCityFromBytes(database)
		if err != nil {
			panic(err)
		}
		ipdbCity = loadCity
		return
	}
	data = database
	dataLen = uint32(len(data))
}

// QueryIPByDat 从dat查询IP，仅加载dat格式数据库时使用
func QueryIPByDat(ipv4 string) (location *Location, err error) {
	ip := net.ParseIP(ipv4).To4()
	if ip == nil {
		return nil, errors.New("ip is not ipv4")
	}
	ip32 := binary.BigEndian.Uint32(ip)
	posA := binary.LittleEndian.Uint32(data[:4])
	posZ := binary.LittleEndian.Uint32(data[4:8])
	var offset uint32 = 0
	for {
		mid := posA + (((posZ-posA)/indexLen)>>1)*indexLen
		buf := data[mid : mid+indexLen]
		_ip := binary.LittleEndian.Uint32(buf[:4])
		if posZ-posA == indexLen {
			offset = byte3ToUInt32(buf[4:])
			buf = data[mid+indexLen : mid+indexLen+indexLen]
			if ip32 < binary.LittleEndian.Uint32(buf[:4]) {
				break
			} else {
				offset = 0
				break
			}
		}
		if _ip > ip32 {
			posZ = mid
		} else if _ip < ip32 {
			posA = mid
		} else if _ip == ip32 {
			offset = byte3ToUInt32(buf[4:])
			break
		}
	}
	if offset <= 0 {
		return nil, errors.New("ip not found")
	}
	posM := offset + 4
	mode := data[posM]
	var ispPos uint32
	var addr, isp string
	switch mode {
	case redirectMode1:
		posC := byte3ToUInt32(data[posM+1 : posM+4])
		mode = data[posC]
		posCA := posC
		if mode == redirectMode2 {
			posCA = byte3ToUInt32(data[posC+1 : posC+4])
			posC += 4
		}
		for i := posCA; i < dataLen; i++ {
			if data[i] == 0 {
				addr = string(data[posCA:i])
				break
			}
		}
		if mode != redirectMode2 {
			posC += uint32(len(addr) + 1)
		}
		ispPos = posC
	case redirectMode2:
		posCA := byte3ToUInt32(data[posM+1 : posM+4])
		for i := posCA; i < dataLen; i++ {
			if data[i] == 0 {
				addr = string(data[posCA:i])
				break
			}
		}
		ispPos = offset + 8
	default:
		posCA := offset + 4
		for i := posCA; i < dataLen; i++ {
			if data[i] == 0 {
				addr = string(data[posCA:i])
				break
			}
		}
		ispPos = offset + uint32(5+len(addr))
	}
	if addr != "" {
		addr = strings.TrimSpace(gb18030Decode([]byte(addr)))
	}
	ispMode := data[ispPos]
	if ispMode == redirectMode1 || ispMode == redirectMode2 {
		ispPos = byte3ToUInt32(data[ispPos+1 : ispPos+4])
	}
	if ispPos > 0 {
		for i := ispPos; i < dataLen; i++ {
			if data[i] == 0 {
				isp = string(data[ispPos:i])
				if isp != "" {
					if strings.Contains(isp, "CZ88.NET") {
						isp = ""
					} else {
						isp = strings.TrimSpace(gb18030Decode([]byte(isp)))
					}
				}
				break
			}
		}
	}
	location = SplitCZResult(addr, isp, ipv4)
	locationCache.Store(ipv4, location)
	return location, nil
}

// QueryIPByIpdb 从ipdb查询IP，仅加载ipdb格式数据库时使用
func QueryIPByIpdb(ip string) (location *Location, err error) {
	ret, err := ipdbCity.Find(ip, "CN")
	if err != nil {
		return
	}
	location = SplitCZResult(ret[0], ret[1], ip)
	locationCache.Store(ip, location)
	return location, nil
}

// QueryIP 从内存或缓存查询IP
func QueryIP(ip string) (location *Location, err error) {
	if v, ok := locationCache.Load(ip); ok {
		return v.(*Location), nil
	}
	switch dataType {
	case dataTypeDat:
		return QueryIPByDat(ip)
	case dataTypeIpdb:
		return QueryIPByIpdb(ip)
	default:
		return nil, errors.New("data type not support")
	}
}
