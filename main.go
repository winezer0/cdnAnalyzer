package main

import (
	"flag"
	"fmt"
	"github.com/winezer0/cdnAnalyzer/ecs_query"
	"github.com/winezer0/cdnAnalyzer/file_utils"
	"os"
	"strings"
)

func main() {
	// 定义命令行标志
	singleDomain := flag.String("d", "", "input single domain")
	inputFile := flag.String("f", "", "input domain file path")
	outputFile := flag.String("o", "output.csv", "output domain file path")

	// 解析命令行标志
	flag.Parse()

	//读取域名进行操作
	var domains []string
	if *singleDomain != "" {
		domains = append(domains, *singleDomain)
	}

	// 读取文件并将内容添加到 domains 列表中
	if *inputFile != "" {
		domains2, err := file_utils.ReadFileToList(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file [%s]: %v\n", *inputFile, err)
			os.Exit(1)
		}
		domains = append(domains, domains2...)
	}

	// 检查最终的 domains 列表是否为空
	if len(domains) == 0 {
		fmt.Fprintln(os.Stderr, "No domains provided. Please specify either --domain or --inputFile.")
		flag.Usage()
		os.Exit(1)
	}

	// 遍历每个域名并调用 DoEcsQuery 函数
	var results []map[string]string
	for index, domain := range domains {
		adders, err := ecs_query.DoEcsQuery(domain)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error querying domain %s: %v\n", domain, err)
			continue
		}

		result := map[string]string{
			"domain":    domain,
			"ipNumbers": fmt.Sprintf("%d", len(adders)),
			"likeCdn":   fmt.Sprintf("%v", len(adders) > 1),
			"ipAddress": strings.Join(adders, ", "),
		}

		fmt.Printf("[%d/%d] Domain:[%s] ipNum:[%d] likeCdn:[%v]\n", index+1, len(domains), domain, len(adders), len(adders) > 1)
		results = append(results, result)
	}

	// 将结果写入输出文件
	headers := []string{"domain", "ipNumbers", "likeCdn", "ipAddress"}

	if err := file_utils.WriteCSV(*outputFile, results, true, "a+", headers); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("Result success writed to file: [%s]\n", *outputFile)

	}
}
