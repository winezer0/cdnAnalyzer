# cdnAnalyzer
CDN Check On Golang

目前最完善的CDN分析工具 便于快速筛选CDN资产信息 基于IP|ASN|CNAME|IPlocate等多种方案

## 功能介绍
- 使用GO编写
- 支持输入(URL|Domain|IP)等多种格式目标
- 会进行(域名解析|IP分析|CDN|WAF|Cloud)等信息分析.
- 自定义多个DNS服务器进行解析
- 自定义城市IP地址进行EDNS分析
- 支持通过ASN, qqwry, ipv6数 据库识别IP归属信息
- 通过数据源对资产信息进行CDN|WAF|CLoud信息分析
- 输出CSV、JSON、TXT、SYS日志格式


## TODO
- 实现自动化依赖资源更新，目前可以通过transfer_test.go手动更新资源库.


## 安装方式
安装可执行程序后, 还需要补充数据文件.

### 源码安装
```
git clone --depth 1 https://github.com/winezer0/cdnAnalyzer
go build -ldflags="-s -w" -o cdnAnalyzer.exe ./cmd/docheck/main.go
```

### release安装
```
通过workflow编译的程序将自动发布到release中:
https://github.com/winezer0/cdnAnalyzer/releases
```

### 依赖数据库文件下载和更新(暂未实现自动更新)
```
DNS服务器 (基本无需更新)
   https://github.com/winezer0/cdnAnalyzer/blob/main/asset/resolvers.txt
   
城市对应IP示例 (基本无需更新)
   https://github.com/winezer0/cdnAnalyzer/blob/main/asset/city_ip.csv
      
CDN|WAF|云数据库 (后续实现自动更新)
   https://github.com/winezer0/cdnAnalyzer/blob/main/asset/source.json

ASN信息数据库 (依赖于其他项目)
   ipv4: https://github.com/sapics/ip-location-db/blob/main/geolite2-asn-mmdb/geolite2-asn-ipv4.mmdb
   ipv6: https://github.com/sapics/ip-location-db/blob/main/geolite2-asn-mmdb/geolite2-asn-ipv6.mmdb 

IP定位数据库 (建议周期性更新)
   ipv4: https://github.com/metowolf/qqwry.dat/releases/latest 
   ipv6: https://github.com/winezer0/cdnAnalyzer/blob/main/asset/zxipv6wry.db   (库文件停止更新)

提示: 目前需要将数据库文件存放在程序目录的asset文件夹中.
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


### 使用示例
提示: 在window下使用-t /t 是相同的,只是会自动根据操作系统来显示参数标志符.
```
λ cdnAnalyzer.exe -h
Usage:
  cdnAnalyzer [OPTIONS]

CDN信息分析检查工具, 用于检查(URL|Domain|IP)等格式目标所使用的(域名解析|IP分析|CDN|WAF|Cloud)等信息.
Application Options:
  /t, /target:                              目标文件路径|目标字符串列表(逗号分隔)
  /T, /target-type:[string|file|sys]        目标数据类型: string/file/sys (default: string)
  /r, /resolvers:                           DNS解析服务器配置文件路径 (default: asset/resolvers.txt)
  /n, /resolvers-num:                       选择用于解析的最大DNS服务器数量 (default: 5)
  /c, /city-map:                            EDNS城市IP映射文件路径 (default: asset/city_ip.csv)
  /m, /city-num:                            随机选择的城市数量 (default: 5)
  /d, /dns-concurrency:                     DNS并发数 (default: 5)
  /e, /edns-concurrency:                    EDNS并发数 (default: 5)
  /w, /timeout:                             超时时间(秒) (default: 5)
  /C, /query-edns-cnames                    启用EDNS CNAME查询
  /S, /query-edns-use-sys-ns                启用EDNS系统NS查询
  /a, /asn-ipv4:                            IPv4 ASN数据库路径 (default: asset/geolite2-asn-ipv4.mmdb)
  /A, /asn-ipv6:                            IPv6 ASN数据库路径 (default: asset/geolite2-asn-ipv6.mmdb)
  /4, /ipv4-db:                             IPv4地理位置数据库路径 (default: asset/qqwry.dat)
  /6, /ipv6-db:                             IPv6地理位置数据库路径 (default: asset/zxipv6wry.db)
  /s, /source:                              CDN源数据配置文件路径 (default: asset/source.json)
  /o, /output-file:                         输出结果文件路径
  /y, /output-type:[csv|json|txt|sys]       输出文件类型: csv/json/txt/sys (default: sys)
  /l, /output-level:[default|quiet|detail]  输出详细程度: default/quiet/detail (default: default)

Help Options:
  /?                                        Show this help message
  /h, /help                                 Show this help message

```
### 使用管道符传入
```
λ echo www.baidu.com | cdnAnalyzer.exe -T sys
[
  {
    "raw": "www.baidu.com",
    "fmt": "www.baidu.com",
    "is_cdn": true,
    "cdn_company": "百度旗下业务地域负载均衡系统",
    "is_waf": false,
    "waf_company": "",
    "is_cloud": false,
    "cloud_company": "",
    "ip_size_is_cdn": true,
    "ip_size": 10
  }
]

```
### 传入目标字符串
```
λ cdnAnalyzer.exe -t www.baidu.com,www.google.com
[
  {
    "raw": "www.baidu.com",
    "fmt": "www.baidu.com",
    "is_cdn": true,
    "cdn_company": "百度旗下业务地域负载均衡系统",
    "is_waf": false,
    "waf_company": "",
    "is_cloud": false,
    "cloud_company": "",
    "ip_size_is_cdn": true,
    "ip_size": 10
  },
  {
    "raw": "www.google.com",
    "fmt": "www.google.com",
    "is_cdn": false,
    "cdn_company": "",
    "is_waf": false,
    "waf_company": "",
    "is_cloud": false,
    "cloud_company": "",
    "ip_size_is_cdn": true,
    "ip_size": 10
  }
]
```

### 其他思路(未实现)：
CDN API查询【参考 [YouChenJun/CheckCdn](https://github.com/YouChenJun/CheckCdn)】


### 开发或数据库参考
cdncheck  | nali | nemo_go | ip-location-db

