package generate

import (
	"cdnCheck/fileutils"
	"fmt"
	"testing"
)

func TestTransferCdnCnameYamlToJson(t *testing.T) {
	// https://github.com/4ft35t/cdn/blob/master/src/cdn.yml
	path := "C:\\Users\\WINDOWS\\Desktop\\CDNCheck\\asset\\cdn_cname.yml"
	TransferCdnCnameYamlToJson(path)
}

func TestTransferSourceJsonToJson(t *testing.T) {
	//国外源：https://github.com/projectdiscovery/cdncheck/blob/main/sources_data.json
	//国内源：https://github.com/hanbufei/isCdn/blob/main/client/data/sources_china.json

	path := "C:\\Users\\WINDOWS\\Downloads\\sources_data.json"
	cdnCheckSourceData, err := LoadCdnCheckSourceDataJson(path)

	if err != nil {
		fmt.Println("加载配置失败:", err)
		return
	}

	// 打印结果
	fmt.Printf("CDN: %+v\n", len(cdnCheckSourceData.CDN))
	fmt.Printf("WAF: %+v\n", len(cdnCheckSourceData.WAF))
	fmt.Printf("Cloud: %+v\n", len(cdnCheckSourceData.Cloud))
	fmt.Printf("Common: %+v\n", len(cdnCheckSourceData.Common))

	cdnData := ConvertConfigToCDNData(cdnCheckSourceData)
	outFile := "C:\\Users\\WINDOWS\\Downloads\\cdn_data.json"
	fileutils.WriteJsonFromStruct(outFile, cdnData)
}
