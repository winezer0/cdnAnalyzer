package analyzer

import (
	"fmt"
	"github.com/winezer0/cdnAnalyzer/pkg/maputils"
	"strings"
)

const (
	FieldIP    = "ip"
	FieldASN   = "asn"
	FieldCNAME = "cname"
	FieldKEYS  = "keys"
)

const (
	CategoryCDN   = "cdn"
	CategoryWAF   = "waf"
	CategoryCloud = "cloud"
)

// normalizeASN 清洗 ASN 字符串
func normalizeASN(asn string) string {
	asn = strings.TrimSpace(asn)
	asn = strings.ToUpper(asn)
	if strings.HasPrefix(asn, "AS") {
		asn = strings.TrimPrefix(asn, "AS")
	}
	return asn
}

// NormalizeASNList 批量清洗 ASN 字符串列表
func NormalizeASNList(asns []string) []string {
	result := make([]string, 0, len(asns))
	for _, asn := range asns {
		result = append(result, normalizeASN(asn))
	}
	return result
}

// MergeCdnDataList 实现多个cdnData数据的合并
func MergeCdnDataList(cdnDataList ...CDNData) (*CDNData, error) {
	mergedCdnData := NewEmptyCDNData()
	var mergedMap map[string]interface{}

	for _, cdnData := range cdnDataList {
		//转换为Json对象后进行通用合并操作
		cdnDataString := maputils.AnyToJsonStr(cdnData)
		cdnDataMap, err := maputils.ParseJSON(cdnDataString)
		if err != nil {
			return mergedCdnData, err
		}

		if mergedMap == nil {
			mergedMap = cdnDataMap
		} else {
			mergedMap = maputils.DeepMerge(mergedMap, cdnDataMap)
		}
	}

	//转换回对象格式
	if err := maputils.ConvertMapsToStructs(mergedMap, mergedCdnData); err != nil {
		return nil, err
	}
	return mergedCdnData, nil
}

// AddDataToCdnDataCategory 将 dataList 中的数据添加到 CDNData 的指定 Category 和字段中
func AddDataToCdnDataCategory(cdnData *CDNData, categoryName string, fieldName string, providerKey string, dataList []string) error {
	// 获取对应 Category 的 map[string][]string 字段
	targetMap := GetCategoryField(cdnData, categoryName, fieldName)
	if targetMap == nil {
		return fmt.Errorf("failed to get target map for category: %s, field: %s", categoryName, fieldName)
	}

	//是否需要格式化处理
	shouldNormalize := fieldName == FieldASN

	// 遍历 dataList 添加数据（去重后插入）
	for _, item := range dataList {
		modifiedItem := item
		if shouldNormalize {
			modifiedItem = normalizeASN(item)
		}

		found := false
		for _, existingList := range targetMap {
			for _, existing := range existingList {
				comparison := existing
				if shouldNormalize {
					comparison = normalizeASN(existing)
				}
				if comparison == modifiedItem {
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			if targetMap[providerKey] == nil {
				targetMap[providerKey] = make([]string, 0)
			}
			targetMap[providerKey] = append(targetMap[providerKey], item)
		}
	}

	return nil
}

// GetCategoryField 根据 categoryName 和 fieldName 返回对应的 map[string][]string
func GetCategoryField(cdnData *CDNData, categoryName, fieldName string) map[string][]string {
	var category *Category

	switch categoryName {
	case "cdn":
		category = &cdnData.CDN
	case "waf":
		category = &cdnData.WAF
	case "cloud":
		category = &cdnData.CLOUD
	default:
		return nil
	}

	switch fieldName {
	case "ip":
		return category.IP
	case "asn":
		return category.ASN
	case "cname":
		return category.CNAME
	case "keys":
		return category.KEYS
	default:
		return nil
	}
}
