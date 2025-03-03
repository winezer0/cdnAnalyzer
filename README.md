# cdnAnalyzer
cdnAnalyzer


基于 智能DNS技术 实现的CDN判断程序

代码参考： https://github.com/burpheart/cdnlookup

### Edns-Client-Subnet (ECS) 介绍
Edns-Client-Subnet (ECS) 是一种DNS扩展机制，允许递归DNS服务器在查询权威DNS服务器时附加上客户端的子网信息。这项技术主要用于优化内容分发网络(CDN)和其他基于地理位置的服务，通过更精确地定位用户位置来提升服务性能和用户体验。

### 工作原理
当用户的DNS请求到达递归解析器时，递归解析器使用ECS选项将客户端的部分IP地址（通常是/24或/32掩码）附加到对权威DNS服务器的请求中。权威DNS服务器利用这些信息确定最接近用户的资源位置，并返回相应的IP地址。这意味着，通过ECS技术，无需使用代理，只需提供要模拟的客户端IP地址，即可获取指定IP地址地理位置的DNS解析结果。

### ECS技术在CDN检测中的应用
ECS作为CDN技术的基础组件之一，具有较高的稳定性和不易变动性。

这使得它成为判断域名是否使用了CDN的理想选择。

具体步骤如下：

获取不同IP区域的解析结果：首先利用ECS技术准确获取域名在不同IP区域的解析结果。

数量判断：通过对解析出的IP进行数量分析，初步判断域名是否采用了CDN。

归属信息分析：进一步对解析得到的IP进行归属地分析，从而有效过滤出使用了CDN的域名。

这种方法不仅提高了准确性，还简化了流程。

### ECS技术与多地Ping技术的对比
相较于传统的多地Ping等技术，ECS具有显著优势：

响应速度更快：由于ECS是基于DNS解析的过程，因此其响应速度通常比执行多地Ping测试要快。
更高的可用性：ECS查询同时也是DNS解析过程的一部分，这减少了重复进行DNS解析的需求，提升了整体效率。
### ECS技术在CDN检查中的利用：

由于该技术属于CDN技术的底层技术，拥有较强的不易变动性，

当将其用于域名CDN判断时， 可以先准确的获取到域名在不同IP区域的Ip解析结果,

通过对解析IP进行数量判断，能够初步判断出域名是否不为cdn

当再对DNS解析结果IP进行IP归属信息分析，可基本完整的过滤出CDN域名。

### ECS技术和多地ping技术技术对比：

相较于多地ping等技术，ECS是基于DNS解析，响应速度更快，可用性更强。

ECS查询过程,同时也是DNS解析过程,可以节省部分重复性DNS解析IP的工作。

## TODO

```
1、 通过 域名 解析数量 进行  cdn 域名 判断 【已实现 基于ECS DNS解析】
2、 对解析的IP进行进一步判断 [IP归属地查询、cdn asn数据库判断、]


参考实现：
    国外源：https://github.com/projectdiscovery/cdncheck/blob/main/sources_data.json
    国内源：https://github.com/hanbufei/isCdn/blob/main/client/data/sources_china.json
    https://github.com/zu1k/nali 基于IP数据库
    https://github.com/projectdiscovery/cdncheck  基于IP范围、ASN、域名
    https://github.com/hanbufei/isCdn   基于IP范围和官方API  2024年5月
    https://github.com/YouChenJun/CheckCdn 基于官方API      2024年11月
    https://github.com/alwaystest18/cdnChecker  2024年11月
```
