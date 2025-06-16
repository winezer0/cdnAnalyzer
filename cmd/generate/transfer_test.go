package generate

import (
	"cdnCheck/fileutils"
	"cdnCheck/models"
	"testing"
)

func TestAddDataToCdnCategory(t *testing.T) {
	// 1.加载数据源
	sourceJson := "C:\\Users\\WINDOWS\\Downloads\\sources.json"
	sourceData := models.NewEmptyCDNDataPointer()
	if err := fileutils.ReadJsonToStruct(sourceJson, sourceData); err != nil {
		panic(err)
	}

	// 2.读取 ASN 文本文件内容
	asnFile := "cdn_asn.txt"
	asnList, err := fileutils.ReadTextToList(asnFile)
	if err != nil {
		panic(err)
	}
	err = models.AddDataToCdnDataCategory(sourceData, models.CategoryCDN, models.FieldASN, "UNKNOWN", asnList)
	if err != nil {
		panic(err)
	}

	// 3.读取 IP 文本文件内容
	ipsFile := "cdn_ips.txt"
	ipsList, err := fileutils.ReadTextToList(ipsFile)
	if err != nil {
		panic(err)
	}
	err = models.AddDataToCdnDataCategory(sourceData, models.CategoryCDN, models.FieldIP, "UNKNOWN", ipsList)
	if err != nil {
		panic(err)
	}
	// 4.读取 CNAME 文本文件内容
	cnameFile := "cdn_ips.txt"
	cnameList, err := fileutils.ReadTextToList(cnameFile)
	if err != nil {
		panic(err)
	}
	err = models.AddDataToCdnDataCategory(sourceData, models.CategoryCDN, models.FieldCNAME, "UNKNOWN", cnameList)
	if err != nil {
		panic(err)
	}
	// 5. 写入文件
	outFile := sourceJson + ".update.json"
	fileutils.WriteJsonFromStruct(outFile, *sourceData)
}

func TestMergeSameData(t *testing.T) {
	// 加载并转换 cloud yaml
	cloudYamlFile := "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\cloud_keys.yml"
	cloudYamlTransData := TransferCloudKeysYaml(cloudYamlFile)
	//fileutils.WriteJsonFromStruct("cloudYamlTransData.json", cloudYamlTransData)

	// 加载并转换 cdn.yml
	// https://github.com/4ft35t/cdn/blob/master/src/cdn.yml
	cdnYamlFile := "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\cdn.yml"
	cdnYamlTransData := TransferNaliCdnYaml(cdnYamlFile)
	//fileutils.WriteJsonFromStruct("cdnYamlTransData.json", cdnYamlTransData)

	// 国外源：https://github.com/projectdiscovery/cdncheck/blob/main/sources_data.json
	// 加载sources_data.json 数据的合并
	sourceDataJson := "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\sources_data.json"
	sourceData := TransferPDCdnCheckJson(sourceDataJson)
	//fileutils.WriteJsonFromStruct("sourceData.json", sourceData)

	// 国内源：https://github.com/hanbufei/isCdn/blob/main/client/data/sources_china.json
	// 加载 sources_china.json
	sourceChinaJson := "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\sources_china.json"
	sourceChina := TransferPDCdnCheckJson(sourceChinaJson)
	//fileutils.WriteJsonFromStruct("sourceChina.json", sourceChina)

	// 合并写入文件
	sourceMerge, _ := models.CdnDataMergeSafe(*sourceData, *sourceChina, *cdnYamlTransData, *cloudYamlTransData)
	fileutils.WriteJsonFromStruct("source.json", sourceMerge)
}
