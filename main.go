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
	domain := "example.com"
	timeout := 3 * time.Second

	resolvers, err := filetools.ReadFileToList("resolvers.txt")
	if err != nil {
		fmt.Println("加载DNS服务器失败:", err)
		os.Exit(1)
	}

	result := dnsquery.QueryAllDNSWithMultiResolvers(domain, resolvers, timeout, 5)
	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(b))
}
