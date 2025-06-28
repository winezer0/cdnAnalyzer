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
- 实现IP数据库自动下载, 目前是直接从github下载

## TODO
- 编写代码实现命令行更新source.json信息 
- 编写工作流实现source.json自动推送到github
- 编写工作流实现更新ip数据库等大文件

## 安装方式
安装可执行程序后, 还需要补充数据文件.

### go install 安装
```
go install github.com/winezer0/cdnAnalyzer/cmd/docheck@latest
```

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



### 依赖数据库文件下载和更新
```
依赖数据库文件已实现自动更新 具体下载路径 请查看config.yaml文件
```

### 配置文件使用
现在支持通过YAML配置文件来管理所有参数和数据库下载配置。使用 `-c` 参数指定配置文件路径：

```
cdnAnalyzer.exe -c config.yaml
```
命令行参数会覆盖配置文件中的设置，因此您可以同时使用配置文件和命令行参数。


### 使用示例
提示: 在window下使用-t /t 是相同的,只是会自动根据操作系统来显示参数标志符.
```
λ cdnAnalyzer.exe -h
Usage:
  cdnAnalyzer [OPTIONS]

CDN信息分析检查工具, 用于检查(URL|Domain|IP)等格式目标所使用的(域名解析|IP分析|CDN|WAF|Cloud)等信息.
Application Options:
  /c, /config:                              YAML配置文件路径
  /t, /target:                              目标文件路径|目标字符串列表(逗号分隔)
  /T, /target-type:[string|file|sys]        目标数据类型: string/file/sys (default: string)
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

