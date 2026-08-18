# 取线程数thcount,zabbix5自定义监控项

Centos7.6 zabbix\_agent2 (Zabbix) 5.0.2

```
#!/bin/bash
#igo

help(){
    echo "${0} <wip|app|report>"
    exit 1
}
case "${1}" in
    wip)
ps -ef |grep -i wip.jar|grep -v grep |awk '{print $2}'|xargs -I {} cat /proc/{}/status | grep Threads|awk '{print $2}'
        ;;
    app)
ps -ef |grep -i wip.jar|grep -v grep |awk '{print $2}'|xargs -I {} cat /proc/{}/status | grep Threads|awk '{print $2}'
        ;;
    report)
ps -ef |grep -i report.jar|grep -v grep |awk '{print $2}'|xargs -I {} cat /proc/{}/status | grep Threads|awk '{print $2}'
        ;;
    *)
        help
        ;;

esac
```

```
zabbix
# cat /etc/zabbix/zabbix_agent2.d/thcount.conf
UserParameter=thcount[*],bash /data/igo/thcount.sh $1

Zabbix_WEB_UI
主机,监控项,创建监控项
名称: thcount_wip
键值: thcount[wip]
点击添加
```
