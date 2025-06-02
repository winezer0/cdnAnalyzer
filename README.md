# CDNCheck
CDN Check On Golang

1、进行DNS解析
1.1 使用常规方案进行A记录、CNAME 记录查询
1.2 使用ECS方案进行A记录、CNAME 记录查询

2、进行IP信息查询

IP 归属地信息 纯真IP库、IP2region、ipv6
IP 属性信息  
IP ASN信息


3、进行域名信息查询
基于本地数据库进行域名CDN查询 cdn.yml 



## DNS记录的常见类型

1. A记录
A记录也称为主机记录，A记录的基本作用就是一个主机域名对应的ip地址是多少，即是域名和ip地址的对应关系。

例如：
www.baidu.com. IN A 1.1.1.2
mx1.baidu.com. IN A 1.1.1.3
mx2.baidu.com. IN A 1.1.1.3

2. NS记录
NS记录称为域名服务器记录，用来指定该域名由哪个DNS服务器来进行解析。假设baidu.com区域有两个DNS服务器负责解析，ns1.baidu.com是主服务器，ns2.baidu.com是辅助服务器，ns1.baidu.com的ip是202.99.16.1，ns2.baidu.com的ip是202.99.16.2。那么我们应该创建两条NS记录，当然，NS记录依赖A记录的解析，我们首先应该为ns1.baidu.com和ns2.baidu.com创建两条A记录
注：ns记录说明，在这个区域里，有多少个服务器承担解析的任务

例如：
baidu.com. IN NS ns1.baidu.com. 
baidu.com. IN NS ns2.baidu.com.


3. SOA记录
起始授权记录，用于一个区域的开始，SOA记录后的所有信息均是用于控制这个区域的，每个区域数据库文件都必须包含一个SOA记录，并且必须是其中的第一个资源记录，用以标识DNS服务器管理的起始位置，SOA说明能解析这个区域的dns服务器中哪个是主服务器
例如，NS记录说明了有两个DNS服务器负责baidu.com的域名解析，但哪个是主服务器呢？这个任务由SOA记录来完成

4. CNAME记录
又称为别名记录，其实就是让一个服务器有多个域名，大致相当于给一个人起个外号。

为什么需要Cname记录呢？一方面是照顾用户的使用习惯，例如我们习惯把邮件服务器命名为mail，把ftp服务器命名为ftp；
那如果只有一台服务器，同时提供邮件服务和FTP服务，那我们究竟该么命名呢？
我们可以把服务器命名为mail.baidu.com，然后再创建一个Cname记录叫ftp.baidu.com就可以两者兼顾了。
另外使用Cname记录也有安全方面的考虑因素？
例如我们不希望别人知道某个网站的真实域名，那我们可以让用户访问网站的别名，例如我们访问的百度网站的真实域名就是www.a.shifen.com，
我们使用的www.baidu.com只是www.a.shifen.com的别名而已
例如：
web.sangfor.com. IN CNAME www.sangfor.com

5. MX记录
又称为邮件交换记录，MX记录用于说明哪台服务器是当前区域的邮件服务器，
例如在baidu.com区域中，mail.baidu.com是邮件服务器，而且IP地址是202.99.16.125。那么我们就可以在DNS服务器中进行下列处理：
1、为邮件服务器创建A记录，我们首先为邮件服务器创建一条A记录，这是因为MX记录中描述邮件服务器时不能使用IP地址，只能使用完全合格域名
例如：
magedu.com. IN MX 10 mx1.magedu.com. 
IN MX 20 mx2.magedu.com

6. PRT记录
又称为逆向查询记录，用于从ip地址中查询域名。
PRT记录是A记录的逆向记录，作用是把IP地址解析为域名
例如：
4.3.2.1.in-addr.arpa. IN PRT www.sangfor.com