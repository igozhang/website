# dns_高可用

```
env
centos7.6
```

```
1,hosts缓存
/etc/hosts 添加本地条目 将dns内网解析条目添加到服务器本地

2,resolv策略
/etc/resolv.conf 添加条目，添加完成后效果如下：
使用内网dns服务器：
options timeout:1 attempts:1 rotate
使用公网dns服务器：
options timeout:1 attempts:1 rotate

3,dnsmasq添加配置 
min-cache-ttl=3600
no-negcache
log-async=100
cache-size=15000
address=/qq.qq.qq/68.68.68.248

4,添加/etc/resolv.conf 自切换脚本（利用zabbix监控项）每10秒出发运行一次并收集dnsmasq时延数据
监控项: UserParameter=dns.change,sudo bash /etc/zabbix/script/dnsChange.sh
$ cat /etc/zabbix/script/dnsChange.sh
cat /etc/zabbix/script/dnsChange.sh
#bin/bash
Delay249=`dig qq.qq.qq @10.12.240.249 +time=1 |grep Query|awk '{print $4}'`
if [ $"$Delay250"x = x ]; then Delay250=99;fi
Delay250=`dig qq.qq.qq @10.12.240.250 +time=1 |grep Query|awk '{print $4}'`
if [ $"$Delay250"x = x ]; then Delay250=99;fi
echo $Delay249
echo $Delay250

if [ $Delay249 -gt 50 ] && [ $Delay250 -gt 50 ];
then
echo -e "nameserver 223.5.5.5 \nnameserver 114.114.114.114\noptions timeout:1 attempts:1 rotate"> /run/resolvconf/resolv.conf
elif [ $Delay249 -gt 50 ] && [ $Delay250 -lt 50 ];
then
echo -e "nameserver 10.12.240.250"> /run/resolvconf/resolv.conf
else
echo -e "nameserver 10.12.240.249\nnameserver 10.12.240.250\noptions timeout:1 attempts:1 rotate"> /run/resolvconf/resolv.conf
fi

5,监控项及触发器
5.1 Zabbix_serv/proxy添加公网dns解析监控项及触发器
监控项UserParameter=chk.dns,/bin/bash /opt/zabbix/script/chkdns.sh
脚本：# cat /opt/zabbix/script/chkdns.sh
#!/bin/bash
URL=`dig www.baidu.com +time=1 +short @223.5.5.5|head -1`
# 以223.5.5.5为dns服务器，解析www.baidu.com 超时1s。
if [ $"$URL"x = 'www.a.shifen.com.'x ]
then
  echo 1
else
  echo 0
fi

5.2 zabbix_serv/proxy添加内网dns解析监控及触发器
监控项:
UserParameter=chk.dns250,dig qq.qq.qq +time=1 @10.12.240.250 |grep Query|awk '{print $4}'
UserParameter=chk.dns249,dig qq.qq.qq +time=1 @10.12.240.249 |grep Query|awk '{print $4}'
脚本：# cat /opt/zabbix/script/chklocaldns248.sh
dig qq.qq.qq +time=1 @172.22.240.248 |grep Query|awk '{print $4}'

6,自动重启(防止dnsmasq挂掉)
6.1
每分钟检查一次
* * * * * bash /opt/script/chkdns.sh

# cat /opt/script/chkdns.sh
#!/bin/bash
DNS=10.12.240.250
tmp=/tmp/dns.log
t=`date "+%Y-%m-%d %H:%M:%S"`
log=/opt/script/chkdns.log

cat /dev/null > $tmp
timeout 3 dig chkdns.ops.vrviu.com @$DNS > $tmp

if [ ! "`grep '111.111.111.111' $tmp`" ]
   then
       echo $t dns fail >> $log
       /usr/bin/supervisorctl restart webproc
fi
6.2 开启dns日志并每天清理
59 23 * * * cat /dev/null > /opt/webproc/log/dnsmasq.log
```
