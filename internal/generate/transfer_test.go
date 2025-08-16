package generate

import (
	"github.com/winezer0/cdnAnalyzer/internal/analyzer"
	"github.com/winezer0/cdnAnalyzer/pkg/fileutils"
	"testing"
)

func TestAddDataToCdnCategory(t *testing.T) {
	// 1.加载数据源
	sourceJson := "C:\\Users\\WINDOWS\\Downloads\\sources.json"
	sourceData := analyzer.NewEmptyCDNData()
	if err := fileutils.ReadJsonToStruct(sourceJson, sourceData); err != nil {
		panic(err)
	}

	// 2.读取 ASN 文本文件内容
	asnFile := "cdn_asn.txt"
	asnList, err := fileutils.ReadTextToList(asnFile)
	if err != nil {
		panic(err)
	}
	err = analyzer.AddDataToCdnDataCategory(sourceData, analyzer.CategoryCDN, analyzer.FieldASN, "UNKNOWN", asnList)
	if err != nil {
		panic(err)
	}

	// 3.读取 IP 文本文件内容
	ipsFile := "cdn_ips.txt"
	ipsList, err := fileutils.ReadTextToList(ipsFile)
	if err != nil {
		panic(err)
	}
	err = analyzer.AddDataToCdnDataCategory(sourceData, analyzer.CategoryCDN, analyzer.FieldIP, "UNKNOWN", ipsList)
	if err != nil {
		panic(err)
	}
	// 4.读取 CNAME 文本文件内容
	cnameFile := "cdn_ips.txt"
	cnameList, err := fileutils.ReadTextToList(cnameFile)
	if err != nil {
		panic(err)
	}
	err = analyzer.AddDataToCdnDataCategory(sourceData, analyzer.CategoryCDN, analyzer.FieldCNAME, "UNKNOWN", cnameList)
	if err != nil {
		panic(err)
	}
	// 5. 写入文件
	outFile := sourceJson + ".update.json"
	fileutils.WriteJson(outFile, *sourceData)
}

func TestMergeSameData(t *testing.T) {
	// 加载并转换 cloud yaml
	cloudYamlFile := "C:\\Users\\WINDOWS\\Desktop\\cdnAnalyzer\\assets\\cloud_keys.yml"
	cloudYamlTransData := TransferCloudKeysYaml(cloudYamlFile)
	//fileutils.WriteJson("cloudYamlTransData.json", cloudYamlTransData)

	// 加载并转换 cdn.yml
	// https://github.com/4ft35t/cdn/blob/master/src/cdn.yml
	cdnYamlFile := "C:\\Users\\WINDOWS\\Desktop\\cdnAnalyzer\\assets\\cdn.yml"
	cdnYamlTransData := TransferCdnDomainsYaml(cdnYamlFile)
	//fileutils.WriteJson("cdnYamlTransData.json", cdnYamlTransData)

	// 国外源：https://github.com/projectdiscovery/cdncheck/blob/main/sources_data.json
	// 加载sources_foreign.json 数据的合并
	sourceDataJson := "C:\\Users\\WINDOWS\\Desktop\\cdnAnalyzer\\assets\\sources_foreign.json"
	sourceData := TransferPDCdnCheckJson(sourceDataJson)
	//fileutils.WriteJson("sourceData.json", sourceData)

	// 国内源：https://github.com/hanbufei/isCdn/blob/main/client/data/sources_china.json
	// 加载 sources_china.json
	sourceChinaJson := "C:\\Users\\WINDOWS\\Desktop\\cdnAnalyzer\\assets\\sources_china.json"
	sourceChina := TransferPDCdnCheckJson(sourceChinaJson)
	//fileutils.WriteJson("sourceChina.json", sourceChina)

	// 合并写入文件
	sourceMerge, _ := analyzer.MergeCdnDataList(*sourceData, *sourceChina, *cdnYamlTransData, *cloudYamlTransData)
	fileutils.WriteJson("sources.json", sourceMerge)
}

func TestTransferProviderYAML(t *testing.T) {
	providerYaml := "C:\\Users\\WINDOWS\\Desktop\\provider.yaml"
	sourceData := TransferProviderYAML(providerYaml)
	outFile := providerYaml + ".update.json"
	fileutils.WriteJson(outFile, *sourceData)
}

func TestMergeChinaCdnData(t *testing.T) {
	// 国内源：https://github.com/hanbufei/isCdn/blob/main/client/data/sources_china.json
	sourceChinaJson := "C:\\Users\\WINDOWS\\Desktop\\sources_china.json"
	sourceChina := TransferPDCdnCheckJson(sourceChinaJson)

	// 国内源：https://github.com/mabangde/cdncheck_cn/blob/main/sources_data.json
	sourceChinaJson2 := "C:\\Users\\WINDOWS\\Desktop\\sources_china2.json"
	sourceChina2 := TransferPDCdnCheckJson(sourceChinaJson2)

	// 合并写入文件
	sourceMerge, _ := analyzer.MergeCdnDataList(*sourceChina, *sourceChina2)
	fileutils.WriteJson("C:\\Users\\WINDOWS\\Desktop\\sources_china.json", sourceMerge)
}

func TestGenEmptyCdnData(t *testing.T) {
	cdnData := analyzer.NewEmptyCDNData()
	fileutils.WriteJson("C:\\Users\\WINDOWS\\Desktop\\sources_added.json", cdnData)

}
