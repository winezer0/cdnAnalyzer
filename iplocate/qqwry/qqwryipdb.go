package qqwry

import (
	"errors"
	"github.com/ipipdotnet/ipdb-go"
	"os"
	"strings"
	"sync"
)

var (
	data          []byte
	ipdbCity      *ipdb.City
	locationCache = &sync.Map{}
)

type Location struct {
	Country  string // 国家
	Province string // 省份
	City     string // 城市
	District string // 区县
	ISP      string // 运营商
	IP       string // IP地址
}

// QueryIP 从内存或缓存查询IP
func QueryIP(ip string) (location *Location, err error) {
	if v, ok := locationCache.Load(ip); ok {
		return v.(*Location), nil
	}
	return QueryIPIpdb(ip)
}

// QueryIPIpdb 从ipdb查询IP，仅加载ipdb格式数据库时使用
func QueryIPIpdb(ip string) (location *Location, err error) {
	ret, err := ipdbCity.Find(ip, "CN")
	if err != nil {
		return
	}
	location = SplitResult(ret[0], ret[1], ip)
	locationCache.Store(ip, location)
	return location, nil
}

// LoadData 从内存加载IP数据库
func LoadData(database []byte) {
	if string(database[6:11]) == "build" {
		loadCity, err := ipdb.NewCityFromBytes(database)
		if err != nil {
			panic(err)
		}
		ipdbCity = loadCity
		return
	}
	data = database
}

// LoadFile 从文件加载IP数据库
func LoadFile(filepath string) (err error) {
	// 判断文件是否存在
	info, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return errors.New("file not exist")
	}

	// 判断是否是空文件
	if info.Size() == 0 {
		return errors.New("file is empty") // 或者返回一个自定义错误，例如：errors.New("file is empty")
	}

	body, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	LoadData(body)
	return
}

// SplitResult 按照调整后的纯真社区版IP库地理位置格式返回结果
func SplitResult(addr string, isp string, ipv4 string) (location *Location) {
	location = &Location{ISP: isp, IP: ipv4}
	splitList := strings.Split(addr, "–")
	for i := 0; i < len(splitList); i++ {
		switch i {
		case 0:
			location.Country = splitList[i]
		case 1:
			location.Province = splitList[i]
		case 2:
			location.City = splitList[i]
		case 3:
			location.District = splitList[i]
		}
	}
	if location.Country == "局域网" {
		location.ISP = location.Country
	}
	return
}
