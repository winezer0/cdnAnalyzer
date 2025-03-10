# cdnAnalyzer
cdnAnalyzer 基于智能DNS.ECS扩展技术实现的CDN判断程序

代码层原理请参考： burpheart 大佬的 https://github.com/burpheart/cdnlookup

### Install 安装
```
go install github.com/winezer0/cdnAnalyzer@latest
```
### Edns-Client-Subnet (ECS) 介绍
Edns-Client-Subnet (ECS) 是一种DNS扩展机制，允许递归DNS服务器在查询权威DNS服务器时附加上客户端的子网信息。这项技术主要用于优化内容分发网络(CDN)和其他基于地理位置的服务，通过更精确地定位用户位置来提升服务性能和用户体验。

当用户的DNS请求到达递归解析器时，递归解析器使用ECS选项将客户端的部分IP地址（通常是/24或/32掩码）附加到对权威DNS服务器的请求中。权威DNS服务器利用这些信息确定最接近用户的资源位置，并返回相应的IP地址。这意味着，通过ECS技术，无需使用代理，只需提供要模拟的客户端IP地址，即可获取指定IP地址地理位置的DNS解析结果。

### ECS技术在CDN检测中的应用
ECS作为CDN技术的基础组件之一，具有较高的稳定性和不易变动性。 这使得它成为判断域名是否使用了CDN的理想选择。 

相较于传统的多地Ping等技术，ECS具有显著优势： 响应速度更快、稳定性更好、可用性更强。

具体实现思路如下：

```
1、获取不同IP区域的解析结果
2、解析结果IP数量判断
3、综合归属信息分析
```



## TODO
```
1、 实现ECS扩展DNS解析域名 实现基于结果数量的CDN判断 【完成】
2、 对解析的IP进行进一步判断 
    IP归属地查询、
    ASN|IP数据库判断、
    CNAME信息补充、
    解析结果C段整合分析
```

## 相关项目及后续参考
```
    基础代码：
        https://github.com/burpheart/cdnlookup
        
    后续参考：
        https://github.com/zu1k/nali 基于IP|CDN信息数据库
        https://github.com/projectdiscovery/cdncheck  基于IP范围、ASN、域名
            国外源：https://github.com/projectdiscovery/cdncheck/blob/main/sources_data.json
            国内源：https://github.com/hanbufei/isCdn/blob/main/client/data/sources_china.json
    
        https://github.com/hanbufei/isCdn   基于IP范围和官方API  2024年5月
        https://github.com/YouChenJun/CheckCdn 基于官方API      2024年11月
        https://github.com/alwaystest18/cdnChecker  2024年11月
```
