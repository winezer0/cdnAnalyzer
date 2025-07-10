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


## ✅ 命令行参数说明文档

### 📋 工具参数说明

以下是支持的所有命令行参数及其含义说明：

---

### 🔧 配置文件相关

| 参数             | 短格式  | 长格式               | 描述                   | 默认值         |
|----------------|------|-------------------|----------------------|-------------|
| `ConfigFile`   | `-c` | `--config-file`   | 指定配置文件路径（YAML 格式）    | 空字符串 (自动设置) |
| `UpdateConfig` | `-C` | `--update-config` | 从远程 URL 更新配置内容覆盖本地配置 | false       |

---

### 📥 输入相关

| 参数          | 短格式  | 长格式            | 描述                                          | 默认值             |
|-------------|------|----------------|---------------------------------------------|-----------------|
| `Input`     | `-i` | `--input`      | 输入数据（可以是字符串或文件路径，多个用逗号分隔）                   | 必填 (除非使用-I sys) |
| `InputType` | `-I` | `--input-type` | 输入类型：`string`（字符串）、`file`（文件路径）、`sys`（标准输入） | `"string"`      |

---

### 📤 输出相关

| 参数            | 短格式  | 长格式               | 描述                                          | 默认值                      |
|---------------|------|-------------------|---------------------------------------------|--------------------------|
| `Output`      | `-o` | `--output`        | 输出文件路径                                      | `"analyser_output.json"` |
| `OutputType`  | `-O` | `--output-type`   | 输出格式：`csv` / `json` / `txt` / `sys`（标准输出）   | `"sys"`                  |
| `OutputLevel` | `-l` | `--output-level`  | 输出详细程度：`2 default` / `1 quiet` / `3 detail` | `2`                    |
| `OutputNoCDN` | `-n` | `--output-no-cdn` | 只输出非 CDN 和 WAF 的结果                          | false                    |

---

### 🛢️ 数据库更新相关

| 参数         | 短格式  | 长格式           | 描述                                | 默认值   |
|------------|------|---------------|-----------------------------------|-------|
| `Proxy`    | `-p` | `--proxy`     | 下载数据库时使用的代理（支持 `http` 或 `socks5`） | 空字符串  |
| `Folder`   | `-d` | `--folder`    | 数据库存储目录（默认用户主目录）                  | 用户主目录 |
| `UpdateDB` | `-u` | `--update-db` | 自动更新数据库文件（定期检查）                   | false |

---

## 📚 使用示例

```bash
# 基础使用 - 指定输入并输出到 JSON 文件
./cdnAnalyzer -i  example.com,google.com -o results.json -O json

# 检查更新数据库
./cdnAnalyzer -u

# 通过代理下载数据库并指定存储路径
./cdnAnalyzer -p http://127.0.0.1:8080 -u
```
---

### 使用示例
提示: 在window下使用-t /t 是相同的,只是会自动根据操作系统来显示参数标志符.

### 使用管道符传入
```
λ echo www.baidu.com | cdnAnalyzer.exe -I sys
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
λ cdnAnalyzer.exe -i www.baidu.com,www.google.com
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