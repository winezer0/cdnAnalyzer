package generate

import (
	"cdnCheck/fileutils"
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

	inFile := "C:\\Users\\WINDOWS\\Downloads\\sources_data.json"
	outFile := inFile + ".new.json"
	cdnData := TransferCdnCheckJson(inFile)
	fileutils.WriteJsonFromStruct(outFile, cdnData)
}
