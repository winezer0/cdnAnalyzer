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
	datas := []string{
		"8.8.8.8",
		"119.29.29.52",
		"2405:6f00:c602::1",
		"2409:8c1e:75b0:1120::27",
		"2402:3c00:1000:4::1",
		"2408:8652:200::c101",
		"2409:8900:103f:14f:d7e:cd36:11af:be83",
		"fe80::5c12:27dc:93a4:3426", // 链路本地地址，可能查不到地理位置
	}

	for _, queryIp := range datas {
		t.Run(queryIp, func(t *testing.T) {
			location, err := QueryIP(queryIp)
			if err != nil {
				t.Logf("查询失败：%v", err)
				t.FailNow()
			}

			emptyVal := func(val string) string {
				if val != "" {
					return val
				}
				return "UNKNOWN"
			}

			t.Logf("IP: %s -> 国家：%s，省份：%s，城市：%s，区县：%s，运营商：%s",
				queryIp,
				emptyVal(location.Country),
				emptyVal(location.Province),
				emptyVal(location.City),
				emptyVal(location.District),
				emptyVal(location.ISP),
			)
		})
	}
}
