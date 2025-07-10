## 安装方式
安装可执行程序后, 需要补充依赖数据库文件, 可执行命令或手动下载解压assets目录.

### go install 安装
```
go install github.com/winezer0/cdnAnalyzer/cmd/cdnAnalyzer@latest
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


## 命令行参数说明文档

### 工具参数

```
Usage: cdnAnalyzer [OPTIONS]

Options:

# 配置文件相关
-c, --config-file FILE       指定配置文件路径 (YAML)
-C, --update-config          从远程 URL 更新配置内容

# 基本参数（覆盖 App Config）
-i, --input INPUT            输入目标列表，支持文件或逗号分隔的字符串
-I, --input-type TYPE        输入数据类型: string(直接输入)/file(文件)/sys(stdin) [default: string]

# 输出配置（覆盖 App Config）
-o, --output FILE            输出文件路径 [default: analyser_output.json]
-O, --output-type TYPE       输出文件类型: csv/json/txt/sys [default: sys]
-l, --output-level LEVEL     输出详细级别：1=安静模式 / 2=默认 / 3=详细模式 [default: 2] [choices: 1, 2, 3]
-n, --output-no-cdn          只输出非 CDN/WAF 的信息

# 数据库更新配置
-p, --proxy URL              使用代理下载文件（支持 http/socks5）
-d, --folder DIR             数据库存储目录（默认为用户目录）
-u, --update-db              自动更新数据库文件（定期检查）

# DNS 相关参数（有值时会覆盖配置文件）
-t, --dns-timeout SEC        设置 DNS 查询超时时间（秒）[default: 0]
-r, --resolvers-num NUM      设置使用的 resolver 数量 [default: 0]
-m, --city-map-num NUM       设置城市地图 worker 数量 [default: 0]
-w, --dns-concurrency NUM    设置并发 DNS 查询数 [default: 0]
-W, --edns-concurrency NUM   设置并发 EDNS 查询数 [default: 0]
-q, --query-ednscnames BOOL  是否启用通过 EDNS 解析 CNAME [allow: "", false, true]
-s, --query-edns-use-sys-ns BOOL  是否使用系统 DNS 服务器解析 EDNS [allow: "", false, true]
```


### 工具参数说明

以下是支持的所有命令行参数及其含义说明：

---

#### 配置文件相关

| 参数             | 短格式  | 长格式               | 描述                   | 默认值         |
|----------------|------|-------------------|----------------------|-------------|
| `ConfigFile`   | `-c` | `--config-file`   | 指定配置文件路径（YAML 格式）    | 空字符串 (自动设置) |
| `UpdateConfig` | `-C` | `--update-config` | 从远程 URL 更新配置内容覆盖本地配置 | false       |

---

#### 输入相关

| 参数          | 短格式  | 长格式            | 描述                                          | 默认值             |
|-------------|------|----------------|---------------------------------------------|-----------------|
| `Input`     | `-i` | `--input`      | 输入数据（可以是字符串或文件路径，多个用逗号分隔）                   | 必填 (除非使用-I sys) |
| `InputType` | `-I` | `--input-type` | 输入类型：`string`（字符串）、`file`（文件路径）、`sys`（标准输入） | `"string"`      |

---

#### 输出相关

| 参数            | 短格式  | 长格式               | 描述                                          | 默认值                      |
|---------------|------|-------------------|---------------------------------------------|--------------------------|
| `Output`      | `-o` | `--output`        | 输出文件路径                                      | `"analyser_output.json"` |
| `OutputType`  | `-O` | `--output-type`   | 输出格式：`csv` / `json` / `txt` / `sys`（标准输出）   | `"sys"`                  |
| `OutputLevel` | `-l` | `--output-level`  | 输出详细程度：`2 default` / `1 quiet` / `3 detail` | `2`                    |
| `OutputNoCDN` | `-n` | `--output-no-cdn` | 只输出非 CDN 和 WAF 的结果                          | false                    |

---

#### 数据库更新相关

| 参数         | 短格式  | 长格式           | 描述                                | 默认值   |
|------------|------|---------------|-----------------------------------|-------|
| `Proxy`    | `-p` | `--proxy`     | 下载数据库时使用的代理（支持 `http` 或 `socks5`） | 空字符串  |
| `Folder`   | `-d` | `--folder`    | 数据库存储目录（默认用户主目录）                  | 用户主目录 |
| `UpdateDB` | `-u` | `--update-db` | 自动更新数据库文件（定期检查）                   | false |

---


#### DNS 相关参数（覆盖配置文件）

| 参数 | 短选项 | 长选项 | 描述 | 默认值 | 可选值 |
|------|--------|--------|------|--------|--------|
| DNSTimeout | `-t` | `--dns-timeout` | 设置 DNS 查询超时时间（秒） | `0` | - |
| ResolversNum | `-r` | `--resolvers-num` | 设置使用的 resolver 数量 | `0` | - |
| CityMapNum | `-m` | `--city-map-num` | 设置城市地图 worker 数量 | `0` | - |
| DNSConcurrency | `-w` | `--dns-concurrency` | 设置并发 DNS 查询数 | `0` | - |
| EDNSConcurrency | `-W` | `--edns-concurrency` | 设置并发 EDNS 查询数 | `0` | - |
| QueryEDNSCNAMES | `-q` | `--query-ednscnames` | 是否启用通过 EDNS 解析 CNAME | `""` | `""`, `"false"`, `"true"` |
| QueryEDNSUseSysNS | `-s` | `--query-edns-use-sys-ns` | 是否使用系统 DNS 服务器解析 EDNS | `""` | `""`, `"false"`, `"true"` |

---

## 使用示例

```bash

# 检查更新数据库
./cdnAnalyzer -u

# 通过代理下载数据库并指定存储路径
./cdnAnalyzer -p http://127.0.0.1:8080 -u

# 基础使用 - 指定输入并输出到 JSON 文件
./cdnAnalyzer -i  example.com,google.com -o results.json -O json
```

---
### 使用管道符传入
```
echo www.baidu.com | cdnAnalyzer.exe -I sys
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
cdnAnalyzer.exe -i www.baidu.com,www.google.com
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
---

