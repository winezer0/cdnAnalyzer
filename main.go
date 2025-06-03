package main

import (
	"cdnCheck/dnsquery"
	"cdnCheck/filetools"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func main() {
	domain := "www.example.com"
	timeout := time.Second * 5

	resolvers, err := filetools.ReadFileToList("resolvers.txt")
	fmt.Printf("load resolvers: %v\n", len(resolvers))
	if err != nil {
		fmt.Println("加载DNS服务器失败:", err)
		os.Exit(1)
	}

	result := dnsquery.QueryAllDNSWithMultiResolvers(domain, resolvers, 5, timeout)
	fmt.Printf("DNS result: %v\n", result)
	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(b))
}
