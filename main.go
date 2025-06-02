package main

import (
	"cdnCheck/dns_query"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func main() {
	domain := "example.com"
	timeout := 3 * time.Second

	resolvers, err := dns_query.LoadFilesToList("resolvers.txt")
	if err != nil {
		fmt.Println("加载DNS服务器失败:", err)
		os.Exit(1)
	}

	result := dns_query.QueryAllDNSWithMultiResolvers(domain, resolvers, timeout, 5)
	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(b))
}
