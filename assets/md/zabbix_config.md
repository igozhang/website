# zabbix配置文件

```
# cat /etc/zabbix/zabbix_agentd.conf
PidFile=/tmp/zabbix_agentd.pid
LogFile=/tmp/zabbix_agentd.log
LogFileSize=0
Server=10.10.240.168
ServerActive=10.10.240.168
Hostname=hostname-mysql-01
HostMetadata=vm
```
