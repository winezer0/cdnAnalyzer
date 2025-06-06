# CDNCheck
CDN Check On Golang

### MIND
```
│
│
├── 输入类型判断
│   ├── 域名
│   │   ├── DNS/EDNS 解析
│   │   │   ├── 获取 IP 信息
│   │   │   └── 获取 CNAMES 信息
│   │   ├── IP 归属地信息查询
│   │   │   ├── 查询 IP 的地理位置信息
│   │   │   └── 查询 IP 的 ASN 信息
│   │   ├── CDN 判断 (基于 WAF|CDN|Could 的 IP|ASN|CNAME)
│   │   │   ├── 分析 IP 地址
│   │   │   ├── 分析 ASN 信息
│   │   │   └── 分析 CNAME 信息
│   │   └── IP 归属地信息综合判断
│   │       └── 根据 IP 归属地进行特定逻辑处理

```



### 1、进行DNS解析（A记录、CNAME 记录查询）
- 【已完成】1. 使用常规方案
- 【已完成】2. 使用EDNS方案

### 2、进行IP信息查询

- IP地址库收集整理 
  - IP_database: https://github.com/adysec/IP_database
  - ip-location-db：https://github.com/sapics/ip-location-db

- 综合IP查询工具参考
  - zu1k/nali https://github.com/zu1k/nali
  - sjzar/ips https://github.com/sjzar/ips

#### IPv4 数据库比较
```
纯真IP数据库2025新版
    https://github.com/metowolf/qqwry.dat
    提示: 部分老版本qqwry库可能不兼容新版本的qqwry.dat
    
    https://github.com/xiaoqidun/qqwry
    提示: xiaoqidun/qqwry dat格式仅支持ipv4查询。
    提示: xiaoqidun/qqwry ipdb格式支持ipv4和ipv6查询。 但IPv6查询结果不够详细.

    开发参考：https://github.com/xiaoqidun/qqwry
    
其他:
  纯真IP数据库官方版 2024年10月停止更新
    https://github.com/FW27623/qqwry

  lionsoul2014/ip2region 数据库更新频率较低
      https://github.com/lionsoul2014/ip2region
  
  IPIP.NET city.free.ipdb数据库精确度较低   
      https://ipip.net/
  
  [mmdb](https://maxmind.com/) CITY数据库
    CN增强版 更新ing https://github.com/alecthw/mmdb_china_ip_list
    开发库 https://github.com/oschwald/maxminddb-golang
    开发库 https://github.com/oschwald/geoip2-golang
```

#### IPv6数据库比较
```
[zxinc IPv6 only](https://ip.zxinc.org/) 目前已经不再进行更新
下载地址： https://raw.githubusercontent.com/ZX-Inc/zxipdb-python/main/data/ipv6wry.db
提示: 其他IPv6数据库内容更不够详细

开发参考：[zu1k/nali](https://github.com/zu1k/nali)
```


#### ASN数据库
```
GeoLite2数据库以高精度著称 支持 IPV4+IPV6
geolite2 ASN 数据库整合
  https://github.com/sapics/ip-location-db/tree/main/geolite2-asn-mmdb
  开发库 https://github.com/oschwald/maxminddb-golang
  提示: ASN整合后的数据库已经和原版本存在差异无法使用geoip2-gloang打开
  提示: 默认不支持通过ASN反查IP查询 需要自己实现
 
其他:
  ip2asn
    https://github.com/libp2p/go-libp2p-asn-util
```

其他IP数据库
```
[awdb](https://ipplus360.com/)

```

- IP 属性信息
- 【已完成】IP 归属地查询 ipv4 纯真IP库
- 【已完成】IP 归属地查询 ipv6 ipv6wry.db
- 【已完成】IP查询ASN和组织信息  geolite2-asn-ipv4/ipv6.mmdb
- 【已完成】ASN反向查询IP段落  geolite2-asn-ipv4/ipv6.mmdb



### 3、CDN信息判断

【已完成】基于本地数据进行域名CDN查询 nali cdn.yml 
【已完成】基于本地数据进行域名CDN查询 cdncheck source_data.json/source_china.json

### 4.其他思路：
CDN API查询【参考 [YouChenJun/CheckCdn](https://github.com/YouChenJun/CheckCdn)】


### DNS记录类型支持情况
```
[OK]A记录 域名对应的ip地址 指示域名和ip地址的对应关系
[OK]AAAA 域名对应的IPv6解析地址
[OK]CNAME 别名记录 其实就是让一个服务器有多个域名

[OK]NS记录 域名服务器记录 指定该域名由哪个DNS服务器来进行解析
[OK]MX记录 邮件交换记录 说明哪台服务器是当前区域的邮件服务器
[OK]TXT记录 在DNS中存储任意的文本信息 常用于 域名验证 SPF记录 DMARC 等

[NO]SOA记录 起始授权记录 用于指示解析这个区域的主dns服务器
[NO]PRT记录 IP逆向查询记录 从ip地址中查询域名。

PS:
  不需要对任意域名都查询 PTR / SOA记录，尤其是子域名
  PTR 是对 IP 地址的反向查询（如 115.46.235.103.in-addr.arpa）。
  SOA 应对 主域名（如 baidu.com、shifen.com）进行查询，而不是子域名。
```


	


