# cdnAnalyzer

**CDN Check On Golang**

目前最完善的CDN分析工具, 便于快速筛选CDN资产信息, 基于IP、ASN、CNAME、IP归属地等多种方案进行分析.

## 免责声明

继续阅读文章或使用工具视为您已同意《 NOVASEC免责声明》: [NOVASEC免责声明](https://mp.weixin.qq.com/s/iRWRVxkYu7Fx5unxA34I7g)


## 更新说明

### 2025-09-20
- 添加 IP2Region 支持，提供超大 IPv6 地址库，切换IPlocate 可参考或直接使用 config_ip2region.yaml (注意:新版需要更新配置文件)
- 优化 geolite-asn 支持，将开始IPv6 ASN和IPv4 ASN 进行合并，只需要一个ASN文件


### 当前支持的IP数据库类型

1. **QQWry数据库** (默认IPv4数据库)
   - 用于IPv4地址的地理位置查询
   - 文件名: `qqwry.dat`
   - 特点: 轻量级，更新频繁

2. **ZXWry数据库** (默认IPv6数据库)
   - 用于IPv6地址的地理位置查询
   - 文件名: `zxipv6wry.db`
   - 特点: 专门针对IPv6地址

3. **IP2Region数据库** (可选)
   - 代码库同时支持IPv4和IPv6地址查询
   - 文件名: `ip2region_v4.xdb` 和 `ip2region_v6.xdb` (较大600M)
   - 特点: IPv6数据更全, 例如:"2405:6f~00:c602::1": "中国|北京市|北京市|专线用户"

### 如何切换数据库

1. **使用IP2Region数据库**

   修改配置文件 `config.yaml` 中的数据库ipv4locate|ipv6locate的实际下载URL部分
   ```yaml
   download-items:
     # 启用IP2Region IPv4数据库
     - module: ipv4locate
       filename: ip2region_v4.xdb
       download-urls:
         - https://github.com/lionsoul2014/ip2region/blob/master/data/ip2region_v4.xdb

     # 启用IP2Region IPv6数据库
     - module: ipv6locate
       filename: ip2region_v6.xdb
       download-urls:
         - https://github.com/lionsoul2014/ip2region/blob/master/data/ip2region_v6.xdb
   ```

2. **使用组合数据库（QQWry ipv4 + IP2Region ipv6)**

   ```yaml
   download-items:
     # QQWry数据库 IP信息带有部分文字，可能有助于CDN判断
     - module: ipv4locate
       filename: qqwry.dat
       download-urls:
         - https://github.com/metowolf/qqwry.dat/releases/latest/download/qqwry.dat

     # ip2region_v6.xdb 数据更全，但是文件更大
     - module: ipv6locate
       filename: ip2region_v6.xdb
       download-urls:
         - https://github.com/lionsoul2014/ip2region/blob/master/data/ip2region_v6.xdb
   ```


## TODO
-   [x] 整理 unknown-cdn-cname 资产和其他源数据
-   [x] 实现已知CDN域名IP、疑似CDN域名自动分析脚本
-   [ ] 将CDN信息脱敏后备份, 并节建立Issue允许用户上传疑似CDN信息用于补充CDN IP和CNAME数据库
-   [ ] 优化代码 实现快速分析模式, 默认的DNS查询次数过多, 导致批量查询时回显较慢, 临时调节可以修改config.yaml中的超时/线程配置
-   [ ] 优化代码 实现多个CDN数据源合并时，能够自动进行IP级去重操作,当前仅实现字符串去重
-   [ ] 考虑优化数据源格式 增加service键,用于标记资产属于厂商的公共服务域名 (好像没什么用)


## 功能介绍

- 支持多种输入格式：URL、域名、IP.
- 全面分析：进行域名解析、IP分析，识别CDN、WAF、Cloud服务.
- 自定义 DNS 服务器进行解析.
- 支持通过 EDNS 和自定义城市IP进行精准地域解析.
- 通过 ASN、纯真IP库、IPv6数据库识别IP归属信息.
- 丰富的输出格式：CSV、JSON、TXT、标准输出.
- 数据库自动更新：自动从网络下载最新的IP和CDN数据库.
- CDN数据源自动更新：通过 Github Actions 每天自动更新.

## 安装方式

### 1. Go Install

