package qqwry

import (
	"testing"
)

func init() {
	//支持直接加载 qqwry.dat 或 qqwry.ipdb 文件
	dbpath := "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\qqwry.ipdb"
	if err := LoadFile(dbpath); err != nil {
		panic(err)
	}
}

func TestQueryIP(t *testing.T) {
	queryIp := "2409:8929:52b:36d9:8f6e:2e8b:a35:1148"
	location, err := QueryIP(queryIp)
	if err != nil {
		t.Fatal(err)
	}
	emptyVal := func(val string) string {
		if val != "" {
			return val
		}
		return "未知"
	}
	t.Logf("国家：%s，省份：%s，城市：%s，区县：%s，运营商：%s",
		emptyVal(location.Country),
		emptyVal(location.Province),
		emptyVal(location.City),
		emptyVal(location.District),
		emptyVal(location.ISP),
	)
}
