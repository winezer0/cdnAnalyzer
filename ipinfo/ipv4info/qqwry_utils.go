package ipv4info

import (
	"bytes"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"strings"
)

// LocationToStr Location 结构转字符串
func LocationToStr(loc Location) string {
	parts := make([]string, 0)

	if loc.Country != "" {
		parts = append(parts, loc.Country)
	}
	if loc.Province != "" {
		parts = append(parts, loc.Province)
	}
	if loc.City != "" {
		parts = append(parts, loc.City)
	}
	if loc.District != "" {
		parts = append(parts, loc.District)
	}
	if loc.ISP != "" {
		parts = append(parts, loc.ISP)
	}
	//if loc.IP != "" {
	//	parts = append(parts, "IP:"+loc.IP)
	//}
	return strings.Join(parts, " ")
}

// SplitCZResult 按照调整后的纯真社区版IP库地理位置格式返回结果
func SplitCZResult(addr string, isp string, ipv4 string) (location *Location) {
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

// byte3ToUInt32 将 3 个字节的数据转换为一个 uint32 类型的整数（用于处理小端序数据）
func byte3ToUInt32(data []byte) uint32 {
	i := uint32(data[0]) & 0xff
	i |= (uint32(data[1]) << 8) & 0xff00
	i |= (uint32(data[2]) << 16) & 0xff0000
	return i
}

// gb18030Decode 将 GB18030 编码的字节切片解码为 UTF-8 编码的字符串。
func gb18030Decode(src []byte) string {
	in := bytes.NewReader(src)
	out := transform.NewReader(in, simplifiedchinese.GB18030.NewDecoder())
	d, _ := io.ReadAll(out)
	return string(d)
}