```bash
go install github.com/winezer0/cdnAnalyzer/cmd/cdnAnalyzer@latest
```
*安装可执行程序后, 需要补充依赖数据库文件, 可执行 `-u` 命令或手动下载解压 `assets` 目录.*

### 2. 源码安装

```bash
git clone --depth 1 https://github.com/winezer0/cdnAnalyzer
cd cdnAnalyzer
go build -ldflags="-s -w" -o cdnAnalyzer.exe ./cmd/cdnAnalyzer/main.go
```

### 3. Release 安装

从 Github Releases 页面下载预编译好的可执行文件：
[https://github.com/winezer0/cdnAnalyzer/releases](https://github.com/winezer0/cdnAnalyzer/releases)

---

## 使用说明

### 命令行参数

`cdnAnalyzer` 支持丰富的命令行参数以自定义其行为.

```
Usage: cdnAnalyzer [OPTIONS]
```

#### **配置文件相关**

| 参数 | 短格式 | 长格式 | 描述 |
| :--- | :--- | :--- | :--- |
| `ConfigFile` | `-c` | `--config-file` | 指定配置文件路径 (YAML) |
| `UpdateConfig` | `-C` | `--update-config` | 从远程 URL 更新配置内容 |

#### **输入相关**

| 参数 | 短格式 | 长格式 | 描述 | 默认值 |
| :--- | :--- | :--- | :--- | :--- |
| `Input` | `-i` | `--input` | 输入目标，支持文件或逗号分隔的字符串 | 必填 (除非使用`-I sys`)|
| `InputType` | `-I` | `--input-type` | 输入类型: `string`(直接输入)/`file`(文件)/`sys`(stdin) | `string` |

#### **输出相关**

| 参数 | 短格式 | 长格式 | 描述 | 默认值 |
| :--- | :--- | :--- | :--- | :--- |
| `Output` | `-o` | `--output` | 输出文件路径 | `analyser_output.json` |
| `OutputType` | `-O` | `--output-type` | 输出文件类型: `csv`/`json`/`txt`/`sys` | `sys` |
| `OutputLevel` | `-l` | `--output-level` | 输出详细级别：1=安静 / 2=默认 / 3=详细 | `2` |
| `OutputNoCDN` | `-n` | `--output-no-cdn` | 只输出非 CDN/WAF 的信息 | `false` |

#### **数据库更新相关**

| 参数 | 短格式 | 长格式 | 描述 | 默认值 |
| :--- | :--- | :--- | :--- | :--- |
| `Proxy` | `-p` | `--proxy` | 使用代理下载文件 (支持 http/socks5) | - |
| `Folder` | `-d` | `--folder` | 数据库存储目录 (默认为用户目录) | 用户主目录 |
| `UpdateDB` | `-u` | `--update-db` | 自动更新数据库文件 (定期检查) | `false` |

#### **DNS 相关参数**

这些参数会覆盖配置文件中的设置.

| 参数 | 短格式 | 长格式 | 描述 | 默认值 |
| :--- | :--- | :--- | :--- | :--- |
| `DNSTimeout` | `-t` | `--dns-timeout` | DNS 查询超时时间 (秒) | `0` |
| `ResolversNum` | `-r` | `--resolvers-num` | 使用的 resolver 数量 | `0` |
| `CityMapNum` | `-m` | `--city-map-num` | 城市地图 worker 数量 | `0` |
| `DNSConcurrency` | `-w` | `--dns-concurrency` | 并发 DNS 查询数 | `0` |
| `EDNSConcurrency` | `-W` | `--edns-concurrency`| 并发 EDNS 查询数 | `0` |
| `QueryEDNSCNAMES` | `-q` | `--query-ednscnames`| 是否启用通过 EDNS 解析 CNAME | `false` |
| `QueryEDNSUseSysNS` | `-s` | `--query-edns-use-sys-ns`| 是否使用系统 DNS 服务器解析 EDNS | `false` |

### 使用示例

#### 检查更新数据库
```bash
# 从默认源更新
./cdnAnalyzer -u

# 通过代理下载数据库
./cdnAnalyzer -p http://127.0.0.1:8080 -u
```

#### 分析目标
```bash
# 分析单个目标，结果输出到控制台
./cdnAnalyzer -i example.com

# 分析多个目标，结果输出到 JSON 文件
./cdnAnalyzer -i example.com,google.com -o results.json -O json

# 从文件读取目标
./cdnAnalyzer -i targets.txt -I file

# 通过管道传入目标
echo www.baidu.com | ./cdnAnalyzer -I sys
```

---

## 工作原理

```
[输入] -> [类型判断]
              |
              +-- 域名 -> [DNS/EDNS 解析] -> 获取 IP 和 CNAME
              |
              +-- IP   -> [IP归属地查询] -> 获取地理位置和 ASN 信息
                           |
                           V
                      [信息分析] -> [基于 CNAME, IP, ASN 等判断 CDN/WAF/Cloud] -> [输出]
```

1.  **域名解析**: 实现标准 DNS (查询 CNAME/A/AAAA) 和 EDNS (查询 A/AAAA) 解析.
2.  **IP归属地查询**:
    -   IPv4: 基于纯真IP库.
    -   IPv6: 基于 ipv6wry 数据库.
3.  **ASN信息查询**:
    -   IPv4/IPv6: 基于 GeoLite2 ASN 数据库.
4.  **CDN/WAF/Cloud 识别**:
    -   综合多个数据源，通过 CNAME、IP段、ASN等信息进行交叉验证.

---

## 数据源

`cdnAnalyzer` 整合了多个公开的数据源以提供准确的分析结果.
-   更新中 新增Ip2Region IPv4数据库 : [ip2region_v4.xdb](https://github.com/lionsoul2014/ip2region/blob/master/data/ip2region_v4.xdb)
-   更新中 新增Ip2Region IPv6数据库 : [ip2region_v6.xdb](https://github.com/lionsoul2014/ip2region/blob/master/data/ip2region_v6.xdb) 600M+
-   更新中 IPv4数据库 `qqwry.dat`: [metowolf/qqwry.dat](https://github.com/metowolf/qqwry.dat)
-   已停止 IPv6数据库 `zxipv6wry.db`: [内置](https://github.com/winezer0/cdnAnalyzer/blob/main/assets/zxipv6wry.db)
-   更新中 ASNvx数据库 `geolite2-asn.mmdb ipv4+-ipv6`: [sapics/ip-location-db](https://github.com/sapics/ip-location-db/blob/main/geolite2-asn-mmdb/geolite2-asn.mmdb)
-   维护中 DNS服务器 `dns-resolvers`: [自定义](https://github.com/winezer0/cdnAnalyzer/blob/main/assets/resolvers.txt)
-   维护中 对应城市IP `city_ip.csv`: [自定义](https://github.com/winezer0/cdnAnalyzer/blob/main/assets/city_ip.csv)
-   维护中 云资产关键字 `cloud_keys.yml`: [自定义](https://github.com/winezer0/cdnAnalyzer/blob/main/assets/cloud_keys.yml)
-   更新中 CDN域名信息 `4ft35t cdn.yml`: [4ft35t/cdn](https://github.com/4ft35t/cdn/blob/master/src/cdn.yml)
-   更新中 国外CDN信息 `sources_foreign.json`: [projectdiscovery/cdncheck](https://github.com/projectdiscovery/cdncheck/blob/main/sources_data.json)
-   更新中 国外CDN信息 `provider_foreign.yaml`: [projectdiscovery/cdncheck](https://github.com/projectdiscovery/cdncheck/blob/main/cmd/generate-index/provider.yaml)
-   已停止 国内CDN信息 `sources_china.json`: [hanbufei/isCdn](https://github.com/hanbufei/isCdn/blob/main/client/data/sources_china.json) 
-   已停止 国内CDN信息 `sources_china2.json`: [mabangde/cdncheck_cn](https://github.com/mabangde/cdncheck_cn/blob/main/sources_data.json) **已合并到 sources_china.json**
-   已停止 粗略CDN域名 `unknown-cdn-cname.txt`: [alwaystest18/cdnChecker](https://github.com/alwaystest18/cdnChecker/blob/master/cdn_cname) **已合并到 sources_china.json**
-   更新中 用户提交信息 `sources_added.json`: [自定义](https://github.com/winezer0/cdnAnalyzer/blob/main/assets/sources_added.json) **由用户提交**

---

## 开发或数据库参考
cdncheck  | nali | nemo_go | ip-location-db 等等

## 联系方式

如需获取更多信息、技术支持或定制服务，请通过以下方式联系我们：

**NOVASEC微信公众号** 或通过社交信息联系开发者 **【酒零】**

![NOVASEC0](https://raw.githubusercontent.com/winezer0/mypics/refs/heads/main/NOVASEC0.jpg)