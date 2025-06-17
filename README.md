# CDNCheck
CDN Check On Golang

## TODO 
1. 实现命令行参数|配置文件 格式化输入
2. 实现格式化输出
3. 代码逻辑优化
4. 自动化资源更新

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

### 其他思路：
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


	


