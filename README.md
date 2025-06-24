# CDNCheck
CDN Check On Golang

GO编写的CDN信息分析检查工具 用于检查(URL|Domain|IP)等格式目标所使用的(域名解析|IP分析|CDN|WAF|Cloud)等信息.

## 功能介绍
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
git clone --depth 1 https://github.com/winezer0/cdncheck
go build -ldflags="-s -w" -o cdncheck.exe ./cmd/docheck/main.go
```

### release安装
```
通过workflow编译的程序将自动发布到release中:
https://github.com/winezer0/cdncheck/releases
```

### 依赖数据库文件下载和更新(暂未实现自动更新)
```
DNS服务器 (基本无需更新)
   https://github.com/winezer0/cdncheck/blob/main/asset/resolvers.txt
   
城市对应IP示例 (基本无需更新)
   https://github.com/winezer0/cdncheck/blob/main/asset/city_ip.csv
      
CDN|WAF|云数据库 (后续实现自动更新)
   https://github.com/winezer0/cdncheck/blob/main/asset/source.json

ASN信息数据库 (依赖于其他项目)
   ipv4: https://github.com/sapics/ip-location-db/blob/main/geolite2-asn-mmdb/geolite2-asn-ipv4.mmdb
   ipv6: https://github.com/sapics/ip-location-db/blob/main/geolite2-asn-mmdb/geolite2-asn-ipv6.mmdb 

IP定位数据库 (建议周期性更新)
   ipv4: https://github.com/metowolf/qqwry.dat/releases/latest 
   ipv6: https://github.com/winezer0/cdncheck/blob/main/asset/zxipv6wry.db   (库文件停止更新)

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
   - 纯真IP库 https://github.com/xiaoqidun/qqwry
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

