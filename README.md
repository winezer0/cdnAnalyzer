# cdnAnalyzer
CDN Check On Golang

目前最完善的CDN分析工具 便于快速筛选CDN资产信息 基于IP|ASN|CNAME|IPlocate等多种方案

## 免责声明
继续阅读文章或使用工具视为您已同意NOVASEC免责声明

[NOVASEC免责声明](https://mp.weixin.qq.com/s/iRWRVxkYu7Fx5unxA34I7g)


## 功能介绍
- 使用GO编写
- 支持输入(URL|Domain|IP)等多种格式目标
- 会进行(域名解析|IP分析|CDN|WAF|Cloud)等信息分析.
- 自定义多个DNS服务器进行解析
- 自定义城市IP地址进行EDNS分析
- 支持通过ASN, qqwry, ipv6数 据库识别IP归属信息
- 通过数据源对资产信息进行CDN|WAF|CLoud信息分析
- 输出CSV、JSON、TXT、SYS日志格式
- 实现IP数据库自动下载, 目前是直接从github下载
- 实现CDN解析数据自动更新, 目前是直接通过Github工作流实现每天自动更新

## TODO
- 寻找更多的CDN数据源信息

## 已集成数据源
```
qqwry.dat
- https://github.com/metowolf/qqwry.dat

zxipv6wry.db
- https://github.com/winezer0/cdnAnalyzer/blob/main/assets/zxipv6wry.db

geolite2-asn-ipv4.mmdb
- https://github.com/sapics/ip-location-db/blob/main/geolite2-asn-mmdb/geolite2-asn-ipv4.mmdb

geolite2-asn-ipv6.mmdb
- https://github.com/sapics/ip-location-db/blob/main/geolite2-asn-mmdb/geolite2-asn-ipv6.mmdb

dns-resolvers(自定义的)
- https://github.com/winezer0/cdnAnalyzer/blob/main/assets/resolvers.txt

city_ip.csv
- https://github.com/winezer0/cdnAnalyzer/blob/main/assets/city_ip.csv 

nali cdn.yml
- https://raw.githubusercontent.com/4ft35t/cdn/master/src/cdn.yml

cloud_keys.yml (自定义的)
- https://github.com/winezer0/cdnAnalyzer/blob/main/assets/cloud_keys.yml

sources_china.json
- https://github.com/hanbufei/isCdn/blob/main/client/data/sources_china.json

sources_foreign.json
- https://github.com/projectdiscovery/cdncheck/blob/main/sources_data.json

provider_foreign.yaml
- https://github.com/projectdiscovery/cdncheck/blob/main/cmd/generate-index/provider.yaml

unknown-cdn-cname.txt
- https://github.com/alwaystest18/cdnChecker/blob/master/cdn_cname

sources_china2.json
- https://github.com/mabangde/cdncheck_cn/blob/main/sources_data.json
```


### MIND
```
│
├── 输入类型判断
│   ├── 域名
│   │   ├── DNS/EDNS 解析
│   │   │   ├── 获取 IP 信息
│   │   │   └── 获取 CNAME 信息
│   ├── IP   
│   │   ├── IP 归属地信息查询
│   │   │   ├── 查询 IP 地理位置信息
│   │   │   └── 查询 IP ASN 信息
│   ├── INFO  
│   │   ├── 基于 IP|ASN|CNAME 分析 CDN|WAF|Could
│   │   │   ├── 基于 CNAME 判断 CDN|WAF|Could
│   │   │   ├── 基于 IPLocate 判断 CDN|WAF|Could
│   │   │   ├── 基于 ASN 判断 CDN|WAF|Could
│   │   │   ├── 基于 IP CIDR 判断 CDN|WAF|Could
│   │   │   └── 基于 IP 解析结果数量分析是否存在CDN

```

### 功能实现
1. 实现 域名 DNS 解析 查询 CNAME|A|AAAA 记录
2. 实现 域名 EDNS 解析 查询 A|AAAA 记录
3. 实现 IPv4 归属地查询
   - 纯真IP库 https://github.com/zu1k/nali
4. 实现 IPv6 归属地查询
   - ipv6wry https://github.com/zu1k/nali
5. 实现 IPv4/IPv6 ASN号|ASN组织查询
    - geolite2-asn https://github.com/sapics/ip-location-db/tree/main/geolite2-asn-mmdb
6. 实现 CDN | WAF | CLOUD 信息判断
   - nali cdn.yml
   - cdncheck  source_data.json
   - cdncheck  source_china.json
   - other custom keys




### 其他思路(未实现)：
CDN API查询【参考 [YouChenJun/CheckCdn](https://github.com/YouChenJun/CheckCdn)】


### 开发或数据库参考
cdncheck  | nali | nemo_go | ip-location-db 等等

### 联系方式
如需获取更多信息、技术支持或定制服务，请通过以下方式联系我们：
NOVASEC微信公众号或通过社交信息联系开发者【酒零】

![NOVASEC0](https://raw.githubusercontent.com/winezer0/mypics/refs/heads/main/NOVASEC0.jpg)


