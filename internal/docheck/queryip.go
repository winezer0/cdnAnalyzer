package docheck

import (
	"github.com/winezer0/cdninfo/internal/analyzer"
	"github.com/winezer0/ipinfo/pkg/queryip"
	"github.com/winezer0/xutils/logging"
)

// QueryIPInfo 进行IP信息查询
func QueryIPInfo(ipDbConfig *queryip.IPDbConfig, checkInfos []*analyzer.CheckInfo) []*analyzer.CheckInfo {
	// 初始化IP数据库引擎
	ipEngines, err := queryip.InitDBEngines(ipDbConfig)
	if err != nil {
		logging.Fatalf("初始化数据库失败: %v", err)
	}
	defer ipEngines.Close()

	//对 checkInfos 中的A/AAAA记录进行IP信息查询，并赋值回去
	for _, checkInfo := range checkInfos {
		if len(checkInfo.A) > 0 || len(checkInfo.AAAA) > 0 {
			ipInfo, err := ipEngines.QueryIPInfo(checkInfo.A, checkInfo.AAAA)
			if err != nil {
				logging.Warnf("查询IP信息失败: %v", err)
			} else {
				checkInfo.Ipv4Locate = convertIPLocationsToMap(ipInfo.IPv4Locations)
				checkInfo.Ipv4Asn = ipInfo.IPv4AsnInfos
				checkInfo.Ipv6Locate = convertIPLocationsToMap(ipInfo.IPv6Locations)
				checkInfo.Ipv6Asn = ipInfo.IPv6AsnInfos
			}
		}
	}

	return checkInfos
}

func convertIPLocationsToMap(locations []queryip.IPLocation) []map[string]string {
	if len(locations) == 0 {
		return nil
	}

	result := make([]map[string]string, 0, len(locations))
	for _, loc := range locations {
		locationStr := ""
		if loc.IPLocate != nil {
			locationStr = loc.IPLocate.Location
		}
		result = append(result, map[string]string{loc.IP: locationStr})
	}

	return result
}
