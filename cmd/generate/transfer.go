package generate

import (
	"cdnCheck/fileutils"
	"cdnCheck/maputils"
	"cdnCheck/models"
	"fmt"
	"strings"
)

// TransferNaliCdnYaml  实现nali cdn.yml到json的转换
func TransferNaliCdnYaml(path string) *models.CDNData {
	// 数据来源 https://github.com/4ft35t/cdn/blob/master/src/cdn.yml
	// 初始化 CDNData 结构
	cdnData := models.NewCDNData()

	// 1. 读取 YAML 到 NaliCdnData
	var yamlData NaliCdnData
	err := fileutils.ReadYamlToStruct(path, &yamlData)
	if err != nil {
		panic(err)
	}

	// 2. 构建 cname map[string][]string 并赋值给 cdnData.CDN.CNAME
	for domain, info := range yamlData {
		cdnData.CDN.CNAME[info.Name] = append(cdnData.CDN.CNAME[info.Name], domain)
	}

	return cdnData
}

// TransferCdnCheckJson  实现cdncheck数据源到json的转换
func TransferCdnCheckJson(path string) *models.CDNData {
	//1、加载cdncheck数据源
	var cdnCheckData CdnCheckData
	err := fileutils.ReadJsonToStruct(path, &cdnCheckData)
	if err != nil {
		panic(err)
	}

	//2、转换数据
	cdnData := models.NewCDNData()
	// 将 cdn/waf/cloud 的值作为 IP 数据填充到对应字段
	cdnData.CDN.IP = maputils.CopyMap(cdnCheckData.CDN)
	cdnData.WAF.IP = maputils.CopyMap(cdnCheckData.WAF)
	cdnData.Cloud.IP = maputils.CopyMap(cdnCheckData.Cloud)

	// 合并 common 到 cdn.cname
	for provider, cnames := range cdnCheckData.Common {
		cdnData.CDN.CNAME[provider] = append([]string{}, cnames...)
	}
	return cdnData
}

// DataType 表示要操作的数据类型
type DataType int

const (
	DataTypeIP DataType = iota
	DataTypeASN
	DataTypeCNAME
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

// AddDataToCdnCategory 将文本文件中的数据添加到 CDNData 的对应字段中
func AddDataToCdnCategory(
	sourceConfig string,
	dataFile string,
	providerKey string,
	dataType DataType,
) (*models.CDNData, error) {

	// 1. 加载原始配置
	cdnData := models.NewCDNData()
	if err := fileutils.ReadJsonToStruct(sourceConfig, cdnData); err != nil {
		return nil, err
	}

	// 2. 读取文本文件内容
	dataList, err := fileutils.ReadTextToList(dataFile)
	if err != nil {
		return nil, err
	}

	// 3. 根据 dataType 选择对应的 map 和比较方式
	var targetMap map[string][]string
	var shouldNormalize bool

	switch dataType {
	case DataTypeIP:
		targetMap = cdnData.CDN.IP
		shouldNormalize = false
	case DataTypeASN:
		targetMap = cdnData.CDN.ASN
		shouldNormalize = true
	case DataTypeCNAME:
		targetMap = cdnData.CDN.CNAME
		shouldNormalize = false
	default:
		return nil, fmt.Errorf("unsupported data type")
	}

	// 4. 遍历 dataList 添加数据
	for _, item := range dataList {
		if shouldNormalize {
			item = normalizeASN(item)
		}

		found := false
		for _, existingList := range targetMap {
			for _, existing := range existingList {
				if shouldNormalize {
					existing = normalizeASN(existing)
				}

				if existing == item {
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			targetMap[providerKey] = append(targetMap[providerKey], item)
		}
	}

	return cdnData, nil
}
