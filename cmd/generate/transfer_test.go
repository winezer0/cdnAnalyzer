package generate

import (
	"cdnCheck/fileutils"
	"cdnCheck/models"
	"testing"
)

func TestTransferCdnCnameYamlToJson(t *testing.T) {
	// https://github.com/4ft35t/cdn/blob/master/src/cdn.yml
	inFile := "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\cdn_cname.yml"
	outFile := inFile + ".new.json"
	cdnData := TransferNaliCdnYaml(inFile)
	fileutils.WriteJsonFromStruct(outFile, cdnData)
}

func TestTransferCdnCheckJson(t *testing.T) {
	//国外源：https://github.com/projectdiscovery/cdncheck/blob/main/sources_data.json
	//国内源：https://github.com/hanbufei/isCdn/blob/main/client/data/sources_china.json

	inFile := "C:\\Users\\WINDOWS\\Downloads\\sources_china.json"
	outFile := inFile + ".new.json"
	cdnData := TransferCdnCheckJson(inFile)
	fileutils.WriteJsonFromStruct(outFile, cdnData)
}

func TestAddDataToCdnCategory(t *testing.T) {
	inFile := "C:\\Users\\WINDOWS\\Downloads\\sources_data.json.new.json"

	// 创建空结构体指针，由 ReadJsonToStruct 自动填充
	cdnData := models.NewEmptyCDNDataPointer()
	if err := fileutils.ReadJsonToStruct(inFile, cdnData); err != nil {
		panic(err)
	}

	// 2. 读取 ASN 文本文件内容
	asnFile := "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\cdn_asn.txt"
	asnList, err := fileutils.ReadTextToList(asnFile)
	if err != nil {
		panic(err)
	}
	AddDataToCdnDataCategory(cdnData, asnList, "UNKNOWN", DataTypeASN)

	// 3. 读取 IP 文本文件内容
	ipsFile := "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\cdn_ips.txt"
	ipsList, err := fileutils.ReadTextToList(ipsFile)
	if err != nil {
		panic(err)
	}
	AddDataToCdnDataCategory(cdnData, ipsList, "UNKNOWN", DataTypeIP)

	// 4. 读取 CNAMEs 文本文件内容
	cnameFile := "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\cdn_ips.txt"
	cnameList, err := fileutils.ReadTextToList(cnameFile)
	if err != nil {
		panic(err)
	}
	AddDataToCdnDataCategory(cdnData, cnameList, "UNKNOWN", DataTypeCNAME)

	// 5. 写入文件
	outFile := inFile + ".add.nemo.json"
	fileutils.WriteJsonFromStruct(outFile, *cdnData)
}
