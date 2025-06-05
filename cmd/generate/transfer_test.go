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
	dataFile := "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\cdn_asn.txt"

	// 创建空结构体指针，由 ReadJsonToStruct 自动填充
	cdnData := models.NewEmptyCDNDataAddress()
	if err := fileutils.ReadJsonToStruct(inFile, cdnData); err != nil {
		panic(err)
	}

	// 2. 读取文本文件内容
	dataList, err := fileutils.ReadTextToList(dataFile)
	if err != nil {
		panic(err)
	}

	AddDataToCdnCategory(cdnData, dataList, "UNKNOWN", DataTypeASN)
	outFile := inFile + ".cdn_asn.json"
	fileutils.WriteJsonFromStruct(outFile, *cdnData)
}
