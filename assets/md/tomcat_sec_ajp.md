# Apache Tomcat AJP漏洞

```
CVE-2020-1938
在 AJP 连接器中发现文件读取/包含漏洞。未经身份验证的远程攻击者可利用此漏洞从易受攻击的服务器中读取 Web 应用程序文件。如果易受攻击的服务器允许上传文件，则攻击者可以借助各类文件上传恶意 JavaServer Pages (JSP) 代码并获得远程代码执行 (RCE)
```

```
修复方法01: 关闭AJP服务
确认是否需要正常使用AJP协议，如若不使用，可以关闭该协议
（1）编辑 <CATALINA_BASE>/conf/server.xml，找到如下行（<CATALINA_BASE> 为 Tomcat 的工作目录）：<Connector port="8009" protocol="AJP/1.3" redirectPort="8443" />

（2）将此行注释掉（或直接删掉此行）：<!--<Connector port="8009" protocol="AJP/1.3" redirectPort="8443" />-->

（3）更改完毕后，重启 Tomcat 即可
```

```
修复方法02: 升级tomcat
将 Tomcat 服务器升级至 7.0.100、8.5.51、9.0.31 或更高版本
```
